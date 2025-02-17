package user

import (
	"fmt"
	"strings"

	"github.com/archimoebius/hexer/app/cache"
	"github.com/archimoebius/hexer/tui/constant"
	"github.com/archimoebius/hexer/util/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

type ModelAdministrateUser struct {
	lg      *lipgloss.Renderer
	styles  *constant.Styles
	form    *huh.Form
	Width   int
	Height  int
	Session *ssh.Session
}

type UserSelection struct {
	Id       string
	Username string
}

var selectedUserItems []UserSelection

func NewModelModelAdministrateUser(session *ssh.Session, mui *constant.MakeAdministrateUserInput) (*ModelAdministrateUser, error) {

	m := &ModelAdministrateUser{
		Session: session,
	}

	m.Width = mui.Width
	m.Height = mui.Height
	m.lg = lipgloss.DefaultRenderer()
	m.styles = constant.NewStyles(m.lg)

	users, err := database.GetUsers()
	if err != nil {
		return nil, err
	}

	var options []huh.Option[UserSelection]

	for _, user := range users {
		var username = fmt.Sprintf("N %s - %s", user.Name, user.EMail)
		if user.Verified {
			username = fmt.Sprintf("V %s - %s", user.Name, user.EMail)
		}
		var selection = UserSelection{Id: user.Id, Username: username}
		options = append(options, huh.NewOption(username, selection))
	}

	var userActionOptions []huh.Option[string]

	userActionOptions = append(userActionOptions, huh.NewOption("verify", "verify"))
	userActionOptions = append(userActionOptions, huh.NewOption("delete", "delete"))

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[UserSelection]().
				Title("Select 1..N to users").
				Description("There shouldn't be any users with the same name/email/id").
				Height(10).
				Value(&selectedUserItems).
				Key("Id").
				Options(options...),
			huh.NewSelect[string]().
				Key("Action").
				Description("The action to take on the selected users").
				Title("Action").
				Options(userActionOptions[:]...).
				WithHeight(8),
			huh.NewConfirm().
				Key("Apply").
				Title("Apply?").
				Affirmative("Let's GO!").
				Negative("Abort!"),
		),
	).
		WithWidth(m.Width / 2).
		WithShowHelp(false).
		WithShowErrors(false)

	return m, nil
}

func (m ModelAdministrateUser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			(*m.Session).Write([]byte("K'bye!\n")) // #nosec G104
			(*m.Session).Exit(0)                   // #nosec G104
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {

		if m.form.GetBool("Apply") {
			action := m.form.GetString("Action")

			if action == "verify" {
				for _, entry := range selectedUserItems {

					user, err := database.GetUserById(entry.Id)

					if err != nil {
						(*m.Session).Write([]byte(fmt.Sprintf("Failed to find user: %v\n", err))) // #nosec G104
						continue
					}

					cache.AddUserKeyToCache(user.Name, user.SSHPublicKey)

					err = database.VerifyUser(entry.Id)

					if err != nil {
						(*m.Session).Write([]byte(fmt.Sprintf("Failed to remove user: %v due to error: %v\n", user, err))) // #nosec G104
					}
				}
			}

			if action == "delete" {
				for _, entry := range selectedUserItems {

					user, err := database.GetUserById(entry.Id)

					if err != nil {
						(*m.Session).Write([]byte(fmt.Sprintf("Failed to find user: %v\n", err))) // #nosec G104
						continue
					}
					cache.RemoveUserPublicSSHKeyFromCache(user.Name)

					err = database.RemoveUser(entry.Id)

					if err != nil {
						(*m.Session).Write([]byte(fmt.Sprintf("Failed to remove user: %v due to error: %v\n", user, err))) // #nosec G104
					}
				}
			}

			(*m.Session).Write([]byte("K'bye!\n")) // #nosec G104
			(*m.Session).Exit(0)                   // #nosec G104
		} else {
			(*m.Session).Write([]byte("K'bye!\n")) // #nosec G104
			(*m.Session).Exit(0)                   // #nosec G104
		}
	}

	if m.form.State == huh.StateAborted {
		(*m.Session).Write([]byte("K'bye!\n")) // #nosec G104
		(*m.Session).Exit(0)                   // #nosec G104
	}

	return m, tea.Batch(cmds...)
}

func (m ModelAdministrateUser) Init() tea.Cmd {
	return m.form.Init()
}

func (m ModelAdministrateUser) View() string {
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

func (m ModelAdministrateUser) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}
