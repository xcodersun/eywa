package models

import (
	"errors"
	"fmt"
	. "github.com/vivowares/octopus/persistence"
	. "github.com/vivowares/octopus/utils"
	"strings"
)

var SupportedDataTypes = []string{"float64", "int64", "boolean", "string"}

// use channel to query
// channel metadata is stored in persistant db
// need some validations
type Channel struct {
	// we don't allow duplicate channel names
	Name        string
	Description string
	// we still wanna restrict the tags/fields creation. Since this is used to monitor hardwares. these settings shouldn't change much
	// however we do allow modifications on channel settings
	Tags      []string          // max to 255 tags, tag datatype is always string
	Fields    map[string]string // max to 255 fields, different fields may have different datatype
	datastore DataStore
}

func NewChannel() *Channel {
	return &Channel{
		Tags:      make([]string, 0),
		Fields:    make(map[string]string),
		datastore: DB,
	}
}

func (c *Channel) Validate() (map[string]error, bool) {
	ers := make(map[string]error)
	if len(c.Name) == 0 {
		ers["name"] = errors.New("empty channel name")
	}

	if len(c.Description) == 0 {
		ers["description"] = errors.New("empty channel description")
	}

	if len(c.Tags) > 255 {
		ers["tags"] = errors.New("too many tags, max 255 supported")
	}

	for _, tagName := range c.Tags {
		if v, found := c.Fields[tagName]; found {
			ers["tags"] = errors.New(fmt.Sprintf("conflicting tag name: %s defined in fields too", v))
		}
	}

	if len(c.Fields) > 255 || len(c.Fields) == 0 {
		ers["fields"] = errors.New(fmt.Sprintf("the number of fields must be between 1 ~ 255 instead of %d", len(c.Fields)))
	}

	for k, v := range c.Fields {
		if StringSliceContains(SupportedDataTypes, v) {
			ers["fields"] = errors.New(fmt.Sprintf("supported datatype on %s: %s, supported datatypes are %s", k, v, strings.Join(SupportedDataTypes, ",")))
		}
	}

	return ers, len(ers) == 0
}

func (c *Channel) Store() (map[string]error, bool) {
	if ers, valid := c.Validate(); !valid {
		return ers, false
	}

	err := c.datastore.Save()
	if err != nil {
		return map[string]error{"datastore": err}, false
	}

	return map[string]error{}, true
}
