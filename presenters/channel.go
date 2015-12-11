package presenters

import (
	. "github.com/vivowares/octopus/models"
)

type ChannelBrief struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewChannelBrief(c *Channel) *ChannelBrief {
	return &ChannelBrief{
		Id:          c.Id,
		Name:        c.Name,
		Description: c.Description,
	}
}
