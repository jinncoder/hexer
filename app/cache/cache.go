package cache

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/ssh"
)

var userToSSHPublicKey sync.Map

func AddUserKeyToCache(hashedUsername string, userSSHPublicKey string) {
	parsed, _, _, _, _ := ssh.ParseAuthorizedKey(
		[]byte(userSSHPublicKey),
	)
	userToSSHPublicKey.Store(hashedUsername, parsed)
}

func GetUserPublicSSHKeyFromCache(hashedUsername string) (ssh.PublicKey, error) {
	v, ok := userToSSHPublicKey.Load(hashedUsername)

	if !ok {
		return nil, fmt.Errorf("failed to load SSH Public Key for %s", hashedUsername)
	}

	return v.(ssh.PublicKey), nil
}

func RemoveUserPublicSSHKeyFromCache(hashedUsername string) {
	userToSSHPublicKey.Delete(hashedUsername)
}
