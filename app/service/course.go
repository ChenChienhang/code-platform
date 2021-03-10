// @Author: 陈健航
// @Date: 2021/1/12 16:39
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/code"
	"code-platform/library/common/response"
	"math"
)

var CourseService = new(courseService)

type courseService struct{}

// ListCourseByTeacherId 根据老师id获取该老师所开设的课程信息
// @receiver s
// @params current
// @params size
// @params teacherId
// @return *model.CoursePageResp
// @return error
// @date 2021-01-15 19:46:37
func (s *courseService) ListCourseByTeacherId(req *model.ListByTeacherIdReq) (resp *response.PageResp, err error) {
	// 按创建时间降序
	d := dao.Course.
		// 创建时间降序
		Order(dao.Course.Columns.CreatedAt+" desc").
		// 剔除不必要字段
		FieldsEx(dao.LabComment.Columns.DeletedAt,
			dao.Course.Columns.CourseDes,
			dao.Course.Columns.TeacherName,
			dao.Course.Columns.TeacherId).
		Where(dao.Course.Columns.TeacherId, req.TeacherId)
	records := make([]*model.CourseResp, 0)
	if err = d.Page(req.PageCurrent, req.PageSize).Scan(&records); err != nil {
		return nil, err
	}

	// 分页信息查询
	count, err := d.Count()
	if err != nil {
		return nil, err
	}

	// 分页信息整合
	resp = &response.PageResp{
		Records: records,
		PageInfo: &response.PageInfo{
			Size:    len(records),
			Total:   count,
			Current: req.PageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(req.PageCurrent))),
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
func (s *courseService) ListCourseByStuId(req *model.ListCourseByStuIdReq) (resp *response.PageResp, err error) {
	courseIds, count, err := dao.Course.ListCourseIdByStuId(req.PageCurrent, req.PageSize, req.StudentId)
	if err != nil {
		return nil, err
	}
	// 按创建时间降序
	records := make([]*model.CourseResp, 0)
	if err = dao.Course.Order(dao.Course.Columns.CreatedAt+" desc").
		// 剔除不必要字段
		FieldsEx(dao.Course.Columns.DeletedAt, dao.Course.Columns.CourseDes, dao.Course.Columns.SecretKey).
		Where(dao.Course.Columns.CourseId, courseIds).Scan(&records); err != nil {
		return nil, err
	}

	// 分页信息整合
	resp = &response.PageResp{
		Records: records,
		PageInfo: &response.PageInfo{
			Size:    len(records),
			Total:   count,
			Current: req.PageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(req.PageSize))),
		}}
	return resp, nil
}

// AttendCourse 加入课程
// @receiver s
// @params req
// @date 2021-01-15 19:19:07
func (s *courseService) AttendCourse(req *model.AttendCourseReq) (err error) {
	// 检查学生是否已经先完善学号，真实姓名等信息
	one, err := dao.SysUser.WherePri(req.StudentId).Fields(dao.SysUser.Columns.RealName, dao.SysUser.Columns.Num).FindOne()
	if err != nil {
		return err
	}
	if one == nil {
		return code.UserNotExistError
	}
	if one.Num == "" || one.RealName == "" {
		return code.InfoNotCompleteError
	}
	// 找出对应课程id
	secretKey, err := dao.Course.
		Where(dao.Course.Columns.CourseId, req.CourseId).
		And(dao.Course.Columns.IsClose, 0).
		FindValue(dao.Course.Columns.SecretKey)
	if err != nil {
		return err
	}
	// 不存在该课程
	if secretKey.IsEmpty() {
		return code.CourseNotExitError
	}
	// 选课密码错误
	if req.SecretKey != secretKey.Int() {
		return code.CourseKeyError
	}
	// 插入选课记录
	if err = dao.Course.AttendCourse(req.StudentId, req.CourseId); err != nil {
		return err
	}
	return nil
}

// ListStuByCourseId 根据课程id分页查询所有修读该课程的学生信息
// @receiver s
// @params courseId
// @params pageCurrent
// @params pageSize
// @return *model.SysUserPageResp
// @return error
// @date 2021-02-06 23:16:34
func (s *courseService) ListStuByCourseId(req *model.ListStuByCourseIdReq) (resp *response.PageResp, err error) {
	// 查出所有修读该门课程的学生学号
	userIds, count, err := dao.Course.ListUserIdByCourseId(req.PageCurrent, req.PageSize, req.CourseId)
	if err != nil {
		return nil, err
	}

	// 查出所有学生的具体信息
	d := dao.SysUser.Order(dao.SysUser.Columns.Num).WherePri(userIds).
		FieldsEx(dao.SysUser.Columns.Password, dao.SysUser.Columns.DeletedAt)
	records := make([]*model.SysUserResp, 0)
	if err := d.Scan(&records); err != nil {
		return nil, err
	}

	resp = &response.PageResp{
		Records: records,
		PageInfo: &response.PageInfo{
			Size:    len(records),
			Total:   count,
			Current: req.PageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(req.PageSize))),
		}}
	return resp, nil
}

func (s *courseService) Update(req *model.UpdateCourseReq) (err error) {
	removeFlag := false
	// 封面发生过变化
	if req.PicUrl != nil {
		//查出旧封面
		if oldPicUrl, err := dao.Course.WherePri(req.CourseId).FindValue(dao.Course.Columns.PicUrl); err != nil {
			return err
		} else if !oldPicUrl.IsEmpty() {
			removeFlag = true
			defer func(flag *bool) {
				//goland:noinspection GoUnhandledErrorResult
				if *flag {
					go FileService.RemoveDirtyFile(*req.PicUrl)
					go FileService.RemoveObject(oldPicUrl.String())
				}
			}(&removeFlag)
		}
	}

	// 保存
	if _, err = dao.Course.WherePri(req.CourseId).OmitEmpty().Save(req); err != nil {
		// 保存不成功则不删除旧的封面
		removeFlag = false
		return err
	}

	return nil
}

func (s *courseService) InsertCourse(req *model.InsertCourseReq) (err error) {
	// 获取教师名称
	if teacherName, err := dao.SysUser.WherePri(req.TeacherId).FindValue(dao.SysUser.Columns.RealName); err != nil {
		return err
	} else {
		req.TeacherName = teacherName.String()
	}
	// 保存
	if _, err = dao.Course.OmitEmpty().Insert(req); err != nil {
		return err
	}
	if req.PicUrl != nil {
		go FileService.RemoveDirtyFile(*req.PicUrl)
	}
	return nil
}

func (s *courseService) SearchList(req *model.SearchListByCourseNameReq) (resp *response.PageResp, err error) {
	// 模糊搜索
	d := dao.Course.Where(dao.Course.Columns.CourseName+" like ?", "%"+req.CourseName+"%").
		// 字段剔除
		FieldsEx(
			dao.Course.Columns.DeletedAt,
			dao.Course.Columns.SecretKey,
		)
	records := make([]*model.CourseResp, 0)
	if err = d.Page(req.PageCurrent, req.PageSize).Scan(&records); err != nil {
		return nil, err
	}
	count, err := d.Count()
	if err != nil {

	}
	// 分页信息整合
	resp = &response.PageResp{
		Records: records,
		PageInfo: &response.PageInfo{
			Size:    len(records),
			Total:   count,
			Current: req.PageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(req.PageSize))),
		}}
	return resp, nil
}

func (s *courseService) Delete(teacherId int, courseId int) (err error) {
	// 删除课程信息,不递归删除课程下的资源
	if _, err = dao.Course.WherePri(courseId).And(dao.Course.Columns.TeacherId, teacherId).Delete(); err != nil {
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
//	response.Succ(r, resp)
//}
