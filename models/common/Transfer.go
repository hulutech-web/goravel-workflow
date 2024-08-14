package common

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goravel/framework/facades"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FieldValue []string
type RuleItem struct {
	RuleName  string `json:"rule_name" form:"rule_name"`
	RuleTitle string `json:"rule_title" form:"rule_title"`
	RuleValue string `json:"rule_value" form:"rule_value"`
}
type Rule []RuleItem

func (t *Rule) Scan(value interface{}) error {
	bytesValue, _ := value.([]byte)
	return json.Unmarshal(bytesValue, t)
}

func (t Rule) Value() (driver.Value, error) {
	//如果t为nil,返回nil
	return json.Marshal(t)
}

func (t *FieldValue) Scan(value interface{}) error {
	bytesValue, _ := value.([]byte)
	return json.Unmarshal(bytesValue, t)
}

func (t FieldValue) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// 地图位置
type Coordinates struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type CoordRes struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

func (c CoordRes) GormDataType() string {
	return "geometry"
}

// GormValue 根据坐标数据生成 SQL 表达式
func (c CoordRes) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	//如果c为nil,返回nil
	if c.Coordinates == nil {
		//赋值为nil
		return clause.Expr{SQL: "NULL", Vars: []interface{}{}}
	}

	// 生成 SQL 表达式
	return clause.Expr{
		SQL:  "ST_GeomFromText(?,4326)",
		Vars: []interface{}{fmt.Sprintf("POINT(%f %f)", c.Coordinates[1], c.Coordinates[0])},
	}
}

// Scan 方法实现了 sql.Scanner 接口，用于从数据库中扫描和解码坐标数据
func (c *CoordRes) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid value type, expected []byte")
	}
	if len(bytes) == 0 {
		c = nil
		return nil
	}
	var coordRes string
	querySql := "SELECT ST_AsGeoJSON(?) as coord"
	param := string(bytes)
	facades.Orm().Query().Raw(querySql, param).Pluck("coord", &coordRes)
	err := json.Unmarshal([]byte(coordRes), c)
	return err
}
