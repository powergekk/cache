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

type Hashvalue struct { //哈希缓存结构
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
	_, ok := this.value.Load(key)
	if !ok {
		do_hash(this.key, map[string]interface{}{key.(string): value}, this.patch, 0, "hash_set")
	}
	this.value.Store(key, value)

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

type kvvalue struct { //kv缓存结构
	value string
	time  int64
}

const (
	cache_ext = "33hao" //缓存前缀
)

var (
	//thread int                  //线程计数
	hashcache   sync.Map                      //哈希储存变量
	hashcache_q []hashqueue                   //哈希写入队列
	hashdelete  map[int64][]map[string]string //哈希待删除变量
	kvcache     map[string]kvvalue            //kv缓存
	kvcache_q   []map[string]kvvalue          //kv写入队列
	h_q         sync.Mutex                    //哈希写入队列锁
	hash_no     int                           //哈希文件序号
	filepatch   string                        //哈希持久化文件夹
	kv_mutex    sync.RWMutex                  //kv操作锁
)

//kv写入
func Set(key string, value string, param ...int64) {
	var expire int64
	expire = -1
	if len(param) == 1 {
		expire = param[0]
	}
	if expire == 0 {
		expire = -1
	}
	do_kv(key, value, expire, "set")
}

//kv读出
func Get(key string) string {
	return do_kv(key, "", 0, "get")
}

//kv删除
func Del(key string) {
	do_kv(key, "", 0, "del")
}

//kv操作函数
func do_kv(key string, value string, expire int64, t string) string {
	switch t {
	case "get":
		kv_mutex.RLock()
		defer kv_mutex.RUnlock()
		if kvcache[key].time == -1 || kvcache[key].time > Timestampint() {
			return kvcache[key].value
		}
		if kvcache[key].time == 0 {
			return ""
		}
		fallthrough
	case "del":
		kv_mutex.Lock()
		defer kv_mutex.Unlock()
		delete(kvcache, key)
	case "set":
		kv_mutex.Lock()
		defer kv_mutex.Unlock()
		write := false //写入标识
		tmp := kvcache[key]
		if expire > 0 {
			expire = Timestampint() + expire
		} else {
			expire = -1
		}
		if tmp.value == value { //保存的值与原始值相等
			if expire > 0 { //设置超时时间
				tmp.time = expire
				write = true
			} else { //永久有效
				if tmp.time != -1 {
					tmp.time = -1
					write = true
				}
			}
		} else {
			tmp.value = value
			tmp.time = expire
			write = true
		}
		if write { //赋值，写入持久化
			kvcache[key] = tmp
			wireteString := make(map[string]map[string]interface{})
			for k, v := range kvcache {
				wireteString[k] = make(map[string]interface{})
				wireteString[k]["value"] = v.value
				wireteString[k]["time"] = strconv.FormatInt(v.time, 10)
			}
			f, err1 := os.Create(filepatch + "/kv_db.cache")
			if err1 != nil {
				fmt.Println(err1)
			}
			_, err1 = io.WriteString(f, Msgpack_pack(wireteString))
			if err1 != nil {
				fmt.Println(err1)
			}
			f.Close()
		}
	}
	return ""
}

/**hash执行函数,对于读写都在此完成，加锁以免冲突
 * 写入value两种方式，map[string]interface{},sync.Map
 **/
func do_hash(key string, value map[string]interface{}, patch string, expire int64, t string) (hash Hashvalue) {
	//if t == "hdel" {
	//fmt.Println(key, patch, "删除")
	//panic("存在删除")
	//}
	var value_v Hashvalue
	var patch_v sync.Map
	switch t {
	case "hset_r":
		fallthrough
	case "hash_set": //临时set以便于保持session持续可读,此方法不会写入本地文件
		fallthrough
	case "hset":

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
		if t == "hset_r" {
			hash_write(patch, key, tmp_witre)
		} else {
			hash_queue(key, tmp_witre, patch, expire, t)
		}

	case "expire_del":
		patch_v_i, ok := hashcache.Load(patch)
		if ok {
			patch_v = patch_v_i.(sync.Map)
			patch_v.Delete(key)
		}
	case "hget":
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
		value_v = Hashvalue{patch: patch, key: key}
		patch_v.Store(key, value_v)
		hashcache.Store(patch, patch_v)
		return value_v
		//do_hash(key, nil, patch, 0, "expire_del") //超时删除
	case "hdel":
		patch_v_i, ok := hashcache.Load(patch)
		if ok {
			patch_v = patch_v_i.(sync.Map)
			patch_v.Delete(key)
			go hash_write(patch, key, Hashvalue{time: -2})
		}
	case "hdel_all":
		_, ok := hashcache.Load(patch)
		if ok {
			hashcache.Delete(patch)
			go hash_write(patch, key, Hashvalue{time: -3})
		}
	case "write_db":
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
	return
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

var read sync.Mutex

//hash读取
func Hget(key string, patch string) Hashvalue {
	return do_hash(key, nil, patch, 0, "hget")

}

//hash删除
func Hdel(key string, patch string) bool {
	//go func(){
	do_hash(key, nil, patch, 0, "hdel")
	//}()
	return true
}

//hash删除patch下所有key
func Hdel_all(patch string) bool {
	go func() {
		do_hash("", nil, patch, 0, "hdel_all")
	}()
	return true
}

func test1() {
	fmt.Println()
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
			do_hash("", nil, "", 0, "write_db")
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

/*从单文件读取kv缓存数据
 *传入文件路径
 */
func makekvfromfile(v string) bool {
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

		for key, val := range result.(map[string]interface{}) {
			if key != "" {
				time, err := strconv.ParseInt((val.(map[string]interface{})["time"]).(string), 10, 64)
				if err != nil || time < -1 {
					continue
				}
				if time != -1 && time < Timestampint() {
					continue
				}
				kvcache[key] = kvvalue{value: (val.(map[string]interface{})["value"]).(string), time: time}
			}
		}
		return true
	}
	return false
}

func init() {
	filepatch = "./cache_hash"
	//hashcache = make(map[string]map[string]Hashvalue)
	kvcache = make(map[string]kvvalue)

	hashdelete = make(map[int64][]map[string]string)

	makehashfromfile(filepatch + "/h_db.cache") //加载持久化缓存
	makekvfromfile(filepatch + "/kv_db.cache")  //加载持久化缓存
	makecachefromfiles()                        //加载与整理碎片缓存

	go func() {
		for true {
			time.Sleep(time.Millisecond * 999)
			t := Timestampint()
			go func(t int64) {
				if len(hashdelete[t]) > 0 {
					for _, v := range hashdelete[t] {
						go do_hash(v["key"], nil, v["patch"], 0, "expire_del")
					}
				}

				hash_queue("", Hashvalue{}, "", 0, "sync")

			}(t)

		}
	}()

}
