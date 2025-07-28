package app

import (
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
)

func onlyVerifiedUsersMiddleware() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {

			if !validateSessionUser(s) {
				wish.Fatalln(s, "\033cYou're account is not verified - please contact your administrator")
				return
			}

			next(s)
		}
	}
}
