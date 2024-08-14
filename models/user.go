package models

import (
	"github.com/goravel/framework/contracts/database/factory"
	"github.com/goravel/framework/database/orm"
	"github.com/hulutech-web/goravel-workflow/database/factories"
)

type User struct {
	orm.Model
	Name      string `gorm:"column:name;type:varchar(255);null;default:'';comment:'姓名'" form:"name" json:"name"`
	AvatarUrl string `gorm:"column:avatarUrl;type:varchar(255);null;default:'';comment:'头像地址'" form:"avatarUrl" json:"avatarUrl"`
	Mobile    string `gorm:"column:mobile;type:varchar(255);null;default:'';comment:'手机号'" form:"mobile" json:"mobile"`
	Password  string `gorm:"column:password;type:varchar(255);null;default:'';comment:'密码'" form:"password" json:"password"`
	Email     string `gorm:"column:email;type:varchar(255);null;default:'';comment:'邮箱'" form:"email" json:"email"`
	IdNumber  string `gorm:"column:id_number;type:varchar(255);null;default:'';comment:'证件号'" form:"idNumber" json:"idNumber"`
	IsMember  int    `gorm:"column:is_member;type:int;default:1;comment:'是否会员1非会员，2会员'" form:"is_member" json:"is_member"`
	State     int    `gorm:"column:state;type:int;default:1;comment:'状态1正常，2禁用'" form:"state" json:"state"`
	Dept      Dept   `gorm:"-" json:"dept"`
	orm.SoftDeletes
}

func (*User) Factory() factory.Factory {
	return &factories.UserFactory{}
}
