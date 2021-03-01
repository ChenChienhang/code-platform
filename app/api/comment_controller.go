// @Author: 陈健航
// @Date: 2021/1/16 0:32
// @Description:
package api

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service"
	"code-platform/library/common/response"
	"github.com/gogf/gf/net/ghttp"
)

var CommentController = new(commentController)

type commentController struct{}

// InsertCourseComment 新增课程评论
// @receiver c
// @params r
// @date 2021-01-16 20:38:52
func (c *commentController) InsertCourseComment(r *ghttp.Request) {
	var req *model.InsertCourseCommentReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if err := service.CourseCommentService.InsertCourseComment(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

// InsertLabComment 新增课程评论
// @receiver c
// @params r
// @date 2021-01-16 20:38:52
func (c *commentController) InsertLabComment(r *ghttp.Request) {
	var req *model.InsertLabCommentReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if err := service.CourseCommentService.InsertLabComment(req); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

// ListCourseComment 分页查询课程评论
// @receiver c
// @params r
// @date 2021-01-30 21:43:14
func (c *commentController) ListCourseComment(r *ghttp.Request) {
	var req *model.ListCourseCommentReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	resp, err := service.CourseCommentService.ListCourseCommentByCourseId(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

// ListLabComment 分页查询实验评论
// @receiver c
// @params r
// @date 2021-01-30 21:43:14
func (c *commentController) ListLabComment(r *ghttp.Request) {
	var req *model.ListLabCommentReq
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	resp, err := service.CourseCommentService.ListLabCommentByLabId(req)
	if err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, resp)
}

// DeleteCourseComment 删除课程评论,伪删除
// @receiver c
// @params r
// @date 2021-01-16 00:42:58
func (c *commentController) DeleteCourseComment(r *ghttp.Request) {
	userId := r.GetVar(dao.SysUser.Columns.UserId).Int()
	commentId := r.GetInt("commentId")
	if err := service.CourseCommentService.DeleteCourseComment(commentId, userId); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}

// DeleteLabComment 删除实验评论,伪删除
// @receiver c
// @params r
// @date 2021-01-16 00:42:58
func (c *commentController) DeleteLabComment(r *ghttp.Request) {
	userId := r.GetVar(dao.SysUser.Columns.UserId).Int()
	commentId := r.GetInt("commentId")
	if err := service.CourseCommentService.DeleteLabComment(commentId, userId); err != nil {
		response.Exit(r, err)
	}
	response.Succ(r, true)
}
