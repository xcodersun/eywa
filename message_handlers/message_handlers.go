package message_handlers

import (
	"errors"
	. "github.com/vivowares/eywa/connections"
	"github.com/vivowares/eywa/models"
)

var SupportedMessageHandlers = map[string]*Middleware{"indexer": Indexer, "logger": Logger}

var channelNotFound = errors.New("channel not found when indexing data")

func findCachedChannel(idStr string) (*models.Channel, bool) {
	id := models.DecodeHashId(idStr)
	ch, found := models.FetchCachedChannelById(id)
	return ch, found
}
