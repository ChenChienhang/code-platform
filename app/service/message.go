// @Author: 陈健航
// @Date: 2021/3/22 0:14
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/response"
	"github.com/gogf/gf/frame/g"
	"math"
)

var messageService = newMessageService()

type MessageService struct {
	redisHeaderIsNewMessage string
}

func newMessageService() *MessageService {
	return &MessageService{"code.platform:ide:message"}
}

func (m *MessageService) IsNewMessageNotify(userId int) (bool, error) {
	r := g.Redis()
	// 新消息集合中是否有该用户
	v, err := r.DoVar("SISMEMBER ", m.redisHeaderIsNewMessage, userId)
	if err != nil {
		return false, err
	}
	// 存在该用户id，即存在消息
	return v.Int() == 1, err
}

func (m *MessageService) removeNewMessageNotify(userId int) (err error) {
	r := g.Redis()
	// 新消息集合中是否有该用户
	if _, err = r.DoVar("SREM ", m.redisHeaderIsNewMessage, userId); err != nil {
		return err
	}
	return nil
}

func (m *MessageService) ListMessage(req *model.ListMessageReq) (resp *response.PageResp, err error) {
	records := make([]*model.ListMessageResp, 0)
	d := dao.Message.Where(dao.Message.Columns.ReceiverId, req.UserId)
	if err = d.Page(req.PageCurrent, req.PageSize).Order(dao.Message.Columns.CreatedAt + " desc").Scan(&records); err != nil {
		return nil, err
	}
	count, err := d.Count()
	if err != nil {
		return nil, err
	}
	resp = &response.PageResp{
		Records: records,
		PageInfo: &response.PageInfo{
			Size:    len(records),
			Total:   count,
			Current: req.PageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(req.PageSize))),
		}}
	return resp, nil
}

func (m *MessageService) InsertMessage(msg *model.Message) (err error) {
	if _, err = dao.Message.Insert(msg); err != nil {
		return err
	}
	r := g.Redis()
	// 通知接受者有新消息
	if _, err = r.DoVar("SADD ", m.redisHeaderIsNewMessage, msg.ReceiverId); err != nil {
		return err
	}
	return nil
}
