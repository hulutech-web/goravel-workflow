package prochook

import (
	"fmt"
	"github.com/hulutech-web/goravel-workflow/models"
	"reflect"
)

// Hookable 接口定义了 Passhook 和 UnPasshook 方法。
type Hookable interface {
	Passhook()
	UnPasshook()
}

// BaseHooker 提供了一个实现 Hookable 接口的基础结构体。
// 它嵌入了 Emp 结构体，并在方法中调用 Emp 的方法。
type BaseHooker struct {
	models.Emp
}

// Passhook 方法实现了 Hookable 接口。
// 它先调用 Emp 的 Passhook 方法，然后再调用自己的方法。
func (bh *BaseHooker) Passhook() {
	bh.Emp.Passhook()
	fmt.Println("BaseHooker Passhook called.")
}

// UnPasshook 方法实现了 Hookable 接口。
// 它先调用 Emp 的 UnPasshook 方法，然后再调用自己的方法。
func (bh *BaseHooker) UnPasshook() {
	bh.Emp.UnPasshook()
	fmt.Println("BaseHooker UnPasshook called.")
}

// Hooker 是一个通用的结构体，它能够自动检测并调用嵌入的 BaseHooker 的方法。
type Hooker struct {
	BaseHooker
}

// Passhook 方法实现了 Hookable 接口。
// 它会自动调用 BaseHooker 的 Passhook 方法。
func (h *Hooker) Passhook() {
	h.callBaseHookerMethod("Passhook")
}

// UnPasshook 方法实现了 Hookable 接口。
// 它会自动调用 BaseHooker 的 UnPasshook 方法。
func (h *Hooker) UnPasshook() {
	h.callBaseHookerMethod("UnPasshook")
}

// callBaseHookerMethod 是一个辅助函数，用于调用 BaseHooker 中的方法。
func (h *Hooker) callBaseHookerMethod(methodName string) {
	method := reflect.ValueOf(h.BaseHooker).MethodByName(methodName)
	if method.IsValid() && method.Kind() == reflect.Func {
		method.Call(nil) // nil 表示没有参数6
	}
}
