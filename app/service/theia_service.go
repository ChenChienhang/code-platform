// @Author: 陈健航
// @Date: 2021/2/25 20:14
// @Description:
package service

import (
	"code-platform/app/dao"
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/util/gconv"
	"golang.org/x/crypto/ssh"
	"path/filepath"
	"strconv"
	"time"
)

var TheiaService = newTheiaService()

type theiaService struct {
	sshClient *ssh.Client
}

func newTheiaService() (s *theiaService) {
	s = new(theiaService)
	// 添加定时任务凌晨12点重新部署容器
	//if _, err := gcron.Add("@daily",
	//	func() {
	//		redeploymentUrls := g.Cfg().GetStrings("k3s.redeployment.redeploymentUrl")
	//		token := g.Cfg().GetString("k3s.redeployment.bearToken")
	//		// 添加任务，每天向k3s发布重新部署任务
	//		for _, v := range redeploymentUrls {
	//			if _, err := g.Client().Header(map[string]string{
	//				"Authorization":  "Bearer " + token,
	//				"Accept":         "application/json",
	//				"Content-Type":   "application/json",
	//				"Content-Length": "2",
	//			}).Post(v); err != nil {
	//				glog.Errorf("定时重新部署err:%s", err.Error())
	//			}
	//		}
	//	},
	//); err != nil {
	//	panic(err)
	//}
	//// 创建ssh client
	//config := &ssh.ClientConfig{
	//	User:            g.Cfg().GetString("k3s.storage.user"),
	//	Auth:            []ssh.AuthMethod{ssh.Password(g.Cfg().GetString("k3s.storage.pass"))},
	//	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	//}
	//client, err := ssh.Dial("tcp", g.Cfg().GetString("k3s.storage.host")+":22", config)
	//if err != nil {
	//	println(err)
	//}
	//s.sshClient = client
	return s
}

// GetTheiaWorkSpaceByCourseId 获取可访问的theia url
// @receiver s
// @params language 语言 枚举
// @params userId 用户id
// @params labId 实验id,当-1是为自由project区
// @return workSpaceUrl
// @date 2021-02-25 21:54:23
func (s *theiaService) GetTheiaWorkSpaceByCourseId(userId int, labId int) (workSpaceUrl string, err error) {
	// 查出语言类型
	courseId, err := dao.Course.Cache(time.Hour * 1).WherePri(labId).FindValue(dao.Lab.Columns.CourseId)
	if err != nil {
		return "", err
	}
	language, err := dao.Course.Cache(time.Hour * 1).WherePri(courseId.Int()).FindValue(dao.Course.Columns.Language)
	if err != nil {
		return "", err
	}
	var baseUrl string
	switch language.Int() {
	case 1:
		baseUrl = g.Cfg().GetString("k3s.theia.theia-full")
	case 2:
		baseUrl = g.Cfg().GetString("k3s.theia.theia-cpp")
	case 3:
		baseUrl = g.Cfg().GetString("k3s.theia.theia-python")
	case 4:
		baseUrl = g.Cfg().GetString("k3s.theia.theia-go")
	case 5:
		baseUrl = g.Cfg().GetString("k3s.theia.theia-web")
	}
	// 预先创建挂载文件夹，否则报错
	if err = s.execMkdirInStorage(userId, gconv.String(labId)); err != nil {
		return "", err
	}
	return baseUrl + "/#/home/project/" + filepath.Join(strconv.Itoa(userId), strconv.Itoa(labId)), err
}

// GetTheiaWorkSpaceByLanguage 打开自由编辑区
// @receiver s
// @params language
// @params userId
// @return workSpaceUrl
// @return err
// @date 2021-02-26 13:23:17
func (s *theiaService) GetTheiaWorkSpaceByLanguage(language int, userId int) (workSpaceUrl string, err error) {
	var baseUrl string
	switch language {
	case 1:
		baseUrl = g.Cfg().GetString("k3s.theia.theia-full")
	case 2:
		baseUrl = g.Cfg().GetString("k3s.theia.theia-cpp")
	case 3:
		baseUrl = g.Cfg().GetString("k3s.theia.theia-python")
	case 4:
		baseUrl = g.Cfg().GetString("k3s.theia.theia-go")
	case 5:
		baseUrl = g.Cfg().GetString("k3s.theia.theia-web")
	}
	// 预先创建挂载文件夹，否则报错
	if err = s.execMkdirInStorage(userId, "project"); err != nil {
		return "", err
	}
	return baseUrl + "/#/home/project/" + strconv.Itoa(userId) + "/project", nil
}

func (s *theiaService) execMkdirInStorage(userId int, dirName string) error {
	session, err := s.sshClient.NewSession()
	if err != nil {
		return nil
	}
	// defer 释放
	defer func(ssh *ssh.Session) {
		_ = session.Close()
	}(session)
	// 完整文件目录
	path := filepath.Join(g.Cfg().GetString("k3s.storage.path"), strconv.Itoa(userId), dirName)
	// 如果不存在则创建目录
	if err = session.Run(fmt.Sprintf("if [ ! -d %s ];then %s fi", path, "mkdir "+path)); err != nil {
		return err
	}
	return nil
}
