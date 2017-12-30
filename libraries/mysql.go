package libraries

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

var o orm.Ormer

type Mysql struct {
}

//mysql结构
type Mysql_columns struct {
	Name        string
	Sql_type    string
	Null        string
	Sql_default interface{}
	Primary     bool
	Autoinc     bool
}

type Transaction struct {
	Connect orm.Ormer
}

/*执行select专用
 *返回数据结构模式[]map[string]string
 */
func (this *Mysql) Select(sql string, master string, t *Transaction) (maps []orm.Params, err error) {
	if master != "slave" && master != "default" {
		master = "default"
	}
	s := o
	if t != nil && t.Connect != nil && master == "default" {
		s = t.Connect
	} else {
		s.Using(master)
	}
	_, err = s.Raw(sql).Values(&maps)
	return
}

/*执行sql语句
 *返回新增ID和error
 *result包含LastInsertId()与RowsAffected()方法
 */
func (this *Mysql) Insert(sql string, master string, t *Transaction) (LastInsertId int64, err error) {
	if master != "slave" && master != "default" {
		master = "default"
	}
	s := o
	if t != nil && t.Connect != nil {
		s = t.Connect
	} else {
		s.Using(master)
	}
	res, err := s.Raw(sql).Exec()
	if err == nil {
		LastInsertId, _ = res.LastInsertId()
	}
	return
}

/*执行sql语句
 *返回error
 *result包含LastInsertId()与RowsAffected()方法
 */
func (this *Mysql) Query(sql string, master string, t *Transaction) (result bool, err error) {
	if master != "slave" && master != "default" {
		master = "default"
	}
	s := o
	if t != nil && t.Connect != nil {
		s = t.Connect
	} else {
		s.Using(master)
	}
	_, err = o.Raw(sql).Exec()
	result = false
	if err == nil {
		result = true
	}
	return
}

//执行语句并取受影响行数
func (this *Mysql) Query_getaffected(sql string, master string, t *Transaction) (num int64, err error) {
	if master != "slave" && master != "default" {
		master = "default"
	}
	s := o
	if t != nil && t.Connect != nil {
		s = t.Connect
	} else {
		s.Using(master)
	}
	res, err := o.Raw(sql).Exec()
	if err == nil {
		num, _ = res.RowsAffected()
	}
	return
}

//列出所有表
func (this *Mysql) ShowTables(master string) (list orm.ParamsList) {
	if master != "slave" && master != "default" {
		master = "default"
	}
	s := o
	s.Using(master)
	sql := "SHOW TABLES"
	s.Raw(sql).ValuesFlat(&list)
	return
}

//列出表结构
func (this *Mysql) ShowColumns(table string, master string) map[string]Mysql_columns {
	sql := "SHOW COLUMNS FROM `" + table + "`"
	result, err := this.Select(sql, master, new(Transaction))
	Errorlog(err, "初始化错误，无法列出表结构")
	re := make(map[string]Mysql_columns)
	for _, tmp := range result {
		re[tmp["Field"].(string)] = Mysql_columns{Name: tmp["Field"].(string), Sql_type: tmp["Type"].(string), Null: tmp["Null"].(string), Sql_default: tmp["Default"], Primary: (tmp["Key"].(string) == "PRI"), Autoinc: (tmp["Extra"].(string) == "auto_increment")}
	}
	return re
}

//开始事务
func (this *Mysql) Begin() (re orm.Ormer) {
	re = orm.NewOrm()
	re.Begin()
	return
}

//提交事务
func (this *Mysql) Commit(t *Transaction) error {
	return t.Connect.Commit()
}

//回滚事务
func (this *Mysql) Rollback(t *Transaction) error {
	return t.Connect.Rollback()
}

func (this *Mysql) Mysql_init(db_config map[string]string) {
	maxIdle := 30 //最大空闲链接
	maxConn := 50 //最大连接数
	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", db_config["db.master.user"]+":"+db_config["db.master.pwd"]+"@tcp("+db_config["db.master.host"]+":"+db_config["db.master.port"]+")/"+db_config["db.master.name"]+"?charset="+db_config["db.master.charset"], maxIdle, maxConn)
	orm.RegisterDataBase("slave", "mysql", db_config["db.slave.user"]+":"+db_config["db.slave.pwd"]+"@tcp("+db_config["db.slave.host"]+":"+db_config["db.slave.port"]+")/"+db_config["db.slave.name"]+"?charset="+db_config["db.slave.charset"], maxIdle, maxConn)
	o = orm.NewOrm()
}
