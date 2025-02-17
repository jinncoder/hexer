package database

import (
	"github.com/archimoebius/hexer/util"
	"github.com/google/uuid"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/tools/types"
)

type User struct {
	Id           string         `db:"id" json:"id"`
	EMail        string         `db:"email" json:"email"`
	Name         string         `db:"name" json:"name"`
	Verified     bool           `db:"verified" json:"verified"`
	Password     string         `db:"password" json:"password"`
	SSHPublicKey string         `db:"ssh_public_key" json:"ssh_public_key"`
	Created      types.DateTime `db:"created" json:"created"`
	Updated      types.DateTime `db:"updated" json:"updated"`
	TokenKey     string         `db:"tokenKey" json:"tokenKey"`
}

func GetUsers() ([]User, error) {
	users := []User{}

	db, err := GetDatabaseInstance()

	if err != nil {
		return users, err
	}

	err = db.
		Select("*").
		From("users").
		All(&users)

	if err != nil {
		return users, err
	}

	return users, nil
}

func GetNewUsers() ([]User, error) {
	user := []User{}

	db, err := GetDatabaseInstance()

	if err != nil {
		return user, err
	}

	err = db.
		Select("*").
		From("users").
		Where(dbx.HashExp{
			"verified": false,
		}).
		All(&user)

	if err != nil {
		return user, err
	}

	return user, nil
}

func GetUsersByName(username string) ([]User, error) {
	user := []User{}

	db, err := GetDatabaseInstance()

	if err != nil {
		return user, err
	}

	err = db.
		Select("*").
		From("users").
		Where(dbx.HashExp{
			"name": util.SHA256SUM(username),
		}).
		All(&user)

	if err != nil {
		return user, err
	}

	return user, nil
}

func GetUserById(id string) (User, error) {
	user := User{}

	db, err := GetDatabaseInstance()

	if err != nil {
		return user, err
	}

	err = db.
		Select("*").
		From("users").
		Where(dbx.HashExp{
			"id": id,
		}).
		One(&user)

	if err != nil {
		return user, err
	}

	return user, nil
}
func VerifyUser(id string) error {
	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Update(
		"users",
		dbx.Params{
			"verified": true,
		},
		dbx.NewExp(
			"id = {:id}",
			dbx.Params{"id": id},
		),
	).Execute()

	if err != nil {
		return err
	}

	return nil
}

func AddUser(
	email string,
	username string,
	password string,
	ssh_public_key string,

) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	uuid4, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	_, err = db.Insert("users", dbx.Params{
		"name":           util.SHA256SUM(username),
		"password":       password,
		"email":          email,
		"ssh_public_key": ssh_public_key,
		"created":        types.NowDateTime(),
		"updated":        types.NowDateTime(),
		"TokenKey":       uuid4.String(),
	}).Execute()

	if err != nil {
		return err
	}

	// TODO: LastInsertId does not work - as id isn't auto increment... see model_query.go in dbx ~ line 90
	// newUserId, err := result.LastInsertId()
	// if err != nil {
	// 	return user, err
	// }

	// err = db.
	// 	Select("users.*").
	// 	From("users").
	// 	Where(dbx.HashExp{
	// 		"id": fmt.Sprintf("%d", newUserId),
	// 	}).
	// 	One(&user)

	// if err != nil {
	// 	return user, err
	// }

	return nil
}

func DoesUserValueExist(column string, value string) error {
	user := User{}

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	if column == "name" {
		value = util.SHA256SUM(value)
	}

	err = db.
		Select("*").
		From("users").
		Where(dbx.HashExp{
			column: value,
		}).
		One(&user)

	if err != nil {
		return err
	}

	return nil
}

func IsUserVerified(username string) (bool, error) {
	user := User{}

	db, err := GetDatabaseInstance()

	if err != nil {
		return false, err
	}

	err = db.
		Select("*").
		From("users").
		Where(dbx.HashExp{
			"name": util.SHA256SUM(username),
		}).
		One(&user)

	if err != nil {
		return false, err
	}

	return user.Verified, nil
}

func RemoveUser(id string) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.
		Delete("users", dbx.NewExp("id = {:id}", dbx.Params{"id": id})).
		Execute()

	if err != nil {
		return err
	}

	return nil
}
