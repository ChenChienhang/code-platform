// @Author: 陈健航
// @Date: 2021/2/19 22:06
// @Description:
package component

import (
	"fmt"
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/gogf/gf/frame/g"
)

var RedisLock = newRedisLock()

func newRedisLock() (r *redsync.Redsync) {
	if g.Cfg().GetBool("server.Multiple") {
		client := goredislib.NewClient(&goredislib.Options{
			Addr:     fmt.Sprintf("%s:%s", g.Cfg().GetString("redis.host"), g.Cfg().GetString("redis.port")),
			Password: g.Cfg().GetString("redis.pass"),
		})
		pool := goredis.NewPool(client)
		return redsync.New(pool)
	} else {
		return nil
	}
}
