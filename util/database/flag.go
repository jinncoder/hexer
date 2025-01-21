package database

import (
	"fmt"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/tools/types"
)

type Flag struct {
	Id              string         `db:"id" json:"id"`
	Name            string         `db:"name" json:"name"`
	Value           string         `db:"value" json:"value"`
	Note            string         `db:"note" json:"note"`
	AuthenticatorId string         `db:"link_authenticator" json:"link_authenticator"`
	Created         types.DateTime `db:"created" json:"created"`
	Updated         types.DateTime `db:"updated" json:"updated"`
}

func (flag Flag) ListDescription() string {
	if len(flag.Note) > 0 {
		return fmt.Sprintf("%s (%s)", flag.Value, flag.Note)
	}

	return flag.Value
}

func (flag Flag) ListTitle() string {
	authenticator, err := GetAuthenticatorById(flag.AuthenticatorId)

	if err != nil {
		return flag.Name
	} else {
		auth_extra := strings.ToUpper(authenticator.FQDN)

		if len(auth_extra) <= 0 {
			auth_extra = authenticator.Name
		}

		return fmt.Sprintf("[%s] %s", auth_extra, flag.Name)
	}
}

func GetFlagById(id string) (Flag, error) {
	flag := Flag{}

	db, err := GetDatabaseInstance()

	if err != nil {
		return flag, err
	}

	err = db.
		Select("*").
		From("flag").
		Where(dbx.HashExp{
			"id": id,
		}).
		One(&flag)

	if err != nil {
		return flag, err
	}

	return flag, nil
}

func GetAuthenticatorFlagList(authenticatorId string) ([]Flag, error) {

	flag_list := []Flag{}

	if len(authenticatorId) <= 0 {
		return flag_list, fmt.Errorf("authenticatorId required")
	}

	db, err := GetDatabaseInstance()

	if err != nil {
		return flag_list, err
	}

	err = db.
		Select("*").
		From("flag").
		Where(dbx.HashExp{"link_authenticator": authenticatorId, "hidden": false}).
		OrderBy("created ASC", "updated DESC").
		AndOrderBy("name ASC").
		All(&flag_list)

	if err != nil {
		return flag_list, err
	}

	return flag_list, nil
}

func GetProjectFlagList(projectId string) ([]Flag, error) {

	flag_list := []Flag{}

	if len(projectId) <= 0 {
		return flag_list, fmt.Errorf("projectId required")
	}

	authenticator_list, err := GetAuthenticatorList(projectId)

	if err != nil {
		return flag_list, err
	}

	for _, authenticator := range authenticator_list {

		authenticator_flag_list, err := GetAuthenticatorFlagList(authenticator.Id)

		if err != nil {
			return flag_list, err
		}

		flag_list = append(flag_list, authenticator_flag_list[:]...)
	}

	return flag_list, nil
}

func UpdateFlag(
	flagId string,
	authenticatorId string,
	name string,
	value string,
	note string,
) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Update("flag", dbx.Params{
		"name":               name,
		"value":              value,
		"note":               note,
		"link_authenticator": authenticatorId,
		"updated":            types.NowDateTime(),
	},
		dbx.NewExp("id = {:id}", dbx.Params{"id": flagId})).Execute()

	if err != nil {
		return err
	}

	return nil
}

func AddFlag(
	authenticatorId string,
	name string,
	value string,
	note string,
) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Insert("flag", dbx.Params{
		"name":               name,
		"value":              value,
		"note":               note,
		"link_authenticator": authenticatorId,
		"created":            types.NowDateTime(),
		"updated":            types.NowDateTime(),
	}).Execute()

	if err != nil {
		return err
	}

	return nil
}

func RemoveFlag(id string) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Update("flag", dbx.Params{
		"hidden": true,
	},
		dbx.NewExp("id = {:id}", dbx.Params{"id": id})).Execute()

	if err != nil {
		return err
	}

	return nil
}
