package presenters

import (
	"github.com/vivowares/eywa/configs"
)

func NewConf(cfg *configs.Conf) *configs.Conf {
	_cfg, _ := cfg.DeepCopy()
	_cfg.Security.SSL = nil
	_cfg.Security.Dashboard.Password = ""
	_cfg.Security.Dashboard.AES = nil

	return _cfg
}
