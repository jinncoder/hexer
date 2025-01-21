package authenticator

import (
	"fmt"
	"net"
	"strings"

	"github.com/archimoebius/hexer/tui/constant"
	"github.com/archimoebius/hexer/util/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

type ModelMakeProject struct {
	lg                        *lipgloss.Renderer
	styles                    *constant.Styles
	form                      *huh.Form
	Width                     int
	Height                    int
	ProjectId                 string
	AuthenticatorId           string
	ProjectPickerLastTabIndex int
	Session                   *ssh.Session
}

func NewModelMakeAuthenticator(session *ssh.Session, mai *constant.MakeAuthenticatorInput) (*ModelMakeProject, error) {

	m := &ModelMakeProject{
		Session: session,
	}

	m.Width = mai.Width
	m.Height = mai.Height
	m.lg = lipgloss.DefaultRenderer()
	m.styles = constant.NewStyles(m.lg)
	m.ProjectId = mai.ProjectId
	m.AuthenticatorId = mai.AuthenticatorId
	m.ProjectPickerLastTabIndex = mai.LastTabIndex

	authenticatorName := ""
	authenticatorType := constant.AuthenticatorTypeHost.String()
	authenticatorIP := ""
	authenticatorFQDN := ""
	authenticatorNote := ""
	addOrUpdateAuthenticator := "Add Authenticator?"

	var authenticatorLink database.Authenticator
	var authenticator database.Authenticator
	var err error

	if m.AuthenticatorId != "" {
		authenticator, err = database.GetAuthenticatorById(m.AuthenticatorId)

		if err != nil {
			fmt.Printf("failed to load authenticator") // TODO: user error message
		}

		addOrUpdateAuthenticator = "Update Authenticator?"
		authenticatorName = authenticator.Name
		authenticatorType = authenticator.Type
		authenticatorIP = authenticator.IPv4
		authenticatorFQDN = authenticator.FQDN
		authenticatorNote = authenticator.Note

		if authenticator.AuthenticatorId != "" {
			authenticatorLink, err = database.GetAuthenticatorById(authenticator.AuthenticatorId)

			if err != nil {
				fmt.Printf("failed to load authenticator link") // TODO: user error message
			}
		}
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("Name").
				Title("Name").
				Placeholder("Authenticator Name").
				Validate(func(s string) error {
					if m.form.GetBool("AddAuthenticator") && len(s) <= 1 || len(s) > 255 {
						return fmt.Errorf("%s must be > 1 char and < 255\n", s)
					}
					return nil
				}).
				Value(&authenticatorName).
				Description("The name of the authenticator to track things under"),
			huh.NewSelect[string]().
				Key("Type").
				Description("What does this authenticate for?").
				Title("Type").
				OptionsFunc(func() []huh.Option[string] {
					var options []huh.Option[string]
					options = append(options, huh.NewOption(strings.ToLower(authenticatorType), strings.ToLower(authenticatorType)))

					for _, value := range constant.AuthenticatorTypeList {
						if strings.ToLower(value) == authenticatorType {
							continue
						}
						options = append(options, huh.NewOption(strings.ToLower(value), strings.ToLower(value)))
					}
					return options
				}, nil),
			huh.NewInput().
				Key("IPv4").
				Title("IPv4 Address").
				Placeholder("1.3.3.7").
				Validate(func(s string) error {
					if s == "" {
						return nil
					}

					ip := net.ParseIP(s)
					if ip == nil || (ip.To4() == nil && ip.To16() == nil) {
						return fmt.Errorf("IP address is not valid - only IPv4 and IPv6 are supported")
					}

					return nil
				}).
				Value(&authenticatorIP).
				Description("If known/available - record the IPv4 address here"),
		), huh.NewGroup(
			huh.NewInput().
				Key("FQDN").
				Title("Fully Qualified Domain Name (FQDN)").
				Placeholder("hostname.subdomain.topleveldomain").
				Validate(func(s string) error {
					if s == "" {
						return nil
					}

					if m.form.GetBool("AddAuthenticator") && len(s) <= 1 || len(s) > 255 {
						return fmt.Errorf("%s must be > 1 char and < 255\n", s)
					}
					return nil
				}).
				Value(&authenticatorFQDN).
				Description("The absolute domain name for this authenticator"),
			huh.NewInput().
				Key("Note").
				Title("Note").
				Placeholder("A short blurb about this authenticator... For example: how does one reach it?").
				Value(&authenticatorNote).
				Description("Brief, critical information - put that here - it'll show in the listing"),
			huh.NewSelect[string]().
				Key("Authenticator").
				Description("Is authentication processed elsewhere?").
				Title("Authenticator").
				OptionsFunc(func() []huh.Option[string] {
					var options []huh.Option[string]
					authenticators, _ := database.GetAuthenticatorList(m.ProjectId)

					if authenticator.AuthenticatorId != "" {
						options = append(options, huh.NewOption(authenticatorLink.ListTitle(), authenticatorLink.Id).Selected(true))
					}

					options = append(options, huh.NewOption("None...", "").Selected(true))

					for _, authenticator := range authenticators {
						if &authenticatorLink != nil && authenticator.Id == authenticatorLink.Id {
							continue
						}
						options = append(options, huh.NewOption(authenticator.ListTitle(), authenticator.Id))
					}

					return options
				}, nil),
		), huh.NewGroup(
			huh.NewConfirm().
				Key("AddAuthenticator").
				Title(addOrUpdateAuthenticator).
				Affirmative("Yes").
				Negative("No").
				Validate(func(c bool) error {
					if c { // all this crap simply b/c m.form.GetBool("AddAuthenticator") value isn't updated -> followed by a final form validate check... pita...
						s := m.form.GetString("Name")
						if len(s) <= 1 || len(s) > 255 {
							return fmt.Errorf("authenticator name to short")
						}
					}
					return nil
				}),
		),
	).
		WithWidth(m.Width / 2).
		WithShowHelp(false).
		WithShowErrors(false)

	return m, nil
}

func (m ModelMakeProject) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, constant.SwitchModeCmd(constant.ModePickProject,
				constant.NewPickProjectInput(m.Width, m.Height, m.ProjectId, m.ProjectPickerLastTabIndex),
			)
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {

		if m.form.GetBool("AddAuthenticator") {
			if m.AuthenticatorId != "" {
				err := database.UpdateAuthenticator(
					m.ProjectId,
					m.AuthenticatorId,
					m.form.GetString("Name"),
					m.form.GetString("Type"),
					m.form.GetString("IPv4"),
					m.form.GetString("FQDN"),
					m.form.GetString("Note"),
					m.form.GetString("Authenticator"),
				)
				if err != nil {
					return m, constant.ErrorMessage(err)
				}
			} else {
				err := database.AddAuthenticator(
					m.ProjectId,
					m.form.GetString("Name"),
					m.form.GetString("Type"),
					m.form.GetString("IPv4"),
					m.form.GetString("FQDN"),
					m.form.GetString("Note"),
					m.form.GetString("Authenticator"),
				)
				if err != nil {
					return m, constant.ErrorMessage(err)
				}
			}
		}

		return m, constant.SwitchModeCmd(constant.ModePickProject,
			constant.NewPickProjectInput(m.Width, m.Height, m.ProjectId, m.ProjectPickerLastTabIndex),
		)
	}

	return m, tea.Batch(cmds...)
}

func (m ModelMakeProject) Init() tea.Cmd {
	return m.form.Init()
}

func (m ModelMakeProject) View() string {
	s := m.styles
	header := constant.AppBoundaryView(m.styles, m.Width, "", "")

	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(v)

	errors := m.form.Errors()

	if len(errors) > 0 {
		header = constant.AppErrorBoundaryView(m.styles, m.Width, m.errorView())
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, form)

	footer := constant.AppBoundaryView(m.styles, m.Width, m.form.Help().ShortHelpView(m.form.KeyBinds()), "")
	if len(errors) > 0 {
		footer = constant.AppErrorBoundaryView(m.styles, m.Width, "")
	}
	m.styles.Base.Width(m.Width).Height(m.Height)

	return s.Base.Render(header + "\n" + body + "\n\n" + footer)
}

func (m ModelMakeProject) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}
