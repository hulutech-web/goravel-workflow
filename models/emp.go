package models

import (
	"fmt"
	"github.com/goravel/framework/database/orm"
)

type Emp struct {
	orm.Model
	Name     string `gorm:"column:name;not null" json:"name" form:"name"`
	Email    string `gorm:"column:email;not null;unique_index:users_email_unique" json:"email" form:"email"`
	Password string `gorm:"column:password;not null" json:"password" form:"password"` // Exclude password from JSON response
	WorkNo   string `gorm:"column:workno;not null;unique_index:users_workno_unique" json:"workno" form:"workno"`
	DeptID   int    `gorm:"column:dept_id;not null;default:0" json:"dept_id" form:"dept_id"`
	Leave    int    `gorm:"column:leave;not null;default:0" json:"leave" form:"leave"`
	UserID   uint   `gorm:"column:user_id;" json:"user_id" form:"user_id"`
	Dept     Dept   `json:"Dept"`
}

// Passhook 方法的默认实现。
func (e *Emp) Passhook() {
	fmt.Println("Emp Passhook called.")
}

// UnPasshook 方法的默认实现。
func (e *Emp) UnPasshook() {
	fmt.Println("Emp UnPasshook called.")
}

func (u *Emp) Register() string {
	return "Emp"
}

func (u *Emp) Action() func(string) error {
	return nil
}
