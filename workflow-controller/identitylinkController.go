package controller

import (
	"encoding/json"
	"fmt"
	"github.com/mumushuiding/util"
	"go-workflow/workflow-engine/service"
	"net/http"
)

// FindParticipantByProcInstID 根据流程id查询流程参与者
func FindParticipantByProcInstID(writer http.ResponseWriter, request *http.Request) {
	type Data struct {
		ProcInstID int `json:"procInstID"`
	}
	var data Data
	json.NewDecoder(request.Body).Decode(&data)
	fmt.Println("data:", data)

	if data.ProcInstID == 0 {
		util.ResponseErr(writer, "流程 procInstID 不能为空")
		return
	}
	procInstID := data.ProcInstID

	result, err := service.FindParticipantByProcInstID(procInstID)
	if err != nil {
		util.ResponseErr(writer, err)
		return
	}
	fmt.Fprintf(writer, result)

}
