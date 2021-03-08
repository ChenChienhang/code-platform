// @Author: 陈健航
// @Date: 2021/2/15 16:26
// @Description:
package api

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service"
	"code-platform/library/common/response"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/util/gconv"
	"time"
)

var CheckInController = new(checkInController)

type checkInController struct{}

// CheckinForTeacher 教师启动签到操作
// @receiver c
// @params r
// @date 2021-02-28 10:40:52
func (c *checkInController) CheckinForTeacher(r *ghttp.Request) {
	var req *model.StartCheckInReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if err := service.CheckinService.StartCheckIn(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

func (c *checkInController) CheckinForStudent(r *ghttp.Request) {
	msg := new(model.CheckInMsg)
	ws, err := r.WebSocket()
	if err != nil {
		response.Exit(r, err)
	}
	for {
		cancelPolling := false
		var userId, courseId int
		_, msgByte, err := ws.ReadMessage()
		if err != nil {
			// ws断开
			glog.Debug("ws 断开")
			// 取消协程的轮询
			cancelPolling = true
			// 从等待池删除
			break
		}
		if err := gjson.DecodeTo(msgByte, msg); err != nil {
			glog.Errorf("err in ws decode:%s", err.Error())
			continue
		}
		switch msg.Type {
		case 1:
			userId = r.GetVar(dao.SysUser.Columns.UserId).Int()
			courseId = gconv.Int(msg.Data.Get("courseId"))
			// 查看签到情况
			ok, resp, err := service.CheckinService.CheckProcessing(courseId)
			if err != nil {
				glog.Errorf("ws 检查签到情况错误：%s", err.Error())
			}
			// 这里包含两种情况，已经启动签到和未启动或者签到过时
			c.write(ws, response.JsonResponse{Code: 0, Message: "执行成功", Data: resp})
			// 如果未启动签到
			if !ok && err != nil {
				// 启动一个轮询
				go func(cancel *bool, ws *ghttp.WebSocket) {
					// 未取消时
					for !*cancel {
						// 每秒轮询
						time.Sleep(time.Second * 1)
						ok, resp, err := service.CheckinService.Polling(courseId)
						if err != nil {
							glog.Errorf("ws 轮询错误：%s", err.Error())
						}
						// 查到发起了签到，写回客户端，取消轮询
						if ok {
							c.write(ws, response.JsonResponse{Code: 0, Message: "执行成功", Data: resp})
							break
						}
					}
				}(&cancelPolling, ws)
			}
		case 2:
			secretKey := msg.Data.Get("secretKey")
			resp, err := service.CheckinService.CheckIn(userId, courseId, secretKey)
			if err != nil {
				glog.Errorf("ws 签到错误:%s", err.Error())
			}
			c.write(ws, response.JsonResponse{Code: 0, Message: "执行成功", Data: resp})
		}
	}
}

func (c *checkInController) write(ws *ghttp.WebSocket, jsonResponse response.JsonResponse) {
	msgBytes, err := gjson.Encode(jsonResponse)
	if err != nil {
		glog.Errorf("ws json编码失败，err:%s", err.Error())
	}
	if err = ws.WriteMessage(ghttp.WS_MSG_TEXT, msgBytes); err != nil {
		glog.Errorf("ws 写入消息失败 err:%s", err.Error())
	}
}

func (c *checkInController) ListCheckinRecords(r *ghttp.Request) {
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

func (c *checkInController) ListCheckinDetails(r *ghttp.Request) {
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

func (c *checkInController) DeleteRecordsDetails(r *ghttp.Request) {
	checkInRecordId := r.GetInt("checkInRecordId")
	if _, err := dao.CheckinRecord.WherePri(checkInRecordId).Delete(); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

func (c *checkInController) ExportCheckinRecords(r *ghttp.Request) {
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

func (c *checkInController) UpdateCheckinDetail(r *ghttp.Request) {
	var req *model.UpdateCheckinDetail
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if err := service.CheckinService.UpdateCheckinDetail(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}
