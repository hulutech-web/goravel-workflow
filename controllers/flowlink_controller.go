package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
	"github.com/spf13/cast"
	"goravel/app/http/controllers/common"
	"goravel/app/models"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type FlowlinkController struct {
	//Dependent services
}

func NewFlowlinkController() *FlowlinkController {
	return &FlowlinkController{
		//Inject services
	}
}

func (r *FlowlinkController) Update(ctx http.Context) http.Response {
	var flow models.Flow
	err := ctx.Request().Bind(&flow)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "绑定数据错误", err)
	}
	if flow.Jsplumb == "" {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "流程图数据为空", nil)
	}
	//解析流程图数据
	jsMap := common.Plumb{}
	err = json.Unmarshal([]byte(flow.Jsplumb), &jsMap)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "解析流程图数据错误", err)
	}
	tx, _ := facades.Orm().Query().Begin()
	//更新总数
	jsMap.Total = len(jsMap.List)
	for _, node := range jsMap.List {
		//	1-更新process
		var process models.Process
		tx.Model(&models.Process{}).Where("id=?", node.ID).First(&process)
		style := node.Style
		process.Style = style
		//"width:128px;height:30px;line-height:30px;color:#FF8C00;left:461px;top:84px;"使用一个正则匹配到left:461px;top:84px;
		re := regexp.MustCompile(`left:(\d+)px;top:(\d+)px;`)

		matches := re.FindStringSubmatch(style)
		// 检查是否找到匹配项
		if matches != nil && len(matches) > 2 {
			leftValue := matches[1]
			topValue := matches[2]
			// 更新process的位置信息
			process.PositionLeft = fmt.Sprintf("%spx", leftValue)
			process.PositionTop = fmt.Sprintf("%spx", topValue)
		} else {
			tx.Rollback()
			return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "下一节点匹配失败", err)
		}
		tx.Model(&models.Process{}).Where("id=?", process.ID).Save(&process)
		//	更新process的位置信息

		//更新流程轨迹 flowlink表 type=Condition
		var old_process_ids []int
		tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).Where("type=?", "Condition").Where("process_id=?", node.ID).
			Pluck("next_process_id", &old_process_ids)
		if node.ProcessTo != "" {
			p1, err := parseCommaSeparatedInts(node.ProcessTo)
			if err != nil {
				tx.Rollback()
				return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "解析流程图数据错误", err.Error())
			}
			if !slicesEqual(p1, old_process_ids) {
				adds := arrayDiff(p1, old_process_ids)
				for _, add := range adds {
					tx.Model(&models.Flowlink{}).Create(&models.Flowlink{
						FlowID:        flow.ID,
						Type:          "Condition",
						ProcessID:     cast.ToUint(node.ID),
						NextProcessID: add,
						Sort:          100,
					})
				}
				dels := arrayDiff(old_process_ids, p1)
				tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).Where("type=?", "Condition").
					Where("process_id=?", node.ID).Where("next_process_id IN (?)", dels).Delete(&models.Flowlink{})
			}
		} else {
			if len(old_process_ids) > 1 {
				//	只保留一个，因为下一步骤不可能存在2个
				newOldId := old_process_ids[0]
				tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).Where("type=?", "Condition").
					Where("process_id=?", node.ID).Where("next_process_id IN (?)", old_process_ids).Delete(&models.Flowlink{})
				tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).Where("type=?", "Condition").
					Where("process_id=?", newOldId).Update("next_process_id", -1)
			} else {
				var fcount int64
				tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).Where("type=?", "Condition").
					Where("process_id=?", node.ID).Count(&fcount)
				if fcount > 0 {
					tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).Where("type=?", "Condition").
						Where("process_id=?", node.ID).Update("next_process_id", -1)
				} else {
					tx.Model(&models.Flowlink{}).Create(&models.Flowlink{
						FlowID:        flow.ID,
						Type:          "Condition",
						ProcessID:     cast.ToUint(node.ID),
						NextProcessID: -1,
						Sort:          100,
					})
				}
			}
		}
	}
	flow.IsPublish = false
	tx.Model(&models.Flow{}).Where("id=?", flow.ID).Save(&flow)
	tx.Commit()
	return httpfacades.NewResult(ctx).Success("保存成功", nil)
}

func slicesEqual(a, b []int) bool {
	sort.Ints(a)
	sort.Ints(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// 并返回一个整数切片。
func parseCommaSeparatedInts(s string) ([]int, error) {
	// 先去除字符串中的空白字符
	s = strings.TrimSpace(s)

	// 使用 strings.Split 将字符串按照逗号分割成子字符串切片
	parts := strings.Split(s, ",")

	// 创建一个整数切片
	var nums []int

	// 遍历每个子字符串，尝试将其转换为整数
	for _, part := range parts {
		// 去除每个子字符串中的空白字符
		part = strings.TrimSpace(part)

		// 使用 strconv.Atoi 转换子字符串为整数
		num, err := strconv.Atoi(part)
		if err != nil {
			// 如果转换失败，返回错误
			return nil, err
		}
		// 添加转换后的整数到切片中
		nums = append(nums, num)
	}

	// 返回整数切片
	return nums, nil
}

// arrayDiff 接受两个整数切片，返回一个新的切片，包含所有在第一个切片中但不在第二个切片中的元素。
func arrayDiff(slice1, slice2 []int) []int {
	diff := make([]int, 0)

	// 创建一个映射来存储第二个切片中的元素
	elementMap := make(map[int]bool)
	for _, elem := range slice2 {
		elementMap[elem] = true
	}

	// 遍历第一个切片
	for _, elem := range slice1 {
		// 如果元素不在第二个切片中，则添加到结果切片
		if !elementMap[elem] {
			diff = append(diff, elem)
		}
	}

	return diff
}
