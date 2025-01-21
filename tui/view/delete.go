package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"
)

func ConfirmDelete(session *ssh.Session, title string) bool {
	var confirm bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Delete " + title).
				Value(&confirm)),
	)
	// .WithTimeout(time.Duration(5.0 * float64(time.Second)))

	if session != nil {
		form.WithProgramOptions(append(bubbletea.MakeOptions(*session), tea.WithAltScreen())...)
	} else {
		form.WithProgramOptions(tea.WithAltScreen())
	}

	form.Init()       // Setup form/groups
	err := form.Run() // Show it!

	if err != nil {
		return false // TODO: "good enough" default?
	}

	return confirm
}
