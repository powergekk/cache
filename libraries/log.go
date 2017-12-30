package libraries

import (
	"github.com/astaxie/beego/logs"
	"time"
)

type Err_log struct {
	Err       string
	Err_func  string
	Err_param string
}

func Writelog(content string) {
	log := logs.NewLogger()
	log.Async()
	log.Async(1e3)
	log.SetLogger(logs.AdapterFile, `{"filename":"run.log"}`)
	log.Info(content)
}

/*写err log
 *参数一，写入字串符
 *参数二，指定文件名
 */
func Write_errorlog(content string, filename string) {
	log := logs.NewLogger()
	timestamp := time.Now().Unix()
	log.SetLogger(logs.AdapterFile, `{"filename":"log/`+time.Unix(timestamp, 0).Format("2006-01-02")+`_`+filename+`.log"}`)
	log.Error(content)
}

/*检查err
 *参数二，附加备注信息（可省略）
 *参数三，指定文件名（可省略）
 */
func Errorlog(err error, param ...string) {
	if err != nil {
		switch len(param) {
		case 0:
			param = append(param, "", "")
		case 1:
			param[0] = `
		错误备注:` + param[0]
			param = append(param, "")
		}
		content := "	错误信息:" + err.Error() + param[0]
		Write_errorlog(content, param[1])
	}
}

/*把err添加到切片返回
 *参数一，传入errs切片
 *参数二，传入err
 *参数三，传入出错方法(备注)
 *参数四，出错函数执行的函数(可选)
 */
func Adderr(errs *[]Err_log, err error, err_func string, param map[string]interface{}) {

	if err != nil {
		if err.Error() == "" {
			if len(*errs) > 0 {
				*errs = append(*errs, Err_log{Err: err.Error(), Err_func: err_func, Err_param: Json_pack(param)})
			}
		} else {
			*errs = append(*errs, Err_log{Err: err.Error(), Err_func: err_func, Err_param: Json_pack(param)})
		}

	}
}

//展开errs保存到日志
func Save_errlog(errs []Err_log, uri string, param map[string][]string) {
	if len(errs) > 0 {
		Write_errorlog(Msgpack_pack(map[string]interface{}{"errs": errs, "uri": uri, "param": param}), "errs")
	}

}
func init() {

}
