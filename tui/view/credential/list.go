package credential

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

type ModelListCredential struct {
	keys                      *listKeyMap
	lg                        *lipgloss.Renderer
	styles                    *constant.Styles
	list                      list.Model
	Width                     int
	Height                    int
	OnlyCleartext             bool
	ProjectId                 string
	ProjectPickerLastTabIndex int
	AuthenticatorId           string
	Session                   *ssh.Session
}

type listKeyMap struct {
	insertItem key.Binding
	editItem   key.Binding
	removeItem key.Binding
	copyItem   key.Binding
}

type item struct {
	title       string
	description string
	id          string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

func NewModelListCredential(session *ssh.Session, lfi *constant.ListCredentialInput) (*ModelListCredential, error) {
	var lg = lipgloss.DefaultRenderer()

	m := &ModelListCredential{
		Session:                   session,
		Width:                     lfi.Width,
		Height:                    lfi.Height,
		ProjectId:                 lfi.ProjectId,
		ProjectPickerLastTabIndex: lfi.LastTabIndex,
		AuthenticatorId:           lfi.AuthenticatorId,
		lg:                        lg,
		styles:                    constant.NewStyles(lg),
		OnlyCleartext:             false,

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
			copyItem: key.NewBinding(
				key.WithHelp("c", "copy"),
			),
		},
	}
	delegate := list.NewDefaultDelegate()
	delegate.Styles = m.styles.ListDefaultItemStyle

	credentials, err := database.GetProjectCredentialList(m.ProjectId, m.OnlyCleartext)

	if err != nil {
		return nil, err
	}

	items := make([]list.Item, len(credentials))

	for idx, credential := range credentials {

		items[idx] = item{
			title:       credential.ListTitle(),
			description: credential.Description(),
			id:          credential.Id,
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
			key.NewBinding(
				key.WithKeys("alt+c"),
				key.WithHelp("alt+c", "only cleartext"),
			),
		}
	}
	m.list.SetSize(10, 30)

	return m, nil
}

func (m ModelListCredential) Init() tea.Cmd {
	return nil
}

func (m ModelListCredential) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return m, constant.SwitchModeCmd(constant.ModeMakeCredential,
					constant.NewMakeCredentialInput(m.Width, m.Height, m.ProjectId, "", m.ProjectPickerLastTabIndex, m.AuthenticatorId),
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
			return m, constant.SwitchModeCmd(constant.ModeMakeCredential,
				constant.NewMakeCredentialInput(m.Width, m.Height, m.ProjectId, id, m.ProjectPickerLastTabIndex, m.AuthenticatorId),
			)

		case "c":
			return m, constant.SwitchModeCmd(constant.ModeCopyCredential,
				constant.NewCopyCredentialInput(m.Width, m.Height, m.ProjectId, m.ProjectPickerLastTabIndex, m.AuthenticatorId, id),
			)

		case "alt+c":
			m.OnlyCleartext = !m.OnlyCleartext
			credentials, err := database.GetProjectCredentialList(m.ProjectId, m.OnlyCleartext)

			if err != nil {
				return m, constant.ErrorMessage(err)
			}

			items := make([]list.Item, len(credentials))

			for idx, credential := range credentials {

				items[idx] = item{
					title:       credential.ListTitle(),
					description: credential.Description(),
					id:          credential.Id,
				}
			}

			return m, m.list.SetItems(items)
		case "x":

			if view.ConfirmDelete(m.Session, title) {

				err := database.RemoveCredential(id)
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

func (m ModelListCredential) View() string {
	header := constant.AppBoundaryView(m.styles, m.Width, "", " ")

	m.styles.Base.Width(m.Width)
	m.list.SetSize(m.Width, (m.Height/5)*4)

	return m.styles.Base.Render(header + "\n" + m.list.View())
}
