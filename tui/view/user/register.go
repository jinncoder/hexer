package user

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/archimoebius/hexer/app/cache"
	"github.com/archimoebius/hexer/tui/constant"
	"github.com/archimoebius/hexer/util/database"
	"github.com/asaskevich/govalidator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/google/uuid"
)

type ModelMakeUser struct {
	lg      *lipgloss.Renderer
	styles  *constant.Styles
	form    *huh.Form
	Width   int
	Height  int
	Session *ssh.Session
}

func validateUsername(username string) error {
	if len(username) <= 1 || len(username) > 255 {
		return fmt.Errorf("username to short")
	}

	if !govalidator.IsAlphanumeric(username) {
		return fmt.Errorf("username only allows alphanumeric values")
	}

	if database.DoesUserValueExist("name", username) == nil {
		return fmt.Errorf("username already exists - try another")
	}
	return nil
}

func validateEMail(email string) error {
	if len(email) <= 1 || len(email) > 255 {
		return fmt.Errorf("E-Mail required")
	}
	if !govalidator.IsEmail(email) {
		return fmt.Errorf("invalid E-Mail")
	}
	if database.DoesUserValueExist("email", email) == nil {
		return fmt.Errorf("email already exists - try another")
	}
	return nil
}

func validateSSHPublicKey(sshPublicKey string) error {
	_, _, _, _, err := ssh.ParseAuthorizedKey(
		[]byte(sshPublicKey),
	)
	if err != nil {
		return fmt.Errorf("failed to parse SSH Public Key")
	}
	if database.DoesUserValueExist("ssh_public_key", sshPublicKey) == nil {
		return fmt.Errorf("ssh public key already exists - try another")
	}

	return nil
}

func NewModelMakeUser(session *ssh.Session, mui *constant.MakeUserInput) (*ModelMakeUser, error) {

	m := &ModelMakeUser{
		Session: session,
	}

	m.Width = mui.Width
	m.Height = mui.Height
	m.lg = lipgloss.DefaultRenderer()
	m.styles = constant.NewStyles(m.lg)

	publicKey := (*session).PublicKey()
	var sshPublicKeyType = ""
	var sshPublicKeyData = ""

	if publicKey != nil {
		sshPublicKeyType = publicKey.Type()
		sshPublicKeyData = base64.StdEncoding.EncodeToString((*session).PublicKey().Marshal())
	}
	var sshPublicKey = fmt.Sprintf("%s %s", sshPublicKeyType, sshPublicKeyData)

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("EMail").
				Title("E-Mail").
				Placeholder("Your E-Mail address").
				Validate(validateEMail).
				Description("Your Prefered E-Mail address"),
			huh.NewInput().
				Key("Username").
				Title("Username").
				Placeholder("Your Fancy Username Here").
				Validate(validateUsername).
				Description("Your handle while using this application"),
			// TODO: expose the web UI for users?
			// huh.NewInput().
			// 	Title("Password").
			// 	EchoMode(huh.EchoModePassword).
			// 	Description("Set your Password"),
			huh.NewInput().
				Key("SSHPublicKey").
				Title("SSH Public Key").
				Placeholder("The contents, of that weird *.pub file you get, from running `ssh-keygen`").
				Validate(validateSSHPublicKey).
				Value(&sshPublicKey).
				Description("Your public SSH Key (for authentication)"),
			huh.NewConfirm().
				Key("AddUser").
				Title("Done?").
				Affirmative("Let's GO!").
				Negative("Abort!").
				Validate(func(c bool) error {
					if c { // all this crap simply b/c m.form.GetBool("AddProject") value isn't updated -> followed by a final form validate check... pita...
						s := m.form.GetString("Username")
						err := validateUsername(s)
						if err != nil {
							return err
						}

						s = m.form.GetString("EMail")
						err = validateEMail(s)
						if err != nil {
							return err
						}

						s = m.form.GetString("SSHPublicKey")
						err = validateSSHPublicKey(s)
						if err != nil {
							return err
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

func (m ModelMakeUser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			(*m.Session).Write([]byte("K'bye!")) // #nosec G104
			(*m.Session).Exit(0)                 // #nosec G104
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {

		if m.form.GetBool("AddUser") {
			username := m.form.GetString("Username")
			email := m.form.GetString("EMail")
			sshpublickey := m.form.GetString("SSHPublicKey")

			err := database.AddUser(email, username, uuid.New().String(), sshpublickey)

			if err != nil {
				(*m.Session).Write([]byte(fmt.Sprintf("Failed to register: %v", err))) // #nosec G104
				(*m.Session).Exit(0)                                                   // #nosec G104
			}

			if username == "administrate" {
				users, err := database.GetUsersByName("administrate")

				if err != nil {
					(*m.Session).Write([]byte(fmt.Sprintf("Failed to register: %v", err))) // #nosec G104
					(*m.Session).Exit(0)                                                   // #nosec G104
				}

				var user = users[0]

				database.VerifyUser(user.Id)

				if err != nil {
					(*m.Session).Write([]byte(fmt.Sprintf("Failed to register: %v", err))) // #nosec G104
					(*m.Session).Exit(0)                                                   // #nosec G104
				}

				cache.AddUserKeyToCache(user.Name, user.SSHPublicKey)
			}

			(*m.Session).Write([]byte("Registered - now wait to be verified!")) // #nosec G104
			(*m.Session).Exit(0)                                                // #nosec G104
		} else {
			(*m.Session).Write([]byte("K'bye!")) // #nosec G104
			(*m.Session).Exit(0)                 // #nosec G104
		}
	}

	if m.form.State == huh.StateAborted {
		(*m.Session).Write([]byte("K'bye!")) // #nosec G104
		(*m.Session).Exit(0)                 // #nosec G104
	}

	return m, tea.Batch(cmds...)
}

func (m ModelMakeUser) Init() tea.Cmd {
	return m.form.Init()
}

func (m ModelMakeUser) View() string {
	s := m.styles
	header := constant.AppBoundaryView(m.styles, m.Width, "", "")

	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(v)

	// Status (right side)
	var status string
	{
		const statusWidth = 28
		statusMarginLeft := m.Width/8 - statusWidth - lipgloss.Width(form) - s.Status.GetMarginRight()
		status = s.Status.
			Height(lipgloss.Height(form)).
			Width(statusWidth).
			MarginLeft(statusMarginLeft).
			Render(
				s.StatusHeader.Render("Username: ") + "\n  " + m.form.GetString("Username") + "\n" +
					s.StatusHeader.Render("E-Mail: ") + "\n  " + m.form.GetString("EMail") + "\n",
			)
	}

	errors := m.form.Errors()

	if len(errors) > 0 {
		header = constant.AppErrorBoundaryView(m.styles, m.Width, m.errorView())
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, form, status)

	footer := constant.AppBoundaryView(m.styles, m.Width, m.form.Help().ShortHelpView(m.form.KeyBinds()), "")
	if len(errors) > 0 {
		footer = constant.AppErrorBoundaryView(m.styles, m.Width, "")
	}
	m.styles.Base.Width(m.Width).Height(m.Height)

	return s.Base.Render(header + "\n" + body + "\n\n" + footer)
}

func (m ModelMakeUser) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}
