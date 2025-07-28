package note

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

type ModelMakeNote struct {
	lg                        *lipgloss.Renderer
	styles                    *constant.Styles
	form                      *huh.Form
	Width                     int
	Height                    int
	ProjectId                 string
	NoteId                    string
	ProjectPickerLastTabIndex int
	Session                   *ssh.Session
}

func NewModelMakeNote(session *ssh.Session, mfi *constant.MakeNoteInput) (*ModelMakeNote, error) {

	m := &ModelMakeNote{
		Session: session,
	}

	m.Width = mfi.Width
	m.Height = mfi.Height
	m.lg = lipgloss.DefaultRenderer()
	m.styles = constant.NewStyles(m.lg)
	m.ProjectId = mfi.ProjectId
	m.NoteId = mfi.NoteId
	m.ProjectPickerLastTabIndex = mfi.LastTabIndex

	addOrUpdateNote := "Add Note?"
	noteName := ""
	noteValue := ""
	// var noteProject database.Project
	// var note database.Note

	if m.NoteId != "" {
		addOrUpdateNote = "Update Note?"

		note, err := database.GetNoteById(m.NoteId)
		if err != nil {
			fmt.Printf("failed to load note")
		}

		noteName = note.Name
		noteValue = note.Value

		if note.ProjectId != "" {
			_, err = database.GetProjectById(note.ProjectId)
			if err != nil {
				fmt.Printf("failed to load note project") // TODO: user error message
			}
		}
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("Name").
				Title("Name").
				Placeholder("Note Name").
				Validate(func(s string) error {
					if m.form.GetBool("AddNote") && len(s) <= 1 || len(s) > 255 {
						return fmt.Errorf("%s must be > 1 char and < 255\n", s)
					}
					return nil
				}).
				Value(&noteName).
				Description("The name of the authenticator to track things under"),
			// TODO: need https://github.com/charmbracelet/huh/issues/512
			huh.NewText().
				Key("Value").
				CharLimit(5000000).
				Description("That pesky value that we all seek...").
				Value(&noteValue).
				Title("Value").
				ExternalEditor(true). // TODO: enables RCE on the server depending on the env EDITOR - make a configuration bool...
				WithHeight(m.Height/2),
			// TODO: link notes to authenticators?
			// ), huh.NewGroup(
			// huh.NewSelect[string]().
			// 	Key("Authenticator").
			// 	Description("The authenticator you found the note on").
			// 	Title("Authenticator").
			// 	Validate(func(s string) error {
			// 		if s == "" {
			// 			return fmt.Errorf("note must have an authenticator...for now...just make a 'empty' one...damn nullset...")
			// 		}
			// 		return nil
			// 	}).
			// 	OptionsFunc(func() []huh.Option[string] {
			// 		var options []huh.Option[string]
			// 		authenticators, _ := database.GetAuthenticatorList(m.ProjectId)

			// 		options = append(options, huh.NewOption("Nope...", "").Selected(true)) // TODO: add link_project to db for this case?

			// 		for _, authenticator := range authenticators {
			// 			options = append(options, huh.NewOption(authenticator.ListTitle(), authenticator.Id))
			// 		}

			// 		return options
			// 	}, nil),
			huh.NewConfirm().
				Key("AddNote").
				Title(addOrUpdateNote).
				Affirmative("Yes").
				Negative("No").
				Validate(func(c bool) error {
					if c { // all this crap simply b/c m.form.GetBool("AddNote") value isn't updated -> followed by a final form validate check... pita...
						s := m.form.GetString("Name")
						if len(s) <= 1 || len(s) > 255 {
							return fmt.Errorf("note name to short")
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

func (m ModelMakeNote) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		if m.form.GetBool("AddNote") {
			if m.NoteId != "" {
				err := database.UpdateNote(
					m.NoteId,
					m.ProjectId,
					m.form.GetString("Name"),
					m.form.GetString("Value"),
				)
				if err != nil {
					return m, constant.ErrorMessage(err)
				}
			} else {
				err := database.AddNote(
					m.ProjectId,
					m.form.GetString("Name"),
					m.form.GetString("Value"),
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

func (m ModelMakeNote) Init() tea.Cmd {
	return m.form.Init()
}

func (m ModelMakeNote) View() string {
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

func (m ModelMakeNote) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}
