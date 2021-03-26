// @Author: 陈健航
// @Date: 2021/3/16 13:48
// @Description:
package admin

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/response"
	"math"
)

var CourseAdminService = new(courseAdminService)

type courseAdminService struct{}

func (s courseAdminService) ListCourse(req *model.ListCourseAdminReq) (resp *response.PageResp, err error) {
	records := make([]*model.CourseResp, 0)
	d := dao.Course
	if err = d.Page(req.PageCurrent, req.PageSize).Scan(&records); err != nil {
		return nil, err
	}

	count, err := d.Count()
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
