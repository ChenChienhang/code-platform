// @Author: 陈健航
// @Date: 2021/3/1 0:36
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/code"
	"code-platform/library/common/response"
	"math"
)

var CourseResourceService = new(courseResourceService)

type courseResourceService struct{}

func (s *courseResourceService) Insert(req *model.InsertCourseRecourseReq) (err error) {
	if _, err = dao.CourseRecourse.Insert(req); err != nil {
		return err
	}
	// 删除脏文件
	if req.AttachmentUrl != nil {
		go FileService.RemoveDirtyFile(*req.AttachmentUrl)
	}
	return nil
}

func (s *courseResourceService) Update(req *model.UpdateCourseRecourseReq) (err error) {
	removeFlag := false
	if req.AttachmentUrl != nil {
		defer func(flag *bool) {
			if *flag {
				go FileService.RemoveDirtyFile(*req.AttachmentUrl)
				//goland:noinspection GoUnhandledErrorResult
				go FileService.RemoveObject(*req.AttachmentUrl)
			}
		}(&removeFlag)
	}
	if _, err = dao.CourseRecourse.Update(req); err != nil {
		removeFlag = false
		return err
	}
	return nil
}

func (s *courseResourceService) List(req *model.ListCourseResourceReq) (resp *response.PageResp, err error) {
	d := dao.CourseRecourse.Where(dao.CourseRecourse.Columns.CourseId, req.CourseId).
		FieldsEx(dao.CourseRecourse.Columns.DeletedAt, dao.CourseRecourse.Columns.CourseId, dao.CourseRecourse.Columns.AttachmentUrl).
		Order(dao.CourseRecourse.Columns.CreatedAt + " desc")
	records := make([]*model.CourseRecourseResp, 0)
	if err = d.Page(req.PageCurrent, req.PageSize).Scan(&records); err != nil {
		return nil, err
	}
	count, err := d.Count()
	if err != nil {
		return nil, err
	}

	resp = &response.PageResp{
		Records: records,
		PageInfo: &response.PageInfo{
			Size:    len(records),
			Total:   count,
			Current: req.PageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(req.PageSize))),
		},
	}
	return resp, nil
}

func (s *courseResourceService) Delete(userId int, courseRecourseId int) (err error) {
	courseId, err := dao.CourseRecourse.WherePri(courseRecourseId).FindValue(dao.CourseRecourse.Columns.CourseId)
	if err != nil {
		return err
	}
	teacherId, err := dao.Course.WherePri(courseId).FindValue(dao.Course.Columns.TeacherId)
	if err != nil {
		return err
	}
	if teacherId.Int() != userId {
		return code.AuthError
	}
	if _, err = dao.CourseRecourse.WherePri(courseRecourseId).Delete(); err != nil {
		return err
	}
	return nil
}

func (s *courseResourceService) GetOne(courseRecourseId int) (resp *model.CourseRecourseResp, err error) {
	resp = new(model.CourseRecourseResp)
	if err = dao.CourseRecourse.WherePri(courseRecourseId).
		FieldsEx(dao.CourseRecourse.Columns.DeletedAt, dao.CourseRecourse.Columns.CourseId).Scan(&resp); err != nil {
		return nil, err
	}
	return resp, nil
}
