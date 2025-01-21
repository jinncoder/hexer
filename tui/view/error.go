package view

import (
	tea "github.com/charmbracelet/bubbletea"
)

type ModelFatalError struct {
	error error
}

func NewModelFatalError(error error) (*ModelFatalError, error) {

	m := &ModelFatalError{
		error: error,
	}

	return m, nil

}

func (m ModelFatalError) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlD:
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

func (m ModelFatalError) Init() tea.Cmd {
	return nil
}

func (m ModelFatalError) View() string {
	return m.error.Error()
}
