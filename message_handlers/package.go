package message_handlers

import (
	"github.com/vivowares/eywa/models"
)

func findCachedChannel(idStr string) (*models.Channel, bool) {
	id := models.DecodeHashId(idStr)
	ch, found := models.FetchCachedChannelById(id)
	return ch, found
}
