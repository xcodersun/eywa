package presenters

import (
	"encoding/json"
	"github.com/vivowares/eywa/configs"
)

type Conf map[string]interface{}

func NewConf(cfg *configs.Conf) Conf {
	js, _ := json.Marshal(cfg)
	_cfg := map[string]interface{}{}
	json.Unmarshal(js, &_cfg)

	sec := _cfg["security"].(map[string]interface{})
	delete(sec, "ssl")
	dash := sec["dashboard"].(map[string]interface{})
	delete(dash, "password")
	delete(dash, "aes")

	return Conf(_cfg)
}
