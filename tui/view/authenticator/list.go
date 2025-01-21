package authenticator

import (
	"sort"
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

type ModelListAuthenticator struct {
	keys                      *listKeyMap
	lg                        *lipgloss.Renderer
	styles                    *constant.Styles
	list                      list.Model
	Width                     int
	Height                    int
	ProjectId                 string
	ProjectTitle              string
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

func NewModelListAuthenticator(session *ssh.Session, lci *constant.ListAuthenticatorInput) (*ModelListAuthenticator, error) {
	var lg = lipgloss.DefaultRenderer()

	projectTitle, err := database.GetProjectTitle(lci.ProjectId)
	if err != nil {
		return nil, err
	}

	m := &ModelListAuthenticator{
		Session:                   session,
		Width:                     lci.Width,
		Height:                    lci.Height,
		ProjectId:                 lci.ProjectId,
		ProjectTitle:              projectTitle,
		ProjectPickerLastTabIndex: lci.LastTabIndex,
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

	authenticators, err := database.GetAuthenticatorList(m.ProjectId)

	if err != nil {
		return nil, err
	}

	items := make([]list.Item, len(authenticators))

	for idx, authenticator := range authenticators {

		items[idx] = item{
			title:       authenticator.ListTitle(),
			description: authenticator.Note,
			id:          authenticator.Id,
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
				key.WithKeys("ctrl+h"),
				key.WithHelp("ctrl+h", "/etc/hosts"),
			),
		}
	}

	return m, nil
}

func (m ModelListAuthenticator) Init() tea.Cmd {
	return nil
}

func (m ModelListAuthenticator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return m, constant.SwitchModeCmd(constant.ModeMakeAuthenticator,
					constant.NewMakeAuthenticatorInput(m.Width, m.Height, m.ProjectId, "", m.ProjectPickerLastTabIndex),
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
			return m, constant.SwitchModeCmd(constant.ModeMakeAuthenticator,
				constant.NewMakeAuthenticatorInput(m.Width, m.Height, m.ProjectId, id, m.ProjectPickerLastTabIndex),
			)

		case "c":
			return m, constant.SwitchModeCmd(constant.NewModelCopyAuthenticator,
				constant.NewCopyAuthenticatorInput(m.Width, m.Height, m.ProjectId, m.ProjectPickerLastTabIndex, id))

		case "ctrl+h":
			authenticators, err := database.GetAuthenticatorList(m.ProjectId)

			if err != nil {
				return m, m.list.NewStatusMessage(m.styles.Status.Render("Error " + title + " : failed to obtain authenticators"))
			}

			var etc_hosts string
			authenticatorMap := make(map[string]string, len(authenticators))

			for _, authenticator := range authenticators {
				entry := authenticator.Name

				if len(authenticator.FQDN) > 0 {
					entry += " " + authenticator.FQDN
				}

				authenticatorMap[authenticator.IPv4] = entry
			}

			if len(authenticatorMap) <= 0 {
				return m, m.list.NewStatusMessage(m.styles.Status.Render("No authenticators found"))
			}

			// sort output of /etc/hosts file by IP address
			keys := make([]string, 0, len(authenticatorMap))
			for k := range authenticatorMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				etc_hosts += k + "\t" + authenticatorMap[k] + "\n"
			}

			clipboard.Write(clipboard.FmtText, []byte(etc_hosts))

			return m, m.list.NewStatusMessage(m.styles.Status.Render("Copied generated /etc/hosts entries to clipboard"))

		case "x":

			if view.ConfirmDelete(m.Session, title) {

				err := database.RemoveAuthenticator(id)
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
					constant.NewListProjectInput(m.Width, m.Height),
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

func (m ModelListAuthenticator) View() string {
	header := constant.AppBoundaryView(m.styles, m.Width, "", " ")

	m.styles.Base.Width(m.Width)
	m.list.SetSize(m.Width, (m.Height/5)*4)

	return m.styles.Base.Render(header + "\n" + m.list.View())
}
