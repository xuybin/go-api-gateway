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

		err:=insertData(connStr)
		if err!=nil{
			panic(err)
		}else {
			return enforcer
		}
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
		//_, err = engine.Exec("INSERT  INTO `"+dbName+"`.`casbin_rule`(`p_type`,`v0`,`v1`,`v2`) VALUES ('p','admin','/policy/*','(GET)|(POST)|(PUT)|(DELETE)')")
	}else {
		err=fmt.Errorf("dataSourceName:%s doesn't exist dbName",dataSourceName)
	}
	return
}

func insertData(dataSourceName string) (err error) {
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
		engine.Exec("DELETE FROM `"+dbName+"`.`casbin_rule` WHERE `p_type`='p' AND `v0`='admin' AND `v1`='/policy/*' AND `v2`='(GET)|(POST)|(PUT)|(DELETE)|(HEAD)' AND `v3`IS NULL AND `v4`IS NULL AND `v5`IS NULL")
		_, err = engine.Exec("INSERT  INTO `"+dbName+"`.`casbin_rule`(`p_type`,`v0`,`v1`,`v2`) VALUES ('p','admin','/policy/*','(GET)|(POST)|(PUT)|(DELETE)|(HEAD)')")
	}else {
		err=fmt.Errorf("dataSourceName:%s doesn't exist dbName",dataSourceName)
	}
	return
}