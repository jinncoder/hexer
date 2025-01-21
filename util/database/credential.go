package database

import (
	"fmt"
	"strings"

	"github.com/archimoebius/hexer/tui/constant"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/tools/types"
)

type Credential struct {
	Id              string         `db:"id" json:"id"`
	UserName        string         `db:"username" json:"username"`
	Value           string         `db:"value" json:"value"`
	Type            string         `db:"type" json:"type"`
	Format          string         `db:"format" json:"format"`
	Encoding        string         `db:"encoding" json:"encoding"`
	Note            string         `db:"note" json:"note"`
	AuthenticatorId string         `db:"link_authenticator" json:"link_authenticator"`
	CleartextId     string         `db:"cleartext_id" json:"cleartext_id"`
	Created         types.DateTime `db:"created" json:"created"`
	Updated         types.DateTime `db:"updated" json:"updated"`
}

func (credential Credential) Description() string {
	var note = ""

	if len(credential.Note) > 0 {
		note = fmt.Sprintf("%s - ", credential.Note)
	}

	return fmt.Sprintf("%s%s", note, credential.Value)
}

func (credential Credential) ListTitle() string {
	authenticator, err := GetLinkedAuthenticatorRootByAuthenticatorId(credential.AuthenticatorId)
	if err != nil {
		return "UnknownAuthenticator"
	}
	var extra = ""

	if len(authenticator.FQDN) > 0 {
		extra = fmt.Sprintf("@%s", strings.ToUpper(authenticator.FQDN))
	}

	return fmt.Sprintf("%s|%s %s%s", authenticator.Type, credential.Type, credential.UserName, extra)
}

func GetAuthenticatorCredentialList(authenticatorId string, onlyCleartext bool) ([]Credential, error) {

	credential_list := []Credential{}

	if len(authenticatorId) <= 0 {
		return credential_list, fmt.Errorf("authenticatorId required")
	}

	db, err := GetDatabaseInstance()
	if err != nil {
		return credential_list, err
	}

	authenticator, err := GetLinkedAuthenticatorRootByAuthenticatorId(authenticatorId)
	if err != nil {
		return credential_list, err
	}

	query := db.
		Select("*").
		From("credential")

	where := query.Where(dbx.HashExp{"link_authenticator": authenticator.Id, "hidden": false})

	if onlyCleartext {
		where = query.Where(dbx.HashExp{
			"link_authenticator": authenticator.Id,
			"type":               strings.ToLower(constant.CredentialTypeCleartext.String()),
		})
	}

	err =
		where.
			OrderBy("created ASC", "updated DESC").
			AndOrderBy("username ASC").
			All(&credential_list)

	if err != nil {
		return credential_list, err
	}

	return credential_list, nil
}

func GetProjectCredentialList(projectId string, onlyCleartext bool) ([]Credential, error) {
	credential_list := []Credential{}
	seen_credential_map := make(map[string]bool)

	if len(projectId) <= 0 {
		return credential_list, fmt.Errorf("projectId required")
	}

	authenticator_list, err := GetAuthenticatorList(projectId)

	if err != nil {
		return credential_list, err
	}

	for _, authenticator := range authenticator_list {
		authenticator_credential_list, err := GetAuthenticatorCredentialList(authenticator.Id, onlyCleartext)

		if err != nil {
			return credential_list, err
		}

		for _, credential := range authenticator_credential_list {
			if !seen_credential_map[credential.Id] {
				credential_list = append(credential_list, credential)
			}
			seen_credential_map[credential.Id] = true
		}
	}

	return credential_list, nil
}

func UpdateCredential(
	credentialId string,
	authenticatorId string,
	userName string,
	value string,
	authenticatorType string,
	authenticatorFormat string,
	authenticatorEncoding string,
	note string,
	cleartextId string,
) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Update("credential", dbx.Params{
		"username":           userName,
		"value":              value,
		"type":               authenticatorType,
		"format":             authenticatorFormat,
		"encoding":           authenticatorEncoding,
		"note":               note,
		"link_authenticator": authenticatorId,
		"cleartext_id":       cleartextId,
		"updated":            types.NowDateTime(),
	},
		dbx.NewExp("id = {:id}", dbx.Params{"id": credentialId})).Execute()

	if err != nil {
		return err
	}

	return nil
}

func AddCredential(
	authenticatorId string,
	userName string,
	value string,
	authenticatorType string,
	authenticatorFormat string,
	authenticatorEncoding string,
	note string,
	cleartextId string,
) error {

	authenticator, err := GetLinkedAuthenticatorRootByAuthenticatorId(authenticatorId)
	if err != nil {
		return err
	}

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Insert("credential", dbx.Params{
		"username":           userName,
		"value":              value,
		"type":               authenticatorType,
		"format":             authenticatorFormat,
		"encoding":           authenticatorEncoding,
		"note":               note,
		"link_authenticator": authenticator.Id,
		"cleartext_id":       cleartextId,
		"created":            types.NowDateTime(),
		"updated":            types.NowDateTime(),
	}).Execute()

	if err != nil {
		return err
	}

	return nil
}

func RemoveCredential(id string) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Update("credential", dbx.Params{
		"hidden": true,
	},
		dbx.NewExp("id = {:id}", dbx.Params{"id": id})).Execute()

	if err != nil {
		return err
	}

	return nil
}

func GetCredentialById(id string) (Authenticator, Credential, error) {
	credential := Credential{}
	authenticator := Authenticator{}

	db, err := GetDatabaseInstance()

	if err != nil {
		return authenticator, credential, err
	}

	err = db.
		Select("*").
		From("credential").
		Where(dbx.HashExp{
			"id": id,
		}).
		One(&credential)

	if err != nil {
		return authenticator, credential, err
	}

	authenticator, err = GetLinkedAuthenticatorRootByAuthenticatorId(credential.AuthenticatorId)
	if err != nil {
		return authenticator, credential, err
	}

	return authenticator, credential, nil
}

func GetCredentialCopyList(id string) ([]string, error) {
	var credentialForms []string
	authenticator, credential, err := GetCredentialById(id)

	credentialForms = append(credentialForms, credential.Value)
	credentialForms = append(credentialForms, CredentialFormImpacket(authenticator, credential)...)
	credentialForms = append(credentialForms, CredentialFormNetExec(authenticator, credential)...)
	credentialForms = append(credentialForms, CredentialFormEvilWinRM(authenticator, credential)...)

	if err != nil {
		return credentialForms, err
	}

	return credentialForms, nil
}
