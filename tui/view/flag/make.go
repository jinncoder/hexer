package flag

import (
	"fmt"
	"strings"

	"github.com/archimoebius/hexer/tui/constant"
	"github.com/archimoebius/hexer/util/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

type ModelMakeFlag struct {
	lg                        *lipgloss.Renderer
	styles                    *constant.Styles
	form                      *huh.Form
	Width                     int
	Height                    int
	ProjectId                 string
	FlagId                    string
	ProjectPickerLastTabIndex int
	Session                   *ssh.Session
}

func NewModelMakeFlag(session *ssh.Session, mfi *constant.MakeFlagInput) (*ModelMakeFlag, error) {

	m := &ModelMakeFlag{
		Session: session,
	}

	m.Width = mfi.Width
	m.Height = mfi.Height
	m.lg = lipgloss.DefaultRenderer()
	m.styles = constant.NewStyles(m.lg)
	m.ProjectId = mfi.ProjectId
	m.FlagId = mfi.FlagId
	m.ProjectPickerLastTabIndex = mfi.LastTabIndex

	flagName := ""
	flagValue := ""
	flagNote := ""

	addOrUpdateFlag := "Add Flag?"

	var flagAuthenticator database.Authenticator
	var flag database.Flag
	var err error

	if m.FlagId != "" {
		flag, err = database.GetFlagById(m.FlagId)

		if err != nil {
			fmt.Printf("failed to load flag") // TODO: user error message
		}

		addOrUpdateFlag = "Update Flag?"
		flagName = flag.Name
		flagValue = flag.Value
		flagNote = flag.Note

		if flag.AuthenticatorId != "" {
			flagAuthenticator, err = database.GetAuthenticatorById(flag.AuthenticatorId)

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
				Placeholder("Flag Name").
				Validate(func(s string) error {
					if m.form.GetBool("AddFlag") && len(s) <= 1 || len(s) > 255 {
						return fmt.Errorf("%s must be > 1 char and < 255\n", s)
					}
					return nil
				}).
				Value(&flagName).
				Description("The name of the authenticator to track things under"),
			huh.NewInput().
				Key("Value").
				Description("That pesky value that we all seek...").
				Value(&flagValue).
				Title("Value"),
			huh.NewInput().
				Key("Note").
				Title("Note").
				Placeholder("A short blurb about this authenticator... For example: how does one reach it?").
				Value(&flagNote).
				Description("Brief, critical information - put that here - it'll show in the listing"),
		), huh.NewGroup(
			huh.NewSelect[string]().
				Key("Authenticator").
				Description("The authenticator you found the flag on").
				Title("Authenticator").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("flag must have an authenticator...for now...just make a 'empty' one...damn nullset...")
					}
					return nil
				}).
				OptionsFunc(func() []huh.Option[string] {
					var options []huh.Option[string]
					authenticators, _ := database.GetAuthenticatorList(m.ProjectId)

					if flag.AuthenticatorId != "" {
						options = append(options, huh.NewOption(flagAuthenticator.ListTitle(), flagAuthenticator.Id).Selected(true))
					}

					options = append(options, huh.NewOption("None...", "").Selected(true)) // TODO: add link_project to db for this case?

					for _, authenticator := range authenticators {

						if &flagAuthenticator != nil && authenticator.Id == flagAuthenticator.Id {
							continue
						}

						options = append(options, huh.NewOption(authenticator.ListTitle(), authenticator.Id))
					}

					return options
				}, nil),
			huh.NewConfirm().
				Key("AddFlag").
				Title(addOrUpdateFlag).
				Affirmative("Yes").
				Negative("No").
				Validate(func(c bool) error {
					if c { // all this crap simply b/c m.form.GetBool("AddFlag") value isn't updated -> followed by a final form validate check... pita...
						s := m.form.GetString("Name")
						if len(s) <= 1 || len(s) > 255 {
							return fmt.Errorf("flag name to short")
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

func (m ModelMakeFlag) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		if m.form.GetBool("AddFlag") {
			if m.FlagId != "" {
				err := database.UpdateFlag(
					m.FlagId,
					m.form.GetString("Authenticator"),
					m.form.GetString("Name"),
					m.form.GetString("Value"),
					m.form.GetString("Note"),
				)
				if err != nil {
					return m, constant.ErrorMessage(err)
				}
			} else {
				err := database.AddFlag(
					m.form.GetString("Authenticator"),
					m.form.GetString("Name"),
					m.form.GetString("Value"),
					m.form.GetString("Note"),
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

func (m ModelMakeFlag) Init() tea.Cmd {
	return m.form.Init()
}

func (m ModelMakeFlag) View() string {
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

func (m ModelMakeFlag) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}
