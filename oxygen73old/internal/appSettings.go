package internal

import (
	"encoding/json"
	"fmt"
	"github.com/lxn/walk"
	"path/filepath"
	"strconv"
	"time"
)

type AppSetting struct {
	*walk.IniFileSettings
}

func (x *AppSetting) GetTime(key string, defaultValue time.Time) time.Time {
	ret := defaultValue
	if value, ok := x.Get(key); ok {
		t, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			logger.Printf("error %s, get time value [%s]: %s",
				filepath.Base(x.FilePath()),
				key, err.Error())

		} else {
			ret = t
		}
	}
	return ret
}

func (x *AppSetting) GetInt(key string, defaultValue int) int {
	if s, ok := x.Get(key); ok {
		if v, e := strconv.Atoi(s); e == nil {
			return v
		}
	}
	return defaultValue
}

func (x *AppSetting) PutTime(key string, t time.Time) {
	err := x.Put(key, t.Format(time.RFC3339Nano))
	if err != nil {
		logger.Panicf("%s, put time value [%s] <- %v: %s",
			filepath.Base(x.FilePath()),
			key, t, err.Error())
	}

}

func (x *AppSetting) GetJson(key string, v interface{}) error {
	s, found := x.Get(key)
	if !found {
		return fmt.Errorf("setting, json, key not found: %s", key)
	}
	return json.Unmarshal([]byte(s), v)
}

func (x *AppSetting) PutJson(key string, v interface{}) {
	bts, err := json.Marshal(v)
	if err != nil {
		logger.Panic(err)
	}
	err = x.Put(key, string(bts))
	if err != nil {
		logger.Panic(err)
	}
}
