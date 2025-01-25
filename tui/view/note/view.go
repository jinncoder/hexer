package note

import (
	"fmt"

	"github.com/archimoebius/hexer/tui/constant"
	"github.com/archimoebius/hexer/util/database"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

type ModelViewNote struct {
	lg                        *lipgloss.Renderer
	styles                    *constant.Styles
	Width                     int
	Height                    int
	ProjectId                 string
	NoteId                    string
	ProjectPickerLastTabIndex int
	Session                   *ssh.Session
	viewport                  viewport.Model
	rendered                  string
}

func NewModelViewNote(session *ssh.Session, mfi *constant.ViewNoteInput) (*ModelViewNote, error) {

	m := &ModelViewNote{
		Session: session,
	}

	m.Width = mfi.Width
	m.Height = mfi.Height
	m.lg = lipgloss.DefaultRenderer()
	m.styles = constant.NewStyles(m.lg)
	m.ProjectId = mfi.ProjectId
	m.NoteId = mfi.NoteId
	m.ProjectPickerLastTabIndex = mfi.LastTabIndex

	contents, err := database.GetNotesAsMarkdown(m.ProjectId)
	if err != nil {
		return m, fmt.Errorf("error reading note: %v", err)
	}

	md, err := glamour.Render(string(contents), m.styles.Background)
	if err != nil {
		return m, err
	}
	m.rendered = md

	m.viewport = viewport.New(m.Width, m.Height)
	m.viewport.Style = m.styles.Base
	m.viewport.YPosition = m.Height
	m.viewport.SetContent(m.rendered)
	m.viewport.GotoTop()

	return m, nil
}

func (m ModelViewNote) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		m.viewport = viewport.New(msg.Width, msg.Height)
		m.viewport.YPosition = msg.Height
		m.viewport.SetContent(m.rendered)
		m.viewport.GotoTop()

	case tea.KeyMsg:
		switch msg.String() {
		case "home":
			m.viewport.GotoTop()
		case "end":
			m.viewport.GotoBottom()
		case "esc":
			return m, constant.SwitchModeCmd(constant.ModePickProject,
				constant.NewPickProjectInput(m.Width, m.Height, m.ProjectId, m.ProjectPickerLastTabIndex),
			)
		default:
			m.viewport.Update(msg)
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	return m, tea.Batch(cmd)
}

func (m ModelViewNote) Init() tea.Cmd {
	return nil
}

func (m ModelViewNote) View() string {
	s := m.styles
	header := constant.AppBoundaryView(m.styles, m.Width, "", "")

	m.viewport.Height = m.Height / 8 * 7

	body := lipgloss.JoinHorizontal(lipgloss.Top, m.viewport.View())

	footer := constant.AppBoundaryView(m.styles, m.Width, m.helpView(), "")

	m.styles.Base.Width(m.Width).Height(m.Height / 8 * 1)

	return s.Base.Render(header + "\n" + body + "\n\n" + footer)
}

func (m ModelViewNote) helpView() string {
	return "↑ (page up) / ↓ (page down): scroll pager (arrows) • home: goto top • end: goto bottom • esc: close preview"
}
