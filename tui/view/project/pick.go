package project

import (
	"fmt"
	"strings"

	"github.com/archimoebius/hexer/tui/constant"
	authenticator "github.com/archimoebius/hexer/tui/view/authenticator"
	credential "github.com/archimoebius/hexer/tui/view/credential"
	flag "github.com/archimoebius/hexer/tui/view/flag"
	note "github.com/archimoebius/hexer/tui/view/note"
	"github.com/archimoebius/hexer/util/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

type ModelPickProject struct {
	lg           *lipgloss.Renderer
	styles       *constant.Styles
	Width        int
	Height       int
	ProjectId    string
	ProjectTitle string
	Tabs         []tea.Model
	TabTitle     []string
	ActiveTab    int
	Session      *ssh.Session
}

func NewModelPickProject(session *ssh.Session, lpi *constant.PickProjectInput) (*ModelPickProject, error) {
	var lg = lipgloss.DefaultRenderer()

	projectTitle, err := database.GetProjectTitle(lpi.ProjectId)
	if err != nil {
		return nil, err
	}

	tea.SetWindowTitle("Hexer : " + projectTitle)

	m := &ModelPickProject{
		Session:      session,
		Width:        lpi.Width,
		Height:       lpi.Height,
		ProjectId:    lpi.ProjectId,
		ProjectTitle: projectTitle,
		ActiveTab:    lpi.TabIndex,
		lg:           lg,
		styles:       constant.NewStyles(lg),
	}

	// Addition of "Authenticator" tab
	authenticatorTab, err := authenticator.NewModelListAuthenticator(
		m.Session,
		constant.NewListAuthenticatorInput(lpi.Width, lpi.Height, lpi.ProjectId, m.ActiveTab),
	)
	if err != nil {
		return nil, err
	}
	m.Tabs = append(m.Tabs, authenticatorTab)
	m.TabTitle = append(m.TabTitle, "Authenticators")

	// Addition of "Flags" tab
	flagTab, err := flag.NewModelListFlag(
		m.Session,
		constant.NewListFlagInput(lpi.Width, lpi.Height, lpi.ProjectId, m.ActiveTab),
	)
	if err != nil {
		return nil, err
	}
	m.Tabs = append(m.Tabs, flagTab)
	m.TabTitle = append(m.TabTitle, "Flags")

	// Addition of "Credentials" tab
	credentialTab, err := credential.NewModelListCredential(
		m.Session,
		constant.NewListCredentialInput(lpi.Width, lpi.Height, lpi.ProjectId, m.ActiveTab, ""),
	)
	if err != nil {
		return nil, err
	}
	m.Tabs = append(m.Tabs, credentialTab)
	m.TabTitle = append(m.TabTitle, "Credentials")

	// Addition of "Notes" tab
	noteTab, err := note.NewModelListNote(
		m.Session,
		constant.NewListNoteInput(lpi.Width, lpi.Height, lpi.ProjectId, m.ActiveTab),
	)
	if err != nil {
		return nil, err
	}
	m.Tabs = append(m.Tabs, noteTab)
	m.TabTitle = append(m.TabTitle, "Notes")

	// More tabs here if need be...

	return m, nil
}

func (m ModelPickProject) Init() tea.Cmd {
	return nil
}

func (m ModelPickProject) updateTabIndex() tea.Cmd {
	return func() tea.Msg {
		return constant.ProjectTabChangeMsg{
			ActiveTabIndex: m.ActiveTab,
		}
	}
}

func (m ModelPickProject) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {

		case "right", "l", "n", "tab":
			m.ActiveTab = min(m.ActiveTab+1, len(m.Tabs)-1)
			cmds = append(cmds, m.updateTabIndex())
		case "left", "h", "p", "shift+tab":
			m.ActiveTab = max(m.ActiveTab-1, 0)
			cmds = append(cmds, m.updateTabIndex())
		case "esc":
			return m, constant.SwitchModeCmd(constant.ModeListProject,
				constant.NewListProjectInput(m.Width, m.Height),
			)
		}
	}

	var cmd tea.Cmd
	m.Tabs[m.ActiveTab], cmd = m.Tabs[m.ActiveTab].Update(msg)

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m ModelPickProject) View() string {
	header := constant.AppBoundaryView(m.styles, m.Width, fmt.Sprintf("%s (%s)", m.ProjectTitle, m.ProjectId), "")

	m.styles.Base.Width(m.Width)

	doc := strings.Builder{}

	var renderedTabs []string

	for i, t := range m.TabTitle {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.TabTitle)-1, i == m.ActiveTab
		if isActive {
			style = m.styles.ActiveTabStyle
		} else {
			style = m.styles.InactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()

		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}

		style = style.Border(border).Bold(isActive)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")

	return m.styles.Base.Render(doc.String() + header + "\n" + m.Tabs[m.ActiveTab].View())
}
