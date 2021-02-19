// @Author: 陈健航
// @Date: 2021/1/12 16:39
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/response"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/net/ghttp"
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
func (s *courseService) ListCourseByTeacherId(pageCurrent, pageSize, teacherId int) (*model.CoursePageResp, error) {
	// 按创建时间降序
	d := dao.Course.Page(pageCurrent, pageSize).
		// 创建时间降序
		Order(dao.Course.Columns.CreatedAt+" desc").
		// 剔除不必要字段
		FieldsEx(dao.LabComment.Columns.DeletedAt,
			dao.Course.Columns.CourseDes,
			dao.Course.Columns.TeacherName,
			dao.Course.Columns.TeacherId).
		Where(dao.Course.Columns.TeacherId, teacherId)
	all, err := d.FindAll()
	if err != nil {
		return nil, err
	}
	if all == nil {
		all = make([]*model.Course, 0)
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
			Current: pageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(pageSize))),
		}}
	return resp, nil
}

// ListCourseByStuId 根据学生id获取该学生修读的课程信息
// @receiver s
// @params current
// @params size
// @params teacherId
// @return *model.CoursePageResp
// @return error
// @date 2021-01-15 19:46:37
func (s *courseService) ListCourseByStuId(pageCurrent, pageSize, stuId int) (*model.CoursePageResp, error) {
	courseIds, err := dao.ReCourseUser.ListCourseIdByStuId(pageCurrent, pageSize, stuId)
	if err != nil {
		return nil, err
	}
	// 按创建时间降序
	d := dao.Course.Order(dao.Course.Columns.CreatedAt).
		// 剔除不必要字段
		FieldsEx(dao.Course.Columns.DeletedAt, dao.Course.Columns.CourseDes, dao.Course.Columns.SecretKey).
		Where(dao.Course.Columns.CourseId, courseIds)
	all, err := d.FindAll()
	if err != nil {
		return nil, err
	}
	if all == nil {
		all = make([]*model.Course, 0)
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
			Current: pageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(pageSize))),
		}}
	return resp, nil
}

// StuJoinCourse 加入课程
// @receiver s
// @params req
// @date 2021-01-15 19:19:07
func (s *courseService) StuJoinCourse(userId, CourseId, SecretKey int) error {
	// 检查学生是否已经先完善学号，真实姓名等信息
	one, err := dao.SysUser.WherePri(userId).Fields(dao.SysUser.Columns.RealName, dao.SysUser.Columns.Num).FindOne()
	if err != nil {
		return err
	}
	if one.Num == "" || one.RealName == "" {
		return gerror.NewCode(10000, "请先至少完善学号以及真实姓名信息")
	}
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

// ListStuPageByCourseId 根据课程id分页查询所有修读该课程的学生信息
// @receiver s
// @params courseId
// @params pageCurrent
// @params pageSize
// @return *model.SysUserPageResp
// @return error
// @date 2021-02-06 23:16:34
func (s *courseService) ListStuPageByCourseId(courseId int, pageCurrent int, pageSize int) (*model.SysUserPageResp, error) {
	// 查出所有修读该门课程的学生学号
	userIds, err := dao.ReCourseUser.ListUserIdByCourseId(pageCurrent, pageSize, courseId)
	if err != nil {
		return nil, err
	}

	// 查出所有学生的具体信息
	d := dao.SysUser.Order(dao.SysUser.Columns.Num).WherePri(userIds).
		FieldsEx(dao.SysUser.Columns.Password, dao.SysUser.Columns.DeletedAt)
	all, err := d.FindAll()
	if err != nil {
		return nil, err
	}
	if all == nil {
		all = make([]*model.SysUser, 0)
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
			Current: pageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(pageSize))),
		}}
	return resp, nil
}

// DisbandCourseByCourseId 解散课程
// @receiver s
// @params teacherId
// @params courseId
// @return error
// @date 2021-02-06 23:55:17
func (s *courseService) DisbandCourseByCourseId(teacherId int, courseId int, secretKey int) (err error) {
	// 删除该门课程的选课记录
	if err = dao.ReCourseUser.DeleteByCourseId(courseId); err != nil {
		return err
	}

	// 删除课程信息
	if _, err = dao.Course.WherePri(courseId).
		And(dao.Course.Columns.TeacherId, teacherId).And(dao.Course.Columns.SecretKey, secretKey).
		Delete(); err != nil {
		return err
	}
	return nil
}

func (s *courseService) Update(req *model.Course, file *ghttp.UploadFile) (err error) {
	// pic存服务器，存url
	if file != nil {
		pic, err := FileService.UploadPic(file, 128)
		if err != nil {
			return err
		}
		req.PicUrl = pic
	}
	// 保存
	if _, err = dao.Course.Where(dao.Course.Columns.TeacherId, req.TeacherId).
		And(dao.Course.Columns.CourseId, req.CourseId).OmitEmpty().
		Save(req); err != nil {
		return err
	}
	return nil
}

func (s *courseService) InsertCourse(req *model.Course, file *ghttp.UploadFile) (err error) {
	// pic存服务器，存url
	if file != nil {
		pic, err := FileService.UploadPic(file, 128)
		if err != nil {
			return err
		}
		req.PicUrl = pic
	}
	// 获取教师名称
	teacherName, err := dao.SysUser.WherePri(req.TeacherId).FindValue(dao.SysUser.Columns.RealName)
	if err != nil {
		return err
	}
	req.TeacherName = teacherName.String()
	// 保存
	if _, err = dao.Course.Insert(req); err != nil {
		return err
	}
	return nil
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
