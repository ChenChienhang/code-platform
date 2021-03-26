// @Author: 陈健航
// @Date: 2021/3/5 20:09
// @Description:
package model

import "github.com/gogf/gf/os/gtime"

type OpenIDEReq struct {
	UserId int
	LabId  int
}

type CloseIDEReq struct {
	UserId int
	LabId  int
}

type CheckCodeReq struct {
	TeacherId int
	StuId     int
	LabId     int
}

type CompilerErrorLogResp struct {
	StuId       int    `json:"stu_id"`
	StuNum      string `json:"stu_num"`
	CompilerLog string `json:"compiler_log"`
}

type TheiaStat struct {
	Port      int
	StartTime *gtime.Time
	Count     int
}

type ListContainerStat struct {
	Num      int
	realName int
	LabId    int
}

type PlagiarismCheckResp struct {
	UserId1    int
	UserId2    int
	RealName1  string
	RealName2  string
	Num1       string
	Num2       string
	Code1      string
	Code2      string
	Similarity int
}
