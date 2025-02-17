package constant

// --------------------------------------------------------- Make User Input

type MakeUserInput struct {
	Width  int
	Height int
}

func NewMakeUserInput(width int, height int) *MakeUserInput {
	return &MakeUserInput{
		Width:  width,
		Height: height,
	}
}

func (in *MakeUserInput) isSwitchModeInput() {}

// --------------------------------------------------------- Make Administrate User Input

type MakeAdministrateUserInput struct {
	Width  int
	Height int
}

func NewAdministrateUserInput(width int, height int) *MakeAdministrateUserInput {
	return &MakeAdministrateUserInput{
		Width:  width,
		Height: height,
	}
}

func (in *MakeAdministrateUserInput) isSwitchModeInput() {}
