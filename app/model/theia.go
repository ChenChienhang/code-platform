// @Author: 陈健航
// @Date: 2021/3/5 20:09
// @Description:
package model

type GetIDEUrlReq struct {
	UserId int
	LabId  int
}

type CloseIDEReq struct {
	Duration int
	UserId   int
	LabId    int
}

type CompilerErrorLogResp struct {
	StuId       int    `json:"stu_id"`
	StuNum      string `json:"stu_num"`
	CompilerLog string `json:"compiler_log"`
}
