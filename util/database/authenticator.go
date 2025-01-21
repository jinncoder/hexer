package database

import (
	"fmt"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/tools/types"
)

type Authenticator struct {
	Id              string         `db:"id" json:"id"`
	Name            string         `db:"name" json:"name"`
	Type            string         `db:"type" json:"type"`
	IPv4            string         `db:"ipv4" json:"ipv4"`
	FQDN            string         `db:"fqdn" json:"fqdn"`
	Note            string         `db:"note" json:"note"`
	ProjectId       string         `db:"link_project" json:"link_project"`
	AuthenticatorId string         `db:"link_authenticator" json:"link_authenticator"`
	Created         types.DateTime `db:"created" json:"created"`
	Updated         types.DateTime `db:"updated" json:"updated"`
}

func (authenticator Authenticator) DomainName() string {
	var domain = ""

	if len(authenticator.FQDN) > 0 {
		domain = strings.ToUpper(strings.Replace(
			strings.ToLower(authenticator.FQDN),
			strings.ToLower(authenticator.Name)+".",
			"",
			1,
		))
	}

	return domain
}

func (authenticator Authenticator) ListTitle() string {
	var extra = ""

	if len(authenticator.FQDN) > 0 {
		extra = fmt.Sprintf(" - %s", strings.ToUpper(authenticator.FQDN))
	}

	return fmt.Sprintf("[%s] %s%s | %s", authenticator.Type, authenticator.IPv4, extra, authenticator.Name)
}

func GetAuthenticatorById(id string) (Authenticator, error) {
	authenticator := Authenticator{}

	db, err := GetDatabaseInstance()

	if err != nil {
		return authenticator, err
	}

	err = db.
		Select("*").
		From("authenticator").
		Where(dbx.HashExp{
			"id": id,
		}).
		One(&authenticator)

	if err != nil {
		return authenticator, err
	}

	return authenticator, nil
}

func GetAuthenticatorList(projectId string) ([]Authenticator, error) {

	authenticator_list := []Authenticator{}

	if len(projectId) <= 0 {
		return authenticator_list, fmt.Errorf("projectId required")
	}

	db, err := GetDatabaseInstance()

	if err != nil {
		return authenticator_list, err
	}

	err = db.
		Select("*").
		From("authenticator").
		Where(dbx.HashExp{"link_project": projectId, "hidden": false}).
		OrderBy("type ASC", "created ASC", "updated DESC").
		AndOrderBy("name ASC").
		All(&authenticator_list)

	if err != nil {
		return authenticator_list, err
	}

	return authenticator_list, nil
}

func GetLinkedAuthenticatorRootByAuthenticatorId(authenticator_id string) (Authenticator, error) {

	authenticator := Authenticator{}

	if len(authenticator_id) <= 0 {
		return authenticator, nil
	}

	db, err := GetDatabaseInstance()

	if err != nil {
		return authenticator, err
	}

	err = db.
		Select("*").
		From("authenticator").
		Where(dbx.HashExp{"id": authenticator_id}).
		One(&authenticator)

	if err != nil {
		return authenticator, err
	}

	if authenticator.AuthenticatorId != "" && authenticator.Id != authenticator.AuthenticatorId {
		return GetLinkedAuthenticatorRootByAuthenticatorId(authenticator.AuthenticatorId)
	} else {
		return authenticator, nil
	}
}

func UpdateAuthenticator(
	projectId string,
	updateAuthenticatorId string,
	name string,
	authenticatorType string,
	ipv4 string,
	fqdn string,
	note string,
	authenticatorId string,
) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Update("authenticator", dbx.Params{
		"name":               name,
		"type":               strings.ToLower(authenticatorType),
		"ipv4":               ipv4,
		"fqdn":               fqdn,
		"note":               note,
		"link_project":       projectId,
		"link_authenticator": authenticatorId,
		"updated":            types.NowDateTime(),
	},
		dbx.NewExp("id = {:id}", dbx.Params{"id": updateAuthenticatorId})).Execute()

	if err != nil {
		return err
	}

	return nil
}

func AddAuthenticator(
	projectId string,
	name string,
	authenticatorType string,
	ipv4 string,
	fqdn string,
	note string,
	authenticatorId string,
) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Insert("authenticator", dbx.Params{
		"name":               name,
		"type":               strings.ToLower(authenticatorType),
		"ipv4":               ipv4,
		"fqdn":               fqdn,
		"note":               note,
		"link_project":       projectId,
		"link_authenticator": authenticatorId,
		"created":            types.NowDateTime(),
		"updated":            types.NowDateTime(),
	}).Execute()

	if err != nil {
		return err
	}

	return nil
}

func RemoveAuthenticator(id string) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Update("authenticator", dbx.Params{
		"hidden": true,
	},
		dbx.NewExp("id = {:id}", dbx.Params{"id": id})).Execute()

	if err != nil {
		return err
	}

	return nil
}

func GetAuthenticatorCopyList(id string) ([]string, error) {
	var authenticatorForms []string
	authenticator, err := GetAuthenticatorById(id)
	if err != nil {
		return authenticatorForms, err
	}

	authenticatorForms = append(authenticatorForms, authenticator.IPv4)

	if authenticator.FQDN != "" {
		authenticatorForms = append(authenticatorForms, authenticator.FQDN)
	}

	authenticatorForms = append(authenticatorForms, authenticator.Name)

	if err != nil {
		return authenticatorForms, err
	}

	return authenticatorForms, nil
}
