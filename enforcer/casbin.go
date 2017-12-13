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
		_, err = engine.Exec("INSERT  INTO `"+dbName+"`.`casbin_rule`(`p_type`,`v0`,`v1`,`v2`) VALUES ('p','admin','/policy/*','(GET)|(POST)|(PUT)|(DELETE)')")

	}else {
		err=fmt.Errorf("dataSourceName:%s doesn't exist dbName",dataSourceName)
	}
	return
}