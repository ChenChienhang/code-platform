// @Author: 陈健航
// @Date: 2021/2/10 22:34
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/code"
	"code-platform/library/common/response"
	"math"
)

var LabService = new(labService)

type labService struct{}

// Insert 新建实验
// @receiver s
// @params req
// @params attachment
// @return err
// @date 2021-02-20 23:52:17
func (s *labService) Insert(req *model.InsertLabReq) (err error) {
	if req.AttachmentUrl != nil {
		go func() {
			FileService.RemoveDirtyFile(*req.AttachmentUrl)
		}()
	}
	return nil
}

// ListByCourseId 分页列表实验
// @receiver s
// @params pageCurrent
// @params pageSize
// @params courseId
// @return resp
// @return err
// @date 2021-02-20 23:52:31
func (s *labService) ListByCourseId(req *model.ListLabByCourseIdReq) (resp *response.PageResp, err error) {
	// 查找lab信息
	d := dao.Lab.Where(dao.Lab.Columns.CourseId, req.CourseId).FieldsEx(dao.Lab.Columns.DeletedAt, dao.Lab.Columns.AttachmentUrl)

	records := make([]*model.LabResp, 0)
	if err = d.Order(dao.Lab.Columns.CreatedAt).Page(req.PageCurrent, req.PageSize).Scan(&records); err != nil {
		return nil, err
	}

	// 分页信息整合
	count, err := d.Count()
	if err != nil {
		return nil, err
	}

	labSubmits := make([]*model.LabSubmit, 0)
	if err = dao.LabSubmit.Where(dao.LabSubmit.Columns.UserId, req.UserId).
		FieldsEx(dao.LabSubmit.Columns.DeletedAt).Scan(&labSubmits); err != nil {
		return nil, err
	}
	// 查一下每一个实验是否完成
	for _, v := range records {
		for _, v1 := range labSubmits {
			if v.LabId == v1.LabId && v1.IsFinish == 1 {
				v.IsFinish = true
				break
			}
		}
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

func (s *labService) Update(req *model.UpdateLabReq) (err error) {
	removeFlag := false
	// 保存旧附件
	if req.AttachmentUrl != nil {
		if oldAttachmentUrl, err := dao.Lab.WherePri(req.LabId).FindValue(); err != nil {
			return err
		} else if !oldAttachmentUrl.IsEmpty() {
			// 删除旧附件
			removeFlag = true
			defer func(flag *bool) {
				if *flag {
					//goland:noinspection GoUnhandledErrorResult
					go FileService.RemoveObject(oldAttachmentUrl.String())
					go FileService.RemoveDirtyFile(*req.AttachmentUrl)
				}
			}(&removeFlag)
		}
	}
	if _, err = dao.Lab.OmitEmpty().Update(req); err != nil {
		removeFlag = false
		return err
	}
	return nil
}

func (s *labService) GetOne(labId int) (resp *model.LabResp, err error) {
	resp = new(model.LabResp)
	if err = dao.Lab.WherePri(labId).FieldsEx(dao.Lab.Columns.DeletedAt).Scan(&resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *labService) Delete(userId int, labId int) (err error) {
	courseId, err := dao.Lab.WherePri(labId).FindValue(dao.Lab.Columns.CourseId)
	if err != nil {
		return err
	}
	teacherId, err := dao.Course.WherePri(courseId.String()).FindValue(dao.Course.Columns.TeacherId)
	if err != nil {
		return err
	}
	if userId != teacherId.Int() {
		return code.AuthError
	}
	if _, err = dao.Lab.WherePri(labId).Delete(); err != nil {
		return err
	}
	return nil
}

func (s *labService) ListByToken(req *model.ListLabByTokenReq) (resp *response.PageResp, err error) {
	records := make([]*model.LabResp, 0)
	// 找出所有参加的课程
	courseIds, err := dao.Course.ListCourseIdByStuId(req.UserId)
	d := dao.Lab.Where(dao.Lab.Columns.CourseId, courseIds)
	// 找出这些课程所有的实验，按时间降序
	if err = d.Order(dao.Lab.Columns.CreatedAt+" desc").Page(req.PageCurrent, req.PageSize).
		FieldsEx(dao.Lab.Columns.DeletedAt).Scan(&records); err != nil {
		return nil, err
	}
	count, err := d.Count()

	// 查一下每一个实验是否完成
	labSubmits := make([]*model.LabSubmit, 0)
	if err = dao.LabSubmit.Where(dao.LabSubmit.Columns.UserId, req.UserId).
		FieldsEx(dao.LabSubmit.Columns.DeletedAt).Scan(&labSubmits); err != nil {
		return nil, err
	}
	for _, v := range records {
		for _, v1 := range labSubmits {
			if v.LabId == v1.LabId && v1.IsFinish == 1 {
				v.IsFinish = true
				break
			}
		}
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
