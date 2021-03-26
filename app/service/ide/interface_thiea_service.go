// @Author: 陈健航
// @Date: 2021/3/3 23:31
// @Description:
package ide

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/response"
	"github.com/gogf/gf/frame/g"
	"time"
)

type iTheiaService interface {
	OpenIDE(req *model.OpenIDEReq) (url string, err error)
	ClearIDE()
	StopIDE(req *model.CloseIDEReq) (err error)
	CollectCompilerErrorLog(req *model.SelectCompilerErrorLogReq) (resp *response.PageResp, err error)
	CheckCode(req *model.CheckCodeReq) (url string, err error)
	ListContianer()
}

var TheiaService = newTheiaService()

func newTheiaService() (t iTheiaService) {
	// 用k3s
	if g.Cfg().GetBool("ide.isK3S") {

	} else {
		// 用docker
		t = newDockerTheiaService()
	}
	// 每3分钟清理一次容器
	go func() {
		for {
			time.Sleep(3 * time.Minute)
			t.ClearIDE()
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

func getLanguageEnumByLabId(labId int) (languageEnum int, err error) {
	// 查出所用语言
	if labId != 0 {
		courseId, err := dao.Lab.Cache(time.Hour).WherePri(labId).FindValue(dao.Lab.Columns.CourseId)
		if err != nil {
			return 0, err
		}
		languageEnumV, err := dao.Course.Cache(time.Hour).WherePri(courseId.Int()).FindValue(dao.Course.Columns.Language)
		if err != nil {
			return 0, err
		}
		languageEnum = languageEnumV.Int()
	} else {
		// 自由区
		languageEnum = 0
	}
	return languageEnum, nil
}

// 获取容器名
func getImageName(languageEnum int) (imageName string) {
	switch languageEnum {
	case 0:
		imageName = g.Cfg().GetString("theia.docker.image.full")
	case 1:
		imageName = g.Cfg().GetString("theia.docker.image.cpp")
	case 2:
		imageName = g.Cfg().GetString("theia.docker.image.java")
	case 3:
		imageName = g.Cfg().GetString("theia.docker.image.python")
	}
	return imageName
}

func getEnvironmentMount(languageEnum int) (environmentMount string) {
	switch languageEnum {
	case 0:
		//full
		environmentMount = "/home/theia/.theia"
	case 1:
		//cpp
		environmentMount = "/root/.theia"
	case 2:
		//java
		environmentMount = "/root/.theia"
	case 3:
		//python
		environmentMount = "/home/theia/.theia"
	case 4:
		//web
		environmentMount = "/root/.theia"
	}
	return environmentMount
}
