// @Author: 陈健航
// @Date: 2021/1/15 21:30
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/code"
	"code-platform/library/common/response"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/util/gconv"
	"math"
)

var CourseCommentService = new(courseCommentService)

type courseCommentService struct{}

func (s *courseCommentService) ListCourseCommentByCourseId(req *model.ListCourseCommentReq) (*model.CourseCommentEntityPageResp, error) {
	var CommentResps = make([]*model.CourseCommentEntity, 0)
	// 分页，排序
	if err := dao.CourseComment.Page(req.PageCurrent, req.PageSize).Order(dao.CourseComment.Columns.CreatedAt+" desc").
		// 主评，对应课程
		Where(dao.CourseComment.Columns.Pid, 0).And(dao.CourseComment.Columns.CourseId, req.CourseId).
		FieldsEx(dao.CourseComment.Columns.DeletedAt).
		ScanList(&CommentResps, "Comment"); err != nil {
		return nil, err
	}

	// 查子评论
	if err := dao.CourseComment.
		Where(dao.CourseComment.Columns.Pid,
			gdb.ListItemValuesUnique(CommentResps, "Comment", "CourseCommentId")).
		Order(dao.CourseComment.Columns.CreatedAt+" desc").
		FieldsEx(dao.CourseComment.Columns.DeletedAt).
		ScanList(&CommentResps, "ReplyComments", "Comment", "pid:CourseCommentId"); err != nil {
		return nil, err
	}

	for _, v := range CommentResps {
		if v.ReplyComments == nil {
			v.ReplyComments = make([]*model.CourseCommentResp, 0)
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
	resp := &model.CourseCommentEntityPageResp{
		Records: CommentResps,
		PageInfo: &response.PageInfo{
			Size:    len(CommentResps),
			Total:   count,
			Current: req.PageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(req.PageSize))),
		}}
	return resp, nil
}

func (s *courseCommentService) ListLabCommentByLabId(req *model.ListLabCommentReq) (*model.LabCommentEntityPageResp, error) {
	var comments = make([]*model.LabCommentEntity, 0)
	// 分页，排序
	if err := dao.LabComment.Page(req.PageCurrent, req.PageSize).Order(dao.LabComment.Columns.CreatedAt+" desc").
		// 主评，对应课程
		Where(dao.LabComment.Columns.Pid, 0).And(dao.LabComment.Columns.LabId, req.LabId).
		FieldsEx(dao.LabComment.Columns.DeletedAt).
		ScanList(&comments, "Comment"); err != nil {
		return nil, err
	}

	// 查子评论
	if err := dao.LabComment.
		Where(dao.LabComment.Columns.Pid,
			gdb.ListItemValuesUnique(comments, "Comment", "CourseCommentId")).
		FieldsEx(dao.LabComment.Columns.DeletedAt).
		ScanList(&comments, "ReplyComments", "Comment", "pid:CourseCommentId"); err != nil {
		return nil, err
	}
	for _, v := range comments {
		if v.ReplyComments == nil {
			v.ReplyComments = make([]*model.LabCommentResp, 0)
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
	resp := &model.LabCommentEntityPageResp{
		Records: comments,
		PageInfo: &response.PageInfo{
			Size:    len(comments),
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
	var courseComment *model.CourseComment
	if err = gconv.Struct(req, &courseComment); err != nil {
		return err
	}
	if courseComment.ReplyId != 0 {
		// 回复别人的评论,要做处理
		tmp, err := dao.CourseComment.WherePri(courseComment.ReplyId).
			Fields(dao.CourseComment.Columns.Pid, dao.CourseComment.Columns.Username).FindOne()
		if err != nil {
			return err
		}
		// 回复的是主评,则pid为主评id
		if tmp.Pid == 0 {
			courseComment.Pid = courseComment.ReplyId
		} else {
			// 否则，pid与被回复的pid保持一致
			courseComment.Pid = tmp.Pid
		}

		// 被回复的那条评论的用户昵称
		courseComment.ReplyUsername = tmp.Username
	}
	// 回复评论的用户昵称
	if one, err := dao.SysUser.WherePri(courseComment.UserId).
		Fields(dao.SysUser.Columns.NickName,
			dao.SysUser.Columns.AvatarUrl).
		FindOne(); err != nil {
		return err
	} else if one == nil {
		return code.UserNotExistError
	} else {
		courseComment.Username = one.NickName
		courseComment.UserAvatarUrl = one.AvatarUrl
	}

	// 保存
	if _, err = dao.CourseComment.Insert(courseComment); err != nil {
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
	var labComment *model.LabComment
	if err = gconv.Struct(req, &labComment); err != nil {
		return err
	}
	if labComment.ReplyId != 0 {
		// 回复别人的评论,要做处理
		replyComment, err := dao.LabComment.WherePri(labComment.ReplyId).
			Fields(dao.LabComment.Columns.Pid, dao.LabComment.Columns.Username).FindOne()
		if err != nil {
			return err
		}
		// 回复的是主评,则pid为主评id
		if replyComment.Pid == 0 {
			labComment.Pid = labComment.ReplyId
		} else {
			// 否则，pid与被回复的pid保持一致
			labComment.Pid = replyComment.Pid
		}
		// 被回复的那条评论的用户昵称
		labComment.ReplyUsername = replyComment.Username
	}
	// 回复评论的用户昵称
	if one, err := dao.SysUser.WherePri(labComment.UserId).
		Fields(dao.SysUser.Columns.NickName,
			dao.SysUser.Columns.AvatarUrl).
		FindOne(); err != nil {
		return err
	} else if one == nil {
		return code.UserNotExistError
	} else {
		labComment.Username = one.NickName
		labComment.UserAvatarUrl = one.AvatarUrl
	}
	// 保存
	if _, err = dao.LabComment.Insert(labComment); err != nil {
		return err
	}
	return nil
}

func (s *courseCommentService) DeleteLabComment(commentId int, userId int) (err error) {
	if _, err = dao.CourseComment.Where(g.Map{
		dao.LabComment.Columns.LabId:  commentId,
		dao.LabComment.Columns.UserId: userId,
	}).Save(dao.LabComment.Columns.CommentText, "该评论已删除"); err != nil {
		return err
	}
	return nil
}

func (s *courseCommentService) DeleteCourseComment(commentId int, userId int) (err error) {
	if _, err = dao.CourseComment.Where(g.Map{
		dao.CourseComment.Columns.CourseId: commentId,
		dao.CourseComment.Columns.UserId:   userId,
	}).Save(dao.CourseComment.Columns.CommentText, "该评论已删除"); err != nil {
		return err
	}
	return nil
}
