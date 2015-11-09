package models

import (
	"time"
)

// use point to write
// point data is stored in influxdb or elasticsearch
type Point struct {
	Raw       string // goes to archive
	channel   Channel
	Timestamp time.Time
	Tags      map[string]string
	Fields    map[string]interface{}
	// Validate() map[string]error //field/tag -> error
	// Store() error               // -> index db
}

//https://influxdb.com/docs/v0.9/write_protocols/write_syntax.html
