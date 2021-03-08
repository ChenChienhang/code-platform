// @Author: 陈健航
// @Date: 2021/3/3 18:47
// @Description:
package ide

import (
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"time"
)

type iTheiaServiceMap interface {
	SetV(key string, value interface{}) (err error)
	GetV(key string) (value interface{}, err error)
	GetTimeOut() (removeIDE []*gmap.StrStrMap)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type theiaServiceRedisMap struct {
	redisHeader string
}

func (u *theiaServiceRedisMap) GetTimeOut() (removeIDE []*gmap.StrStrMap) {
	r := g.Redis()
	// 找出所有的key
	v, err := r.DoVar("KEYS", u.redisHeader+"*")
	if err != nil {
		glog.Debug("redis获得keys失败")
	}
	if v == nil {
		return
	}
	keys := v.Strings()
	removeIDE = make([]*gmap.StrStrMap, 0)
	for _, key := range keys {
		v, err := r.DoVar("GET", key)
		if err != nil {
			glog.Debug("redis获得key失败")
		}
		var state = theiaState{}
		if err = gconv.Struct(v, &state); err != nil {
			glog.Debug("redis转型失败")
		}
		// 打开容器已经超过24个小时仍未结束，，可认为意料外的，直接清理
		if (state.EndTime == nil && gtime.Now().Sub(state.StartTime) > 24*time.Hour) ||
			// 距离关闭容器已经超过5分钟，可以直接把容器关了
			gtime.Now().Sub(state.EndTime) > 5*time.Minute {
			split := gstr.Split(gstr.TrimStr(key, u.redisHeader), "-")
			language := split[0]
			userId := split[1]
			labId := split[2]
			m := gmap.NewStrStrMap(true)
			m.Set("userId", userId)
			m.Set("language", language)
			m.Set("labId", labId)
			removeIDE = append(removeIDE, m)
		}
	}
	return removeIDE
}

func (u *theiaServiceRedisMap) SetV(key string, value interface{}) (err error) {
	r := g.Redis()
	if _, err = r.Do("SET", u.redisHeader+key, value); err != nil {
		return err
	}
	return nil
}

func (u *theiaServiceRedisMap) GetV(key string) (value interface{}, err error) {
	r := g.Redis()
	v, err := r.DoVar("GET", u.redisHeader+key)
	if err != nil {
		return nil, err
	}
	if v.IsNil() {
		return nil, nil
	}
	return v.Interface(), nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type theiaServiceSimpleMap struct {
	m *gmap.StrAnyMap
}

func (u *theiaServiceSimpleMap) GetTimeOut() (removeIDE []*gmap.StrStrMap) {
	keys := u.m.Keys()
	removeIDE = make([]*gmap.StrStrMap, 0)
	for _, key := range keys {
		v := u.m.Get(key)
		state := theiaState{}
		if err := gconv.Struct(v, &state); err != nil {
			glog.Errorf("RemoveTimeOut转型失败：%s", err.Error())
		}
		// 打开容器已经超过24个小时仍未结束，，可认为意料外的，直接清理
		if (state.EndTime == nil && gtime.Now().Sub(state.StartTime) > 24*time.Hour) ||
			// 距离关闭容器已经超过一个小时，可以直接把容器关了
			gtime.Now().Sub(state.EndTime) > time.Hour {
			split := gstr.Split(key, "-")
			language := split[0]
			userId := split[1]
			labId := split[2]
			m := gmap.NewStrStrMap(true)
			m.Set("userId", userId)
			m.Set("language", language)
			m.Set("labId", labId)
			removeIDE = append(removeIDE, m)
		}
	}
	return removeIDE
}

func (u *theiaServiceSimpleMap) SetV(key string, value interface{}) (err error) {
	u.m.Set(key, value)
	return nil
}

func (u *theiaServiceSimpleMap) GetV(key string) (value interface{}, err error) {
	v := u.m.Get(key)
	return v, nil
}
