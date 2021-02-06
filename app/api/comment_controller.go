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
	"github.com/gogf/gf/net/ghttp"
)

var CourseCommentController = new(courseCommentController)

type courseCommentController struct{}

// DeleteCourseCommentById 删除评论
// @receiver c
// @params r
// @date 2021-01-16 00:42:58
func (c *labController) DeleteCourseCommentById(r *ghttp.Request) {
	userId := component.GetUserId(r)
	commentId := r.GetInt("CommentId")
	if _, err := dao.CourseComment.
		Where(dao.CourseComment.Columns.UserId, userId).And(dao.CourseComment.Columns.CourseCommentId, commentId).
		Delete(); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}

// InsertCourseComment 新增评论
// @receiver c
// @params r
// @date 2021-01-16 20:38:52
func (c *labController) InsertCourseComment(r *ghttp.Request) {
	var req *model.LabComment
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	if err := service.CourseCommentService.InsertLabComment(req); err != nil {
		response.Exit(r, err)
	}
	response.Success(r, true)
}

// ListCourseCommentPageByCourseId 分页查询课程评论
// @receiver c
// @params r
// @date 2021-01-30 21:43:14
func (c *labController) ListCourseCommentPageByCourseId(r *ghttp.Request) {
	pageCurrent, pageSize := response.GetPageReq(r)
	courseId := r.GetInt("courseId")
	resp, err := service.CourseCommentService.ListCourseCommentPageByCourseId(pageCurrent, pageSize, courseId)
	if err != nil {
		response.Exit(r, err)
	}
	response.Success(r, resp)
}
