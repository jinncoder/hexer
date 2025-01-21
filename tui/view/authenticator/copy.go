package authenticator

import (
	"strings"

	"github.com/archimoebius/hexer/tui/constant"
	"github.com/archimoebius/hexer/util/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"golang.design/x/clipboard"
)

type ModelCopyAuthenticator struct {
	lg                        *lipgloss.Renderer
	styles                    *constant.Styles
	form                      *huh.Form
	Width                     int
	Height                    int
	ProjectId                 string
	ProjectPickerLastTabIndex int
	AuthenticatorId           string
	Session                   *ssh.Session
}

func NewModelCopyAuthenticator(session *ssh.Session, cai *constant.CopyAuthenticatorInput) (*ModelCopyAuthenticator, error) {

	m := &ModelCopyAuthenticator{
		Session: session,
	}

	m.Width = cai.Width
	m.Height = cai.Height
	m.lg = lipgloss.DefaultRenderer()
	m.styles = constant.NewStyles(m.lg)
	m.ProjectId = cai.ProjectId
	m.ProjectPickerLastTabIndex = cai.LastTabIndex
	m.AuthenticatorId = cai.AuthenticatorId

	// TODO: handle error state...
	authenticatorCopyList, _ := database.GetAuthenticatorCopyList(m.AuthenticatorId)

	var authenticatorValue string = "failed to copy credential"

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("Format").
				Description("The format to return the selected credential in").
				Title("Copy Format").
				Options(huh.NewOptions(authenticatorCopyList[:]...)...).
				Value(&authenticatorValue).
				WithHeight(8),
		),
	).
		WithWidth(m.Width / 2).
		WithShowHelp(false).
		WithShowErrors(false)

	return m, nil
}

func (m ModelCopyAuthenticator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		clipboard.Write(clipboard.FmtText, []byte(m.form.GetString("Format")))

		return m, constant.SwitchModeCmd(constant.ModePickProject,
			constant.NewPickProjectInput(m.Width, m.Height, m.ProjectId, m.ProjectPickerLastTabIndex),
		)
	}

	return m, tea.Batch(cmds...)
}

func (m ModelCopyAuthenticator) Init() tea.Cmd {
	return m.form.Init()
}

func (m ModelCopyAuthenticator) View() string {
	s := m.styles
	header := constant.AppBoundaryView(m.styles, m.Width, "", " ")

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

func (m ModelCopyAuthenticator) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}
