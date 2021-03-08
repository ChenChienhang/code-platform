// @Author: 陈健航
// @Date: 2021/3/3 23:31
// @Description:
package ide

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gtime"
	"time"
)

type iTheiaService interface {
	GetOrRunIDE(userId int, languageEnum int, labId int) (url string, err error)
	clearTimeOutIDE()
	shutDownIDE(userId int, languageEnum int, labId int) (err error)
	CloseIDE(userId int, languageEnum int, labId int) (err error)
}

var TheiaService = newTheiaService()

func newTheiaService() (t iTheiaService) {
	// 用k3s
	if g.Cfg().GetBool("ide.isK3S") {

	} else {
		// 用docker
		t = newDockerTheiaService()
	}
	// 每5分钟清理一次容器
	go func() {
		for {
			time.Sleep(5 * time.Minute)
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
		languageType = "ts"
	case 6:
		languageType = "php"
	case 7:
		languageType = "swift"
	case 8:
		languageType = "rust"
	}
	return languageType
}

type theiaState struct {
	Url       string
	StartTime *gtime.Time
	EndTime   *gtime.Time
}
