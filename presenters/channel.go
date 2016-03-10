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
	hashId, _ := c.HashId()
	return &ChannelBrief{
		ID:          hashId,
		Name:        c.Name,
		Description: c.Description,
	}
}

func NewChannelDetail(c *Channel) *ChannelDetail {
	hashId, _ := c.HashId()
	return &ChannelDetail{
		ID:      hashId,
		Channel: c,
	}
}
