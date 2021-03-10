// @Author: 陈健航
// @Date: 2021/3/3 23:31
// @Description:
package ide

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/response"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gtime"
	"time"
)

type iTheiaService interface {
	GetOrRunIDE(req *model.GetIDEUrlReq) (url string, err error)
	clearTimeOutIDE()
	shutDownIDE(userId int, languageEnum int, labId int) (err error)
	CloseIDE(req *model.CloseIDEReq) (err error)
	CollectCompilerErrorLog(req *model.SelectCompilerErrorLogReq) (resp *response.PageResp, err error)
}

var TheiaService = newTheiaService()

func newTheiaService() (t iTheiaService) {
	// 用k3s
	if g.Cfg().GetBool("ide.isK3S") {

	} else {
		// 用docker
		t = newDockerTheiaService()
	}
	// 每10分钟清理一次容器
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			t.clearTimeOutIDE()
		}
	}()
	return t
}

// getLanguageString 枚举转字符串
// @params languageEnum
// @return string
// @date 2021-03-05 22:14:21
func getLanguageString(languageEnum int) string {
	var languageType string
	switch languageEnum {
	case 0:
		languageType = "full"
	case 1:
		languageType = "cpp"
	case 2:
		languageType = "java"
	case 3:
		languageType = "python"
	case 4:
		languageType = "go"
	case 5:
		// ts/js
		languageType = "web"
	case 6:
		languageType = "php"
	case 7:
		languageType = "swift"
	case 8:
		languageType = "rust"
	}
	return languageType
}

func getLanguageStringByLabId(labId int) (languageString string, err error) {
	// 查出所用语言
	var languageEnum int
	if labId != 0 {
		courseId, err := dao.Lab.Cache(time.Hour).WherePri(labId).FindValue(dao.Lab.Columns.CourseId)
		if err != nil {
			return "", err
		}
		languageEnumV, err := dao.Course.Cache(time.Hour).WherePri(courseId.Int()).FindValue(dao.Course.Columns.Language)
		if err != nil {
			return "", err
		}
		languageEnum = languageEnumV.Int()
	} else {
		languageEnum = 0
	}
	languageString = getLanguageString(languageEnum)
	return languageString, nil
}

type theiaState struct {
	Url       string
	StartTime *gtime.Time
	EndTime   *gtime.Time
}
