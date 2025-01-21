package app

import (
	"github.com/archimoebius/hexer/util"
	"github.com/archimoebius/hexer/util/database"
	"github.com/charmbracelet/ssh"
)

func validateSessionUser(s ssh.Session) bool {
	usernamePlain, _, _ := util.GetUsernameProjectIfPresent(s.User())

	if usernamePlain == "register" {
		return true
	}

	verified, err := database.IsUserVerified(usernamePlain)

	if err != nil {
		verified = false
		util.Logger.Error(err)
	}

	return verified
}
