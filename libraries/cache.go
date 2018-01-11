package libraries

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Hashvalue struct { //缓存结构
	value sync.Map
	time  int64  //time值说明，为0表示结果为空，为-1表示永久缓存，为正值表示以时间戳为到期时间,-2为删除key,-3为删除patch
	patch string //本条缓存所在的patch
	key   string //本条缓存所在的key
}

func (this *Hashvalue) Load(key interface{}) interface{} {
	result, _ := this.value.Load(key)
	return result
}
func (this *Hashvalue) Len() (length int) {
	this.value.Range(func(k, v interface{}) bool {
		length += 1
		return true
	})
	return
}
func (this *Hashvalue) Range(f func(interface{}, interface{}) bool) {
	this.value.Range(f)
}

/**
 *使用Hset,Hset_r可以保存本地文件持久化
 *使用Store方法，可以临时保存内容到缓存，重启进程失效
 **/
func (this *Hashvalue) Store(key interface{}, value interface{}) {
	_, ok := this.value.LoadOrStore(key, value)
	if !ok {
		//临时写入以保证数据存在，如果ok是false，就是sync.Map里面没有这个值，直接进行Store，下次在读取都会一直是nil与false
		do_hash(this.key, map[string]interface{}{key.(string): value}, this.patch, 0, "hash_set")
	} else {
		this.value.Store(key, value)
	}
}
func (this *Hashvalue) Delete(key interface{}) {
	this.value.Delete(key)
}

type hashqueue struct { //哈希队列结构
	key    string
	patch  string
	value  Hashvalue
	expire int64
}

var (
	hashcache   sync.Map                      //储存变量
	hashcache_q []hashqueue                   //写入队列
	hashdelete  map[int64][]map[string]string //待删除变量
	h_q         sync.Mutex                    //写入队列锁
	hash_no     int                           //文件序号
	filepatch   string                        //持久化文件夹
)

/**hash执行函数,对于读写都在此完成，加锁以免冲突
 * 写入value两种方式，map[string]interface{},sync.Map
 **/
func do_hash(key string, value map[string]interface{}, patch string, expire int64, t string) {
	//if t == "hdel" {
	//fmt.Println(key, patch, "删除")
	//panic("存在删除")
	//}
	var value_v Hashvalue
	var patch_v sync.Map
	patch_v_i, ok := hashcache.Load(patch)
	if !ok {
		patch_v.Store(key, Hashvalue{patch: patch, key: key})
		hashcache.Store(patch, patch_v)
	} else {
		patch_v = patch_v_i.(sync.Map)
	}
	value_v_i, ok := patch_v.Load(key)
	if !ok {
		value_v = Hashvalue{patch: patch, key: key}
	} else {
		value_v = value_v_i.(Hashvalue)
	}

	var tmp_witre Hashvalue //实际要写入 更新的内容
	//exp模式，暂时支持单+或者单—
	if value["exp"] != nil {
		vals := strings.Split(value["exp"].(string), ",")
		for _, val := range vals {
			param := strings.Split(value[val].(string), "+")
			if len(param) > 1 {
				t_result := 0
				for _, v := range param {
					//尝试从缓存取值
					t_n, ok := value_v.value.Load(v)
					t_num := 0
					//初始化原始值
					if t_n != nil && ok {
						switch t_n.(type) {
						case string:
							t_num, _ = strconv.Atoi(t_n.(string))
						case int:
							t_num = t_n.(int)
						case int64:
							t_num = int(t_n.(int64))
						}

					}
					if t_num == 0 {
						//尝试直接转换成数字
						t_num, _ = strconv.Atoi(v)
					}
					t_result += t_num
				}
				//增加后，保持type前后一致性
				switch value[val].(type) {
				case string:
					value[val] = strconv.Itoa(t_result)
				case nil:
					value[val] = strconv.Itoa(t_result)
				case int:
					value[val] = t_result
				case int64:
					value[val] = int64(t_result)
				}
				continue
			}
			param = strings.Split(value[val].(string), "-")
			if len(param) > 1 {
				t_result := 0
				for k, v := range param {
					//尝试从缓存取值
					t_n, ok := value_v.value.Load(v)
					t_num := 0
					//初始化原始值
					if t_n != nil && ok {
						switch t_n.(type) {
						case string:
							t_num, _ = strconv.Atoi(t_n.(string))
						case int:
							t_num = t_n.(int)
						case int64:
							t_num = int(t_n.(int64))
						}

					}
					if t_num == 0 {
						//尝试直接转换成数字
						t_num, _ = strconv.Atoi(v)
					}
					if k == 0 {
						t_result = t_num
					} else {
						t_result -= t_num
					}

				}
				//减少后，保持type前后一致性
				switch value[val].(type) {
				case string:
					value[val] = strconv.Itoa(t_result)
				case nil:
					value[val] = strconv.Itoa(t_result)
				case int:
					value[val] = t_result
				case int64:
					value[val] = int64(t_result)
				}
				continue
			}

		}
		delete(value, "exp")
	}
	for k, val := range value {
		value_v.value.Store(k, val)
		tmp_witre.value.Store(k, val)
	}
	if expire > 0 {
		expire = Timestampint() + expire
	} else {
		expire = -1
	}

	//赋值，写入持久化
	value_v.time = expire
	tmp_witre.time = expire
	patch_v.Store(key, value_v)
	if expire > Timestampint() {
		hashdelete[expire] = append(hashdelete[expire], map[string]string{"key": key, "patch": patch})
	}
	switch t {
	case "hset_r":
		//即刻写入
		hash_write(patch, key, tmp_witre)
	case "hset":
		//加入写队列
		hash_queue(key, tmp_witre, patch, expire, t)
	}
}

/**
 * 仅支持以下switch的type持久化写入，如需增加，则需要在makehashfromfile添加反序列化数据组装方式
 */
func hash_write(patch string, key string, tmp Hashvalue) {
	hash_no++
	wireteString := make(map[string]map[string]map[string]interface{})
	wireteString[patch] = make(map[string]map[string]interface{})
	wireteString[patch][key] = make(map[string]interface{})
	value := make(map[string]string)
	tmp.value.Range(func(kk, vv interface{}) bool {
		write_string := ""
		switch vv.(type) {
		case string:
			write_string = "string|" + vv.(string)
		case int:
			write_string = "int|" + strconv.Itoa(vv.(int))
		case int64:
			write_string = "int64|" + fmt.Sprintf("%d", vv)
		case map[string]string:
			write_string = "mps|" + Msgpack_pack(vv)
		case map[string]interface{}:
			write_string = "mpi|" + Msgpack_pack(vv)
		case []string:
			write_string = "ss|" + Msgpack_pack(vv)
		case []map[string]string:
			write_string = "smps|" + Msgpack_pack(vv)
		case []map[string]interface{}:
			write_string = "smpi|" + Msgpack_pack(vv)
		case map[string]map[string]string:
			write_string = "mpsmps|" + Msgpack_pack(vv)
		case map[string]map[string]interface{}:
			write_string = "mpsmpi|" + Msgpack_pack(vv)

		}
		value[kk.(string)] = write_string
		return true
	})
	wireteString[patch][key]["value"] = value
	wireteString[patch][key]["time"] = strconv.FormatInt(tmp.time, 10)
	f, err1 := os.Create(filepatch + "/h_" + strconv.Itoa(hash_no) + ".cache")
	if err1 != nil {
		fmt.Println(err1, "hash文件创建失败")
	}
	_, err1 = io.WriteString(f, Msgpack_pack(wireteString))
	if err1 != nil {
		fmt.Println(err1, "hash文件写入失败")
	}
	f.Close()
}

/*hash写入入口
 * key 哈希名称
 * value 需要写入的值
 * patch 哈希前缀便于分类
 * expire 字段有效期
 */
func Hset_r(key string, value map[string]interface{}, patch string, expire ...int64) bool {
	if len(value) == 0 {
		return false
	}
	if len(expire) == 0 {
		expire = []int64{-1}
	}
	do_hash(key, value, patch, expire[0], "hset_r")
	return true
}

//写入数据
func Hset(key string, value interface{}, patch string, expire ...int64) bool {
	if value == nil {
		return false
	}
	if len(expire) == 0 {
		expire = []int64{-1}
	}
	var write map[string]interface{}
	switch value.(type) {
	case map[string]interface{}:
		write = value.(map[string]interface{})
	case map[string]string:
		write = make(map[string]interface{}, len(value.(map[string]string)))
		for k, v := range value.(map[string]string) {
			write[k] = v
		}
	default:
		t := reflect.TypeOf(value)
		panic("cache写入不支持类型" + t.Name())
	}
	do_hash(key, write, patch, expire[0], "hset")
	return true
}

//哈希队列操作函数，所有操作加锁
func hash_queue(key string, value Hashvalue, patch string, expire int64, t string) {
	h_q.Lock()
	defer h_q.Unlock()
	switch t {
	case "hset": //将写入请求加入队列
		hashcache_q = append(hashcache_q, hashqueue{key: key, patch: patch, value: value, expire: expire})
	case "sync":
		write := make(map[string]map[string]Hashvalue)
		for _, v := range hashcache_q { //分别将队列取出执行hash写入同步
			if write[v.patch] == nil {
				write[v.patch] = make(map[string]Hashvalue)
			}
			tmp := write[v.patch][v.key]
			v.value.value.Range(func(k, v interface{}) bool {
				tmp.value.Store(k, v)
				return true
			})
			tmp.time = v.value.time
			write[v.patch][v.key] = tmp
		}
		//fmt.Println("即将写入",v.key,v.value,v.patch)
		//do_hash(v.key, v.value, v.patch, v.expire, "hset")
		for patch, value := range write {
			for key, val := range value {
				hash_write(patch, key, val)
			}
		}
		hashcache_q = []hashqueue{}
	}
}

//hash读取
func Hget(key string, patch string) Hashvalue {
	var patch_v sync.Map
	var value_v Hashvalue
	patch_v_i, ok := hashcache.Load(patch)
	if !ok {
		patch_v.Store(key, Hashvalue{patch: patch, key: key})
		hashcache.Store(patch, patch_v)
	} else {
		patch_v = patch_v_i.(sync.Map)
	}
	value_v_i, ok := patch_v.Load(key)
	if !ok {
		value_v = Hashvalue{patch: patch, key: key}
	} else {
		value_v = value_v_i.(Hashvalue)
	}
	if value_v.time == -1 || value_v.time > Timestampint() {
		return value_v
	}
	//超时，重置value_v
	patch_v.Store(key, Hashvalue{patch: patch, key: key})
	hashcache.Store(patch, patch_v)
	return value_v
}

//hash删除
func Hdel(key string, patch string) {
	patch_v_i, ok := hashcache.Load(patch)
	if ok {
		patch_v := patch_v_i.(sync.Map)
		patch_v.Delete(key)
		go hash_write(patch, key, Hashvalue{time: -2})
	}
}

//hash删除patch下所有key
func Hdel_all(patch string) {
	_, ok := hashcache.Load(patch)
	if ok {
		hashcache.Delete(patch)
		go hash_write(patch, "", Hashvalue{time: -3})
	}
}

//将cache零散文件整合
func makecachefromfiles() {

	files, err := (ListDir(filepatch, "cache"))
	write_hash := false

	if err != nil {
		dir, _ := os.Getwd() //当前的目录
		var patch string
		if os.IsPathSeparator('\\') { //前边的判断是否是系统的分隔符
			patch = "\\"
		} else {
			patch = "/"
		}
		err := os.Mkdir(dir+patch+filepatch, os.ModePerm) //在当前目录下生成md目录
		if err != nil {
			fmt.Println("创建cache缓存文件夹错误，请检查", dir, "写入权限")
		}
	} else {
		for _, v := range files {
			if !Preg_match("db", v) {
				if Preg_match("h_\\d+", v) {
					write_h := makehashfromfile(v) //读取文件值到hashcache
					if write_h {
						write_hash = true
					}
					err := os.Remove(v) //删除文件
					if err != nil {
						fmt.Println("删除文件失败", v, err)
					}
				}

			}
		}
		if write_hash {
			//结构体无法序列化，需要转换
			wireteString := make(map[string]map[string]map[string]interface{})
			hashcache.Range(func(patch_i, val_i interface{}) bool {
				patch := patch_i.(string)
				val := val_i.(sync.Map)
				wireteString[patch] = make(map[string]map[string]interface{})
				val.Range(func(key_i, v_i interface{}) bool {
					key := key_i.(string)
					v := v_i.(Hashvalue)
					tmp := make(map[string]string)
					v.value.Range(func(kk, vv interface{}) bool {
						write_string := ""
						switch vv.(type) {
						case string:
							write_string = "string|" + vv.(string)
						case int:
							write_string = "int|" + strconv.Itoa(vv.(int))
						case int64:
							write_string = "int64|" + fmt.Sprintf("%d", vv)
						case map[string]string:
							write_string = "mps|" + Msgpack_pack(vv)
						case map[string]interface{}:
							write_string = "mpi|" + Msgpack_pack(vv)
						case []string:
							write_string = "ss|" + Msgpack_pack(vv)
						case []map[string]string:
							write_string = "smps|" + Msgpack_pack(vv)
						case []map[string]interface{}:
							write_string = "smpi|" + Msgpack_pack(vv)
						case map[string]map[string]string:
							write_string = "mpsmps|" + Msgpack_pack(vv)
						case map[string]map[string]interface{}:
							write_string = "mpsmpi|" + Msgpack_pack(vv)

						}
						tmp[kk.(string)] = write_string
						return true
					})
					wireteString[patch][key] = make(map[string]interface{})
					wireteString[patch][key]["value"] = tmp
					wireteString[patch][key]["time"] = strconv.FormatInt(v.time, 10)
					return true
				})
				return true
			})
			f, err1 := os.Create(filepatch + "/h_db.cache")
			if err1 != nil {
				fmt.Println(err1, "hash文件创建失败")
			}
			_, err1 = io.WriteString(f, Msgpack_pack(wireteString))
			if err1 != nil {
				fmt.Println(err1, "hash文件写入失败")
			}
			f.Close()
		}

	}

}

/*从单文件读取hash缓存数据
 *传入文件路径
 */
func makehashfromfile(v string) bool {
	f, err1 := os.Open(v)
	defer f.Close()
	if err1 != nil {
		return false
	}
	b, e := ioutil.ReadAll(f)
	if e != nil {
		return false
	} else {
		result := Msgpack_unpack(b)
		if result == nil {
			return false
		}
		for patch, val := range result.(map[string]interface{}) {
			if patch != "" {
				for key, v := range val.(map[string]interface{}) {
					i, _ := strconv.ParseInt(((v.(map[string]interface{}))["time"]).(string), 10, 64)
					if i == -2 {
						patch_v_i, ok := hashcache.Load(patch)
						if ok {
							patch_v := patch_v_i.(sync.Map)
							patch_v.Delete(key)
						}
						continue
					}
					if i == -3 {
						_, ok := hashcache.Load(patch)
						if ok {
							hashcache.Delete(patch)
						}
						continue
					}
					if i != -1 && i < Timestampint() {
						continue
					}
					tmp := (v.(map[string]interface{}))["value"] //msgpack解码最底层数据map[string]interface{}
					var va sync.Map                              //需要封装的hash数据格式map[string]string
					for kk, vv := range tmp.(map[string]interface{}) {
						index := strings.Index(vv.(string), "|")
						read_type := Substr(vv.(string), 0, index)
						read_value := string([]byte(vv.(string))[index+1:])
						switch read_type {
						case "int":
							t, _ := strconv.Atoi(read_value)
							va.Store(kk, t)
						case "string":
							va.Store(kk, read_value)
						case "int64":
							t, _ := strconv.ParseFloat(read_value, 64)
							va.Store(kk, t)
						case "mps":
							va.Store(kk, Msgpack_unpack_mps(read_value))
						case "mpi":
							va.Store(kk, Msgpack_unpack_mpi(read_value))
						case "ss":
							va.Store(kk, Msgpack_unpack_ss(read_value))
						case "smps":
							va.Store(kk, Msgpack_unpack_smps(read_value))
						case "smpi":
							va.Store(kk, Msgpack_unpack_smpi(read_value))
						case "mpsmps":
							va.Store(kk, Msgpack_unpack_mpsmps(read_value))
						case "mpsmpi":
							va.Store(kk, Msgpack_unpack_mpsmpi(read_value))
						}

					}
					patch_v_i, ok := hashcache.Load(patch)
					var patch_v sync.Map
					if !ok {
						patch_v.Store(key, Hashvalue{patch: patch, key: key, value: va, time: i})
						hashcache.Store(patch, patch_v)
					} else {
						patch_v = patch_v_i.(sync.Map)
						patch_v.Store(key, Hashvalue{patch: patch, key: key, value: va, time: i})
					}

				}
			}
		}
		return true
	}
	return false
}

func init() {
	list_init() //队列初始化
	filepatch = "./cache_hash"
	hashdelete = make(map[int64][]map[string]string)
	makehashfromfile(filepatch + "/h_db.cache") //加载持久化缓存
	makecachefromfiles()                        //加载与整理碎片缓存
	go func() {
		for true {
			//延时999毫秒执行删除
			time.Sleep(time.Millisecond * 999)
			t := Timestampint()
			go func(t int64) {
				if len(hashdelete[t]) > 0 {
					for _, v := range hashdelete[t] {
						patch_v_i, ok := hashcache.Load(v["patch"])
						if ok {
							patch_v := patch_v_i.(sync.Map)
							patch_v.Delete(v["key"])
						}
					}
				}
			}(t)
		}
	}()
	go func() {
		for true {
			//延时1秒执行本地持久化写入
			time.Sleep(time.Millisecond * 1000)
			go hash_queue("", Hashvalue{}, "", 0, "sync")
		}
	}()

}

/**
 * 以下内容是队列
 **/

var (
	list_map   map[string][]interface{} //队列保存变量
	l_q        sync.Mutex               //队列锁
	list_chans map[string][]chan int
	wg         sync.WaitGroup
)

func list_init() {
	list_map = make(map[string][]interface{})
	list_chans = make(map[string][]chan int)
}

/**
*将一个或多个值插入到列表的尾部(最右边)。
*插入一个值Rpush("mylist","hello")
*插入多个值Rpush("mylist","1","2","3")
*插入[]interface{}切片:
   var list []interface{}
   list = append(list,"1")
   list = append(list,map[string]string{"name":"luyu"})
   list = append(list,100)
   Rpush("mylist",list...)
**/
func RPUSH(key string, list ...interface{}) bool {
	if len(list) == 0 {
		return false
	}
	h_q.Lock()
	defer h_q.Unlock()
	list_map[key] = append(list_map[key], list...)
	if len(list_chans[key]) > 0 {
		var new_chans []chan int
		out := true
		//整理空chan，以及对第一个正在等待的chan进行解锁
		for _, list_chan := range list_chans[key] {
			if len(list_chan) > 0 {
				if out {
					<-list_chan
					out = false
				} else {
					new_chans = append(new_chans, list_chan)
				}
			}
		}
		list_chans[key] = new_chans
	}
	return true
}

/**
 *将一个或多个值插入到列表的头部(最左边)，用法同Rpush
 **/
func LPUSH(key string, list ...interface{}) bool {
	if len(list) == 0 {
		return false
	}
	h_q.Lock()
	defer h_q.Unlock()
	list_map[key] = append(list, list_map[key]...)
	if len(list_chans[key]) > 0 {
		var new_chans []chan int
		out := true
		//整理空chan，以及对第一个正在等待的chan进行解锁
		for _, list_chan := range list_chans[key] {
			if len(list_chan) > 0 {
				if out {
					<-list_chan
					out = false
				} else {
					new_chans = append(new_chans, list_chan)
				}
			}
		}
		list_chans[key] = new_chans
	}
	return true
}

/**
 *取出指定列表的第一个元素，如果列表没有元素会阻塞列表直到等待超时或发现可弹出元素为止。
 *LPOP(list1,100)取出名字为list1的列表，没有会等待100秒
 *LPOP(list1)取出列表,没有直接返回
 *当ok返回值为false，则为超时取队列失败
 */
func LPOP(key string, timeout ...int) (result interface{}, ok bool) {
	h_q.Lock()
	defer func() {
		if len(list_map[key]) > 0 {
			if len(list_map[key]) == 1 {
				list_map[key] = nil
			} else {
				list_map[key] = list_map[key][1:]
			}
		}

		h_q.Unlock()
	}()
	if len(list_map[key]) > 0 {
		result = list_map[key][0]
		ok = true
		return
	} else {
		list_chan := make(chan int, 1)
		list_chans[key] = append(list_chans[key], list_chan)
		//加塞
		list_chan <- 0
		h_q.Unlock()
		if len(timeout) == 1 {
			ok = true
			result = waitchan(key, &ok, timeout[0], list_chan)
		}
		h_q.Lock()
	}
	return
}

/**
 *取出指定列表的最后一个元素，如果列表没有元素会阻塞列表直到等待超时或发现可弹出元素为止。
 *RPOP(list1,100)取出名字为list1的列表，没有会等待100秒
 *RPOP(list1)取出列表,没有直接返回
 *当ok返回值为false，则为超时失败
 */
func RPOP(key string, timeout ...int) (result interface{}, ok bool) {
	h_q.Lock()
	defer func() {
		if len(list_map[key]) > 0 {
			if len(list_map[key]) == 1 {
				list_map[key] = nil
			} else {
				list_map[key] = list_map[key][1:]
			}
		}
		h_q.Unlock()
	}()
	if len(list_map[key]) > 0 {
		ok = true
		result = list_map[key][len(list_map[key])-1]
		return
	} else {
		list_chan := make(chan int, 1)
		list_chans[key] = append(list_chans[key], list_chan)
		//加塞
		list_chan <- 0
		h_q.Unlock()
		if len(timeout) == 1 {
			ok = true
			result = waitchan(key, &ok, timeout[0], list_chan)
		}
		h_q.Lock()
	}
	return
}

func waitchan(key string, ok *bool, timeout int, list_chan chan int) (result interface{}) {
	go func(list_chan chan int) {
		//等待指定时间
		time.Sleep(time.Second * time.Duration(timeout))
		h_q.Lock()
		//超时返回nil与false
		*ok = false
		//解锁
		if len(list_chan) > 0 {
			<-list_chan
		}
		h_q.Unlock()
	}(list_chan)
	//尝试解锁
	list_chan <- 0
	h_q.Lock()
	defer h_q.Unlock()
	if len(list_map[key]) > 0 {
		result = list_map[key][0]
	}
	//释放阻塞
	<-list_chan
	return
}

/**
 * 通过索引获取队列的元素
 * 获取失败返回nil,false
 **/
func LINDEX(key string, index int) (result interface{}, ok bool) {
	h_q.Lock()
	defer h_q.Unlock()
	if len(list_map[key]) < index {
		return
	}
	return list_map[key][index], true
}

/**
 * 获取列表长度
 **/
func LLEN(key string) (int, bool) {
	h_q.Lock()
	defer h_q.Unlock()
	if list_map[key] == nil {
		return 0, false
	}
	return len(list_map[key]), true
}

/**
 * 获取列表指定范围内的元素，起始元素是0
 * 表不存在返回false
 * LRANGE("list",2,3)取第2到3个元素
 * LRANGE("list",5,2)如果start比stop小,调换他们的顺序，取第2到第5个元素
 * LRANGE("list",-2,1)取第1个到倒数第2个元素,假如10个元素，等同于1,8
 * LRANGE("list",2)如果stop为空，则取第0到2个元素
 * LRANGE("list",-3) 取最后3个元素
 * 假如stop超过列表长度，返回空
 **/
func LRANGE(key string, start int, param ...interface{}) ([]interface{}, bool) {
	h_q.Lock()
	defer h_q.Unlock()
	var stop int
	if list_map[key] == nil {
		return nil, false
	}
	if len(param) == 0 {
		if start > 0 {
			stop = 0
		} else {
			stop = len(list_map[key]) - 1
		}
	} else {
		switch param[0].(type) {
		case int:
			stop = param[0].(int)
		}
	}
	if start < 0 {
		start = len(list_map[key]) + start
		if start < 0 {
			start = 0
		}
	}

	if stop < 0 {
		stop = len(list_map[key]) + stop
		if stop < 0 {
			stop = 0
		}
	}
	s := start
	if start > stop {
		start = stop
		stop = s
	}
	//最大值超过最大长度
	if stop > len(list_map[key])-1 {
		return nil, true
	}
	//起始大于最大长度,返回空
	if start > len(list_map[key])-1 {
		return nil, true
	}
	result := list_map[key][start:]
	return result[:stop+1-start], true
}

/**
 *根据参数 COUNT 的值，移除列表中与参数 VALUE 相等的元素。
 *count > 0 : 从表头开始向表尾搜索，移除与 VALUE 相等的元素，数量为 COUNT 。
 *count < 0 : 从表尾开始向表头搜索，移除与 VALUE 相等的元素，数量为 COUNT 的绝对值。
 *count = 0 : 移除表中所有与 VALUE 相等的值。
 */
func LREM(key string, count int, value interface{}) bool {
	h_q.Lock()
	defer h_q.Unlock()
	if list_map[key] == nil {
		return false
	}
	var new_list []interface{}
	l := count
	if l < 0 {
		l = l * -1
	}
	vv := Msgpack_pack(value)
	if count == 0 {
		for k, v := range list_map[key] {
			if Msgpack_pack(v) != vv {
				new_list = append(new_list, list_map[key][k])
			}
		}
	} else if count > 0 {
		for k, v := range list_map[key] {
			if Msgpack_pack(v) != vv {
				new_list = append(new_list, list_map[key][k])
			} else {
				l--
				if l < 0 {
					new_list = append(new_list, list_map[key][k])
				}

			}
		}
	} else if count < 0 {
		for kk, _ := range list_map[key] {
			k := len(list_map[key]) - kk - 1
			if Msgpack_pack(list_map[key][k]) != vv {
				new_list = append([]interface{}{list_map[key][k]}, new_list...)
			} else {
				l--
				if l < 0 {
					new_list = append([]interface{}{list_map[key][k]}, new_list...)
				}
			}
		}
	}
	list_map[key] = new_list
	return true
}

/**
 * LTRIM 对一个列表进行修剪(trim)，就是说，让列表只保留指定区间内的元素，不在指定区间之内的元素都将被删除。
 * start 与 stop定义参照LRANGE
 * 设置超过最大值的start会清空列表
 * 设置超过最大值的stop等同于最大值
 **/
func LTRIM(key string, start int, param ...interface{}) bool {
	h_q.Lock()
	defer h_q.Unlock()
	if list_map[key] == nil {
		return false
	}
	var stop int
	if len(param) == 0 {
		if start > 0 {
			stop = 0
		} else {
			stop = len(list_map[key]) - 1
		}
	} else {
		switch param[0].(type) {
		case int:
			stop = param[0].(int)
		}
	}
	if start < 0 {
		start = len(list_map[key]) + start
		if start < 0 {
			start = 0
		}
	}

	if stop < 0 {
		stop = len(list_map[key]) + stop
		if stop < 0 {
			stop = 0
		}
	}
	s := start
	if start > stop {
		start = stop
		stop = s
	}
	//最大值超过最大长度,等同于最大值
	if stop > len(list_map[key])-1 {
		stop = len(list_map[key]) - 1
	}
	//起始大于最大长度,清空列表
	if start > len(list_map[key])-1 {
		list_map[key] = nil
		return true
	}
	result := list_map[key][start:]
	list_map[key] = result[:stop+1-start]
	return true
}

func pop_test() {
	begin := Timestampint() + 1
	fmt.Println("开始测试")
	//读取左边数据等待100秒
	//线程1
	go func() {
		fmt.Println(LPOP("test", 100))
		fmt.Println("1等待了", Timestampint()-begin, "秒")
	}()
	//延迟1秒执行线程2
	time.Sleep(time.Second * 1)
	//线程2
	go func() {
		fmt.Println(LPOP("test", 10))
		fmt.Println("2等待了", Timestampint()-begin, "秒")
	}()
	//等5秒后再写入
	time.Sleep(time.Second * 5)
	fmt.Println("开始写入1")
	RPUSH("test", "久等了")
	//等待3秒后写入
	time.Sleep(time.Second * 3)
	fmt.Println("开始写入2")
	LPUSH("test", "第二次写入")
}

func lrange_test() {
	LPUSH("test", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9")
	fmt.Println(LRANGE("test", 0, 1))
	fmt.Println(LRANGE("test", 5, 10))
	fmt.Println(LRANGE("test", -2))
}

func lrem_test() {
	LPUSH("test", "5", "2", "2", "3", "3", "3", "4", "5", "6", "7")
	LREM("test", 0, "2")          //去掉所有的2
	fmt.Println(list_map["test"]) //[5 3 3 3 4 5 6 7]
	LREM("test", 2, "3")          //去掉左边两个3
	fmt.Println(list_map["test"]) //[5 3 4 5 6 7]
	LREM("test", -1, "5")         //去掉右边那个5
	fmt.Println(list_map["test"]) //[5 3 4   6 7]
}

func ltrim_test() {
	LPUSH("test", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9")
	fmt.Println(LTRIM("test", 0, 7))
	fmt.Println(list_map["test"]) //[0 1 2 3 4 5 6 7]
	fmt.Println(LTRIM("test", 2, 4))
	fmt.Println(list_map["test"]) //[2 3 4]
	fmt.Println(LTRIM("test", 10))
	fmt.Println(list_map["test"]) //[2 3 4]
}
