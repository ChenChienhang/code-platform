// @Author: 陈健航
// @Date: 2021/1/15 21:30
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/response"
	"github.com/gogf/gf/database/gdb"
	"math"
)

var CourseCommentService = new(courseCommentService)

type courseCommentService struct{}

func (s *courseCommentService) ListCourseCommentPageByCourseId(pageCurrent, pageSize, courseId int) (*model.CourseCommentEntityResp, error) {
	var comments []*model.CourseCommentEntity
	// 分页，排序
	if err := dao.CourseComment.Page(pageCurrent, pageSize).Order(dao.CourseComment.Columns.CreatedAt+" desc").
		// 主评，对应课程
		Where(dao.CourseComment.Columns.Pid, 0).And(dao.CourseComment.Columns.CourseId, courseId).
		FieldsEx(dao.CourseComment.Columns.DeletedAt).
		ScanList(&comments, "Comment"); err != nil {
		return nil, err
	}

	// 查子评论
	if err := dao.CourseComment.
		Where(dao.CourseComment.Columns.Pid,
			gdb.ListItemValuesUnique(comments, "Comment", "CourseCommentId")).
		Order(dao.CourseComment.Columns.CreatedAt+" desc").
		FieldsEx(dao.CourseComment.Columns.DeletedAt).
		ScanList(&comments, "ReplyComments", "Comment", "pid:CourseCommentId"); err != nil {
		return nil, err
	}

	for _, v := range comments {
		if v.ReplyComments == nil {
			v.ReplyComments = make([]*model.CourseComment, 0)
		}
	}

	// 分页信息
	count, err := dao.CourseComment.
		Where(dao.CourseComment.Columns.Pid, 0).
		Where(dao.CourseComment.Columns.CourseId, courseId).Count()
	if err != nil {
		return nil, err
	}

	// 分页信息整合
	resp := &model.CourseCommentEntityResp{
		Records: comments,
		PageInfo: &response.PageInfo{
			Size:    len(comments),
			Total:   count,
			Current: pageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(pageSize))),
		}}
	return resp, nil
}

func (s *courseCommentService) GetLabCommentPageByLabId(pageCurrent, pageSize, labId int) (*model.LabCommentEntityResp, error) {
	var comments []*model.LabCommentEntity
	// 分页，排序
	if err := dao.LabComment.Page(pageCurrent, pageSize).Order(dao.LabComment.Columns.CreatedAt+" desc").
		// 主评，对应课程
		Where(dao.LabComment.Columns.Pid, 0).And(dao.LabComment.Columns.LabId, labId).
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
			v.ReplyComments = make([]*model.LabComment, 0)
		}
	}

	// 分页信息
	count, err := dao.LabComment.
		Where(dao.LabComment.Columns.Pid, 0).
		Where(dao.LabComment.Columns.LabId, labId).Count()
	if err != nil {
		return nil, err
	}

	// 分页信息整合
	resp := &model.LabCommentEntityResp{
		Records: comments,
		PageInfo: &response.PageInfo{
			Size:    len(comments),
			Total:   count,
			Current: pageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(pageSize))),
		}}
	return resp, nil
}

// InsertCourseComment
// @receiver s
// @params saveModel
// @return error
// @date 2021-01-31 18:29:13
func (s *courseCommentService) InsertCourseComment(saveModel *model.CourseComment) error {
	if saveModel.ReplyId != 0 {
		// 回复别人的评论,要做处理
		tmp, err := dao.CourseComment.WherePri(saveModel.ReplyId).
			Fields(dao.CourseComment.Columns.Pid, dao.CourseComment.Columns.UserId).FindOne()
		if err != nil {
			return err
		}
		// 回复的是主评,则pid为主评id
		if tmp.Pid == 0 {
			saveModel.Pid = saveModel.ReplyId
		} else {
			// 否则，pid与被回复的pid保持一致
			saveModel.Pid = tmp.Pid
		}

		// 被回复的那条评论的用户昵称
		saveModel.ReplyUsername = tmp.Username
	}
	// 回复评论的用户昵称
	UserNickname, err := dao.SysUser.WherePri(saveModel.UserId).FindValue(dao.SysUser.Columns.NickName)
	if err != nil {
		return err
	}
	saveModel.Username = UserNickname.String()

	// 保存
	if _, err = dao.CourseComment.Insert(saveModel); err != nil {
		return err
	}
	return nil
}

// InsertLabComment
// @receiver s
// @params saveModel
// @return error
// @date 2021-01-31 18:29:22
func (s *courseCommentService) InsertLabComment(saveModel *model.LabComment) error {
	if saveModel.ReplyId != 0 {
		// 回复别人的评论,要做处理
		tmp, err := dao.LabComment.WherePri(saveModel.ReplyId).
			Fields(dao.LabComment.Columns.Pid, dao.LabComment.Columns.UserId).FindOne()
		if err != nil {
			return err
		}
		// 回复的是主评,则pid为主评id
		if tmp.Pid == 0 {
			saveModel.Pid = saveModel.ReplyId
		} else {
			// 否则，pid与被回复的pid保持一致
			saveModel.Pid = tmp.Pid
		}

		// 被回复的那条评论的用户昵称
		saveModel.ReplyUsername = tmp.Username
	}
	// 回复评论的用户昵称
	UserNickname, err := dao.SysUser.WherePri(saveModel.UserId).FindValue(dao.SysUser.Columns.NickName)
	if err != nil {
		return err
	}
	saveModel.Username = UserNickname.String()

	// 保存
	if _, err = dao.LabComment.Insert(saveModel); err != nil {
		return err
	}
	return nil
}
