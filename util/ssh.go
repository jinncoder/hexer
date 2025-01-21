package util

import (
	"strings"
)

type contextKey struct {
	name string
}

var ContextKeyProjectId = &contextKey{"project-id"}
var ContextKeyFilepath = &contextKey{"note-filepath"}

func GetUsernameProjectIfPresent(entry string) (string, string, string) {
	var projectId = ""
	parts := strings.Split(entry, "-")

	username := SHA256SUM(parts[0])

	if len(parts) == 2 {
		projectId = parts[1]
	}

	return parts[0], username, projectId
}

func FixSSHKeyData(sshkey string) []byte {
	return []byte(strings.ReplaceAll(strings.ReplaceAll(sshkey, "-----END", "\n-----END"), "KEY-----", "KEY-----\n"))
}
