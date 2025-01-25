package tui

import (
	"errors"
	"reflect"

	constant "github.com/archimoebius/hexer/tui/constant"
	"github.com/archimoebius/hexer/tui/view"
	"github.com/archimoebius/hexer/tui/view/authenticator"
	credential "github.com/archimoebius/hexer/tui/view/credential"
	projectFlag "github.com/archimoebius/hexer/tui/view/flag"
	note "github.com/archimoebius/hexer/tui/view/note"
	project "github.com/archimoebius/hexer/tui/view/project"
	"github.com/archimoebius/hexer/tui/view/user"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

var (
	ErrInvalidSwitchMode    = errors.New("invalid SwitchMode")
	ErrInvalidTypeAssertion = errors.New("invalid type assertion")
)

type Input struct {
	mode     constant.Mode
	switchIn constant.SwitchModeInput
	session  *ssh.Session
}

func NewInput(session *ssh.Session, mode constant.Mode, switchIn constant.SwitchModeInput) *Input {
	return &Input{
		session:  session,
		mode:     mode,
		switchIn: switchIn,
	}
}

type Model struct {
	child        tea.Model
	style        lipgloss.Style
	forceQuitKey key.Binding
	width        int
	height       int
	session      *ssh.Session
}

func NewModel(in *Input) (*Model, error) {
	m := &Model{
		style: lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center),
		forceQuitKey: key.NewBinding(
			key.WithKeys(
				[]string{
					"ctrl+d",
				}...,
			),
		),
		session: in.session,
	}

	err := m.setChild(in.mode, in.switchIn)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return m.child.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// NOTE: Windows does not have support for reporting when resizes occur as it does not support the SIGWINCH signal.
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if key.Matches(msg, m.forceQuitKey) {
			return m, tea.Quit
		}
	case constant.ErrMsg:
		child, err := view.NewModelFatalError(msg.E)
		if err != nil {
			panic(err)
		}
		m.child = child

		return m, m.child.Init()
	case constant.SwitchModeMsg:
		err := m.setChild(msg.Target, msg.Input)
		if err != nil {
			panic(err)
		}
		return m, m.child.Init()
	}

	var cmd tea.Cmd
	m.child, cmd = m.child.Update(msg)

	return m, cmd
}

func (m *Model) View() string {
	m.style.Width(m.width).Height(m.height)

	return m.style.Render(
		m.child.View(),
	)
}

func (m *Model) setChild(mode constant.Mode, switchIn constant.SwitchModeInput) error {
	if rv := reflect.ValueOf(switchIn); !rv.IsValid() || rv.IsNil() {
		return errors.New("switchIn is not valid")
	}

	switch mode {
	case constant.ModeListProject:
		lpi, ok := switchIn.(*constant.ListProjectInput)
		if !ok {
			return ErrInvalidTypeAssertion
		}
		child, err := project.NewModelListProject(m.session, lpi)
		if err != nil {
			return err
		}
		m.child = child
	case constant.ModeMakeProject:
		mpi, ok := switchIn.(*constant.MakeProjectInput)
		if !ok {
			return ErrInvalidTypeAssertion
		}
		child, err := project.NewModelMakeProject(m.session, mpi)
		if err != nil {
			return err
		}
		m.child = child
	case constant.ModePickProject:
		ppi, ok := switchIn.(*constant.PickProjectInput)
		if !ok {
			return ErrInvalidTypeAssertion
		}
		child, err := project.NewModelPickProject(m.session, ppi)
		if err != nil {
			return err
		}
		m.child = child
	case constant.ModeListAuthenticator:
		lai, ok := switchIn.(*constant.ListAuthenticatorInput)
		if !ok {
			return ErrInvalidTypeAssertion
		}
		child, err := authenticator.NewModelListAuthenticator(m.session, lai)
		if err != nil {
			return err
		}
		m.child = child
	case constant.ModeMakeAuthenticator:
		mai, ok := switchIn.(*constant.MakeAuthenticatorInput)
		if !ok {
			return ErrInvalidTypeAssertion
		}
		child, err := authenticator.NewModelMakeAuthenticator(m.session, mai)
		if err != nil {
			return err
		}
		m.child = child
	case constant.NewModelCopyAuthenticator:
		cai, ok := switchIn.(*constant.CopyAuthenticatorInput)
		if !ok {
			return ErrInvalidTypeAssertion
		}
		child, err := authenticator.NewModelCopyAuthenticator(m.session, cai)
		if err != nil {
			return err
		}
		m.child = child
	case constant.ModeMakeFlag:
		mfi, ok := switchIn.(*constant.MakeFlagInput)
		if !ok {
			return ErrInvalidTypeAssertion
		}
		child, err := projectFlag.NewModelMakeFlag(m.session, mfi)
		if err != nil {
			return err
		}
		m.child = child
	case constant.ModeMakeCredential:
		mci, ok := switchIn.(*constant.MakeCredentialInput)
		if !ok {
			return ErrInvalidTypeAssertion
		}
		child, err := credential.NewModelMakeCredential(m.session, mci)
		if err != nil {
			return err
		}
		m.child = child
	case constant.ModeCopyCredential:
		cci, ok := switchIn.(*constant.CopyCredentialInput)
		if !ok {
			return ErrInvalidTypeAssertion
		}
		child, err := credential.NewModelCopyCredential(m.session, cci)
		if err != nil {
			return err
		}
		m.child = child
	case constant.ModeMakeUser:
		mui, ok := switchIn.(*constant.MakeUserInput)
		if !ok {
			return ErrInvalidTypeAssertion
		}
		child, err := user.NewModelMakeUser(m.session, mui)
		if err != nil {
			return err
		}
		m.child = child
	case constant.ModeMakeNote:
		mni, ok := switchIn.(*constant.MakeNoteInput)
		if !ok {
			return ErrInvalidTypeAssertion
		}
		child, err := note.NewModelMakeNote(m.session, mni)
		if err != nil {
			return err
		}
		m.child = child
	case constant.ModeViewNote:
		vni, ok := switchIn.(*constant.ViewNoteInput)
		if !ok {
			return ErrInvalidTypeAssertion
		}
		child, err := note.NewModelViewNote(m.session, vni)
		if err != nil {
			return err
		}
		m.child = child
	default:
		return ErrInvalidSwitchMode
	}
	return nil
}
