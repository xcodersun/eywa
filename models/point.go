package models

import (
	// "errors"
	// "fmt"
	// . "github.com/vivowares/octopus/utils"
	// "net/url"
	// "strconv"
	// "strings"
	"time"
)

// we don't support nested data structures
var SupportedPointFormat = []string{"json", "url"}

// use point to write
// point data is stored in influxdb or elasticsearch
// we also only support unix time epoch as timestamp
type Point struct {
	Format    string   // goes to archive
	Raw       string   // goes to archive
	channel   *Channel // goes to archive with name
	Timestamp time.Time
	Tags      map[string]string
	Fields    map[string]interface{}
	// Validate() map[string]error //field/tag -> error
	// Store() error               // -> index db
}

// func (p *Point) ParseRaw() error {
// 	if p.channel == nil {
// 		return errors.New("a point needs to be associated with a channel")
// 	}

// 	if !StringSliceContains(SupportedPointFormat, p.Format) {
// 		return errors.New(fmt.Sprintf(
// 			"unsupported data format %s, supported formats are %s",
// 			p.Format, strings.Join(SupportedPointFormat, ",")))
// 	}

// 	switch p.Format {
// 	case "url":
// 		values, err := url.ParseQuery(raw)
// 		if err != nil {
// 			return err
// 		}
// 		i, err := strconv.ParseInt(values.Get("timestamp"), 10, 64)
// 		if err != nil {
// 			return err
// 		}
// 		p.Timestamp = time.Unix(i, 0)

// 		tags := make(map[string]string)
// 		for _, tag := range p.channel.Tags {
// 			if len(values.Get(tag)) > 0 {
// 				tags[tag] = values.Get(tag)
// 			}
// 		}
// 		p.Tags = tags

// 		fields := make(map[string]interface{})
// 		for fieldName, fieldType := range p.Fields {
// 			fieldV := values.Get(fieldName)
// 			if len(fieldV) > 0 {

// 			}
// 		}

// 	case "json":
// 	}

// }
