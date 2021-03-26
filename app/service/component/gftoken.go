// @Author: 陈健航
// @Date: 2020/12/31 23:14
// @Description:
package component

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/library/common/code"
	"code-platform/library/common/response"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/goflyfox/gtoken/gtoken"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/util/gconv"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var GfToken = newGfToken()

func newGfToken() *gtoken.GfToken {
	//initRbac()
	return &gtoken.GfToken{
		CacheMode:       2,
		LoginPath:       "/login",
		LogoutPath:      "/logout",
		LoginBeforeFunc: LoginBeforeFunc,
		LoginAfterFunc:  LoginAfterFunc,
		LogoutAfterFunc: LogoutAfterFunc,
		AuthAfterFunc:   AuthAfterFunc,
		AuthExcludePaths: g.SliceStr{
			"/web/user/nickname",
			"/web/user/signup",
			"/web/user/email",
			"/web/user/password",
			"/web/user/verificationCode",
			"/web/user/test/*",
		},
	}
}

// LogoutAfterFunc 重定义退登结果集
// @params r
// @params respData
// @date 2021-01-04 22:14:32
func LogoutAfterFunc(r *ghttp.Request, respData gtoken.Resp) {
	if respData.Success() {
		response.Succ(r)
	} else {
		response.Exit(r, code.LoginError)
	}
}

// AuthAfterFunc 身份认证操作的后续,主要是鉴权
// @params r
// @params respData
// @date 2021-01-04 22:13:39
func AuthAfterFunc(r *ghttp.Request, respData gtoken.Resp) {
	//存在令牌
	if respData.Success() {
		// 鉴权
		if ok, err := authenticate(respData.GetString("userKey"), r.URL.Path, r.Method); err != nil {
			response.Exit(r, code.OtherError)
		} else if !ok {
			// 权限不足
			response.Exit(r, code.PermissionError)
		}
		// 鉴权成功
		r.SetCtxVar(dao.SysUser.Columns.UserId, respData.GetString("userKey"))
		r.Middleware.Next()
		//不存在令牌
	} else {
		var params map[string]interface{}
		if r.Method == "GET" {
			params = r.GetMap()
		} else if r.Method == "POST" {
			params = r.GetMap()
		} else {
			response.Exit(r, code.OtherError)
			return
		}
		no := gconv.String(gtime.TimestampMilli())
		glog.Info("[AUTH_%s][url:%s][params:%s][data:%s]",
			no, r.URL.Path, params, respData.Json())
		response.Exit(r, code.AuthError)
	}
}

//func GetUserId(r *ghttp.Request) int {
//	return GfToken.GetTokenData(r).GetInt("userKey")
//}

// authenticate 鉴权方法
// @params userId
// @params url
// @params method
// @return bool
// @return error
// @date 2021-01-04 21:58:46
func authenticate(userId string, url string, method string) (ok bool, err error) {
	type Api struct {
		Api    string
		Method string
	}
	apis := make([]Api, 0)
	if err = g.Table("sys_api").InnerJoin("sys_api_role").InnerJoin("sys_user_role").
		InnerJoin("sys_user").InnerJoin("sys_role").Cache(1*time.Hour).
		Where("sys_user.user_id =", userId).And("sys_user.user_id = sys_user_role.user_id").
		And("sys_user_role.role_id = sys_api_role.role_id").And("sys_api_role.api_id = sys_api.api_id").
		Fields("api", "method").Scan(&apis); err != nil {
		return false, err
	}

	for _, v := range apis {
		// 用正则匹配,在列表里找有没有权限的
		isMatch := gregex.IsMatchString(v.Api, url) && v.Method == method
		if isMatch {
			return true, nil
		}
	}
	// 没有权限
	return false, nil
}

// @summary 登录
// @description 登录，返回token。
// @tags    快速CURD
// @Accept  json
// @produce json
// @param   username    formData  	 string true  "用户名（邮箱）"
// @param   password 	formData     string false "密码"
// @router  /user/login [POST]
// @success 200 {object} string "token"
func LoginBeforeFunc(r *ghttp.Request) (string, interface{}) {
	var req *model.LoginReq
	// 转换成结构体
	if err := r.Parse(&req); err != nil {
		response.Exit(r, err)
	}
	// 在数据库查询用户是否存在(只查出密码,id）
	one, err := dao.SysUser.Fields(dao.SysUser.Columns.Password, dao.SysUser.Columns.UserId).
		FindOne(dao.SysUser.Columns.Email, req.Username)
	if err != nil {
		response.Exit(r, err)
	}
	if one == nil {
		//不存在该用户
		response.Exit(r, code.UserNotExistError)
	}

	// 校验密码 密码错误
	if err = bcrypt.CompareHashAndPassword([]byte(one.Password), []byte(req.Password)); err != nil {
		response.Exit(r, code.PasswordError)
	}

	//校验成功
	return gconv.String(one.UserId), ""
}

// LoginAfterFunc 重定义返回后结果集
// @params r
// @params respData
// @date 2021-01-04 22:14:51
func LoginAfterFunc(r *ghttp.Request, respData gtoken.Resp) {
	if !respData.Success() {
		response.Exit(r, code.LoginError)
	} else {
		role, err := dao.SysUser.GetRoleById(respData.GetInt("userKey"))
		if err != nil {
			response.Exit(r, err)
		}
		// 返回token
		response.Succ(r, g.Map{
			"role":  role,
			"token": respData.GetString("token"),
		})
	}
}

func initRbac() {
	file, err := excelize.OpenFile("./config/rbac.xlsx")
	if err != nil {
		println(err)
	}
	rows, err := file.Rows("front_api")
	if err != nil {
		println(err)
	}
	type api struct {
		ApiId       int
		Api         string
		Method      string
		Description string
	}
	apis := make([]api, 0)
	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			println(err)
		}
		apis = append(apis, api{
			ApiId:       gconv.Int(row[0]),
			Api:         row[1],
			Method:      row[2],
			Description: row[3],
		})
	}
	if _, err = g.Table("sys_api").Data(apis).Batch(len(apis)).Save(); err != nil {
		println(err)
	}
	rows, err = file.Rows("role_api")
	if err != nil {
		println(err)
	}
	type roleApi struct {
		ApiRoleId int
		RoleId    string
		ApiId     string
	}
	roleApis := make([]roleApi, 0)
	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			println(err)
		}
		roleApis = append(roleApis, roleApi{
			ApiRoleId: gconv.Int(row[0]),
			RoleId:    row[1],
			ApiId:     row[2],
		})
	}
	if _, err = g.Table("sys_api_role").Data(roleApis).Batch(len(roleApis)).Save(); err != nil {
		println(err)
	}
}
