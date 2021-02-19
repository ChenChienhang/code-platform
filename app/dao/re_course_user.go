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

// ListCourseIdByStuId 根据学生id分页查出所选课程的id
// @receiver d
// @params pageCurrent
// @params pageSize
// @params courseId
// @return []gdb.Value
// @return error
// @date 2021-02-06 23:03:03
func (d *reCourseUserDao) ListCourseIdByStuId(pageCurrent, pageSize, courseId int) ([]gdb.Value, error) {
	all, err := g.Table("re_course_user").
		// 分页
		Page(pageCurrent, pageSize).
		// 创建时间倒序
		Where("user_id", courseId).FindArray("course_id")
	return all, err
}

// ListUserIdByCourseId 根据课程id分页查询userId
// @receiver d
// @params pageCurrent
// @params pageSize
// @params courseId
// @return []gdb.Value
// @return error
// @date 2021-02-06 23:22:43
func (d *reCourseUserDao) ListUserIdByCourseId(pageCurrent, pageSize, courseId int) ([]gdb.Value, error) {
	all, err := g.Table("re_course_user").Page(pageCurrent, pageSize).
		Where("course_id", courseId).FindArray("user_id")
	if err != nil {
		return nil, err
	}
	return all, err
}

// Insert 插入选课记录
// @receiver d
// @params userId
// @params courseId
// @return error
// @date 2021-02-06 23:22:34
func (d *reCourseUserDao) Insert(stuId, courseId int) error {
	if _, err := g.Table("re_course_user").Insert(g.Map{
		"user_id":   stuId,
		"course_id": courseId,
	}); err != nil {
		return err
	}
	return nil
}

// DeleteByUserIdAndCourseId 删除一个学生的某条选课记录
// @receiver d
// @params userId
// @params courseId
// @return error
// @date 2021-02-06 22:39:39
func (d *reCourseUserDao) DeleteByUserIdAndCourseId(userId, courseId int) error {
	if _, err := g.Table("re_course_user").Where(g.Map{
		"user_id":   userId,
		"course_id": courseId,
	}).Delete(); err != nil {
		return err
	}
	return nil
}

// DeleteByCourseId 删除一门课的全部选课记录
// @receiver d
// @params userId
// @params courseId
// @return error
// @date 2021-02-06 22:39:39
func (d *reCourseUserDao) DeleteByCourseId(courseId int) error {
	if _, err := g.Table("re_course_user").Where("course_id", courseId).Delete(); err != nil {
		return err
	}
	return nil
}
