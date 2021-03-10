// @Author: 陈健航
// @Date: 2020/12/29 11:00
// @Description:
package test

import (
	"code-platform/app/dao"
	"code-platform/app/service/component"
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
	"path"
	"strconv"
	"sync"
	"testing"
	"time"
)

type Sy struct {
	sync.Mutex
}

func TestMoss(t *testing.T) {
	fileList, err := gfile.ScanDirFile("C:/temp/solution_directory", "*.c", true)
	if err != nil {
		fmt.Print(err)
	}
	c, err := component.NewMossClient("c")
	defer func(err error) {
		err = c.Close()
		if err != nil {
			println(err)
		}
	}(err)
	if err != nil {
		println(err)
		return
	}
	err = c.Run()
	for _, f := range fileList {
		err = c.UploadFile(f, false)
		if err != nil {
			println(err)
		}
	}
	_ = c.SendQuery()
	url := c.ResultURL
	fmt.Printf("%T, %v\n", url, url)
}

func TestMinio(t *testing.T) {
	type codingTime struct {
		StuId    int
		Duration int
	}
	var codingTimes = make([]*codingTime, 0)
	if err := g.Table("coding_time").Where("lab_id", 1).And("stu_id", 1).
		Fields("SUM(duration) duration,stu_id").Group("stu_id").Scan(&codingTimes); err != nil {
		println(err)
	}
	println(codingTimes)
}

func TestPage(t *testing.T) {
	all, err := dao.SysUser.Page(1, 2).FindAll()
	if err != nil {
		println(err)
	}
	res, err := dao.SysUser.FindValue(dao.SysUser.Columns.UserId, dao.SysUser.Columns.RealName, 1)
	for _, v := range all {
		println(v)
	}

	fmt.Println(res.IsEmpty())
}

func TestPassword(t *testing.T) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		println(err)
	}
	println(string(hashPassword))
}

func Test4(t *testing.T) {
	config := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password("4125qaswIklo8569")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", "118.178.253.239:22", config)
	if err != nil {
		println(err)
	}
	session, err := client.NewSession()
	if err != nil {
		println(err)
	}
	// 这里不处理error了，有输出会伴随错误,命令时格式化输出，用了go的template语法
	//session.Wait()
	output, err := session.CombinedOutput(fmt.Sprintf("sleep 2 && netstat -an | grep :%d", 3000))
	if err != nil {
		println(err)
	}
	_ = session.Close()
	split := gstr.Split(string(output), "\n")
	for _, v := range split {
		if gstr.HasPrefix(v, "mytheia") {

		}
	}
}

func Test45678(t *testing.T) {

}

func Test78(t *testing.T) {
	config := &ssh.ClientConfig{
		User:            g.Cfg().GetString("theia.docker.user"),
		Auth:            []ssh.AuthMethod{ssh.Password(g.Cfg().GetString("theia.docker.pass"))},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", g.Cfg().GetString("theia.docker.host")+":22", config)
	if err != nil {
		println(err)
	}
	session, err := client.NewSession()
	userId := 123
	if err != nil {
		return
	}

	basePath := "/mnt/theia"
	cmd := fmt.Sprintf("if [ ! -d %s ];then %s; fi", path.Join(basePath, "students", strconv.Itoa(userId)),
		fmt.Sprintf("mkdir -p %s && mkdir -p %s && mkdir -p %s && mkdir -p %s && mkdir -p %s && mkdir -p %s && mkdir -p %s && mkdir -p %s mkdir -p %s",
			path.Join(basePath, "students", strconv.Itoa(userId), ".theia-full"),
			path.Join(basePath, "students", strconv.Itoa(userId), ".theia-cpp"),
			path.Join(basePath, "students", strconv.Itoa(userId), ".theia-java"),
			path.Join(basePath, "students", strconv.Itoa(userId), ".theia-python"),
			path.Join(basePath, "students", strconv.Itoa(userId), ".theia-go"),
			path.Join(basePath, "students", strconv.Itoa(userId), ".theia-js"),
			path.Join(basePath, "students", strconv.Itoa(userId), ".theia-php"),
			path.Join(basePath, "students", strconv.Itoa(userId), ".theia-swift"),
			path.Join(basePath, "students", strconv.Itoa(userId), ".theia-rust")))
	if err = session.Run(cmd); err != nil {
		println(err)
	}
}
func Test789(t *testing.T) {
	r := g.Redis()
	timeout, err := r.DoWithTimeout(5*time.Second, "SET", "2", "123")
	do, err := r.DoVar("GET", "2")
	if err != nil {
		println(err)
	}
	println(do.Int())
	println(timeout)

}
