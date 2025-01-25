package constant

import (
	tea "github.com/charmbracelet/bubbletea"
)

type ErrMsg struct {
	E error
}

func ErrorMessage(error error) func() tea.Msg {
	return func() tea.Msg { return ErrMsg{E: error} }
}

type SwitchModeMsg struct {
	Target Mode
	Input  SwitchModeInput
}

type ProjectTabChangeMsg struct {
	ActiveTabIndex int
}

type SwitchModeInput interface {
	isSwitchModeInput()
}

func SwitchModeCmd(target Mode, in SwitchModeInput) tea.Cmd {
	return func() tea.Msg {
		return SwitchModeMsg{
			Target: target,
			Input:  in,
		}
	}
}

type Mode int

const (
	ModeListProject = Mode(iota)
	ModeListProjectFlag
	ModeListAuthenticator
	ModeMakeProject
	ModeMakeAuthenticator
	ModeMakeFlag
	ModeMakeNote
	ModeMakeCredential
	ModeMakeUser
	ModeViewProject
	ModeCopyCredential
	ModeViewNote
	ModePickProject
	NewModelCopyAuthenticator
)

// SwitchModeInput values --------------------------------------------------

// Project -----------------------------------------------------------------

// -------------------------------------------------- List Project Input

type ListProjectInput struct {
	Width  int
	Height int
}

func NewListProjectInput(width int, height int) *ListProjectInput {
	return &ListProjectInput{
		Width:  width,
		Height: height,
	}
}

func (in *ListProjectInput) isSwitchModeInput() {}

// -------------------------------------------------- View Project Input

type ViewProjectInput struct {
	Mode  Mode
	ID    int
	Title string
}

func NewViewProjectInput(mode Mode, id int, title string) *ViewProjectInput {
	return &ViewProjectInput{
		Mode:  mode,
		ID:    id,
		Title: title,
	}
}

func (in *ViewProjectInput) isSwitchModeInput() {}

// -------------------------------------------------- Make Project Input

type MakeProjectInput struct {
	Width     int
	Height    int
	ProjectId string
}

func NewMakeProjectInput(width int, height int, projectId string) *MakeProjectInput {
	return &MakeProjectInput{
		Width:     width,
		Height:    height,
		ProjectId: projectId,
	}
}

func (in *MakeProjectInput) isSwitchModeInput() {}

// -------------------------------------------------- Pick Project Input

type PickProjectInput struct {
	Width     int
	Height    int
	ProjectId string
	TabIndex  int
}

func NewPickProjectInput(width int, height int, projectId string, tabIndex int) *PickProjectInput {
	return &PickProjectInput{
		Width:     width,
		Height:    height,
		ProjectId: projectId,
		TabIndex:  tabIndex,
	}
}

func (in *PickProjectInput) isSwitchModeInput() {}

// Flag --------------------------------------------------------------------

// -------------------------------------------------- List Flag Input

type ListFlagInput struct {
	Width        int
	Height       int
	ProjectId    string
	LastTabIndex int
}

func NewListFlagInput(width int, height int, projectId string, lastTabIndex int) *ListFlagInput {
	return &ListFlagInput{
		Width:        width,
		Height:       height,
		ProjectId:    projectId,
		LastTabIndex: lastTabIndex,
	}
}

func (in *ListFlagInput) isSwitchModeInput() {}

// -------------------------------------------------- Make Flag Input

type MakeFlagInput struct {
	Width        int
	Height       int
	ProjectId    string
	FlagId       string
	LastTabIndex int
}

func NewMakeFlagInput(width int, height int, projectId string, flagId string, lastTabIndex int) *MakeFlagInput {
	return &MakeFlagInput{
		Width:        width,
		Height:       height,
		ProjectId:    projectId,
		FlagId:       flagId,
		LastTabIndex: lastTabIndex,
	}
}

func (in *MakeFlagInput) isSwitchModeInput() {}

// Authenticator------------------------------------------------------------

// -------------------------------------------------- List Authenticator Input

type ListAuthenticatorInput struct {
	Width        int
	Height       int
	ProjectId    string
	LastTabIndex int
}

func NewListAuthenticatorInput(width int, height int, projectId string, lastTabIndex int) *ListAuthenticatorInput {
	return &ListAuthenticatorInput{
		Width:        width,
		Height:       height,
		ProjectId:    projectId,
		LastTabIndex: lastTabIndex,
	}
}

func (in *ListAuthenticatorInput) isSwitchModeInput() {}

// -------------------------------------------------- Make Authenticator Input

type MakeAuthenticatorInput struct {
	Width           int
	Height          int
	ProjectId       string
	AuthenticatorId string
	LastTabIndex    int
}

func NewMakeAuthenticatorInput(width int, height int, projectId string, authenticatorId string, lastTabIndex int) *MakeAuthenticatorInput {
	return &MakeAuthenticatorInput{
		Width:           width,
		Height:          height,
		ProjectId:       projectId,
		AuthenticatorId: authenticatorId,
		LastTabIndex:    lastTabIndex,
	}
}

func (in *MakeAuthenticatorInput) isSwitchModeInput() {}

// -------------------------------------------------- Copy Authenticator Input

type CopyAuthenticatorInput struct {
	Width           int
	Height          int
	ProjectId       string
	LastTabIndex    int
	AuthenticatorId string
}

func NewCopyAuthenticatorInput(width int, height int, projectId string, lastTabIndex int, authenticatorId string) *CopyAuthenticatorInput {
	return &CopyAuthenticatorInput{
		Width:           width,
		Height:          height,
		ProjectId:       projectId,
		LastTabIndex:    lastTabIndex,
		AuthenticatorId: authenticatorId,
	}
}

func (in *CopyAuthenticatorInput) isSwitchModeInput() {}

// Credential---------------------------------------------------------------

// -------------------------------------------------- List Credential Input

type ListCredentialInput struct {
	Width           int
	Height          int
	ProjectId       string
	LastTabIndex    int
	AuthenticatorId string
}

func NewListCredentialInput(width int, height int, projectId string, lastTabIndex int, authenticatorId string) *ListCredentialInput {
	return &ListCredentialInput{
		Width:           width,
		Height:          height,
		ProjectId:       projectId,
		LastTabIndex:    lastTabIndex,
		AuthenticatorId: authenticatorId,
	}
}

func (in *ListCredentialInput) isSwitchModeInput() {}

// -------------------------------------------------- Make Credential Input

type MakeCredentialInput struct {
	Width           int
	Height          int
	ProjectId       string
	CredentialId    string
	LastTabIndex    int
	AuthenticatorId string
}

func NewMakeCredentialInput(width int, height int, projectId string, credentialId string, lastTabIndex int, authenticatorId string) *MakeCredentialInput {
	return &MakeCredentialInput{
		Width:           width,
		Height:          height,
		ProjectId:       projectId,
		CredentialId:    credentialId,
		LastTabIndex:    lastTabIndex,
		AuthenticatorId: authenticatorId,
	}
}

func (in *MakeCredentialInput) isSwitchModeInput() {}

// -------------------------------------------------- Copy Credential Input

type CopyCredentialInput struct {
	Width           int
	Height          int
	ProjectId       string
	LastTabIndex    int
	AuthenticatorId string
	CredentialId    string
}

func NewCopyCredentialInput(width int, height int, projectId string, lastTabIndex int, authenticatorId string, credentialId string) *CopyCredentialInput {
	return &CopyCredentialInput{
		Width:           width,
		Height:          height,
		ProjectId:       projectId,
		LastTabIndex:    lastTabIndex,
		AuthenticatorId: authenticatorId,
		CredentialId:    credentialId,
	}
}

func (in *CopyCredentialInput) isSwitchModeInput() {}

// Note --------------------------------------------------------------------

// --------------------------------------------------------- List Note Input

type ListNoteInput struct {
	Width        int
	Height       int
	ProjectId    string
	LastTabIndex int
}

func NewListNoteInput(width int, height int, projectId string, lastTabIndex int) *ListNoteInput {
	return &ListNoteInput{
		Width:        width,
		Height:       height,
		ProjectId:    projectId,
		LastTabIndex: lastTabIndex,
	}
}

func (in *ListNoteInput) isSwitchModeInput() {}

// --------------------------------------------------------- Make Note Input

type MakeNoteInput struct {
	Width        int
	Height       int
	ProjectId    string
	NoteId       string
	LastTabIndex int
}

func NewMakeNoteInput(width int, height int, projectId string, noteId string, lastTabIndex int) *MakeNoteInput {
	return &MakeNoteInput{
		Width:        width,
		Height:       height,
		ProjectId:    projectId,
		NoteId:       noteId,
		LastTabIndex: lastTabIndex,
	}
}

func (in *MakeNoteInput) isSwitchModeInput() {}

// --------------------------------------------------------- View Note Input

type ViewNoteInput struct {
	Width        int
	Height       int
	ProjectId    string
	NoteId       string
	LastTabIndex int
}

func NewViewNoteInput(width int, height int, projectId string, noteId string, lastTabIndex int) *ViewNoteInput {
	return &ViewNoteInput{
		Width:        width,
		Height:       height,
		ProjectId:    projectId,
		NoteId:       noteId,
		LastTabIndex: lastTabIndex,
	}
}

func (in *ViewNoteInput) isSwitchModeInput() {}
