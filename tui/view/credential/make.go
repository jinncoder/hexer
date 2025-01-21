package credential

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

type ModelMakeCredential struct {
	lg                        *lipgloss.Renderer
	styles                    *constant.Styles
	form                      *huh.Form
	Width                     int
	Height                    int
	ProjectId                 string
	ProjectPickerLastTabIndex int
	AuthenticatorId           string
	CredentialId              string
	Session                   *ssh.Session
}

func NewModelMakeCredential(session *ssh.Session, mci *constant.MakeCredentialInput) (*ModelMakeCredential, error) {

	m := &ModelMakeCredential{
		Session: session,
	}

	m.Width = mci.Width
	m.Height = mci.Height
	m.lg = lipgloss.DefaultRenderer()
	m.styles = constant.NewStyles(m.lg)
	m.ProjectId = mci.ProjectId
	m.CredentialId = mci.CredentialId
	m.ProjectPickerLastTabIndex = mci.LastTabIndex
	m.AuthenticatorId = mci.AuthenticatorId

	addOrUpdateCredential := "Add Credential?" // #nosec G101 - rofl
	credentialType := ""
	credentialFormat := ""
	credentialEncoding := ""
	credentialUsername := ""
	credentialValue := ""
	credentialNote := ""

	var credential database.Credential
	var credentialAuthenticator database.Authenticator
	var credentialCleartext database.Credential

	if m.CredentialId != "" {
		_, credential, err := database.GetCredentialById(m.CredentialId)

		if err != nil {
			fmt.Printf("failed to load flag") // TODO: user error message
		}

		addOrUpdateCredential = "Update Credential?" // #nosec G101 - rofl
		credentialType = credential.Type
		credentialFormat = credential.Format
		credentialEncoding = credential.Encoding
		credentialUsername = credential.UserName
		credentialValue = credential.Value
		credentialNote = credential.Note

		if credential.AuthenticatorId != "" {
			credentialAuthenticator, err = database.GetAuthenticatorById(credential.AuthenticatorId)
			if err != nil {
				fmt.Printf("failed to load credential authenticator") // TODO: user error message
			}
		}

		if credential.CleartextId != "" {
			_, credentialCleartext, err = database.GetCredentialById(credential.CleartextId)
			if err != nil {
				fmt.Printf("failed to load credential cleartext") // TODO: user error message
			}
		}
	}

	var credentialTypeOptions []huh.Option[string]

	if credentialType != "" {
		credentialTypeOptions = append(credentialTypeOptions, huh.NewOption(strings.ToLower(credentialType), strings.ToLower(credentialType)).Selected(true))
	}

	for _, entry := range constant.CredentialTypeList {

		if strings.EqualFold(strings.ToLower(entry), strings.ToLower(credentialType)) {
			continue
		}

		credentialTypeOptions = append(credentialTypeOptions, huh.NewOption(strings.ToLower(entry), strings.ToLower(entry)))
	}

	var credentialFormatOptions []huh.Option[string]

	if credentialFormat != "" {
		credentialFormatOptions = append(credentialFormatOptions, huh.NewOption(strings.ToLower(credentialFormat), strings.ToLower(credentialFormat)).Selected(true))
	}

	for _, entry := range constant.CredentialFormatList {

		if strings.EqualFold(strings.ToLower(entry), strings.ToLower(credentialFormat)) {
			continue
		}

		credentialFormatOptions = append(credentialFormatOptions, huh.NewOption(strings.ToLower(entry), strings.ToLower(entry)))
	}

	var credentialEncodingOptions []huh.Option[string]

	if credentialEncoding != "" {
		credentialEncodingOptions = append(credentialEncodingOptions, huh.NewOption(strings.ToLower(credentialEncoding), strings.ToLower(credentialEncoding)).Selected(true))
	}

	for _, entry := range constant.CredentialEncodingList {

		if strings.EqualFold(strings.ToLower(entry), strings.ToLower(credentialEncoding)) {
			continue
		}

		credentialEncodingOptions = append(credentialEncodingOptions, huh.NewOption(strings.ToLower(entry), strings.ToLower(entry)))
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("Type").
				Description("The type of credential material provided").
				Title("Type").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("one must have a type...")
					}
					return nil
				}).
				Options(credentialTypeOptions[:]...).
				WithHeight(8),
			huh.NewSelect[string]().
				Key("Format").
				Description("The format of credential material provided").
				Title("Format").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("one must have a format - raw for plaintext")
					}
					return nil
				}).
				Options(credentialFormatOptions[:]...).
				WithHeight(8),
			huh.NewSelect[string]().
				Key("Encoding").
				Description("The encoding of credential material provided").
				Title("Encoding").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("one must have an encoding - raw for plaintext")
					}
					return nil
				}).
				Options(huh.NewOptions(constant.CredentialEncodingList[:]...)...).
				Options(credentialEncodingOptions[:]...).
				WithHeight(8),
		), huh.NewGroup(
			huh.NewSelect[string]().
				Key("Authenticator").
				Description("The authenticator which can validate this credential").
				Title("Authenticator").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("A credential must have an authenticator...for now...just make a 'empty' one...damn nullset...")
					}
					return nil
				}).
				OptionsFunc(func() []huh.Option[string] {
					var options []huh.Option[string]
					authenticators, _ := database.GetAuthenticatorList(m.ProjectId)

					if credential.AuthenticatorId != "" {
						options = append(options, huh.NewOption(credentialAuthenticator.ListTitle(), credentialAuthenticator.Id))
					}

					options = append(options, huh.NewOption("None...", "").Selected(true)) // TODO: add link_project to db for this case?

					for _, authenticator := range authenticators {
						if credential.AuthenticatorId != "" && authenticator.Id == credential.AuthenticatorId {
							continue
						}
						options = append(options, huh.NewOption(authenticator.ListTitle(), authenticator.Id))
					}

					return options
				}, nil).
				WithHeight(20),
		), huh.NewGroup(
			huh.NewSelect[string]().
				Key("Cleartext").
				Description("If this credential has either been cracked, discovered, or obtained in the clear - ").
				Title("Cleartext Credential").
				OptionsFunc(func() []huh.Option[string] {
					var options []huh.Option[string]
					credentials, _ := database.GetProjectCredentialList(m.ProjectId, true)

					if credential.CleartextId != "" {
						options = append(options, huh.NewOption(credentialCleartext.ListTitle(), credentialCleartext.Id))
					}

					options = append(options, huh.NewOption("None...", "").Selected(true)) // TODO: handle this case?

					for _, credential := range credentials {
						if credential.CleartextId != "" && credential.CleartextId == credential.Id {
							continue
						}
						options = append(options, huh.NewOption(credential.ListTitle(), credential.Id))
					}

					return options
				}, nil).
				WithHeight(20),
		), huh.NewGroup(
			huh.NewInput().
				Key("Username").
				Title("Username").
				Validate(func(s string) error {
					if m.form.GetBool("AddCredential") && len(s) <= 1 || len(s) > 255 {
						return fmt.Errorf("%s must be > 1 char and < 255\n", s)
					}
					return nil
				}).
				Value(&credentialUsername).
				Description("The username used when authentication occures"),
			huh.NewInput().
				Key("Value").
				Description("That pesky value that we all seek...").
				Value(&credentialValue).
				Title("Value"),
			huh.NewInput().
				Key("Note").
				Title("Note").
				Placeholder("You've 255 characters to provide origination context, if available - or just be snarky").
				Value(&credentialNote).
				Description("Brief, critical information - put that here - it'll show in the listing"),
		), huh.NewGroup(
			huh.NewConfirm().
				Key("AddCredential").
				Title(addOrUpdateCredential).
				Affirmative("Yes").
				Negative("No").
				Validate(func(c bool) error {
					if c { // all this crap simply b/c m.form.GetBool("AddFlag") value isn't updated -> followed by a final form validate check... pita...
						s := m.form.GetString("Username")
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

func (m ModelMakeCredential) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		if m.form.GetBool("AddCredential") {
			if m.CredentialId != "" {
				err := database.UpdateCredential(
					m.CredentialId,
					m.form.GetString("Authenticator"),
					m.form.GetString("Username"),
					m.form.GetString("Value"),
					m.form.GetString("Type"),
					m.form.GetString("Format"),
					m.form.GetString("Encoding"),
					m.form.GetString("Note"),
					m.form.GetString("Cleartext"),
				)
				if err != nil {
					return m, constant.ErrorMessage(err)
				}
			} else {
				err := database.AddCredential(
					m.form.GetString("Authenticator"),
					m.form.GetString("Username"),
					m.form.GetString("Value"),
					m.form.GetString("Type"),
					m.form.GetString("Format"),
					m.form.GetString("Encoding"),
					m.form.GetString("Note"),
					m.form.GetString("Cleartext"),
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

func (m ModelMakeCredential) Init() tea.Cmd {
	return m.form.Init()
}

func (m ModelMakeCredential) View() string {
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

func (m ModelMakeCredential) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}
