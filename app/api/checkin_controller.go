// @Author: 陈健航
// @Date: 2021/2/15 16:26
// @Description:
package api

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service"
	"code-platform/library/common/response"
	"github.com/gogf/gf/net/ghttp"
)

var CheckInController = new(checkInController)

type checkInController struct{}

// CheckinForTeacher 教师启动签到操作
// @receiver c
// @params r
// @date 2021-02-28 10:40:52
func (receiver *checkInController) CheckinForTeacher(r *ghttp.Request) {
	var req *model.StartCheckInReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if err := service.CheckinService.StartCheckIn(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

func (receiver *checkInController) CheckinForStudent(r *ghttp.Request) {
	var req *model.CheckinForStudentReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	req.StuId = r.GetCtxVar(dao.SysUser.Columns.UserId).Int()
	if err := service.CheckinService.CheckIn(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

func (receiver *checkInController) ListCheckinRecords(r *ghttp.Request) {
	var req *model.ListCheckinRecordsReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	resp, err := service.CheckinService.ListCheckinRecords(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

func (receiver *checkInController) ListCheckinDetails(r *ghttp.Request) {
	var req *model.ListCheckinDetailsReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	resp, err := service.CheckinService.ListCheckinDetail(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

func (receiver *checkInController) DeleteRecordsDetails(r *ghttp.Request) {
	checkInRecordId := r.GetInt("checkInRecordId")
	if _, err := dao.CheckinRecord.WherePri(checkInRecordId).Delete(); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

func (receiver *checkInController) ExportCheckinRecords(r *ghttp.Request) {
	courseId := r.GetInt("courseId")
	file, err := service.CheckinService.ExportCheckInRecords(courseId)
	if err != nil {
		response.Exit(r, err)
	}
	if err = file.Write(r.Response.Writer); err != nil {
		response.Exit(r, err)
	}
	r.Response.Header().Set("Content-disposition", "attachment;filename="+"签到表.xlsx")
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Exit()
}

func (receiver *checkInController) UpdateCheckinDetail(r *ghttp.Request) {
	var req *model.UpdateCheckinDetail
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if err := service.CheckinService.UpdateCheckinDetail(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}
