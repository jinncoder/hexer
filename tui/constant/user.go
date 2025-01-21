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
