// @Author: 陈健航
// @Date: 2021/1/16 21:54
// @Description:
package api

import (
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/net/ghttp"
)

var WebSocket = new(webSocketController)

type webSocketController struct{}

var stu = gmap.New(true) // 使用默认的并发安全Map

func (c *webSocketController) Stu(r *ghttp.Request) {
	//ws, err := r.WebSocket()
	//if err != nil {
	//	response.Exit(r, err)
	//}
	//msg := &model.
	//stu.Set(ws, component.GetUserId(r))
	//for {
	//	// 阻塞读取ws消息
	//	_, msgByte, err := ws.ReadMessage()
	//	// 读取消息异常
	//	if err != nil {
	//		stu.Remove(ws)
	//		break
	//	}
	//
	//}
}
