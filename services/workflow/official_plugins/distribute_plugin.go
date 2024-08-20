package official_plugins

import "fmt"

// 分配任务插件
type DistributePlugin struct {
}

func NewDistributePlugin() *DistributePlugin {
	return &DistributePlugin{}
}

func (c *DistributePlugin) Register() string {
	fmt.Println("register distribute plugin called")
	return "official_plugins.distribute"
}

func (c *DistributePlugin) Action() func(string) error {
	fmt.Println("distribute plugin action called")
	return func(task string) error {
		fmt.Println("distribute task", task)
		return nil
	}
}
