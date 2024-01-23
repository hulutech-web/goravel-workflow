package controller

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"go-workflow/workflow-engine/service"

	"github.com/mumushuiding/util"
)

// SaveProcdefByToken SaveProcdefByToken
func SaveProcdefByToken(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		util.ResponseErr(writer, "只支持Post方法！！Only support Post ")
		return
	}
	token, err := GetToken(request)
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	var procdef = service.Procdef{}
	err = util.Body2Struct(request, &procdef)
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	if len(procdef.Name) == 0 {
		util.ResponseErr(writer, "流程名称 name 不能为空")
		return
	}
	if procdef.Resource == nil || len(procdef.Resource.Name) == 0 {
		util.ResponseErr(writer, "字段 resource 不能为空")
		return
	}
	id, err := procdef.SaveProcdefByToken(token)
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	util.Response(writer, fmt.Sprintf("%d", id), true)
}

// SaveProcdef save new procdefnition
// 保存流程定义
func SaveProcdef(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		util.ResponseErr(writer, "只支持Post方法！！Only support Post ")
		return
	}
	var procdef = service.Procdef{}
	err := util.Body2Struct(request, &procdef)
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	if len(procdef.Userid) == 0 {
		util.ResponseErr(writer, "字段 userid 不能为空")
		return
	}
	if len(procdef.Company) == 0 {
		util.ResponseErr(writer, "字段 company 不能为空")
		return
	}
	if len(procdef.Name) == 0 {
		util.ResponseErr(writer, "流程名称 name 不能为空")
		return
	}
	if procdef.Resource == nil || len(procdef.Resource.Name) == 0 {
		util.ResponseErr(writer, "字段 resource 不能为空")
		return
	}
	id, err := procdef.SaveProcdef()
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	util.Response(writer, fmt.Sprintf("%d", id), true)
}

// FindAllProcdefPage find by page
// 分页查询
func FindAllProcdefPage(writer http.ResponseWriter, request *http.Request) {

	var procdef = service.Procdef{PageIndex: 1, PageSize: 10}
	err := util.Body2Struct(request, &procdef)
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	datas, err := procdef.FindAllPageAsJSON()
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", datas)
}

// writer http.ResponseWriter, request *http.Request
// @袁浩 根据id查询流程定义
func FindProcdefById(writer http.ResponseWriter, request *http.Request) {
	type ResId struct {
		ID string `form:"id" json:"id"`
	}
	var resId ResId
	if err := util.Body2Struct(request, &resId); err != nil {
		util.ResponseErr(writer, "request param 【id】 is not valid , id 不存在 ")
		return
	}

	id, err := strconv.Atoi(resId.ID)
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	procdef, err := service.FindProcdefByID(id)
	procdefJson, _ := json.Marshal(procdef)
	procdefJsonStr := string(procdefJson)
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	util.ResponseData(writer, procdefJsonStr)
}

// DelProcdefByID del by id
// 根据 id 删除
func DelProcdefByID(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	var ids []string = request.Form["id"]
	if len(ids) == 0 {
		util.ResponseErr(writer, "request param 【id】 is not valid , id 不存在 ")
		return
	}
	id, err := strconv.Atoi(ids[0])
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	err = service.DelProcdefByID(id)
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	util.ResponseOk(writer)
}

// @袁浩 模拟生成token
func MockGetToken(writer http.ResponseWriter, request *http.Request) {
	//生成一个随机token，并返回
	tokenMsg := map[string]string{"token": makeToken(32)}
	tokenStr, _ := json.Marshal(tokenMsg)
	util.ResponseData(writer, string(tokenStr))
}

func makeToken(len int) string {
	//生成len长度的随机字符串
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < len; i++ {
		result = append(result, bytes[r.Intn(len)])
	}
	return string(result)
}
