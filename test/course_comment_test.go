// @Author: 陈健航
// @Date: 2021/1/31 11:27
// @Description:
package test

import (
	"code-platform/app/model"
	"code-platform/app/service"
	"testing"
)

func TestGetCourseCommentPageByCourseId(t *testing.T) {

}

func TestInsertCourseComment(t *testing.T) {
	course := &model.CourseComment{ReplyId: 3, UserId: 2, CourseId: 1, CommentText: "Awe"}
	if err := service.CourseCommentService.InsertCourseComment(course); err != nil {
		return
	}

}
