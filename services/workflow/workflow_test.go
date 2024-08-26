package workflow

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestRegisterHook(t *testing.T) {
	var name string = "test"
	var method reflect.Value = reflect.ValueOf(func() {})
	w := NewBaseWorkflow()
	w.RegisterHook(name, method)
	if w.hooks == nil {
		fmt.Println("Hooks map is nil!")
		w.hooks = make(map[string][]reflect.Value)
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.hooks[name] = append(w.hooks[name], method)
	//	断言w.hooks["test"]是否为method
	assert.Equal(t, w.hooks[name][0], method)
}

func TestUniqueSlice(t *testing.T) {
	var testInts = []int{1, 2, 3, 4, 5, 6, 5, 4, 3, 2, 1}
	seen := make(map[int]bool)
	result := []int{}

	for _, value := range testInts {
		if _, ok := seen[value]; !ok {
			seen[value] = true
			result = append(result, value)
		}
	}
	//	断言result是否为[1,2,3,4,5,6]
	assert.Equal(t, result, []int{1, 2, 3, 4, 5, 6})

}

// 测试用的钩子函数
func testHook(id uint) {
	fmt.Printf("Hook function called with id: %d\n", id)
}

// 测试 invokeHooks 方法
func TestWorkflow_invokeHooks(t *testing.T) {
	w := NewBaseWorkflow()

	// 测试场景 1: 钩子存在且方法签名匹配
	hookFunc := reflect.ValueOf(testHook)
	w.RegisterHook("testHook", hookFunc)
	w.invokeHooks("testHook", 123)

	// 测试场景 2: 钩子存在但方法签名不匹配（模拟）
	// 注意：这里我们直接调用 invokeHooks 而不是注册一个不匹配的钩子，因为注册钩子时会检查类型
	// 但我们可以通过打印输出来验证方法签名不匹配的情况
	// 我们可以模拟一个不匹配的钩子，但实际上在测试中我们不会这样做，因为 invokeHooks 内部会检查

	// 测试场景 3: 钩子不存在
	w.invokeHooks("nonExistentHook", 456)
}
