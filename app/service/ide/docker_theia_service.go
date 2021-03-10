// @Author: 陈健航
// @Date: 2021/2/25 20:14
// @Description:
package ide

import (
	"code-platform/app/dao"
	"code-platform/app/model"
	"code-platform/app/service"
	"code-platform/app/service/component"
	"code-platform/library/common/code"
	"code-platform/library/common/response"
	"fmt"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"golang.org/x/crypto/ssh"
	"math"
	"math/rand"
	"path"
	"strconv"
	"sync"
	"time"
)

type dockerTheiaService struct {
	sshClient      *ssh.Client
	applyPortMutex iTheiaServiceLock // 申请端口时上的锁
	getTheiaMutex  iTheiaServiceLock // 在缓存中取theia时上的锁
	theiaPortCache iTheiaServiceMap
	basePath       string
}

// newDockerTheiaService 构造函数
// @return s
// @date 2021-03-06 22:28:41
func newDockerTheiaService() (s *dockerTheiaService) {
	s = new(dockerTheiaService)
	// 创建ssh client
	config := &ssh.ClientConfig{
		User:            g.Cfg().GetString("theia.docker.user"),
		Auth:            []ssh.AuthMethod{ssh.Password(g.Cfg().GetString("theia.docker.pass"))},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", g.Cfg().GetString("theia.docker.host")+":22", config)
	if err != nil {
		println(err)
	}
	s.sshClient = client
	if g.Cfg().GetBool("server.Multiple") {
		s.applyPortMutex = &redisLock{
			component.RedisLock.NewMutex("code.platform:signin:ide:port"),
		}
		s.getTheiaMutex = &redisLock{
			component.RedisLock.NewMutex("code.platform:signin:ide:get"),
		}
		s.theiaPortCache = &theiaServiceRedisMap{redisHeader: "code.platform:signin:ide:ide.state:"}
		s.basePath = g.Cfg().GetString("theia.docker.basePath")
	} else {
		s.applyPortMutex = &simpleLock{mu: new(sync.Mutex)}
		s.getTheiaMutex = &simpleLock{mu: new(sync.Mutex)}
		s.theiaPortCache = &theiaServiceSimpleMap{m: gmap.NewStrAnyMap(true)}
		s.basePath = g.Cfg().GetString("theia.docker.basePath")
		// 用内存缓存必须清掉所有之前启动过的容器
		_ = s.clearAllIDE()
	}
	return s
}

// CloseIDE 关闭IDE
// @receiver s
// @params userId 用户id
// @params languageEnum 语言枚举
// @return err
// @date 2021-03-06 22:09:22
func (s *dockerTheiaService) CloseIDE(req *model.CloseIDEReq) (err error) {
	// 存入时间
	if err = service.LabSummitService.InsertCodingTime(req.LabId, req.Duration, req.UserId); err != nil {
		return err
	}
	// 查出所用语言
	languageString, err := getLanguageStringByLabId(req.LabId)
	if err != nil {
		return err
	}
	// 取出缓存
	s.getTheiaMutex.MyLock()
	defer s.getTheiaMutex.MyUnLock()
	v, err := s.theiaPortCache.GetV(fmt.Sprintf("%s-%d-%d", languageString, req.UserId, req.LabId))
	if err != nil {
		return err
	}
	var stat = &theiaState{}
	if err = gconv.Struct(v, &stat); err != nil {
		return err
	}
	// 修改结束时间，注意不是直接关了，等待另一个方法定时清理闲置容器
	stat.EndTime = gtime.Now()
	if err = s.theiaPortCache.SetV(fmt.Sprintf("%s-%d-%d", languageString, req.UserId, req.LabId), stat); err != nil {
		return err
	}
	return nil
}

// GetOrRunIDE 启动IDE
// @receiver s
// @params userId 用户Id
// @params languageEnum 语言类型
// @params labId 实验id,自由区传0
// @return url
// @return err
// @date 2021-03-06 22:10:24
func (s *dockerTheiaService) GetOrRunIDE(req *model.GetIDEUrlReq) (url string, err error) {
	ddl, err := dao.Lab.WherePri(req.LabId).FindValue(dao.Lab.Columns.DeadLine)
	if err != nil {
		return "", err
	}
	// 过了截止时间
	if !ddl.IsNil() && ddl.GTime().Before(gtime.Now()) {
		return "", code.DDLError
	}
	languageString, err := getLanguageStringByLabId(req.LabId)
	if err != nil {
		return "", err
	}
	// 上锁
	s.getTheiaMutex.MyLock()
	defer s.getTheiaMutex.MyUnLock()
	// 查看有没有没有关闭的容器
	v, err := s.theiaPortCache.GetV(fmt.Sprintf("%s-%d-%d", languageString, req.UserId, req.LabId))
	if err != nil {
		return "", err
	}
	if v != nil {
		// 之前的还没有关闭,直接打开之前的,返回之前的端口
		var stat = &theiaState{}
		if err = gconv.Struct(v, &stat); err != nil {
			return "", err
		}
		// 刷新时间，重新存
		stat.StartTime = gtime.Now()
		stat.EndTime = nil
		if err = s.theiaPortCache.SetV(fmt.Sprintf("%s-%d-%d", languageString, req.UserId, req.LabId), stat); err != nil {
			return "", err
		}
		return stat.Url, nil
	}
	// 之前的已经关闭,重新开一个新的容器,并存入缓存
	url, err = s.execRunTheiaDocker(req.UserId, languageString, req.LabId)
	if err != nil {
		return "", err
	}
	return url, nil
}

// clearTimeOutIDE 清理"过期"的IDE
// @receiver s
// @date 2021-03-06 22:14:06
func (s *dockerTheiaService) clearTimeOutIDE() {
	s.getTheiaMutex.MyLock()
	defer s.getTheiaMutex.MyUnLock()
	timeOut := s.theiaPortCache.GetTimeOut()
	for _, v := range timeOut {
		if err := s.execStopAndRemoveTheiaDocker(gconv.Int(v.Get("userId")),
			v.Get("language"),
			gconv.Int(v.Get("labId"))); err != nil {
			glog.Errorf("关闭theia失败")
		}
	}
}

// shutDownIDE 关闭IDE
// @receiver s
// @params userId
// @params languageEnum
// @params labId
// @return err
// @date 2021-03-06 22:12:57
func (s *dockerTheiaService) shutDownIDE(userId int, languageEnum int, labId int) (err error) {
	language := getLanguageString(languageEnum)
	if err = s.execStopAndRemoveTheiaDocker(userId, language, labId); err != nil {
		return err
	}
	return nil
}

// execGetAvailablePort 获得一个可用的端口
// @receiver s
// @return randPort
// @return err
// @date 2021-03-06 22:19:30
func (s *dockerTheiaService) execGetAvailablePort() (randPort int, err error) {
	rand.Seed(time.Now().UnixNano())
loop:
	for {
		// 生成1024-65535的一个随机数
		randPort = rand.Intn(65535-1024) + 1024
		session, err := s.sshClient.NewSession()
		if err != nil {
			return -1, err
		}
		// 这里不处理error了，有输出会伴随错误
		output, err := session.CombinedOutput(fmt.Sprintf("netstat -an | grep :%d", randPort))
		_ = session.Close()
		if err != nil {
			return 0, err
		}
		// 端口可用
		if output == nil {
			break loop
		}
		// 端口不可用，继续拿新的端口
	}
	return randPort, nil
}

// execStopAndRemoveTheiaDocker 执行删除并移除容器
// @receiver s
// @params userId
// @params language
// @params labId
// @return err
// @date 2021-03-06 22:19:38
func (s *dockerTheiaService) execStopAndRemoveTheiaDocker(userId int, language string, labId int) (err error) {
	session, err := s.sshClient.NewSession()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer session.Close()
	// 删除容器
	if err = session.Run(fmt.Sprintf("docker stop mytheia-%s-%d-%d && docker rm mytheia-%s-%d-%d",
		language, userId, labId, language, userId, labId)); err != nil {
		return err
	}
	return nil
}

// execRunTheiaDocker 真正启动一个docker容器，以实验为单位去映射
// @receiver s
// @params userId
// @params language
// @params labId
// @return url
// @return err
// @date 2021-03-06 22:23:42
func (s *dockerTheiaService) execRunTheiaDocker(userId int, language string, labId int) (url string, err error) {
	// 需要上锁,不然并发下端口可能会有错
	s.applyPortMutex.MyLock()
	defer s.applyPortMutex.MyUnLock()
	// 得到可用端口
	port, err := s.execGetAvailablePort()
	if err != nil {
		return
	}
	var theiaType string
	if language == "web" {
		theiaType = "theiaide/theia"
	} else {
		theiaType = "theiaide/theia-" + language
	}
	session, err := s.sshClient.NewSession()
	if err != nil {
		return "", nil
	}
	//goland:noinspection GoUnhandledErrorResult
	defer session.Close()
	// 启动容器
	if err = session.Run(fmt.Sprintf(
		// 设端口
		"docker run -itd --init -p %d:3000 -u root "+
			// 工作目录挂载
			"-v %s:/home/project "+
			// 环境目录挂载
			"-v %s:/home/theia/.theia "+
			// 命名，例如mytheia-java-56-12,，56是userId,12是labId
			"--name=\"mytheia-%s-%d-%d\" "+
			// 语言版本
			"%s", port,
		path.Join(s.basePath, "students", strconv.Itoa(userId), fmt.Sprintf("workspace-%d", labId)),
		path.Join(s.basePath, "students", strconv.Itoa(userId), ".theia-"+language),
		language,
		userId,
		labId,
		theiaType)); err != nil {
		return "", err
	}
	//session1, err := s.sshClient.NewSession()
	//if err != nil {
	//	return "", err
	//}
	////goland:noinspection GoUnhandledErrorResult
	//defer session1.Close()
	//// 如果插件目录不存在,初始化插件目录，使用git下载
	//if err = session1.Run(fmt.Sprintf("if [ ! -d %s ] then git clone %s %s; fi;",
	//	path.Join(s.basePath, "students", strconv.Itoa(userId), ".theia-"+language, "extensions"),
	//	g.Cfg().GetString("gitExtensions"),
	//	path.Join(s.basePath, "students", strconv.Itoa(userId), ".theia-"+language, "extensions"))); err != nil {
	//	return "", err
	//}

	// 返回类似的118.178.253.239:3001
	url = fmt.Sprintf("%s:%d", g.Cfg().GetString("theia.docker.host"), port)
	// 把端口启动的theia状态存下来
	stat := &theiaState{
		Url:       url,
		StartTime: gtime.Now(),
	}
	if err = s.theiaPortCache.SetV(fmt.Sprintf("%s-%d-%d", language, userId, labId), stat); err != nil {
		return "", err
	}
	// 返回启动的地址
	return url, err
}

// clearAllIDE 清理所有已经启动的容器
// @receiver s
// @return error
// @date 2021-03-06 22:51:18
func (s *dockerTheiaService) clearAllIDE() error {
	session, err := s.sshClient.NewSession()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer session.Close()
	// 命令时格式化输出，用了go的template语法
	output, err := session.CombinedOutput(fmt.Sprintf("docker ps --format \"table {{.Names}}\""))
	if err != nil {
		return err
	}
	split := gstr.Split(string(output), "\n")
	for _, v := range split {
		if gstr.HasPrefix(v, "mytheia") {
			split1 := gstr.Split(v, "-")
			language := split1[0]
			userId := split1[1]
			labId := split1[2]
			if err = s.execStopAndRemoveTheiaDocker(gconv.Int(userId), language, gconv.Int(labId)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *dockerTheiaService) CollectCompilerErrorLog(req *model.SelectCompilerErrorLogReq) (resp *response.PageResp, err error) {
	records := make([]*model.CompilerErrorLogResp, 0)
	courseId, err := dao.Lab.WherePri(req.LabId).Cache(time.Hour).FindValue(dao.Lab.Columns.CourseId)
	if err != nil {
		return nil, err
	}
	// 所有选了这门课的学生id
	stuIds, count, err := dao.Course.ListUserIdByCourseId(req.PageCurrent, req.PageSize, courseId.Int())
	compilerErrorLogChan := make(chan *model.CompilerErrorLogResp, len(stuIds))
	// 开协程收集所有日志
	for _, v := range stuIds {
		go s.execCollectCompilerLog(v.Int(), compilerErrorLogChan, req.LabId)
	}

	// 收集结果
	for {
		select {
		case log := <-compilerErrorLogChan:
			records = append(records, log)
		}
		if len(records) == len(stuIds) {
			break
		}
	}

	// 装填学号
	stuDetail := make([]*model.SysUser, 0)
	if err = dao.SysUser.Where(stuIds).Fields(dao.SysUser.Columns.Num, dao.SysUser.Columns.UserId).Scan(&stuDetail); err != nil {
		return nil, err
	}
	for _, v := range records {
		for _, v1 := range stuDetail {
			if v.StuId == v1.UserId {
				v.StuNum = v1.Num
				break
			}
		}
	}
	return &response.PageResp{
		Records: records,
		PageInfo: &response.PageInfo{
			Size:    len(records),
			Total:   count,
			Current: req.PageCurrent,
			Pages:   int(math.Ceil(float64(count) / float64(req.PageSize))),
		}}, nil
}

func (s *dockerTheiaService) execCollectCompilerLog(stuId int, logChan chan *model.CompilerErrorLogResp, labId int) {
	session, err := s.sshClient.NewSession()
	if err != nil {
		return
	}
	//goland:noinspection GoUnhandledErrorResult
	defer session.Close()
	compilerLog := &model.CompilerErrorLogResp{
		StuId: stuId,
	}
	defer func() {
		logChan <- compilerLog
	}()
	// 第一个%s是basePath,第二个是%d是stuId,第三个是%d是labId
	output, err := session.CombinedOutput(fmt.Sprintf(" logPath = %s; if [ ! -e $logPath ] then cat $logPath ;fi;",
		path.Join(s.basePath, "students", strconv.Itoa(stuId), fmt.Sprintf("workspace-%d", labId), ".compilerError.log")))
	if err != nil || output == nil {
		return
	}
	compilerLog.CompilerLog = string(output)
	return
}
