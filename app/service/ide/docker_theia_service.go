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
	role2 "code-platform/library/common/role"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/go-redsync/redsync/v4"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/html"
	"math"
	"math/rand"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"
)

type dockerTheiaService struct {
	sshClient             *ssh.Client
	runTheiaMutex         *redsync.Mutex
	redisTheiaCacheHeader string
	basePath              string
	redisUsedPortHeader   string
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
	s.runTheiaMutex = component.RedisLock.NewMutex("code.platform:ide:Port")
	s.redisTheiaCacheHeader = "code.platform:ide:ide.state:"
	s.basePath = g.Cfg().GetString("theia.docker.basePath")
	s.redisUsedPortHeader = "code.platform:ide:ide.usedPort"
	return s
}

func (s *dockerTheiaService) ListContianer() {
	panic("implement me")
}

func (s *dockerTheiaService) OpenIDE(req *model.OpenIDEReq) (url string, err error) {
	role, err := dao.SysUser.GetRoleById(req.UserId)
	if err != nil {
		return "", err
	}
	switch role {
	case role2.Teacher:
		port, err := s.getIDE(req.UserId, 0, false)
		if err != nil {
			return "", err
		}
		// 教师挂载的目录高一级，所以要特别处理
		url = fmt.Sprintf("%s:%d/#/home/project/%d/workspace-%d",
			g.Cfg().GetString("theia.docker.host"), port, req.UserId, req.LabId)
	case role2.Student:
		port, err := s.getIDE(req.UserId, req.LabId, true)
		if err != nil {
			return "", err
		}
		url = fmt.Sprintf("%s:%d/#/home/project/workspace-%d", g.Cfg().GetString("theia.docker.host"), port, req.LabId)
		// 打开的时候顺便创建提交记录的空条目,表示已经开始实验
	}
	if err = service.LabSummitService.InsertBlankRecord(req.UserId, req.LabId); err != nil {
		return "", err
	}
	return url, nil
}

func (s *dockerTheiaService) CheckCode(req *model.CheckCodeReq) (url string, err error) {
	languageEnum, err := getLanguageEnumByLabId(req.LabId)
	if err != nil {
		return "", err
	}
	port, err := s.getIDE(req.TeacherId, languageEnum, false)
	if err != nil {
		return "", err
	}
	url = fmt.Sprintf("%s:%d/#/home/project/%d/workspace-%d",
		g.Cfg().GetString("theia.docker.host"), port, req.StuId, req.LabId)
	return url, err
}

// CloseIDE 关闭IDE
// @receiver s
// @params userId 用户id
// @params languageEnum 语言枚举
// @return err
// @date 2021-03-06 22:09:22
func (s *dockerTheiaService) StopIDE(req *model.CloseIDEReq) (err error) {
	// 取出缓存
	_ = s.runTheiaMutex.Lock()
	//goland:noinspection GoUnhandledErrorResult
	defer s.runTheiaMutex.Unlock()
	// 查出所用语言
	languageEnum, err := getLanguageEnumByLabId(req.LabId)
	if err != nil {
		return err
	}
	stat, err := s.getTheiaStat(languageEnum, req.UserId)
	if err != nil {
		return err
	}
	// 存入编码时间
	duration := gtime.Now().Sub(stat.StartTime).Minutes()
	if err = service.LabSummitService.InsertCodingTime(req.LabId, int(duration), req.UserId); err != nil {
		return err
	}

	// 减少打开实例的个数
	if err = s.reduceDockerAndStat(req.UserId, languageEnum); err != nil {
		return err
	}
	return nil
}

func (s *dockerTheiaService) reduceDockerAndStat(languageEnum int, userId int) (err error) {
	// 移除docker
	if err = s.execStopAndRemoveTheiaDocker(languageEnum, userId); err != nil {
		return err
	}
	stat, err := s.getTheiaStat(languageEnum, userId)
	if err != nil {
		return err
	}
	stat.Count -= 1
	// 全部实例已经关闭
	if stat.Count == 0 {
		// 归还端口
		if err = s.returnPort(stat.Port); err != nil {
			return err
		}
		// 删除缓存
		if err = s.removeTheiaStat(gconv.Int(userId), gconv.Int(languageEnum)); err != nil {
			return err
		}
	} else {
		// 重新放回
		if err = s.setTheiaStat(languageEnum, userId, stat); err != nil {
			return err
		}
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
func (s *dockerTheiaService) getIDE(userId int, labId int, isStudent bool) (port int, err error) {
	ddl, err := dao.Lab.WherePri(labId).FindValue(dao.Lab.Columns.DeadLine)
	if err != nil {
		return 0, err
	}
	isNotExpire := true
	// 过了截止时间,将不可编辑
	if !ddl.IsNil() && gtime.Now().After(ddl.GTime()) {
		isNotExpire = false
	}
	//获得语言类型
	var languageEnum int
	if labId == 0 {
		languageEnum = 0
	} else {
		languageEnum, err = getLanguageEnumByLabId(labId)
		if err != nil {
			return 0, err
		}
	}

	// 上锁
	_ = s.runTheiaMutex.Lock()
	//goland:noinspection GoUnhandledErrorResult
	defer s.runTheiaMutex.Unlock()
	// 创建工作目录
	if err = s.execMkDir(path.Join(
		s.basePath,
		"codeSpace",
		strconv.Itoa(userId),
		fmt.Sprintf("workspace-%d", labId),
	)); err != nil {
		return 0, err
	}
	// 查看有没有没有关闭的容器
	stat, err := s.getTheiaStat(languageEnum, userId)
	if err != nil {
		return 0, err
	}
	// 未关闭，直接获得实例
	if stat != nil {
		// 刷新时间，增加计数，重新存
		port = stat.Port
		stat.StartTime = gtime.Now()
		stat.Count += 1
		if err = s.setTheiaStat(languageEnum, userId, stat); err != nil {
			return 0, err
		}
	} else {
		// 之前的已经关闭,重新开一个新的容器,并存入缓存
		port, err = s.execRunTheiaDocker(userId, languageEnum, isStudent, isNotExpire)
		if err != nil {
			return 0, err
		}
		stat = &model.TheiaStat{
			Port:      port,
			StartTime: gtime.Now(),
			Count:     1,
		}
		if err = s.setTheiaStat(languageEnum, userId, stat); err != nil {
			return 0, err
		}
		if err = s.addPort(port); err != nil {
			return 0, err
		}
	}
	return port, err
}

// clearIDE 清理"过期"的IDE
// @receiver s
// @date 2021-03-06 22:14:06
func (s *dockerTheiaService) ClearIDE() {
	_ = s.runTheiaMutex.Lock()
	//goland:noinspection GoUnhandledErrorResult
	defer s.runTheiaMutex.Unlock()
	r := g.Redis()
	v, _ := r.DoVar("KEYS", s.redisTheiaCacheHeader+"*")
	keys := v.Strings()
	// 启动协程检查
	wg := &sync.WaitGroup{}
	wg.Add(len(keys))
	for _, key := range keys {
		go func(key string) {
			defer wg.Done()
			v, err := r.DoVar("GET", key)
			if err != nil {
				glog.Debug("ClearIDE，获得key失败", err)
			}
			var stat = &model.TheiaStat{}
			if err = gconv.Struct(v, &stat); err != nil {
				glog.Debug("ClearIDE，转型失败", err)
			}
			split := gstr.Split(gstr.TrimStr(key, s.redisTheiaCacheHeader), ":")
			languageEnum := split[0]
			userId := split[1]
			// 打开容器已经超过24个小时仍未stop，可认为意料外的，直接stop
			if gtime.Now().Sub(stat.StartTime) > 24*time.Hour {
				if err := s.reduceDockerAndStat(gconv.Int(languageEnum), gconv.Int(userId)); err != nil {
					glog.Errorf("关闭容器失败:%s", err.Error())
				}
			}
		}(key)
	}
	wg.Wait()
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
		randPort = rand.Intn(2000) + 30000
		session, err := s.sshClient.NewSession()
		if err != nil {
			return -1, err
		}
		// 这里不处理error了，有输出会伴随错误
		cmd := fmt.Sprintf("netstat -an | grep %d", randPort)
		output, _ := session.CombinedOutput(cmd)
		_ = session.Close()
		if err != nil {
			return 0, err
		}
		// 端口可用,但可能被stop的容器使用
		if output == nil {
			// 端口未被stop的容器使用
			isUsed, err := s.isPortUsed(randPort)
			if err != nil {
				return 0, err
			}
			if !isUsed {
				// 可用，记录
				if err = s.addPort(randPort); err != nil {
					return 0, err
				}
				break loop
			}
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
func (s *dockerTheiaService) execStopAndRemoveTheiaDocker(languageEnum int, userId int) (err error) {
	session, err := s.sshClient.NewSession()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer session.Close()
	// 删除容器
	cmd := fmt.Sprintf("docker stop mytheia-%d-%d && docker rm mytheia-%d-%d",
		languageEnum, userId, languageEnum, userId)
	if err = session.Run(cmd); err != nil {
		return err
	}
	return nil
}

// execRunTheiaDocker 真正启动一个docker容器
// @receiver s
// @params userId
// @params language
// @params labId
// @return url
// @return err
// @date 2021-03-06 22:23:42
func (s *dockerTheiaService) execRunTheiaDocker(userId int, languageEnum int, isStudent bool, isEditable bool) (port int, err error) {
	// 得到可用端口
	port, err = s.execGetAvailablePort()
	if err != nil {
		return
	}
	// 镜像地址
	imageName := getImageName(languageEnum)
	session, err := s.sshClient.NewSession()
	if err != nil {
		return 0, nil
	}
	//goland:noinspection GoUnhandledErrorResult
	defer session.Close()
	var mountWorkSpace string
	// 如果是学生，只能挂载到自己的目录下
	if isStudent {
		mountWorkSpace = path.Join(s.basePath, "codeSpace", strconv.Itoa(userId))
	} else {
		mountWorkSpace = path.Join(s.basePath, "codeSpace")
	}
	var isEditableString string
	if isEditable {
		isEditableString = "-u root"
	} else {
		isEditableString = ""
	}
	cmd := fmt.Sprintf(
		// 设端口
		"docker run -itd --init -p %d:3000 "+
			"%s "+
			"-v %s:/home/project "+
			// 工作目录挂载
			"-v %s:%s "+
			// 环境目录挂载
			"--parse=\"mytheia-%d-%d\" "+
			// 命名，例如mytheia-java-56-12,，56是userId,12是labId
			"%s",
		// 语言版本
		port,
		isEditableString,
		mountWorkSpace,
		// 主机上的工作目录
		path.Join(s.basePath, "codeSpace", strconv.Itoa(userId), ".theia"),
		// ide的环境目录
		getEnvironmentMount(languageEnum),
		// docker里的环境目录
		languageEnum, userId,
		// 名字里用于做标识
		imageName,
		// ide的名称
	)
	// 启动容器
	if err = session.Run(cmd); err != nil {
		return 0, err
	}
	return port, err
}

func (s *dockerTheiaService) CollectCompilerErrorLog(req *model.SelectCompilerErrorLogReq) (resp *response.PageResp, err error) {
	records := make([]*model.CompilerErrorLogResp, 0)
	courseId, err := dao.Lab.WherePri(req.LabId).Cache(time.Hour).FindValue(dao.Lab.Columns.CourseId)
	if err != nil {
		return nil, err
	}
	// 所有选了这门课的学生id
	stuIds, err := dao.Course.ListUserIdByCourseId(courseId.Int())
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
	d := dao.SysUser.Where(stuIds)
	if err = d.Fields(dao.SysUser.Columns.Num, dao.SysUser.Columns.UserId).Scan(&stuDetail); err != nil {
		return nil, err
	}

	count, err := d.Count()
	if err != nil {
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

func (s *dockerTheiaService) execMkDir(dir string) (err error) {
	session, err := s.sshClient.NewSession()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer session.Close()
	cmd := fmt.Sprintf("if [ ! -d %s ]; then mkdir -p %s; fi;", dir, dir)
	if err = session.Run(cmd); err != nil {
		return err
	}
	return nil
}

func (s *dockerTheiaService) setTheiaStat(languageEnum int, userId int, value *model.TheiaStat) (err error) {
	r := g.Redis()
	key := fmt.Sprintf("%s%d:%d", s.redisTheiaCacheHeader, languageEnum, userId)
	if _, err = r.Do("SET", key, value); err != nil {
		return err
	}
	return nil
}

func (s *dockerTheiaService) getTheiaStat(languageEnum int, userId int) (stat *model.TheiaStat, err error) {
	r := g.Redis()
	key := fmt.Sprintf("%s%d:%d", s.redisTheiaCacheHeader, languageEnum, userId)
	v, err := r.DoVar("GET", key)
	if err != nil {
		return nil, err
	}
	if v.IsNil() {
		return nil, nil
	}
	stat = &model.TheiaStat{}
	if err = v.Struct(&stat); err != nil {
		return nil, err
	}
	return stat, err
}

func (s *dockerTheiaService) returnPort(port int) (err error) {
	r := g.Redis()
	// 归还端口
	if _, err = r.Do("SREM", s.redisUsedPortHeader, port); err != nil {
		return err
	}
	return nil
}

func (s *dockerTheiaService) isPortUsed(port int) (ret bool, err error) {
	r := g.Redis()
	isUsed, err := r.DoVar("SISMEMBER", s.redisUsedPortHeader, port)
	if err != nil {
		return false, err
	}
	return isUsed.Int() == 1, nil
}

func (s *dockerTheiaService) addPort(port int) (err error) {
	// 存入端口
	r := g.Redis()
	if _, err = r.Do("SADD", s.redisUsedPortHeader, port); err != nil {
		return err
	}
	return nil
}

func (s *dockerTheiaService) removeTheiaStat(languageEnum int, userId int) (err error) {
	r := g.Redis()
	// 删除stat
	key := fmt.Sprintf("%s%d:%d", s.redisTheiaCacheHeader, gconv.Int(userId), gconv.Int(languageEnum))
	if _, err = r.Do("DEL", key); err != nil {
		glog.Debug("ClearIDE，删除stat失败")
	}
	return err
}

func (s dockerTheiaService) plagiarismCheck(labId int) (resp []*model.PlagiarismCheckResp, err error) {
	session, err := s.sshClient.NewSession()
	if err != nil {
		return nil, nil
	}
	//goland:noinspection GoUnhandledErrorResult
	defer session.Close()
	// 找出课程id
	courseId, err := dao.Lab.WherePri(labId).FindValue(dao.Lab.Columns.CourseId)
	if err != nil {
		return nil, nil
	}
	// 找出userId
	userIds, err := dao.Course.ListUserIdByCourseId(courseId.Int())
	if err != nil {
		return nil, err
	}
	// 找出语言类型
	languageEnum, err := dao.Course.WherePri(courseId).FindValue(dao.Course.Columns.Language)
	if err != nil {
		return nil, err
	}
	// 组织cmd
	var cmd string
	switch languageEnum.Int() {
	case 1:
		language := "cc"
		ext := "cpp"
		cmd = fmt.Sprintf("%s/moss -l %s -d ", s.basePath, language)
		for _, v := range userIds {
			cmd += path.Join(s.basePath, "codeSpace", v.String(), fmt.Sprintf("workspace-%d", labId), fmt.Sprintf("*.%s", ext)) + " "
		}
	case 2:
		language := "java"
		ext := "java"
		cmd = fmt.Sprintf("%s/moss -l %s -d ", s.basePath, language)
		for _, v := range userIds {
			cmd += path.Join(s.basePath, "codeSpace", v.String(), fmt.Sprintf("workspace-%d", labId), fmt.Sprintf("*.%s", ext)) + " "
		}
	case 3:
		language := "python"
		ext := "py"
		cmd = fmt.Sprintf("%s/moss -l %s -d ", s.basePath, language)
		for _, v := range userIds {
			cmd += path.Join(s.basePath, "codeSpace", v.String(), fmt.Sprintf("workspace-%d", labId), fmt.Sprintf("*.%s", ext)) + " "
		}
	default:
		return nil, code.UnSupportLanguageTypeError
	}
	// 执行
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return nil, nil
	}
	// 找出查重后的链接
	strings := gstr.Split(gconv.String(output), "\n")
	var url string
	for _, v := range strings {
		if gstr.HasPrefix(v, "http") {
			url = v
			break
		}
	}
	// 解析结果
	resp, err = s.parsePlagiarismCheck(url)
	if err != nil {
		return nil, err
	}
	// 装填字段
	userDetail := make([]*model.SysUser, 0)
	if err = dao.SysUser.WherePri(userIds).
		Fields(dao.SysUser.Columns.RealName, dao.SysUser.Columns.UserId, dao.SysUser.Columns.Num).
		FindScan(&userDetail); err != nil {
		return nil, nil
	}
	for _, v := range resp {
		count := 0
		for _, v1 := range userDetail {
			if v.UserId1 == v1.UserId {
				v.RealName1 = v1.RealName
				v.Num1 = v1.Num
				count += 1
			}
			if v.UserId2 == v1.UserId {
				v.RealName2 = v1.RealName
				v.Num2 = v1.Num
				count += 1
			}
			if count == 2 {
				break
			}
		}
	}
	return resp, err
}

func (s *dockerTheiaService) parsePlagiarismCheck(url string) (resp []*model.PlagiarismCheckResp, err error) {
	htmlResp, err := http.Get("url")
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer htmlResp.Body.Close()
	root, err := htmlquery.Parse(htmlResp.Body)
	if err != nil {
		return nil, err
	}
	tr := htmlquery.Find(root, "/html/body/table/tbody/tr[*]")
	resp = make([]*model.PlagiarismCheckResp, 0)
	resChannel := make(chan *model.PlagiarismCheckResp, len(tr)-1)
	for i, v := range tr {
		if i == 0 {
			continue
		}
		// 开多个协程爬取信息
		go func(node *html.Node, index int, resChannel chan *model.PlagiarismCheckResp) {
			detailUrl := node.FirstChild.FirstChild.Attr[0].Val
			dir1 := node.FirstChild.FirstChild.FirstChild.Data
			dir2 := node.FirstChild.FirstChild.FirstChild.Data
			userId1 := gstr.SubStr(gstr.TrimStr(dir1, path.Join(s.basePath, "codeSpace")), 1, gstr.PosR(gstr.TrimStr(dir1, path.Join(s.basePath, "codeSpace")), "/")-1)
			userId2 := gstr.SubStr(gstr.TrimStr(dir2, path.Join(s.basePath, "codeSpace")), 1, gstr.PosR(gstr.TrimStr(dir2, path.Join(s.basePath, "codeSpace")), "/")-1)
			similarity := gstr.SubStr(dir1, gstr.PosI(dir1, "(")+1, gstr.PosI(dir1, "%")-gstr.PosI(dir1, "(")-1)
			ret := &model.PlagiarismCheckResp{
				UserId1:    gconv.Int(userId1),
				UserId2:    gconv.Int(userId2),
				Similarity: gconv.Int(similarity),
			}
			htmlResp1, err1 := http.Get(gstr.Replace(url, fmt.Sprintf("match%d.html", index-1), fmt.Sprintf("match%d-0.html", index-1)))
			if err1 != nil {
				return
			}

			root1, err1 := htmlquery.Parse(htmlResp1.Body)
			if err1 != nil {
				return
			}
			ret.Code1 = htmlquery.OutputHTML(root1, false)
			err1 = htmlResp1.Body.Close()
			if err1 != nil {
				return
			}

			htmlResp1, err1 = http.Get(gstr.Replace(detailUrl, fmt.Sprintf("match%d.html", index-1), fmt.Sprintf("match%d-1.html", index-1)))
			if err1 != nil {
				return
			}
			root1, err1 = htmlquery.Parse(htmlResp1.Body)
			if err1 != nil {
				return
			}
			ret.Code2 = htmlquery.OutputHTML(root1, false)
			err1 = htmlResp1.Body.Close()
			if err1 != nil {
				return
			}
			// 通过channel返回
			resChannel <- ret
		}(v, i, resChannel)
	}

	for i := 0; i < len(tr)-1; i++ {
		select {
		case ret := <-resChannel:
			resp = append(resp, ret)
		}
	}
	return resp, nil
}
