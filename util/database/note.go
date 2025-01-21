package database

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/tools/types"
)

type Note struct {
	Id        string         `db:"id" json:"id"`
	Name      string         `db:"name" json:"name"`
	Value     string         `db:"value" json:"value"`
	ProjectId string         `db:"link_project" json:"link_project"`
	Created   types.DateTime `db:"created" json:"created"`
	Updated   types.DateTime `db:"updated" json:"updated"`
}

func (note Note) ListDescription() string {
	i := len(note.Value)

	if i > 0 {

		if i > 20 {
			i = 20
		}

		return fmt.Sprintf("%s...", note.Value[0:i])
	}

	return note.Value
}

func (note Note) ListTitle() string {
	i := len(note.Name)

	if i > 0 {

		if i > 20 {
			i = 20
		}

		return fmt.Sprintf("%s...", note.Name[0:i])
	}

	return note.Name
}

func GetNoteById(id string) (Note, error) {
	note := Note{}

	db, err := GetDatabaseInstance()

	if err != nil {
		return note, err
	}

	err = db.
		Select("*").
		From("note").
		Where(dbx.HashExp{
			"id": id,
		}).
		One(&note)

	if err != nil {
		return note, err
	}

	return note, nil
}

func GetProjectNoteList(projectId string) ([]Note, error) {
	note_list := []Note{}

	if len(projectId) <= 0 {
		return note_list, fmt.Errorf("projectId required")
	}

	db, err := GetDatabaseInstance()

	if err != nil {
		return note_list, err
	}

	err = db.
		Select("*").
		From("note").
		Where(dbx.HashExp{"link_project": projectId, "hidden": false}).
		OrderBy("created ASC", "updated DESC").
		AndOrderBy("name ASC").
		All(&note_list)

	if err != nil {
		return note_list, err
	}

	return note_list, nil
}

func UpdateNote(
	noteId string,
	projectId string,
	name string,
	value string,
) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Update("note", dbx.Params{
		"name":         name,
		"value":        value,
		"updated":      types.NowDateTime(),
		"link_project": projectId,
	},
		dbx.NewExp("id = {:id}", dbx.Params{"id": noteId})).Execute()

	if err != nil {
		return err
	}

	return nil
}

func AddNote(
	projectId string,
	name string,
	value string,
) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Insert("note", dbx.Params{
		"name":         strings.Replace(name, "\n", "", -1),
		"value":        value,
		"created":      types.NowDateTime(),
		"updated":      types.NowDateTime(),
		"link_project": projectId,
	}).Execute()

	if err != nil {
		return err
	}

	return nil
}

func RemoveNote(id string) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Update("note", dbx.Params{
		"hidden": true,
	},
		dbx.NewExp("id = {:id}", dbx.Params{"id": id})).Execute()

	if err != nil {
		return err
	}

	return nil
}

func AddNoteFromFilepath(projectId string, filepath string) error {
	data, err := os.ReadFile(filepath) // #nosec G304 - s.root is used with a random UUID in app/sftp.go:125
	if err != nil {
		return err
	}

	lines := bytes.Split(data, []byte{'\n'})

	if len(lines) < 2 {
		return fmt.Errorf("bad file format")
	}

	title := lines[0]
	joinedData := bytes.Join(lines[1:], []byte{'\n'})

	err = AddNote(projectId, string(title), string(joinedData))
	if err != nil {
		return err
	}

	return nil
}

func GetNotesAsMarkdown(projectId string) (string, error) {

	notes, err := GetProjectNoteList(projectId)
	if err != nil {
		return "", err
	}

	var data = ""

	for _, note := range notes {
		data += fmt.Sprintf("\n# %s\n\n", note.Name)
		data += fmt.Sprintf("%s\n", note.Value)
	}

	return data, nil
}
