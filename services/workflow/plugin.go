package workflow

import (
	"sync"
)

type Plugin interface {
	Register() string
	Action() func(string) error
}

type Collector struct {
	hooks   []string //
	mutex   sync.Mutex
	plugins []Plugin
}

func NewCollector(plugins []Plugin) *Collector {
	return &Collector{
		plugins: plugins,
	}
}

func (c *Collector) Boot() {
	// 加载插件
}

func (c *Collector) RegisterPlugin() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, plugin := range c.plugins {
		plugin.Register()
		plugin.Action()
	}
}
