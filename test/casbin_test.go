// @Author: 陈健航
// @Date: 2021/2/6 22:05
// @Description:
package test

import (
	"github.com/gogf/gf/frame/g"
	"testing"
)

func TestAddCasBin(t *testing.T) {
	_, err := g.Table("casbin_rule").Data(
		g.Map{
			// 给角色添加权限
			"ptype": "p",
			// 角色type,1教师，2学生，3系统管理员
			"v0": "2",
			// 后台路径
			"v1": "/web/course/quit/:courseId",
			// method
			"v2": "DELETE",
		}).Save()
	if err != nil {
		println(err)
	}
}
func TestCabinGroup(t *testing.T) {
	_, err := g.Table("casbin_rule").Data(g.Map{
		// 给角色添加权限
		"ptype": "g",
		// 用户id
		"v0": "3",
		// 角色type,1教师，2学生，3系统管理员
		"v1": "1",
	}).Save()
	if err != nil {
		println(err)
	}
}
