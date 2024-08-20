package workflow

import (
	"sync"
)

type Plugin interface {
	Register() string
	Action() func(string) error
	AddHook(hook string)
	GetHooks() []string
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

func (c *Collector) RegisterPlugin(plugin_name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, plugin := range c.plugins {
		plugin.Register()
	}
	for _, plugin := range c.plugins {
		action := plugin.Action()
		action(plugin_name)
	}
}

func (c *Collector) AddHook(hook string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.hooks = append(c.hooks, hook)
}

func (c *Collector) GetHooks() []string {
	return c.hooks
}
