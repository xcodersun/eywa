package message_handlers

import (
	. "github.com/vivowares/eywa/connections"
	"github.com/vivowares/eywa/models"
)

var SupportedMessageHandlers = map[string]*Middleware{"indexer": Indexer}

func findCachedChannel(idStr string) (*models.Channel, bool) {
	id := models.DecodeHashId(idStr)
	ch, found := models.FetchCachedChannelById(id)
	return ch, found
}
