package presenters

import (
	. "github.com/vivowares/eywa/models"
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
		ID:          c.Base64Id(),
		Name:        c.Name,
		Description: c.Description,
	}
}

func NewChannelDetail(c *Channel) *ChannelDetail {
	return &ChannelDetail{
		ID:      c.Base64Id(),
		Channel: c,
	}
}
