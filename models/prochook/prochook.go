package prochook

import (
	"fmt"
	"reflect"
)

// Emp 表示员工模型，包含 Passhook 和 UnPasshook 方法。
type Emp struct{}

// Passhook 方法的默认实现，接受一个 uint 类型的 id 作为参数。
func (e Emp) Passhook(id uint) {
	fmt.Printf("Emp Passhook called with id: %d.\n", id)
}

// UnPasshook 方法的默认实现，接受一个 uint 类型的 id 作为参数。
func (e Emp) UnPasshook(id uint) {
	fmt.Printf("Emp UnPasshook called with id: %d.\n", id)
}

// Hookable 接口定义了 Passhook 和 UnPasshook 方法，接受一个 uint 类型的 id 作为参数。
type Hookable interface {
	Passhook(id uint)
	UnPasshook(id uint)
}

// BaseHooker 提供了一个实现 Hookable 接口的基础结构体。
// 它嵌入了 Emp 结构体，并在方法中调用 Emp 的方法。
type BaseHooker struct {
	Emp
}

// Passhook 方法实现了 Hookable 接口。
// 它先调用 Emp 的 Passhook 方法，然后再调用自己的方法。
func (bh *BaseHooker) Passhook(id uint) {
	bh.Emp.Passhook(id)
	fmt.Printf("BaseHooker Passhook called with id: %d.\n", id)
}

// UnPasshook 方法实现了 Hookable 接口。
// 它先调用 Emp 的 UnPasshook 方法，然后再调用自己的方法。
func (bh *BaseHooker) UnPasshook(id uint) {
	bh.Emp.UnPasshook(id)
	fmt.Printf("BaseHooker UnPasshook called with id: %d.\n", id)
}

// Hooker 是一个通用的结构体，它能够自动检测并调用嵌入的 BaseHooker 的方法。
type Hooker struct {
	BaseHooker
}

// Passhook 方法实现了 Hookable 接口。
// 它会自动调用 BaseHooker 的 Passhook 方法，并传递 id 参数。
func (h *Hooker) Passhook(id uint) {
	h.callBaseHookerMethod("Passhook", id)
}

// UnPasshook 方法实现了 Hookable 接口。
// 它会自动调用 BaseHooker 的 UnPasshook 方法，并传递 id 参数。
func (h *Hooker) UnPasshook(id uint) {
	h.callBaseHookerMethod("UnPasshook", id)
}

// callBaseHookerMethod 是一个辅助函数，用于调用 BaseHooker 中的方法，并传递参数。
func (h *Hooker) callBaseHookerMethod(methodName string, id uint) {
	method := reflect.ValueOf(h.BaseHooker).MethodByName(methodName)
	if method.IsValid() && method.Kind() == reflect.Func {
		method.Call([]reflect.Value{reflect.ValueOf(id)}) // 传递 id 参数
	}
}

// User 结构体是外部使用者定义的结构体，它可以嵌入 Hooker。
// 用户可以根据需要覆盖 Passhook 和 UnPasshook 方法。
type User struct {
	Hooker
	Name string
}

// Passhook 方法实现了 Hookable 接口。
// 它会自动调用 Hooker 的 Passhook 方法，然后再调用自己的方法。
func (u *User) Passhook(id uint) {
	u.Hooker.Passhook(id) // 调用嵌入的 Hooker 的 Passhook 方法
	fmt.Printf("User %s Passhook called with id: %d.\n", u.Name, id)
}

// UnPasshook 方法实现了 Hookable 接口。
// 它会自动调用 Hooker 的 UnPasshook 方法，然后再调用自己的方法。
func (u *User) UnPasshook(id uint) {
	u.Hooker.UnPasshook(id) // 调用嵌入的 Hooker 的 UnPasshook 方法
	fmt.Printf("User %s UnPasshook called with id: %d.\n", u.Name, id)
}
