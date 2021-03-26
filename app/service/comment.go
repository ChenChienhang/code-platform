// @Author: 陈健航
// @Date: 2021/1/15 21:30
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/response"
	"github.com/gogf/gf/container/gset"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/frame/g"
	"math"
)

var CourseCommentService = new(courseCommentService)

type courseCommentService struct{}

// ListCourseCommentByCourseId 列出所有课程评论
// @receiver s
// @params req
// @return *response.PageResp
// @return error
// @date 2021-03-16 14:25:54
func (s *courseCommentService) ListCourseCommentByCourseId(req *model.ListCourseCommentReq) (resp *response.PageResp, err error) {
	var records = make([]*model.CourseCommentEntity, 0)
	// 分页，排序
	if err = dao.CourseComment.Page(req.PageCurrent, req.PageSize).Order(dao.CourseComment.Columns.CreatedAt+" desc").
		// 主评，对应课程
		Where(dao.CourseComment.Columns.Pid, 0).And(dao.CourseComment.Columns.CourseId, req.CourseId).
		FieldsEx(dao.CourseComment.Columns.DeletedAt).
		ScanList(&records, "Comment"); err != nil {
		return nil, err
	}

	// 查子评论
	if err = dao.CourseComment.
		Where(dao.CourseComment.Columns.Pid,
			gdb.ListItemValuesUnique(records, "Comment", "CourseCommentId")).
		Order(dao.CourseComment.Columns.CreatedAt+" desc").
		FieldsEx(dao.CourseComment.Columns.DeletedAt).
		ScanList(&records, "ReplyComments", "Comment", "pid:CourseCommentId"); err != nil {
		return nil, err
	}

	for _, v := range records {
		if v.ReplyComments == nil {
			v.ReplyComments = make([]*model.CourseCommentResp, 0)
		}
	}

	// 收集userId
	userIds := gset.NewIntSet(false)
	for _, v := range records {
		userIds.Add(v.Comment.UserId)
		for _, v1 := range v.ReplyComments {
			userIds.Add(v1.UserId)
			userIds.Add(v1.ReplyUserId)
		}
	}

	userDetails := make([]*model.SysUser, 0)
	if err = dao.SysUser.WherePri(userIds.Slice()).Fields(
		dao.SysUser.Columns.UserId,
		dao.SysUser.Columns.RealName,
		dao.SysUser.Columns.AvatarUrl,
	).FindScan(&userDetails); err != nil {
		return nil, err
	}
	// 装配
	for _, v := range records {
		for _, u := range userDetails {
			if v.Comment.UserId == u.UserId {
				v.Comment.Username = u.RealName
				v.Comment.UserAvatarUrl = u.AvatarUrl
			}
		}
		for _, v1 := range v.ReplyComments {
			count := 0
			for _, u := range userDetails {
				if v1.UserId == u.UserId {
					v1.Username = u.RealName
					count++
					if count == 2 {
						break
					}
				}
				if v1.ReplyUserId == u.UserId {
					v1.ReplyUsername = u.RealName
					count++
					if count == 2 {
						break
					}
				}
			}
		}
	}

	// 分页信息
	count, err := dao.CourseComment.
		Where(dao.CourseComment.Columns.Pid, 0).
		Where(dao.CourseComment.Columns.CourseId, req.CourseId).Count()
	if err != nil {
		return nil, err
	}

	// 分页信息整合
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

// ListLabCommentByLabId 列出所有实验评论
// @receiver s
// @params req
// @return *response.PageResp
// @return error
// @date 2021-03-16 14:26:05
func (s *courseCommentService) ListLabCommentByLabId(req *model.ListLabCommentReq) (resp *response.PageResp, err error) {
	var records = make([]*model.LabCommentEntity, 0)
	// 分页，排序
	if err = dao.LabComment.Page(req.PageCurrent, req.PageSize).Order(dao.LabComment.Columns.CreatedAt+" desc").
		// 主评，对应课程
		Where(dao.LabComment.Columns.Pid, 0).And(dao.LabComment.Columns.LabId, req.LabId).
		FieldsEx(dao.LabComment.Columns.DeletedAt).
		ScanList(&records, "Comment"); err != nil {
		return nil, err
	}

	// 查子评论
	if err = dao.LabComment.
		Where(dao.LabComment.Columns.Pid,
			gdb.ListItemValuesUnique(records, "Comment", "CourseCommentId")).
		FieldsEx(dao.LabComment.Columns.DeletedAt).
		ScanList(&records, "ReplyComments", "Comment", "pid:CourseCommentId"); err != nil {
		return nil, err
	}
	for _, v := range records {
		if v.ReplyComments == nil {
			v.ReplyComments = make([]*model.LabCommentResp, 0)
		}
	}

	// 收集userId
	userIds := gset.NewIntSet(false)
	for _, v := range records {
		userIds.Add(v.Comment.UserId)
		for _, v1 := range v.ReplyComments {
			userIds.Add(v1.UserId)
			userIds.Add(v1.ReplyUserId)
		}
	}

	// 装配字段
	userDetails := make([]*model.SysUser, 0)
	if err = dao.SysUser.WherePri(userIds.Slice()).Fields(
		dao.SysUser.Columns.AvatarUrl,
		dao.SysUser.Columns.UserId,
		dao.SysUser.Columns.RealName,
	).FindScan(&userDetails); err != nil {
		return nil, err
	}
	for _, v := range records {
		for _, u := range userDetails {
			if v.Comment.UserId == u.UserId {
				v.Comment.Username = u.RealName
				v.Comment.UserAvatarUrl = u.AvatarUrl
			}
		}
		for _, v1 := range v.ReplyComments {
			count := 0
			for _, u := range userDetails {
				if v1.UserId == u.UserId {
					v1.Username = u.RealName
					v1.UserAvatarUrl = u.AvatarUrl
					count++
					if count == 2 {
						break
					}
				}
				if v1.ReplyUserId == u.UserId {
					v1.ReplyUsername = u.RealName
					count++
					if count == 2 {
						break
					}
				}
			}
		}
	}

	// 分页信息
	count, err := dao.LabComment.
		Where(dao.LabComment.Columns.Pid, 0).
		Where(dao.LabComment.Columns.LabId, req.LabId).Count()
	if err != nil {
		return nil, err
	}

	// 分页信息整合
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

// InsertCourseComment
// @receiver s
// @params req
// @return err
// @date 2021-01-31 18:29:13
func (s *courseCommentService) InsertCourseComment(req *model.InsertCourseCommentReq) (err error) {
	// 保存模型
	var courseComment = &model.CourseComment{
		CommentText: req.CommentText,
		UserId:      req.UserId,
	}
	if req.ReplyId != 0 {
		// 回复别人的评论,要做处理
		tmp, err := dao.CourseComment.WherePri(req.ReplyId).
			Fields(dao.CourseComment.Columns.Pid, dao.CourseComment.Columns.ReplyUserId).FindOne()
		if err != nil {
			return err
		}
		// 回复的是主评,则pid为主评id
		if tmp.Pid == 0 {
			courseComment.Pid = req.ReplyId
		} else {
			// 否则，pid与被回复的pid保持一致
			courseComment.Pid = tmp.Pid
		}

		// 被回复的那条评论的用户id
		courseComment.ReplyUserId = tmp.ReplyUserId
	}

	// 保存
	if _, err = dao.CourseComment.Data(courseComment).Insert(); err != nil {
		return err
	}
	return nil
}

// InsertLabComment
// @receiver s
// @params saveModel
// @return error
// @date 2021-01-31 18:29:22
func (s *courseCommentService) InsertLabComment(req *model.InsertLabCommentReq) (err error) {
	var labComment = &model.LabComment{
		CommentText: req.CommentText,
		UserId:      req.UserId,
	}
	if req.ReplyId != 0 {
		// 回复别人的评论,要做处理
		replyComment, err := dao.LabComment.WherePri(req.ReplyId).
			Fields(dao.LabComment.Columns.Pid, dao.LabComment.Columns.ReplyUserId).FindOne()
		if err != nil {
			return err
		}
		// 回复的是主评,则pid为主评id
		if replyComment.Pid == 0 {
			labComment.Pid = req.ReplyId
		} else {
			// 否则，pid与被回复的pid保持一致
			labComment.Pid = replyComment.Pid
		}
		// 被回复的那条评论的用户昵称
		labComment.ReplyUserId = replyComment.ReplyUserId
	}

	// 保存
	if _, err = dao.LabComment.Insert(labComment); err != nil {
		return err
	}
	return nil
}

// DeleteLabComment 删除实验评论
// @receiver s
// @params commentId
// @params userId
// @return err
// @date 2021-03-16 14:26:23
func (s *courseCommentService) DeleteLabComment(commentId int, userId int) (err error) {
	if _, err = dao.LabComment.Where(g.Map{
		dao.LabComment.Columns.LabCommentId: commentId,
		dao.LabComment.Columns.UserId:       userId,
	}).Data(g.Map{
		dao.LabComment.Columns.CommentText: "该评论已删除",
	}).Update(); err != nil {
		return err
	}
	return nil
}

// DeleteCourseComment 删除课程评论
// @receiver s
// @params commentId
// @params userId
// @return err
// @date 2021-03-16 14:26:32
func (s *courseCommentService) DeleteCourseComment(commentId int, userId int) (err error) {
	if _, err = dao.CourseComment.Where(g.Map{
		dao.CourseComment.Columns.CourseCommentId: commentId,
		dao.CourseComment.Columns.UserId:          userId,
	}).Data(g.Map{
		dao.CourseComment.Columns.CommentText: "该评论已删除",
	}).Update(); err != nil {
		return err
	}
	return nil
}
