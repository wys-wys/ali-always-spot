package setting

import (
	"sync/atomic"

	"github.com/alibabacloud-go/tea/tea"
)

var gConfig atomic.Value

func C() *Config {
	return gConfig.Load().(*Config)
}

type Config struct {
	AccessKey *string
	SecretKey *string
	RegionId  *string
}

func InitConfig(c *Config) {
	if c.RegionId == nil {
		c.RegionId = tea.String("cn-hongkong")
	}
	gConfig.Store(c)
}
