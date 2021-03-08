// @Author: 陈健航
// @Date: 2021/3/3 18:44
// @Description:
package ide

import (
	"github.com/go-redsync/redsync/v4"
	"sync"
)

// 用于此处的lock,要看是用分布式锁还是单体锁
type iTheiaServiceLock interface {
	MyLock()
	MyUnLock()
}

// 单体锁
///////////////////////////////////////////////////////////////

type simpleLock struct {
	mu *sync.Mutex
}

func (s *simpleLock) MyLock() {
	s.mu.Lock()
}

func (s *simpleLock) MyUnLock() {
	s.mu.Unlock()
}

// redis 锁
/////////////////////////////////////////////////////////////

type redisLock struct {
	*redsync.Mutex
}

func (s *redisLock) MyLock() {
	_ = s.Lock()
}

func (s *redisLock) MyUnLock() {
	_, _ = s.Unlock()
}
