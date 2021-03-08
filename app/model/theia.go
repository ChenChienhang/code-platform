// @Author: 陈健航
// @Date: 2021/3/5 20:09
// @Description:
package model

type GetIdeUrlReq struct {
	LanguageEnum int
	UserId       int
	LabId        int
}

type CloseIdeReq struct {
	LanguageEnum int
	UserId       int
	LabId        int
}
