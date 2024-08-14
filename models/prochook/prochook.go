package prochook

import (
	"fmt"
	"reflect"
)

// Emp 表示员工模型，包含 passhook 和 unpasshook 方法。
type Emp struct{}

// passhook 方法的默认实现。
func (e Emp) passhook() {
	fmt.Println("Emp passhook called.")
}

// unpasshook 方法的默认实现。
func (e Emp) unpasshook() {
	fmt.Println("Emp unpasshook called.")
}

// Hookable 接口定义了 passhook 和 unpasshook 方法。
type Hookable interface {
	passhook()
	unpasshook()
}

// BaseHooker 提供了一个实现 Hookable 接口的基础结构体。
// 它嵌入了 Emp 结构体，并在方法中调用 Emp 的方法。
type BaseHooker struct {
	Emp
}

// passhook 方法实现了 Hookable 接口。
// 它先调用 Emp 的 passhook 方法，然后再调用自己的方法。
func (bh *BaseHooker) passhook() {
	bh.Emp.passhook()
	fmt.Println("BaseHooker passhook called.")
}

// unpasshook 方法实现了 Hookable 接口。
// 它先调用 Emp 的 unpasshook 方法，然后再调用自己的方法。
func (bh *BaseHooker) unpasshook() {
	bh.Emp.unpasshook()
	fmt.Println("BaseHooker unpasshook called.")
}

// Hooker 是一个通用的结构体，它能够自动检测并调用嵌入的 BaseHooker 的方法。
type Hooker struct {
	BaseHooker
}

// passhook 方法实现了 Hookable 接口。
// 它会自动调用 BaseHooker 的 passhook 方法。
func (h *Hooker) passhook() {
	h.callBaseHookerMethod("passhook")
}

// unpasshook 方法实现了 Hookable 接口。
// 它会自动调用 BaseHooker 的 unpasshook 方法。
func (h *Hooker) unpasshook() {
	h.callBaseHookerMethod("unpasshook")
}

// callBaseHookerMethod 是一个辅助函数，用于调用 BaseHooker 中的方法。
func (h *Hooker) callBaseHookerMethod(methodName string) {
	method := reflect.ValueOf(h.BaseHooker).MethodByName(methodName)
	if method.IsValid() && method.Kind() == reflect.Func {
		method.Call(nil) // nil 表示没有参数6
	}
}

// User 结构体是外部使用者定义的结构体，它可以嵌入 Hooker。
// 用户可以根据需要覆盖 passhook 和 unpasshook 方法。
type User struct {
	Hooker
	Name string
}

// passhook 方法实现了 Hookable 接口。
// 它会自动调用 Hooker 的 passhook 方法，然后再调用自己的方法。
func (u *User) passhook() {
	fmt.Printf("User %s passhook called.\n", u.Name)
}

// unpasshook 方法实现了 Hookable 接口。
// 它会自动调用 Hooker 的 unpasshook 方法，然后再调用自己的方法。
func (u *User) unpasshook() {
	fmt.Printf("User %s unpasshook called.\n", u.Name)
}

// TestHook 调用 passhook 和 unpasshook 方法进行测试。
func TestHook() {
	user := &User{
		Name: "John Doe",
	}

	user.passhook()
	user.unpasshook()
}
