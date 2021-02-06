// @Author: 陈健航
// @Date: 2021/1/12 16:39
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/response"
	"github.com/gogf/gf/errors/gerror"
	"math"
)

var CourseService = new(courseService)

type courseService struct{}

//goland:noinspection GoNilness
//func (s *courseService) FindStuNotInCourse(req *model.StuNotInCourseReq) (*model.StuNotInCourseResp, error) {
//	userIds, err := dao.ReCourseUser.ListUserIdByCourseId(req.CourseId)
//	if err != nil {
//		return nil, err
//	}
//	// 查出所有学生的真实姓名和学号比对
//	all, err := dao.SysUser.Fields(
//		dao.SysUser.Columns.RealName,
//		dao.SysUser.Columns.Num,
//	).FindAll(dao.SysUser.SysUserDao.Columns.UserId, userIds)
//	if err != nil {
//		return nil, err
//	}
//	var resp *model.StuNotInCourseResp
//	if err = gconv.Struct(req, resp); err != nil {
//		return nil, err
//	}
//	for i, v := range resp.StuNo {
//		for _, v2 := range all {
//			if v2.Num == v {
//				resp.IsStuNo[i] = true
//			}
//		}
//	}
//	for i, v := range resp.StuRealName {
//		for _, v2 := range all {
//			if v2.Num == v {
//				resp.IsStuNo[i] = true
//			}
//		}
//	}
//	return resp, nil
//}

// ListCourseByTeacherId 根据老师id获取该老师所开设的课程信息
// @receiver s
// @params current
// @params size
// @params teacherId
// @return *model.CoursePageResp
// @return error
// @date 2021-01-15 19:46:37
func (s *courseService) ListCourseByTeacherId(current, size, teacherId int) (*model.CoursePageResp, error) {
	// 按创建时间降序
	d := dao.Course.Page(current, size).Order(dao.Course.Columns.CreatedAt+" desc").
		FieldsEx(dao.LabComment.Columns.DeletedAt).Where(dao.Course.Columns.TeacherId, teacherId)
	all, err := d.FindAll()
	if err != nil {
		return nil, err
	}

	// 分页信息查询
	count, err := d.Count()
	if err != nil {
		return nil, err
	}

	// 分页信息整合
	resp := &model.CoursePageResp{
		Records: all,
		PageInfo: &response.PageInfo{
			Size:    len(all),
			Total:   count,
			Current: current,
			Pages:   int(math.Ceil(float64(count) / float64(size))),
		}}
	return resp, nil
}

// GetCourseByStudentId 根据学生id找出加入的课程
// @receiver s
// @params current
// @params size
// @params teacherId
// @return *model.CoursePageResp
// @return error
// @date 2021-01-15 19:46:37
func (s *courseService) GetCourseByStudentId(current, size, teacherId int) (*model.CoursePageResp, error) {
	// 按创建时间降序
	d := dao.Course.Page(current, size).Order(dao.Course.Columns.CreatedAt + " desc").
		FieldsEx(dao.LabComment.Columns.DeletedAt)
	all, err := d.Where(dao.Course.Columns.TeacherId, teacherId).FindAll()
	if err != nil {
		return nil, err
	}

	// 分页信息查询
	count, err := d.Count(dao.Course.Columns.TeacherId, teacherId)
	if err != nil {
		return nil, err
	}

	// 分页信息整合
	resp := &model.CoursePageResp{
		Records: all,
		PageInfo: &response.PageInfo{
			Size:    len(all),
			Total:   count,
			Current: current,
			Pages:   int(math.Ceil(float64(count) / float64(size))),
		}}
	return resp, nil
}

// JoinCourse 加入课程
// @receiver s
// @params req
// @date 2021-01-15 19:19:07
func (s *courseService) JoinCourse(userId, CourseId, SecretKey int) error {
	// 找出对应课程id
	secretKey, err := dao.Course.
		Where(dao.Course.Columns.CourseId, CourseId).
		And(dao.Course.Columns.Closed, 0).
		FindValue(dao.Course.Columns.SecretKey)
	if err != nil {
		return err
	}
	// 不存在该课程
	if secretKey.IsEmpty() {
		return gerror.NewCode(10000, "课程不存在")
	}
	// 选课密码错误
	if SecretKey != secretKey.Int() {
		return gerror.NewCode(10000, "课程密码有误")
	}
	// 插入选课记录
	if err = dao.ReCourseUser.Insert(userId, CourseId); err != nil {
		return err
	}
	return nil
}

func (s *courseService) ListStuPageByCourseId(courseId int, current int, size int) (*model.SysUserPageResp, error) {
	// 查出所有修读该门课程的学生学号
	userIds, err := dao.ReCourseUser.ListUserIdByCourseId(courseId)
	if err != nil {
		return nil, err
	}

	// 查出所有学生的具体信息
	d := dao.SysUser.Page(current, size).Order(dao.SysUser.Columns.Num).WherePri(dao.SysUser.Columns.UserId, userIds).
		FieldsEx(dao.SysUser.Columns.Password, dao.SysUser.Columns.DeletedAt)
	all, err := d.FindAll()
	if err != nil {
		return nil, err
	}

	// 分页信息整合
	count, err := d.Count()
	if err != nil {
		return nil, err
	}
	resp := &model.SysUserPageResp{
		Records: all,
		PageInfo: &response.PageInfo{
			Size:    len(all),
			Total:   count,
			Current: current,
			Pages:   int(math.Ceil(float64(count) / float64(size))),
		}}
	return resp, nil
}

//func (s *courseService) name(pageCurrent, pageSize, courseId int) error {
//	// 查出所有学生学号
//	userIds, err := g.DB().Table("re_course_user").FindArray("user_id", "course_id", courseId)
//	if err != nil {
//		return err
//	}
//
//	// 查出所有学生的具体信息
//	var resp model.SysUserPageResp
//	d := dao.SysUser.Page(pageCurrent, pageSize).Order(dao.SysUser.Columns.Num)
//	if err = d.FindScan(&resp, dao.SysUser.SysUserDao.Columns.UserId, userIds); err != nil {
//		response.Exit(r, err)
//	}
//
//	// 分页信息整合
//	count, err := d.Count(dao.SysUser.SysUserDao.Columns.UserId, userIds)
//	if err != nil {
//		response.Exit(r, err)
//	}
//	resp.PageInfo.Size = len(resp.Records)
//	resp.PageInfo.Total = count
//	resp.PageInfo.Current = pageCurrent
//	resp.PageInfo.Current = int(math.Ceil(float64(count) / float64(pageSize)))
//	response.Success(r, resp)
//}
