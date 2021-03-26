// @Author: 陈健航
// @Date: 2021/3/8 16:55
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/response"
	"fmt"
	"github.com/gogf/gf/frame/g"
	"math"
	"time"
)

var LabSummitService = new(labSummitService)

type labSummitService struct{}

func (s *labSummitService) InsertReport(req *model.SummitReportReq) (err error) {
	successFlag := true
	// 移除脏文件
	defer func(flag *bool) {
		if *flag {
			go FileService.RemoveDirtyFile(req.ReportUrl)
		}
	}(&successFlag)
	one, err := dao.LabSubmit.Where(dao.LabSubmit.Columns.LabId, req.LabId).And(dao.LabSubmit.Columns.UserId).
		Fields(dao.LabSubmit.Columns.ReportUrl).FindOne()
	if err != nil {
		return err
	}
	if one == nil {
		// 原来没有记录，插入新记录
		if _, err = dao.LabSubmit.OmitEmpty().Data(&model.LabSubmit{
			ReportUrl: req.ReportUrl,
			LabId:     req.LabId,
			UserId:    req.StuId,
		}).Insert(); err != nil {
			successFlag = false
			return err
		}
	} else {
		// 保存新报告
		if _, err = dao.LabSubmit.Where(dao.LabSubmit.Columns.LabId, req.LabId).And(dao.LabSubmit.Columns.UserId, req.StuId).
			Data(dao.LabSubmit.Columns.ReportUrl, req.ReportUrl).OmitEmpty().Update(); err != nil {
			successFlag = false
			return err
		}
		// 移除旧文件
		if one.ReportUrl != "" {
			//goland:noinspection GoUnhandledErrorResult
			go FileService.RemoveObject(one.ReportUrl)
		}
	}
	return nil
}

// InsertCodingTime 插入编程时间
// @receiver s
// @params labId
// @params duration
// @return err
// @date 2021-03-08 17:16:30
func (s *labSummitService) InsertCodingTime(labId int, duration int, stuId int) (err error) {
	if _, err = dao.CodingTime.Data(g.Map{
		dao.CodingTime.Columns.UserId:   stuId,
		dao.CodingTime.Columns.Duration: duration,
		dao.CodingTime.Columns.LabId:    labId,
	}).Insert(); err != nil {
		return err
	}
	return nil
}

// InsertCodeFinish 是否已经完成编写代码
// @receiver s
// @params req
// @return err
// @date 2021-03-08 17:08:06
func (s *labSummitService) InsertCodeFinish(req *model.SummitLabFinishReq) (err error) {
	one, err := dao.LabSubmit.Where(dao.LabSubmit.Columns.LabId, req.LabId).And(dao.LabSubmit.Columns.UserId).
		FindOne(dao.LabSubmit.Columns.ReportUrl)
	if err != nil {
		return err
	}
	if one == nil {
		// 原来没有记录，插入新记录
		if _, err = dao.LabSubmit.OmitEmpty().Data(g.Map{
			dao.LabSubmit.Columns.UserId:   req.StuId,
			dao.LabSubmit.Columns.LabId:    req.LabId,
			dao.LabSubmit.Columns.IsFinish: req.IsFinish,
		}).Insert(); err != nil {
			return err
		}
	} else {
		// 保存记录
		if _, err = dao.LabSubmit.Where(dao.LabSubmit.Columns.LabId, req.LabId).And(dao.LabSubmit.Columns.UserId, req.StuId).
			OmitEmpty().Update(dao.LabSubmit.Columns.IsFinish, req.IsFinish); err != nil {
			return err
		}
	}
	return nil
}

func (s *labSummitService) ListLabSummit(req *model.SelectLabSummitReq) (resp *response.PageResp, err error) {
	// 找出所有id
	courseId, err := dao.Lab.WherePri(req.LabId).Cache(time.Hour).FindValue(dao.Lab.Columns.CourseId)
	if err != nil {
		return nil, err
	}
	stuIds, err := dao.Course.ListUserIdByCourseId(courseId.Int())
	if err != nil {
		return nil, err
	}

	var records = make([]*model.LabSummitResp, 0)
	d := dao.SysUser.WherePri(stuIds)

	// 查出所有选课学生的基本信息
	if err = d.Page(req.PageCurrent, req.PageSize).Order(dao.SysUser.Columns.Num).Fields(
		dao.SysUser.Columns.Num,
		dao.SysUser.Columns.RealName,
		dao.SysUser.Columns.UserId,
	).Scan(&records); err != nil {
		return nil, err
	}
	count, err := d.Count()
	if err != nil {
		return nil, err
	}

	labSummit := make([]*model.LabSubmit, 0)
	if err = dao.LabSubmit.Where(dao.LabSubmit.Columns.LabId, req.LabId).And(dao.LabSubmit.Columns.UserId, stuIds).
		FieldsEx(dao.LabSubmit.Columns.DeletedAt).Scan(&labSummit); err != nil {
		return nil, err
	}
	for _, v := range records {
		for _, v2 := range labSummit {
			if v.UserId == v2.UserId {
				v.LabSubmitId = v2.LabSubmitId
				v.ReportUrl = v2.ReportUrl
				v.CreatedTime = v2.CreatedAt
				v.UpdatedTime = v2.UpdatedAt
				// 已完成状态
				if v2.IsFinish == 1 {
					v.FinishStat = 1
				} else {
					// 正在做
					v.FinishStat = 2
				}
				break
			}
		}
	}

	// 聚合查询学生的编码时间
	type codingTime struct {
		UserId   int
		Duration int
	}

	var codingTimes = make([]*codingTime, 0)
	if err = dao.CodingTime.Where(dao.CodingTime.Columns.LabId, req.LabId).And(dao.CodingTime.Columns.UserId, stuIds).
		Fields(fmt.Sprintf("SUM(%s) %s", dao.CodingTime.Columns.Duration, dao.CodingTime.Columns.Duration),
			dao.CodingTime.Columns.UserId).
		Group(dao.CodingTime.Columns.UserId).Scan(&codingTimes); err != nil {
		return nil, err
	}
	// 装填
	for _, v := range records {
		for _, v1 := range codingTimes {
			if v.UserId == v1.UserId {
				v.CodingTime = v1.Duration
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

func (s *labSummitService) InsertBlankRecord(userId int, labId int) (err error) {
	if one, err := dao.LabSubmit.Where(dao.LabSubmit.Columns.UserId, userId).
		And(dao.LabSubmit.Columns.LabId, labId).FindOne(); err != nil {
		return err
	} else if one == nil {
		if _, err = dao.LabSubmit.Insert(g.Map{
			dao.LabSubmit.Columns.LabId:    labId,
			dao.LabSubmit.Columns.UserId:   userId,
			dao.LabSubmit.Columns.IsFinish: 0,
		}); err != nil {
			return err
		}
	}
	return nil
}
