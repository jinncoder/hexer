package flag

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
	"golang.design/x/clipboard"
)

type ModelListFlag struct {
	keys                      *listKeyMap
	lg                        *lipgloss.Renderer
	styles                    *constant.Styles
	list                      list.Model
	Width                     int
	Height                    int
	ProjectId                 string
	ProjectPickerLastTabIndex int
	Session                   *ssh.Session
}

type listKeyMap struct {
	insertItem key.Binding
	editItem   key.Binding
	removeItem key.Binding
}

type item struct {
	title       string
	description string
	id          string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

func NewModelListFlag(session *ssh.Session, lfi *constant.ListFlagInput) (*ModelListFlag, error) {
	var lg = lipgloss.DefaultRenderer()

	m := &ModelListFlag{
		Session:                   session,
		Width:                     lfi.Width,
		Height:                    lfi.Height,
		ProjectId:                 lfi.ProjectId,
		ProjectPickerLastTabIndex: lfi.LastTabIndex,
		lg:                        lg,
		styles:                    constant.NewStyles(lg),

		keys: &listKeyMap{
			insertItem: key.NewBinding(
				key.WithKeys("a"),
				key.WithHelp("a", "add item"),
			),
			editItem: key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "edit"),
			),
			removeItem: key.NewBinding(
				key.WithKeys("x"),
				key.WithHelp("x", "delete"),
			),
		},
	}
	delegate := list.NewDefaultDelegate()
	delegate.Styles = m.styles.ListDefaultItemStyle

	flags, err := database.GetProjectFlagList(m.ProjectId)

	if err != nil {
		return nil, err
	}

	items := make([]list.Item, len(flags))

	for idx, flag := range flags {

		items[idx] = item{
			title:       flag.ListTitle(),
			description: flag.ListDescription(),
			id:          flag.Id,
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
			m.keys.editItem,
			m.keys.removeItem,
			key.NewBinding(
				key.WithKeys("c"),
				key.WithHelp("c", "copy"),
			),
			key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "back"),
			),
		}
	}

	return m, nil
}

func (m ModelListFlag) Init() tea.Cmd {
	return nil
}

func (m ModelListFlag) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case constant.ProjectTabChangeMsg:
		m.ProjectPickerLastTabIndex = msg.ActiveTabIndex
		return m, nil

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() != list.Filtering {

			if key.Matches(msg, m.keys.insertItem) {
				return m, constant.SwitchModeCmd(constant.ModeMakeFlag,
					constant.NewMakeFlagInput(m.Width, m.Height, m.ProjectId, "", m.ProjectPickerLastTabIndex),
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
		case "c":
			flag, err := database.GetFlagById(id)
			if err != nil {
				return m, m.list.NewStatusMessage(m.styles.Status.Render("Failed to copy flag value for flag: " + title))
			}

			clipboard.Write(clipboard.FmtText, []byte(flag.Value))

			return m, m.list.NewStatusMessage(m.styles.Status.Render("Copied flag value for flag: " + title))

		case "e":
			return m, constant.SwitchModeCmd(constant.ModeMakeFlag,
				constant.NewMakeFlagInput(m.Width, m.Height, m.ProjectId, id, m.ProjectPickerLastTabIndex),
			)

		case "x":

			if view.ConfirmDelete(m.Session, title) {

				err := database.RemoveFlag(id)
				if err != nil {
					return m, constant.ErrorMessage(err)
				}

				m.list.RemoveItem(index)

				return m, m.list.NewStatusMessage(m.styles.Status.Render("Deleted " + title))
			} else {
				return m, m.list.NewStatusMessage(m.styles.Status.Render("Kept " + title))
			}

		case "esc":
			if m.list.FilterState() == list.Unfiltered {
				return m, constant.SwitchModeCmd(constant.ModePickProject,
					constant.NewPickProjectInput(m.Width, m.Height, m.ProjectId, m.ProjectPickerLastTabIndex),
				)
			}
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m ModelListFlag) View() string {
	header := constant.AppBoundaryView(m.styles, m.Width, "", " ")

	m.styles.Base.Width(m.Width)
	m.list.SetSize(m.Width, (m.Height/5)*4)

	return m.styles.Base.Render(header + "\n" + m.list.View())
}
