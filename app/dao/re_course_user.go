// @Author: 陈健航
// @Date: 2021/1/15 18:06
// @Description:
package dao

import (
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/frame/g"
)

var ReCourseUser = new(reCourseUserDao)

type reCourseUserDao struct{}

func (d *reCourseUserDao) ListUserIdByCourseId(courseId int) ([]gdb.Value, error) {
	all, err := g.Table("re_course_user").Where("course_id", courseId).FindArray("user_id")
	if err != nil {
		return nil, err
	}
	return all, err
}

func (d *reCourseUserDao) Insert(userId, courseId int) error {
	if _, err := g.Table("re_course_user").Insert(g.Map{
		"user_id":   userId,
		"course_id": courseId,
	}); err != nil {
		return err
	}
	return nil
}

func (d *reCourseUserDao) name() {

}
