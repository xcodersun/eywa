package presenters

import (
	"encoding/base64"
	. "github.com/vivowares/octopus/models"
	"strconv"
)

type ChannelBrief struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ChannelDetail struct {
	ID string `json:"id"`
	*Channel
}

func NewChannelBrief(c *Channel) *ChannelBrief {
	return &ChannelBrief{
		ID:          base64.URLEncoding.EncodeToString([]byte(strconv.Itoa(c.Id))),
		Name:        c.Name,
		Description: c.Description,
	}
}

func NewChannelDetail(c *Channel) *ChannelDetail {
	return &ChannelDetail{
		ID:      base64.URLEncoding.EncodeToString([]byte(strconv.Itoa(c.Id))),
		Channel: c,
	}
}
