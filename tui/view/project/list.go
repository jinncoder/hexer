package project

import (
	"time"

	"github.com/archimoebius/hexer/tui/constant"
	"github.com/archimoebius/hexer/tui/view"
	"github.com/archimoebius/hexer/util/database"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

type ModelListProject struct {
	keys      *listKeyMap
	lg        *lipgloss.Renderer
	styles    *constant.Styles
	list      list.Model
	Width     int
	Height    int
	Session   *ssh.Session
	ProjectId string
}

type listKeyMap struct {
	insertItem key.Binding
	chooseItem key.Binding
	removeItem key.Binding
	editItem   key.Binding
}

type item struct {
	title       string
	description string
	id          string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

func NewModelListProject(session *ssh.Session, lpi *constant.ListProjectInput) (*ModelListProject, error) {
	var lg = lipgloss.DefaultRenderer()

	m := &ModelListProject{
		Session: session,
		Width:   lpi.Width,
		Height:  lpi.Height,
		lg:      lg,
		styles:  constant.NewStyles(lg),

		keys: &listKeyMap{
			insertItem: key.NewBinding(
				key.WithKeys("a"),
				key.WithHelp("a", "add item"),
			),
			chooseItem: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "choose"),
			),
			editItem: key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "edit"),
			),
			removeItem: key.NewBinding(
				key.WithKeys("x", "backspace"),
				key.WithHelp("x", "delete"),
			),
		},
	}
	delegate := list.NewDefaultDelegate()
	delegate.Styles = m.styles.ListDefaultItemStyle

	projects, err := database.GetProjectList()

	if err != nil {
		return nil, err
	}

	items := make([]list.Item, len(projects))

	for idx, project := range projects {

		items[idx] = item{
			title:       project.ListTitle(),
			description: project.Note,
			id:          project.Id,
		}
	}

	m.list = list.New(items, delegate, 0, 0)
	m.list.DisableQuitKeybindings()
	m.list.Title = ""
	m.list.Styles.Title = m.styles.TitleStyle
	m.list.StatusMessageLifetime = time.Duration(3.0 * float64(time.Second))
	m.list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			m.keys.insertItem,
			m.keys.chooseItem,
			m.keys.removeItem,
			m.keys.editItem,
			key.NewBinding(
				key.WithKeys("ctrl+d"),
				key.WithHelp("ctrl+d", "exit"),
			),
		}
	}

	return m, nil
}

func (m ModelListProject) Init() tea.Cmd {
	_ = tea.ClearScreen()
	_ = tea.HideCursor()

	tea.SetWindowTitle("Hexer")

	return nil
}

func (m ModelListProject) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() != list.Filtering {

			if key.Matches(msg, m.keys.insertItem) {
				return m, constant.SwitchModeCmd(constant.ModeMakeProject,
					constant.NewMakeProjectInput(m.Width, m.Height, ""),
				)
			}
		}

		var title string
		var id string
		var index int

		if i, ok := m.list.SelectedItem().(item); ok {
			title = i.Title()
			id = i.id
			index = m.list.Index()
		}

		switch msg.String() {

		case "e":
			return m, constant.SwitchModeCmd(constant.ModeMakeProject,
				constant.NewMakeProjectInput(m.Width, m.Height, id),
			)

		case "enter":
			return m, constant.SwitchModeCmd(constant.ModePickProject,
				constant.NewPickProjectInput(m.Width, m.Height, id, 0),
			)

		case "x":

			if view.ConfirmDelete(m.Session, title) {

				err := database.RemoveProject(id)
				if err != nil {
					return m, constant.ErrorMessage(err)
				}

				m.list.RemoveItem(index)

				return m, m.list.NewStatusMessage(m.styles.Status.Render("Deleted " + title))
			} else {
				return m, m.list.NewStatusMessage(m.styles.Status.Render("Kept " + title))
			}

		case "ctrl+d":
			if m.list.FilterState() == list.Unfiltered {
				_ = tea.ClearScreen()

				return m, tea.Quit
			}
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m ModelListProject) View() string {
	header := constant.AppBoundaryView(m.styles, m.Width, "Projects", "")

	m.styles.Base.Width(m.Width)
	m.list.SetSize(m.Width, (m.Height/5)*4)

	return m.styles.Base.Render(header + "\n" + m.list.View())
}
