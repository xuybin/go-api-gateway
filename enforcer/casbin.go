package enforcer

import (
	"github.com/casbin/casbin"
	"github.com/casbin/xorm-adapter"
	"github.com/go-xorm/xorm"
	"strings"
	"fmt"
)

func NewCasbinEnforcer(connStr string) *casbin.Enforcer {
	err:=createMysqlDatabase(connStr)
	if err!=nil{
		panic(err)
	}else {
		Adapter := xormadapter.NewAdapter("mysql", connStr, true)
		enforcer := casbin.NewEnforcer(casbin.NewModel(CasbinConf), Adapter)
		return enforcer
	}
}

func createMysqlDatabase( dataSourceName string) (err error) {
	result:=strings.LastIndex(dataSourceName,"/")
	if result >= 0 && result+1<len(dataSourceName){
		var engine *xorm.Engine
		dbName:=string([]byte(dataSourceName)[result+1:len(dataSourceName)])
		dataSourceName= string([]byte(dataSourceName)[0:result+1])
		engine, err = xorm.NewEngine("mysql", dataSourceName)

		if err != nil {
			return err
		}
		defer engine.Close()
		_, err = engine.Exec("CREATE DATABASE IF NOT EXISTS "+dbName)

	}else {
		err=fmt.Errorf("dataSourceName:%s doesn't exist dbName",dataSourceName)
	}
	return
}

func SubString(str string,begin,length int) (substr string) {
	// 将字符串的转换成[]rune
	rs := []rune(str)
	lth := len(rs)

	// 简单的越界判断
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length
	if end > lth {
		end = lth
	}

	// 返回子串
	return string(rs[begin:end])
}