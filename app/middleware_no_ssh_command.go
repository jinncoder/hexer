package app

import (
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
)

func noSSHCommandMiddleware() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			if len(s.Command()) != 0 {
				wish.Fatalln(s, "No commands are allowed - lulz - you hacker you")
				return
			}
			next(s)
		}
	}
}
