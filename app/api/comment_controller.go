// @Author: 陈健航
// @Date: 2021/1/16 0:32
// @Description:
package api

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service"
	"code-platform/app/service/component"
	"code-platform/library/common/response"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
)

var CourseCommentController = new(courseCommentController)

type courseCommentController struct{}

// InsertCourseComment 新增课程评论
// @receiver c
// @params r
// @date 2021-01-16 20:38:52
func (c *courseCommentController) InsertCourseComment(r *ghttp.Request) {
	var req *model.CourseComment
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if err := service.CourseCommentService.InsertCourseComment(req); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}

// InsertLabComment 新增课程评论
// @receiver c
// @params r
// @date 2021-01-16 20:38:52
func (c *courseCommentController) InsertLabComment(r *ghttp.Request) {
	var req *model.LabComment
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if err := service.CourseCommentService.InsertLabComment(req); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}

// ListCourseComment 分页查询课程评论
// @receiver c
// @params r
// @date 2021-01-30 21:43:14
func (c *courseCommentController) ListCourseComment(r *ghttp.Request) {
	pageCurrent, pageSize := response.GetPageReq(r)
	courseId := r.GetInt("courseId")
	resp, err := service.CourseCommentService.ListCourseCommentByCourseId(pageCurrent, pageSize, courseId)
	if err != nil {
		response.Exit(r, err)
	}
	response.Success(r, resp)
}

// ListLabComment 分页查询实验评论
// @receiver c
// @params r
// @date 2021-01-30 21:43:14
func (c *courseCommentController) ListLabComment(r *ghttp.Request) {
	pageCurrent, pageSize := response.GetPageReq(r)
	courseId := r.GetInt("labId")
	resp, err := service.CourseCommentService.ListLabCommentByLabId(pageCurrent, pageSize, courseId)
	if err != nil {
		response.Exit(r, err)
	}
	response.Success(r, resp)
}

// DeleteCourseComment 删除课程评论
// @receiver c
// @params r
// @date 2021-01-16 00:42:58
func (c *courseCommentController) DeleteCourseComment(r *ghttp.Request) {
	userId := component.GetUserId(r)
	commentId := r.GetInt("commentId")
	if _, err := dao.CourseComment.Where(g.Map{
		dao.CourseComment.Columns.CourseId: commentId,
		dao.CourseComment.Columns.UserId:   userId,
	}).Save(dao.CourseComment.Columns.CommentText, "该评论已删除"); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}

// DeleteLabComment 删除实验评论
// @receiver c
// @params r
// @date 2021-01-16 00:42:58
func (c *courseCommentController) DeleteLabComment(r *ghttp.Request) {
	userId := component.GetUserId(r)
	commentId := r.GetInt("commentId")
	if _, err := dao.CourseComment.Where(g.Map{
		dao.LabComment.Columns.LabId:  commentId,
		dao.LabComment.Columns.UserId: userId,
	}).Save(dao.LabComment.Columns.CommentText, "该评论已删除"); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}
