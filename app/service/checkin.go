// @Author: 陈健航
// @Date: 2021/1/16 22:04
// @Description:
package service

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/code"
	"code-platform/library/common/response"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/text/gstr"
	"math"
	"strconv"
	"time"
)

var CheckinService = newCheckinService()

type checkinService struct {
	// 保存密钥
	redisHeaderSecret string
	// 完成签到的学生
	redisHeaderFinishSignIn string
}

func newCheckinService() (s *checkinService) {
	s = &checkinService{
		redisHeaderFinishSignIn: "code.platform:signin:finish:",
		redisHeaderSecret:       "code.platform:signin:key:",
	}
	return s
}

// StartCheckIn 教师发起签到
// @receiver c
// @params req
// @return err
// @date 2021-03-16 14:25:36
func (receiver *checkinService) StartCheckIn(req *model.StartCheckInReq) (err error) {
	// 存入签到密钥,限时
	if err = receiver.setSecretKey(req.SecretKey, req.CourseId, req.Duration); err != nil {
		return err
	}
	// 清楚上次签到的可能残余的数据
	if err = receiver.removeFinishStu(req.CourseId); err != nil {
		return err
	}
	// 启动一个协程，当倒计时结束，把签到信息保存
	go func() {
		// 等待签到时间结束
		time.Sleep(time.Duration(req.Duration))
		// 取出set
		stuIds, _ := receiver.getFinishStu(req.CourseId)
		defer func() {
			_ = receiver.removeFinishStu(req.CourseId)
		}()
		// 取得已签到的所有学生的id,组装保存的内容
		saveSlice := make([]model.CheckinDetailResp, len(stuIds))
		for i, v := range stuIds {
			saveSlice[i].CourseId = req.CourseId
			saveSlice[i].StuId = v
		}
		// 保存签到信息
		if _, err = dao.CheckinDetail.Batch(len(stuIds)).Update(saveSlice); err != nil {
			glog.Errorf("签到保存数据库错误 :%s")
		}
	}()
	return nil
}

// ListCheckinRecords 列表签到记录
// @receiver c
// @params pageCurrent
// @params pageSize
// @params courseId
// @return resp
// @return err
// @date 2021-02-20 23:55:42
func (receiver *checkinService) ListCheckinRecords(req *model.ListCheckinRecordsReq) (resp *response.PageResp, err error) {
	// 按创建时间降序
	checkInRecordResp := make([]*model.CheckinRecordResp, 0)
	d := dao.CheckinRecord.WherePri(req.CourseId).Order(dao.CheckinRecord.Columns.CreatedAt + " desc")
	if err = d.Page(req.PageCurrent, req.PageSize).Scan(&checkInRecordResp); err != nil {
		return nil, err
	}
	count, err := d.Count()
	if err != nil {
		return nil, err
	}

	for _, v := range checkInRecordResp {
		// 参与签到的人数
		if v.Actual, err = dao.CheckinDetail.Where(dao.CheckinDetail.Columns.CheckinRecordId, v.CheckinRecordId).
			FindCount(); err != nil {
			return nil, err
		}
		// 实际签到的人数
		if v.Total, err = dao.Course.CountByCourseId(v.CourseId); err != nil {
			return nil, err
		}
	}

	// 分页信息整合
	resp = &response.PageResp{
		Records: checkInRecordResp,
		PageInfo: &response.PageInfo{
			Size:    len(checkInRecordResp),
			Total:   count,
			Current: req.PageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(req.PageSize))),
		}}
	return resp, nil
}

// ListCheckinDetail 列表签到详情记录
// @receiver c
// @params pageCurrent
// @params pageSize
// @params checkInRecordId
// @return resp
// @return err
// @date 2021-02-20 23:57:15
func (receiver *checkinService) ListCheckinDetail(req *model.ListCheckinDetailsReq) (resp *response.PageResp, err error) {
	// 查出courseId
	courseId, err := dao.CheckinRecord.WherePri(req.CheckInRecordId).FindValue(dao.CheckinRecord.Columns.CourseId)
	if err != nil {
		return nil, err
	}
	// 查出实际选课学生
	ids, err := dao.Course.ListUserIdByCourseId(courseId.Int())
	if err != nil {
		return nil, err
	}
	checkinDetails := make([]*model.CheckinDetailResp, 0)
	d := dao.SysUser.WherePri(ids)
	if err = d.Fields(
		dao.SysUser.Columns.RealName,
		dao.SysUser.Columns.Num,
		dao.SysUser.Columns.Organization,
	).Page(req.PageCurrent, req.PageSize).Scan(&checkinDetails); err != nil {
		return nil, err
	}

	count, err := d.Count()
	if err != nil {
		return nil, err
	}

	// 查出本次参与签到的学生
	stuIds, err := dao.CheckinDetail.Where(dao.CheckinDetail.Columns.CheckinRecordId, req.CheckInRecordId).
		FindArray(dao.CheckinDetail.Columns.StuId)
	if err != nil {
		return nil, err
	}

	// 标记
	for _, v := range checkinDetails {
		for _, v1 := range stuIds {
			if v1.Int() == v.StuId {
				v.IsCheckIn = true
			}
			break
		}
	}
	// 返回
	resp = &response.PageResp{
		Records: checkinDetails,
		PageInfo: &response.PageInfo{
			Size:    len(checkinDetails),
			Total:   count,
			Current: req.PageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(req.PageSize))),
		}}
	return resp, nil
}

// ExportCheckInRecords 导出签到表
// @receiver c
// @params courseId
// @return file
// @return err
// @date 2021-03-16 14:25:24
func (receiver *checkinService) ExportCheckInRecords(courseId int) (file *excelize.File, err error) {
	// 查出该课程所有选课学生id
	ids, err := g.Table("re_course_user").Where("course_id", courseId).FindArray("user_id")

	// 查出学号，姓名
	students := make([]*model.SysUser, 0)
	if err = dao.SysUser.WherePri(ids).Fields(
		dao.SysUser.Columns.UserId,
		dao.SysUser.Columns.RealName,
		dao.SysUser.Columns.Num,
	).Scan(&students); err != nil {
		return nil, err
	}

	// 查出这门课所有签到记录
	checkinRecords := make([]*model.CheckinRecord, 0)
	if err = dao.CheckinRecord.Where(dao.CheckinRecord.Columns.CourseId, courseId).Fields(
		dao.CheckinRecord.Columns.CheckinRecordId,
		dao.CheckinRecord.Columns.Name,
	).Scan(&checkinRecords); err != nil {
		return nil, err
	}

	// 查出这门课所有签到详细记录
	checkinDetails := make([]*model.CheckinDetail, 0)
	if err = dao.Course.Where(dao.CheckinRecord.Columns.CourseId, gdb.ListItemValuesUnique(
		checkinRecords,
		gstr.CamelCase(dao.CheckinRecord.Columns.CourseId),
	)).FieldsEx(
		dao.CheckinDetail.Columns.StuId,
		dao.CheckinRecord.Columns.CheckinRecordId,
		dao.CheckinRecord.Columns.Name,
	).Scan(&checkinDetails); err != nil {
		return nil, err
	}

	// 创建新excel
	file = excelize.NewFile()

	// 准备表头
	if err = file.SetCellStr("Sheet1", "A1", "学号"); err != nil {
		return nil, err
	}
	if err = file.SetCellStr("Sheet1", "B1", "姓名"); err != nil {
		return nil, err
	}
	if err = file.SetCellStr("Sheet1", "C1", "出勤率"); err != nil {
		return nil, err
	}
	for i, v := range checkinRecords {
		// 在C1,D1,E1...的位置写入,最长到Z1
		if err = file.SetCellStr("Sheet1", fmt.Sprintf("%c1", rune(int('A')+i+3)), v.Name); err != nil {
			return nil, err
		}
	}

	// 表内容
	for i, v := range students {
		// 写入学号，注意编号至少从2开始
		if err = file.SetCellStr("Sheet1", fmt.Sprintf("A%d", i+3), v.Num); err != nil {
			return nil, err
		}
		// 写入姓名
		if err = file.SetCellStr("Sheet1", fmt.Sprintf("B%d", i+3), v.RealName); err != nil {
			return nil, err
		}
		// 出勤次数
		count := 0
		for i1, v1 := range checkinRecords {
			isExist := false
			// 查看学生是否参与了这次签到
			for _, v2 := range checkinDetails {
				if v2.StuId == v.UserId && v1.CheckinRecordId == v2.CheckinRecordId {
					count++
					isExist = true
					break
				}
			}
			// 在C1,D1,E1...的位置写入签到记录的名称,最长到Z1
			if isExist {
				if err = file.SetCellStr("Sheet1", fmt.Sprintf("%receiver%d", rune(int('A')+i1+3), i+2), "√"); err != nil {
					return nil, err
				}
			}
		}
		// 写入出勤率
		rate := float32(count / len(checkinRecords))
		if err = file.SetCellStr("Sheet1", fmt.Sprintf("C%d", i+3), fmt.Sprintf("%.2f", rate)); err != nil {
			return nil, err
		}
	}
	return file, nil
}

// UpdateCheckinDetail 更新checkinDetail
// @receiver c
// @params req
// @return error
// @date 2021-03-16 14:25:12
func (receiver *checkinService) UpdateCheckinDetail(req *model.UpdateCheckinDetail) error {
	if req.IsCheckIn {
		// 新增插入
		if _, err := dao.CheckinDetail.Insert(req); err != nil {
			return err
		}
	} else {
		// 改成没有签到，删掉
		if _, err := dao.CheckinDetail.Where(dao.CheckinDetail.Columns.StuId, req.StuId).And(dao.CheckinDetail.Columns.CheckinRecordId, req.CheckinRecordId).Delete(); err != nil {
			return err
		}
	}
	return nil
}

// CheckIn 学生完成签到
// @receiver c
// @params req
// @return err
// @date 2021-03-16 14:24:58
func (receiver *checkinService) CheckIn(req *model.CheckinForStudentReq) (err error) {
	// 获取签到密钥
	key, err := receiver.getSecretKey(req.CourseId)
	if err != nil {
		return err
	}
	// 无正在进行的签到或者签到已结束
	if key == "" {
		return code.CheckInNotExistError
	} else if key != req.SecretKey {
		//  签到码错误
		return code.CheckInKeyError
	}
	// 签到码正确，加入缓存签到池，等待签到结束后写入数据库
	if err = receiver.setFinishStu(req.StuId, req.CourseId); err != nil {
		return err
	}
	return nil
}

// removeFinishStu 移除在缓存的学生
// @params courseId
// @return err
// @date 2021-03-02 20:26:41
func (receiver *checkinService) removeFinishStu(courseId int) (err error) {
	r := g.Redis()
	if _, err = r.Do("DEL", receiver.redisHeaderFinishSignIn+strconv.Itoa(courseId)); err != nil {
		return err
	}
	return nil
}

// getFinishStu 获得在缓存的学生
// @params courseId
// @return finishStuIds
// @return err
// @date 2021-03-02 20:26:28
func (receiver *checkinService) getFinishStu(courseId int) (finishStuIds []int, err error) {
	// 取出set
	r := g.Redis()
	v, err := r.DoVar("SMEMBERS", receiver.redisHeaderFinishSignIn+strconv.Itoa(courseId))
	if err != nil {
		return nil, err
	}
	return v.Ints(), nil
}

// getSecretKey
// @params courseId
// @return secretKey ""是缓存中不存在值
// @return err
// @date 2021-03-02 19:25:07
func (receiver *checkinService) getSecretKey(courseId int) (secretKey string, err error) {
	// 获取签到密钥
	r := g.Redis()
	v, err := r.DoVar("GET", receiver.redisHeaderSecret+strconv.Itoa(courseId))
	if err != nil {
		return "", err
	}
	// 值不存在
	if v.IsNil() {
		return "", nil
	}
	return v.String(), nil
}

// setSecretKey 密钥
// @params secretKey
// @params duration
// @params courseId
// @return err
// @date 2021-03-02 20:25:46
func (receiver *checkinService) setSecretKey(secretKey string, duration int, courseId int) (err error) {
	r := g.Redis()
	if _, err = r.DoWithTimeout(
		time.Duration(duration)*time.Second,
		"SET",
		receiver.redisHeaderSecret+strconv.Itoa(courseId),
		secretKey,
	); err != nil {
		return err
	}
	return nil
}

// setFinishStu 完成签到的学生放入缓存，等待倒计时结束
// @params stuId
// @params courseId
// @return err
// @date 2021-03-02 20:25:54
func (receiver *checkinService) setFinishStu(stuId int, courseId int) (err error) {
	r := g.Redis()
	if _, err = r.Do("SADD", receiver.redisHeaderFinishSignIn+strconv.Itoa(courseId), stuId); err != nil {
		return err
	}
	return nil
}
