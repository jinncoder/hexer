package database

import (
	"fmt"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/tools/types"
)

type Project struct {
	Id      string         `db:"id" json:"id"`
	Name    string         `db:"name" json:"name"`
	Note    string         `db:"note" json:"note"`
	Created types.DateTime `db:"created" json:"created"`
	Updated types.DateTime `db:"updated" json:"updated"`
}

func (project Project) ListTitle() string {
	return fmt.Sprintf("%s (%s)", project.Name, project.Id)
}

func GetProjectById(id string) (Project, error) {
	project := Project{}

	db, err := GetDatabaseInstance()

	if err != nil {
		return project, err
	}

	err = db.
		Select("*").
		From("project").
		Where(dbx.HashExp{
			"id": id,
		}).
		One(&project)

	if err != nil {
		return project, err
	}

	return project, nil
}

func GetProjectList() ([]Project, error) {

	project_list := []Project{}

	db, err := GetDatabaseInstance()

	if err != nil {
		return project_list, err
	}

	err = db.
		Select("project.*").
		From("project").
		Where(dbx.HashExp{"hidden": false}).
		OrderBy("created ASC", "updated DESC").
		AndOrderBy("name ASC").
		All(&project_list)

	if err != nil {
		return project_list, err
	}

	return project_list, nil
}

func UpdateProject(
	projectId string,
	name string,
	note string,
) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Update("project", dbx.Params{
		"name":    name,
		"note":    note,
		"updated": types.NowDateTime(),
	},
		dbx.NewExp("id = {:id}", dbx.Params{"id": projectId})).Execute()

	if err != nil {
		return err
	}

	return nil
}

func AddProject(name string, note string) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Insert("project", dbx.Params{
		"name":    name,
		"note":    note,
		"created": types.NowDateTime(),
		"updated": types.NowDateTime(),
	}).Execute()

	if err != nil {
		return err
	}

	return nil
}

func GetProjectTitle(id string) (string, error) {
	project := Project{}

	db, err := GetDatabaseInstance()

	if err != nil {
		return "", err
	}

	err = db.
		Select("project.*").
		From("project").
		Where(dbx.NewExp("id = {:id}", dbx.Params{"id": id})).
		One(&project)

	if err != nil {
		return "", err
	}

	return project.Name, nil
}

func RemoveProject(id string) error {

	db, err := GetDatabaseInstance()

	if err != nil {
		return err
	}

	_, err = db.Update("project", dbx.Params{
		"hidden": true,
	},
		dbx.NewExp("id = {:id}", dbx.Params{"id": id})).Execute()

	if err != nil {
		return err
	}

	return nil
}
