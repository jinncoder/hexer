package app

import (
	"fmt"
	"os"

	serveConfig "github.com/archimoebius/hexer/cli/config/serve"
	"github.com/archimoebius/hexer/tui"
	"github.com/archimoebius/hexer/tui/constant"
	"github.com/archimoebius/hexer/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"
)

func hexerMiddleware() wish.Middleware {
	newProgram := func(m *tui.Model, opts ...tea.ProgramOption) *tea.Program {
		return tea.NewProgram(m, opts...)
	}

	teaHandler := func(s ssh.Session) *tea.Program {
		projectId := s.Context().Value(util.ContextKeyProjectId)

		pty, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil
		}

		// default view => list of all projects
		m, err := tui.NewModel(
			tui.NewInput(
				&s,
				constant.ModeListProject,
				constant.NewListProjectInput(pty.Window.Width, pty.Window.Height),
			),
		)

		if err != nil {
			util.Logger.Error(err)
			s.Stderr().Write([]byte("\033cUnable to initialize the project listing - contact your adminstrator")) // #nosec G104
			err = s.Exit(1)
			if err != nil {
				util.Logger.Error(err)
				os.Exit(1)
			}
		}

		if projectId != nil {
			m, err = tui.NewModel(
				tui.NewInput(
					&s,
					constant.ModePickProject,
					constant.NewPickProjectInput(pty.Window.Width, pty.Window.Height, fmt.Sprintf("%v", projectId), 0),
				),
			)

			if err != nil {
				util.Logger.Error(err)
				s.Stderr().Write([]byte("\033cUnable to initialize the project page - contact your adminstrator")) // #nosec G104
				err = s.Exit(1)
				if err != nil {
					util.Logger.Error(err)
					os.Exit(1)
				}
			}
		}

		if s.User() == "register" {
			m, err = tui.NewModel(
				tui.NewInput(
					&s,
					constant.ModeMakeUser,
					constant.NewMakeUserInput(pty.Window.Width, pty.Window.Height),
				),
			)

			if err != nil {
				util.Logger.Error(err)
				s.Stderr().Write([]byte("\033cUnable to initialize the registration page - contact your adminstrator")) // #nosec G104
				err = s.Exit(1)
				if err != nil {
					util.Logger.Error(err)
					os.Exit(1)
				}
			}
		}

		if s.User() == "administrate" {
			if serveConfig.Setting.OpenRegistration {
				s.Stderr().Write([]byte("\033cadministrate user is disabled when open registartion is enabled"))
				err = s.Exit(1)
				return nil
			}

			m, err = tui.NewModel(
				tui.NewInput(
					&s,
					constant.ModeAdministrateUser,
					constant.NewAdministrateUserInput(pty.Window.Width, pty.Window.Height),
				),
			)

			if err != nil {
				util.Logger.Error(err)
				s.Stderr().Write([]byte("\033cUnable to initialize the administrate page - contact your adminstrator")) // #nosec G104
				err = s.Exit(1)
				if err != nil {
					util.Logger.Error(err)
					os.Exit(1)
				}
			}
		}

		return newProgram(m, append(bubbletea.MakeOptions(s), tea.WithAltScreen())...)
	}

	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}
