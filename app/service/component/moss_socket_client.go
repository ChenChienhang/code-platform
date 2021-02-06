// @Author: 陈健航
// @Date: 2021/1/10 21:14
// @Description: stanford moss代码查重的客户端，根据脚本写的，原脚本：http://moss.stanford.edu/general/scripts/mossnet
// 可直接使用，不建议修改
package component

import (
	"errors"
	"fmt"
	"github.com/gogf/gf/container/garray"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/gtcp"
	"github.com/gogf/gf/os/glog"
	"io/ioutil"
	"net/url"
	"strings"
)

type Stage int

const (
	disconnected Stage = iota
	awaitingInitialization
	awaitingLanguage
	awaitingFiles
	awaitingQuery
	awaitingResults
	awaitingEnd
)

type MossClient struct {
	currentStage       Stage
	addr               string
	userID             string
	language           string
	setID              int
	optM               int64
	optD               int
	optX               int
	optN               int
	optC               string
	ResultURL          *url.URL
	supportedLanguages *garray.StrArray
	conn               *gtcp.Conn
}

// NewMossClient 构造函数
// @params language 语言
// @params userId 用户id
// @return *MossClient
// @return error
// @date 2021-01-11 17:03:26
func NewMossClient(language, userId string) (*MossClient, error) {
	supportedLanguages := garray.NewStrArrayFrom(g.SliceStr{"c", "cc", "java", "ml", "pascal", "ada", "lisp", "schema", "haskell", "fortran",
		"ascii", "vhdl", "perl", "matlab", "python", "mips", "prolog", "spice", "vb", "csharp", "modula2", "a8086",
		"javascript", "plsql"})
	if supportedLanguages.Contains(language) {
		return &MossClient{
			currentStage:       disconnected,
			setID:              1,
			optM:               10,
			optD:               1,
			optX:               0,
			optN:               250,
			optC:               "",
			supportedLanguages: supportedLanguages,
			addr:               "moss.stanford.edu:7690",
			userID:             userId,
			language:           language,
		}, nil
	} else {
		return nil, errors.New("MOSS Server does not recognize this programming language")
	}
}

// Close 关闭
// @receiver c
// @return error
// @date 2021-01-11 17:04:12
func (c *MossClient) Close() error {
	c.currentStage = disconnected
	if err := c.sendCommand("end\n"); err != nil {
		return err
	}
	if err := c.conn.Close(); err != nil {
		return err
	}
	return nil
}

// connect 连接斯坦福的moss
// @receiver c
// @return error
// @date 2021-01-11 22:38:17
func (c *MossClient) connect() error {
	if c.currentStage != disconnected {
		return errors.New("client is already connected")
	} else {
		c.conn, _ = gtcp.NewConn(c.addr)
		c.currentStage = awaitingInitialization
	}
	return nil
}

// Run 启动
// @receiver c
// @return error
// @date 2021-01-11 22:38:28
func (c *MossClient) Run() error {
	if err := c.connect(); err != nil {
		return err
	}
	if err := c.sendInitialization(); err != nil {
		return err
	}
	if err := c.sendLanguage(); err != nil {
		return err
	}
	return nil
}

// sendCommand 发送命令行
// @receiver c
// @params objects
// @return error
// @date 2021-01-11 22:38:40
func (c *MossClient) sendCommand(objects ...interface{}) error {
	commandStrings := make([]string, 0, len(objects))

	for var5 := 0; var5 < len(objects); var5++ {
		o := objects[var5]
		s := fmt.Sprintf("%v", o)
		//s := o.(string)
		commandStrings = append(commandStrings, s)
	}
	if err := c.sendCommandStrings(commandStrings); err != nil {
		return err
	}
	return nil
}

// sendCommandStrings 发送命令指令序列
// @receiver c
// @params stringSlice
// @return error
// @date 2021-01-11 22:38:52
func (c *MossClient) sendCommandStrings(stringSlice []string) error {
	if len(stringSlice) > 0 {
		//slice转字符串,空格分隔
		s := strings.Join(stringSlice, " ")
		s += "\n"

		if err := c.conn.Send([]byte(s)); err != nil {
			return errors.New("failed to send command: " + err.Error())
		}
		return nil
	} else {
		return errors.New("failed to send command because it was empty")
	}
}

// sendInitialization 输出化序列值
// @receiver c
// @return error
// @date 2021-01-11 22:39:11
func (c *MossClient) sendInitialization() error {
	if c.currentStage != awaitingInitialization {
		return errors.New("cannot send initialization. Client is either already initialized or not connected yet")
	}
	if err := c.sendCommand("moss", c.userID); err != nil {
		return nil
	}
	if err := c.sendCommand("directory", c.optD); err != nil {
		return nil
	}
	if err := c.sendCommand("X", c.optX); err != nil {
		return nil
	}
	if err := c.sendCommand("maxmatches", c.optM); err != nil {
		return nil
	}
	if err := c.sendCommand("show", c.optN); err != nil {
		return nil
	}
	c.currentStage = awaitingLanguage
	return nil
}

// sendLanguage 发送语言
// @receiver c
// @return error
// @date 2021-01-11 22:39:27
func (c *MossClient) sendLanguage() error {
	if c.currentStage != awaitingLanguage {
		return errors.New("language already sent or client is not initialized yet")
	}
	if err := c.sendCommand("language", c.language); err != nil {
		return err
	}
	serverString, err := c.conn.RecvLine()
	if err != nil {
		return err
	}
	if len(serverString) > 0 && string(serverString) == "yes" {
		c.currentStage = awaitingFiles
	} else {
		return errors.New("MOSS Server does not recognize this programming language")
	}
	return nil
}

// SendQuery 查询结果
// @receiver c
// @return error
// @date 2021-01-11 22:40:00
func (c *MossClient) SendQuery() error {
	if c.currentStage != awaitingQuery {
		return errors.New("cannot send query at this time. Connection is either not initialized or already closed")
	} else if c.setID == 1 {
		return errors.New("you did not upload any files yet")
	} else {
		if err := c.sendCommand(fmt.Sprintf("%s %d %s", "query", 0, c.optC)); err != nil {
			return nil
		}
		c.currentStage = awaitingResults
		result, err := c.conn.RecvLine()
		if err != nil {
			return err
		}
		if len(result) > 0 && strings.HasPrefix(strings.ToLower(string(result)), "http") {
			if c.ResultURL, err = url.Parse(strings.Trim(string(result), " ")); err != nil {
				return err
			}
			c.currentStage = awaitingEnd
		} else {
			return errors.New("MOSS submission failed. The server did not return a valid URL with detection results")
		}
	}
	return nil
}

// UploadFile 上传代码源文件
// @receiver c
// @params filePath
// @params isBaseFile
// @return error
// @date 2021-01-11 22:40:09
func (c *MossClient) UploadFile(filePath string, isBaseFile bool) error {
	if c.currentStage != awaitingFiles && c.currentStage != awaitingQuery {
		return errors.New("cannot upload file. Client is either not initialized properly or the connection is already closed")
	}
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	setID := 0
	if !isBaseFile {
		setID = c.setID
		c.setID++
	}
	filename := strings.ReplaceAll(filePath, "\\", "/")
	uploadString := fmt.Sprintf("file %d %s %d %s\n", setID, c.language, len(fileBytes), filename)
	glog.Info("mossClient uploading file: %s", filename)
	if err = c.conn.Send([]byte(uploadString)); err != nil {
		return err
	}
	if err = c.conn.Send(fileBytes); err != nil {
		return err
	}
	c.currentStage = awaitingQuery
	return nil
}
