// @Author: 陈健航
// @Date: 2021/3/8 16:55
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/response"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/text/gstr"
	"math"
)

var LabSummitService = new(labSummitService)

type labSummitService struct{}

func (s *labSummitService) InsertReport(req *model.SummitReportReq) (err error) {
	// 上传到文件服务器
	successFlag := true
	url, err := FileService.UploadPdf(req.Report)
	// 移除脏文件
	defer func(flag *bool) {
		if *flag {
			go FileService.RemoveDirtyFile(url)
		}
	}(&successFlag)
	one, err := dao.LabSubmit.Where(dao.LabSubmit.Columns.LabId, req.LabId).And(dao.LabSubmit.Columns.StuId).
		FindOne(dao.LabSubmit.Columns.ReportUrl)
	if err != nil {
		return err
	}
	if one == nil {
		// 原来没有记录，插入新记录
		var saveModel = new(model.LabSubmit)
		saveModel.StuId = req.StuId
		saveModel.LabId = req.LabId
		saveModel.ReportUrl = url
		// 查出冗余信息
		if stuDetail, err := dao.SysUser.WherePri(req.StuId).Fields(dao.SysUser.Columns.RealName, dao.SysUser.Columns.Num).FindOne(); err != nil {
			return err
		} else {
			saveModel.StuRealName = stuDetail.RealName
			saveModel.StuNum = stuDetail.Num
			saveModel.IsFinish = 0
		}
		if _, err = dao.LabSubmit.OmitEmpty().Insert(saveModel); err != nil {
			successFlag = false
			return err
		}
	} else {
		// 保存新报告
		if _, err = dao.LabSubmit.Where(dao.LabSubmit.Columns.LabId, req.LabId).And(dao.LabSubmit.Columns.StuId, req.StuId).
			OmitEmpty().Save(dao.LabSubmit.Columns.ReportUrl, url); err != nil {
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
	if _, err = g.Table("coding_time").Data(g.Map{
		"stu_id":   stuId,
		"duration": duration,
		"lab_id":   labId,
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
	one, err := dao.LabSubmit.Where(dao.LabSubmit.Columns.LabId, req.LabId).And(dao.LabSubmit.Columns.StuId).
		FindOne(dao.LabSubmit.Columns.ReportUrl)
	if err != nil {
		return err
	}
	if one == nil {
		// 原来没有记录，插入新记录
		var saveModel = new(model.LabSubmit)
		saveModel.StuId = req.StuId
		saveModel.LabId = req.LabId
		if req.IsFinish {
			saveModel.IsFinish = 1
		} else {
			saveModel.IsFinish = 0
		}
		// 查出冗余信息
		if stuDetail, err := dao.SysUser.WherePri(req.StuId).Fields(dao.SysUser.Columns.RealName, dao.SysUser.Columns.Num).FindOne(); err != nil {
			return err
		} else {
			saveModel.StuRealName = stuDetail.RealName
			saveModel.StuNum = stuDetail.Num
		}
		if _, err = dao.LabSubmit.OmitEmpty().Insert(saveModel); err != nil {
			return err
		}
	} else {
		// 保存记录
		if _, err = dao.LabSubmit.Where(dao.LabSubmit.Columns.LabId, req.LabId).And(dao.LabSubmit.Columns.StuId, req.StuId).
			OmitEmpty().Save(dao.LabSubmit.Columns.IsFinish, req.IsFinish); err != nil {
			return err
		}
	}
	return nil
}

func (s *labSummitService) SelectLabSummit(req *model.SelectLabSummitReq) (resp *response.PageResp, err error) {
	d := dao.LabSubmit.Where(dao.LabSubmit.Columns.LabId, req.LabId)
	var records = make([]*model.LabSummitResp, 0)
	// 分页
	if err = d.Page(req.PageSize, req.PageCurrent).
		// 按学号排序
		Order(dao.LabSubmit.Columns.StuNum).
		FieldsEx(dao.LabSubmit.Columns.DeletedAt).
		Scan(&records); err != nil {
		return nil, err
	}

	// 查询编码时间
	// 列出所有学生id
	stuIds := gdb.ListItemValuesUnique(records, gstr.CamelCase(dao.LabSubmit.Columns.StuId))
	// 聚合查询学生的编码时间
	type codingTime struct {
		StuId    int
		Duration int
	}
	var codingTimes = make([]*codingTime, 0)
	if err = g.Table("coding_time").Where("lab_id", req.LabId).And("stu_id", stuIds).
		Fields("SUM(duration) duration,stu_id").Group("stu_id").Scan(&codingTimes); err != nil {
		return nil, err
	}

	// 装填
	for _, v := range records {
		for _, v1 := range codingTimes {
			if v.StuId == v1.StuId {
				v.CodingTime = v1.Duration
				break
			}
		}
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
		}}
	return resp, nil
}
