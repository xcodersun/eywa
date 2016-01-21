package main

import (
	"encoding/json"
	"flag"
	"fmt"
	. "github.com/vivowares/octopus/configs"
	. "github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"os"
	"path"
)

var messages = `
{
	"messages": {
		"_all": { "enabled": false },
		"_source": { "enabled": false },
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

func main() {
	configFile := flag.String("conf", "", "config file location")
	flag.Parse()
	if len(*configFile) == 0 {
		defaultConf := "/etc/octopus/octopus.yml"
		if _, err := os.Stat(defaultConf); os.IsNotExist(err) {
			home := os.Getenv("OCTOPUS_HOME")
			if len(home) == 0 {
				panic("ENV OCTOPUS_HOME is not set")
			}

			*configFile = path.Join(home, "configs", "octopus_development.yml")
		} else {
			*configFile = defaultConf
		}
	}

	PanicIfErr(InitializeConfig(*configFile))
	InitialLogger()
	PanicIfErr(InitializeIndexClient())

	settings := map[string]interface{}{
		"index.cache.query.enable": true,
		"number_of_shards":         Config().Indices.NumberOfShards,
		"number_of_replicas":       Config().Indices.NumberOfReplicas,
	}

	mappings := map[string]interface{}{}
	PanicIfErr(json.Unmarshal([]byte(messages), &mappings))

	if Config().Indices.TTLEnabled {
		m := mappings["messages"].(map[string]interface{})
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
	resp, err := IndexClient.IndexPutTemplate("channels_template").BodyJson(body).Do()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(resp)
	}
}
