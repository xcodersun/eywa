package models

import (
	"errors"
	"fmt"
	"github.com/vivowares/octopus/connections"
	. "github.com/vivowares/octopus/utils"
	"strings"
)

var SupportedDataTypes = []string{"float", "int", "boolean", "string"}

// we don't support nested data structures
var SupportedPointFormat = []string{"json", "url"}

type Channel struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Format      string            `json:"format"`
	Tags        []string          `json:"tags"`
	Fields      map[string]string `json:"fields"`
}

func (c *Channel) Validate() (map[string]error, bool) {
	ers := make(map[string]error)
	if len(c.Name) == 0 {
		ers["name"] = errors.New("empty channel name")
	}

	if len(c.Description) == 0 {
		ers["description"] = errors.New("empty channel description")
	}

	if len(c.Format) == 0 {
		ers["format"] = errors.New("empty channel format")
	}

	if !StringSliceContains(SupportedPointFormat, c.Format) {
		ers["format"] = errors.New(fmt.Sprintf("unsupported point format %s, supported formats are %s", c.Format, strings.Join(SupportedPointFormat, ",")))
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
		if !StringSliceContains(SupportedDataTypes, v) {
			ers["fields"] = errors.New(fmt.Sprintf("unsupported datatype on %s: %s, supported datatypes are %s", k, v, strings.Join(SupportedDataTypes, ",")))
		}
	}

	return ers, len(ers) == 0
}

func (c *Channel) Insert() (map[string]error, bool) {
	if ers, valid := c.Validate(); !valid {
		return ers, false
	}

	err := MStore.InsertChannel(c)
	if err != nil {
		return map[string]error{"datastore": err}, false
	}

	return map[string]error{}, true
}

func (c *Channel) Update() (map[string]error, bool) {
	if ers, valid := c.Validate(); !valid {
		return ers, false
	}

	err := MStore.UpdateChannel(c)
	if err != nil {
		return map[string]error{"datastore": err}, false
	}

	return map[string]error{}, true
}

func FindChannelByName(name string) (*Channel, bool) {
	return MStore.FindChannelByName(name)
}

func FindChannels() ([]*Channel, error) {
	return MStore.FindChannels()
}

func (c *Channel) Delete() error {
	return MStore.DeleteChannel(c)
}

func (c *Channel) NewPoint(conn connections.Connection, raw string) (*Point, error) {
	p := &Point{
		Raw:     raw,
		channel: c,
		conn:    conn,
	}

	err := p.parseRaw()
	if err != nil {
		return nil, err
	}
	return p, nil
}
