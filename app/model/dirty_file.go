// @Author: 陈健航
// @Date: 2021/2/28 15:35
// @Description:
package model

import (
	"time"
)

// 脏文件
type DirtyFile struct {
	Url        string
	CreateTime time.Time
}
