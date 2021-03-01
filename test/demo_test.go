// @Author: 陈健航
// @Date: 2020/12/29 11:00
// @Description:
package test

import (
	"code-platform/app/dao"
	"code-platform/app/service/component"
	"fmt"
	"github.com/gogf/gf/os/gfile"
	"golang.org/x/crypto/bcrypt"
	"sync"
	"testing"
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
	//fileList, err = gfile.ScanDirFile("C:/temp/base_directory", "*.java", true)
	//if err != nil {
	//	println(err)
	//}
	//for _, f := range fileList {
	//	err = c.UploadFile(f, true)
	//	if err != nil {
	//		println(err)
	//	}
	//}
	_ = c.SendQuery()
	url := c.ResultURL
	fmt.Printf("%T, %v\n", url, url)
}

func TestMinio(t *testing.T) {
	all, err := dao.SysUser.WherePri(122).FindAll()
	if err != nil {
		println(err)
	}
	println(all)
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
	one, err := dao.Course.WherePri(1021).FindOne()
	if err != nil {
		println(err)
	}
	println(one)
}
