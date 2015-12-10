package models

import (
	"errors"
	"fmt"
	. "github.com/vivowares/octopus/utils"
	"strings"
)

var SupportedDataTypes = []string{"float", "int", "boolean", "string"}

type Channel struct {
	Id           int         `sql:"type:integer" json:"id"`
	Name         string      `sql:"type:varchar(255)" json:"name"`
	Description  string      `sql:"type:text" json:"description"`
	Tags         StringSlice `sql:"type:text" json:"tags"`
	Fields       StringMap   `sql:"type:text" json:"fields"`
	AccessTokens StringSlice `sql:"type:text" json:"access_tokens"`
}

func (c *Channel) BeforeSave() error {
	if len(c.Name) == 0 {
		return errors.New("name is empty")
	}

	if len(c.Description) == 0 {
		return errors.New("description is empty")
	}

	if len(c.AccessTokens) == 0 {
		return errors.New("access_tokens are empty")
	}

	if len(c.Tags) > 64 {
		return errors.New("too many tags, at most 64 tags are supported")
	}

	tagMap := make(map[string]bool, 0)

	for _, tagName := range c.Tags {
		if !AlphaNumeric(tagName) {
			return errors.New("invalid tag name, only letters, numbers and underscores are allowed")
		}

		if _, found := tagMap[tagName]; found {
			return errors.New(fmt.Sprintf("duplicate tag name: %s", tagName))
		} else {
			tagMap[tagName] = true

			if _, found = c.Fields[tagName]; found {
				return errors.New(fmt.Sprintf("conflicting tag name: %s defined in fields", tagName))
			}
		}
	}

	if len(c.Fields) == 0 {
		return errors.New("fields are empty")
	}

	if len(c.Fields) > 128 {
		return errors.New("too many fields, at most 128 fields are supported")
	}

	for k, v := range c.Fields {
		if !AlphaNumeric(k) {
			return errors.New("invalid field name, only letters, numbers and underscores are allowed")
		}

		if !StringSliceContains(SupportedDataTypes, v) {
			return errors.New(fmt.Sprintf("unsupported datatype on %s: %s, supported datatypes are %s", k, v, strings.Join(SupportedDataTypes, ",")))
		}
	}

	return nil
}

func (c *Channel) Create() error {
	return DB.Create(c).Error
}

func (c *Channel) Delete() error {
	return DB.Delete(c).Error
}

func (c *Channel) Update() error {
	return DB.Save(c).Error
}

func (c *Channel) FindById(id int) bool {
	DB.First(c, id)
	return DB.NewRecord(c)
}
