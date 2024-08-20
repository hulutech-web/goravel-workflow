package official_plugins

import "fmt"

// 分配任务插件
type DistributePlugin struct {
}

func NewDistributePlugin() *DistributePlugin {
	return &DistributePlugin{}
}

func (c *DistributePlugin) Register() string {
	return "official_plugins.distribute"
}

func (c *DistributePlugin) Action() func(string) error {
	return func(task string) error {
		fmt.Println("distribute task", task)
		return nil
	}
}
