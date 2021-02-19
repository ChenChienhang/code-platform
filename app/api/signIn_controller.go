// @Author: 陈健航
// @Date: 2021/2/15 16:26
// @Description:
package api

import (
	"code-platform/app/model"
	"code-platform/app/service"
	"code-platform/library/common/response"
	"github.com/gogf/gf/container/gset"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/util/gconv"
)

var SignInController = &signInController{
	student: gset.NewStrSet(true),
	teacher: gset.NewStrSet(true),
}

type signInController struct {
	teacher *gset.StrSet
	student *gset.StrSet
}

func (c *signInController) SignUpForTeacher(r *ghttp.Request) {
	msg := &model.OnlineCheckMsg{}
	ws, err := r.WebSocket()
	if err != nil {
		response.Exit(r, err)
	}
	for {
		_, msgByte, err := ws.ReadMessage()
		if err != nil {
			// 说明ws断开
			glog.Info("ws 断开:%s")
			break
		}
		if err := gjson.DecodeTo(msgByte, msg); err != nil {
			glog.Errorf("err in ws decode:%s", err.Error())
			continue
		}
		switch msg.Type {
		case 1:
			secretKey := msg.Data.Get("secretKey")
			courseId := msg.Data.Get("courseId")
			duration := msg.Data.Get("duration")
			name := msg.Data.Get("name")
			ret, err := service.OnlineCheck.StartCheckIn(gconv.Int(courseId), gconv.Int(duration), gconv.Int(secretKey), name)
			if err != nil {
				c.write(ws, response.JsonResponse{
					Code:    10000,
					Message: "启动签到错误:" + err.Error(),
					Data:    false,
				})
			}
			// 写给其他
			c.write(ws, response.JsonResponse{
				Code:    0,
				Message: "执行成功",
				Data:    ret,
			})
		}
	}
}

func (c *signInController) SignUpForStudent(r *ghttp.Request) {
	msg := &model.OnlineCheckMsg{}
	ws, err := r.WebSocket()
	if err != nil {
		response.Exit(r, err)
	}
	for {
		_, msgByte, err := ws.ReadMessage()
		if err != nil {
			// 说明ws断开
			glog.Info("ws 断开:%s")
			break
		}
		if err := gjson.DecodeTo(msgByte, msg); err != nil {
			glog.Errorf("err in ws decode:%s", err.Error())
			continue
		}
		switch msg.Type {
		case 1:
			userId := msg.Data.Get("userId")
			courseId := msg.Data.Get("courseId")
			service.OnlineCheck.StuJoinPool(gconv.Int(userId), gconv.Int(courseId))
		case 2:
			ws.ReadMessage()
		case 3:
		case 4:
		case 5:

		}
	}
}
func (c signInController) write(ws *ghttp.WebSocket, jsonResponse response.JsonResponse) {
	msgBytes, err := gjson.Encode(jsonResponse)
	if err != nil {
		glog.Errorf("ws json decode err:%s", err.Error())
	}
	if err = ws.WriteMessage(ghttp.WS_MSG_TEXT, msgBytes); err != nil {
		glog.Errorf("ws write err:%s", err.Error())
	}
}
