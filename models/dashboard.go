package models

import (
	"errors"
)

type Dashboard struct {
	Id          int    `sql:"type:integer" json:"id"`
	Name        string `sql:"type:varchar(255)" json:"name"`
	Description string `sql:"type:text" json:"description"`
	Definition  string `sql:"type:text" json:"definition"`
}

func (d *Dashboard) BeforeSave() error {
	if len(d.Name) == 0 {
		return errors.New("name is empty")
	}

	if len(d.Description) == 0 {
		return errors.New("description is empty")
	}

	return nil
}

func (d *Dashboard) Create() error {
	return DB.Create(d).Error
}

func (d *Dashboard) Delete() error {
	return DB.Delete(d).Error
}

func (d *Dashboard) Update() error {
	return DB.Save(d).Error
}

func (d *Dashboard) FindById(id int) bool {
	DB.First(d, id)
	return !DB.NewRecord(d)
}

func Dashboards() []*Dashboard {
	dashs := []*Dashboard{}
	DB.Find(&dashs)
	return dashs
}
