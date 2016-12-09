package main

import (
	"encoding/json"
	"fmt"
	. "github.com/eywa/configs"
	. "github.com/eywa/models"
	. "github.com/eywa/utils"
)

var messages = `
{
  "messages": {
    "_all": { "enabled": false },
    "_size" : {"enabled" : true},
    "dynamic_templates": [
      {
        "string_to_doc_values": {
          "match_mapping_type": "string",
          "mapping": {
            "type": "string",
            "index": "not_analyzed",
            "doc_values": true,
            "norms": {
              "enabled": false
            }
          }
        }
      },
      {
        "boolean_to_doc_values": {
          "match_mapping_type": "boolean",
          "mapping": {
            "type": "boolean",
            "doc_values": true
          }
        }
      },
      {
        "date_to_doc_values": {
          "match_mapping_type": "date",
          "mapping": {
            "type": "date",
            "format": "dateOptionalTime",
            "doc_values": true
          }
        }
      },
      {
        "integer_to_doc_values": {
          "match_mapping_type": "integer",
          "mapping": {
            "type": "integer",
            "doc_values": true
          }
        }
      },
      {
        "long_to_doc_values": {
          "match_mapping_type": "long",
          "mapping": {
            "type": "long",
            "doc_values": true
          }
        }
      },
      {
        "float_to_doc_values": {
          "match_mapping_type": "float",
          "mapping": {
            "type": "float",
            "doc_values": true
          }
        }
      },
      {
        "double_to_doc_values": {
          "match_mapping_type": "double",
          "mapping": {
            "type": "double",
            "doc_values": true
          }
        }
      }
    ],
    "properties": {
      "timestamp": {
        "type": "date",
        "format": "epoch_millis",
        "doc_values": true
      }
    }
  }
}
`
var activities = `
{
  "activities": {
    "_all": { "enabled": false },
    "_size" : {"enabled" : false},
    "dynamic_templates": [
      {
        "string_to_doc_values": {
          "match_mapping_type": "string",
          "mapping": {
            "type": "string",
            "index": "not_analyzed",
            "doc_values": true,
            "norms": {
              "enabled": false
            }
          }
        }
      },
      {
        "boolean_to_doc_values": {
          "match_mapping_type": "boolean",
          "mapping": {
            "type": "boolean",
            "doc_values": true
          }
        }
      },
      {
        "date_to_doc_values": {
          "match_mapping_type": "date",
          "mapping": {
            "type": "date",
            "format": "dateOptionalTime",
            "doc_values": true
          }
        }
      },
      {
        "integer_to_doc_values": {
          "match_mapping_type": "integer",
          "mapping": {
            "type": "integer",
            "doc_values": true
          }
        }
      },
      {
        "long_to_doc_values": {
          "match_mapping_type": "long",
          "mapping": {
            "type": "long",
            "doc_values": true
          }
        }
      },
      {
        "float_to_doc_values": {
          "match_mapping_type": "float",
          "mapping": {
            "type": "float",
            "doc_values": true
          }
        }
      },
      {
        "double_to_doc_values": {
          "match_mapping_type": "double",
          "mapping": {
            "type": "double",
            "doc_values": true
          }
        }
      }
    ],
    "properties": {
      "timestamp": {
        "type": "date",
        "format": "epoch_millis",
        "doc_values": true
      }
    }
  }
}
`

func setupES() {
	settings := map[string]interface{}{
		"index.cache.query.enable": true,
		"number_of_shards":         Config().Indices.NumberOfShards,
		"number_of_replicas":       Config().Indices.NumberOfReplicas,
	}

	mappings := map[string]interface{}{}
	FatalIfErr(json.Unmarshal([]byte(messages), &mappings))
	FatalIfErr(json.Unmarshal([]byte(activities), &mappings))

	if Config().Indices.TTLEnabled {
		m := mappings["messages"].(map[string]interface{})
		m["_ttl"] = map[string]interface{}{
			"enabled": true,
			"default": fmt.Sprintf("%.0fs", Config().Indices.TTL.Seconds()),
		}
		m = mappings["activities"].(map[string]interface{})
		m["_ttl"] = map[string]interface{}{
			"enabled": true,
			"default": fmt.Sprintf("%.0fs", Config().Indices.TTL.Seconds()),
		}
	}

	body := map[string]interface{}{
		"template": "channels*",
		"order":    0,
		"settings": settings,
		"mappings": mappings,
	}
	_, err := IndexClient.IndexPutTemplate("channels_template").BodyJson(body).Do()
	FatalIfErr(err)
}
