package common

type Plumb struct {
	Total int             `json:"total"`
	List  map[string]Node `json:"list"`
}

type Node struct {
	ID          int    `json:"id"`
	FlowId      int    `json:"flow_id"`
	ProcessName string `json:"process_name"`
	ProcessTo   string `json:"process_to"`
	Icon        string `json:"icon"`
	Style       string `json:"style"`
}

type ProcessCondition struct {
	Id       int    `json:"id"`
	Index    int    `json:"index"`
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
	Extra    string `json:"extra"`
}

type ProcessRequest struct {
	ProcessName      string             `json:"process_name"`
	ProcessPosition  int                `json:"process_position"`
	ChildFlowId      int                `json:"child_flow_id"`
	ChildAfter       int                `json:"child_after"`
	ChildBackProcess int                `json:"child_back_process"`
	AutoPerson       string             `json:"auto_person"`
	RangeEmpIds      []int              `json:"range_emp_ids"`
	RangeEmpText     []string           `json:"range_emp_text"`
	RangeDeptIds     []int              `json:"range_dept_ids"`
	RangeDeptText    []int              `json:"range_dept_text"`
	ProcessMode      string             `json:"process_mode"`
	ProcessCondition []ProcessCondition `json:"process_condition"`
	StyleWidth       int                `json:"style_width"`
	StyleHeight      int                `json:"style_height"`
	StyleColor       string             `json:"style_color"`
	StyleIcon        string             `json:"style_icon"`
}
