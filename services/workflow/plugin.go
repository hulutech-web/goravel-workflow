package workflow

import (
	"sync"
)

type Plugin interface {
	Register() string
	Action() func(string) error
	AddHook(hook string)
	Execute(plugin_name string, args ...interface{}) error
}

type Collector struct {
	hooks   []string //
	mutex   sync.Mutex
	plugins []Plugin
}

var collector *Collector

func NewCollector(plugins []Plugin) *Collector {
	if collector == nil {
		collector = &Collector{plugins: plugins}
	}
	return collector
}

func GetCollectorIns() *Collector {
	return collector
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
	c.hooks = append(c.hooks, hook)
}

// 执行插件中的Execute方法
func (c *Collector) DoPluginsExec(plugin_name string, args ...interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, plugin := range c.plugins {
		if err := plugin.Execute(plugin_name, args...); err != nil {
			return err
		}
	}
	return nil
}
