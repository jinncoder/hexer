package database

import (
	"fmt"
	"sync"

	"github.com/pocketbase/dbx"
)

var lock = &sync.Mutex{}
var instance *dbx.Builder

func GetDatabaseInstance() (dbx.Builder, error) {
	lock.Lock()
	defer lock.Unlock()

	if instance == nil {
		return nil, fmt.Errorf("no single instance found")
	}

	return *instance, nil
}

func SetDatabaseInstance(db *dbx.Builder) error {
	if instance == nil {
		lock.Lock()
		defer lock.Unlock()
		instance = db
	} else {
		return fmt.Errorf("instance already set")
	}

	return nil
}
