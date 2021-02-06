// @Author: 陈健航
// @Date: 2021/1/16 22:04
// @Description:
package service

import (
	"code-platform/library/common/code"
	"fmt"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/util/gconv"
	"strconv"
	"sync"
	"time"
)

var OnlineCheck = new(onlineCheck)

var once sync.Once

type onlineCheck struct {
	sync.Once
}

var (
	// 签到密钥redis key
	redisSecretCodeHeader = "check_in_code_"
	// 等待签名的学生的redis key
	redisWaitCheckStuPool = "check_wait_in_stu_pool_"
	// 临时存放已经签到成功的学生的学号
	redisCheckStuPool = "check_in_stu_pool_"
)

// StartCheckIn 教师启动签到
// @receiver c
// @params courseId
// @params duration
// @params secretCode
// @return *g.Map
// @return error
// @date 2021-01-17 23:41:08
func (c *onlineCheck) StartCheckIn(courseId, duration, secretCode int) (*g.Map, error) {
	// 存入签到密钥,限时
	if _, err := g.Redis().DoWithTimeout(time.Duration(duration)*time.Second, "SET",
		redisSecretCodeHeader+strconv.Itoa(courseId), secretCode); err != nil {
		return nil, err
	}

	// res:2,表示签到已经开始，expire，剩余签到时间
	resp := &g.Map{"res": 2, "expire": duration}
	return resp, nil
}

// IsChecking 检查是否正在签到
// @receiver c
// @params courseId
// @return *g.Map
// @return error
// @date 2021-01-17 23:40:52
func (c *onlineCheck) IsChecking(courseId int) (*g.Map, error) {
	// 检查签到密钥的过期时间
	v, err := g.Redis().DoVar("TTL", redisSecretCodeHeader+gconv.String(courseId))
	if err != nil {
		return nil, err
		// 键没有过期时间，预料外的错误
	} else if v.Int() == -1 {
		return nil, code.OtherError
	}
	// 该键不存在
	if v.Int() == -2 {
		// res:1,表示未开始签到或签到已结束
		resp := &g.Map{"res": 1}
		return resp, nil
	}
	// 老师已经开始签到，返回相应签到信息
	// res:2,表示签到已经开始，expire，剩余签到时间
	resp := &g.Map{"res": 2, "expire": fmt.Sprintf("%v", v.Int())}
	return resp, nil
}

// CheckIn 学生签到
// @receiver c
// @params courseId
// @params secretCode
// @return *g.Map
// @return error
// @date 2021-01-17 23:41:26
func (c *onlineCheck) CheckIn(stuId, courseId, secretCode int) (*g.Map, error) {
	// 学生签到
	key := redisSecretCodeHeader + strconv.Itoa(courseId)
	// 取出签到密钥
	v, err := g.Redis().DoVar("GET", key)
	if err != nil {
		return nil, err
	}
	// 密钥是空的
	if v.IsEmpty() {
		// 签到过时
		resp := &g.Map{"res": 3}
		return resp, nil
	}
	// 签到码不正确
	if secretCode != v.Int() {
		resp := &g.Map{"res": 4}
		return resp, nil
	}

	// 签到码正确

	resp := &g.Map{"res": 5}
	return resp, nil
}

// stuJoinThePool 学生加入池子
// @receiver c
// @params userId
// @params courseId
// @return int 当前池子里的学生个数
// @return error
// @date 2021-01-17 11:12:50
func (c *onlineCheck) stuJoinThePool(userId, courseId int) (int, error) {
	key := redisWaitCheckStuPool + gconv.String(courseId)
	if _, err := g.Redis().Do("SADD", key, userId); err != nil {
		return 0, err
	}
	v, err := g.Redis().DoVar("SCARD", key, userId)
	if err != nil {
		return 0, err
	}
	return v.Int(), nil
}

// 向客户端写入消息。
// 内部方法不会自动注册到路由中。
func (c *onlineCheck) write(ws *ghttp.WebSocket, msg g.Map) error {
	msgBytes, err := gjson.Encode(msg)
	if err != nil {
		return err
	}
	return ws.WriteMessage(ghttp.WS_MSG_TEXT, msgBytes)
}

// 向所有客户端群发消息。
// 内部方法不会自动注册到路由中。
func (c *onlineCheck) writeStudents(courseId int, msg g.Map) error {
	//b, err := gjson.Encode(msg)
	//if err != nil {
	//	return err
	//}
	//v := CheckPool.Get(courseId)
	//coursePool := v.(*gmap.IntAnyMap)
	//coursePool.RLockFunc()
	//users.RLockFunc(func(m map[interface{}]interface{}) {
	//	for user := range m {
	//		user.(*ghttp.WebSocket).WriteMessage(ghttp.WS_MSG_TEXT, []byte(b))
	//	}
	//})

	return nil
}

// 向客户端返回用户列表。
// 内部方法不会自动注册到路由中。
func (c *onlineCheck) writeUserCountToTeacher(courseId int) error {
	//array := garray.NewSortedStrArray()
	//names.Iterator(func(v string) bool {
	//	array.Add(v)
	//	return true
	//})
	//if err := c.writeStudents(model.ChatMsg{
	//	Type: "list",
	//	Data: array.Slice(),
	//	From: "",
	//}); err != nil {
	//	return err
	//}
	return nil
}
