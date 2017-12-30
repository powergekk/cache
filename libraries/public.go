package libraries

import (
	"bytes"
	"math/rand"
	//"github.com/sillydong/fastimage"
	"compress/zlib"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	//"path/filepath"
)

func DoZlibCompress(src []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(src)
	w.Close()
	return in.Bytes()
}

func Void(params ...interface{}) {}

//返回无序map 原array_under_reset变形函数之一
func Map_under_reset(array interface{}, key string, typ int) map[string]interface{} {
	tmp := make(map[string]interface{})
	switch array.(type) {
	case []map[string]string:

		for _, v := range array.([]map[string]string) {
			if typ == 1 {
				tmp[v[key]] = v
			} else if typ == 2 {
				//tmp[v[key]][] = v
			}
		}
	}
	return tmp

}

//返回map里面的值
func Array_values(array map[string]string) []string {
	re := []string{}
	for _, v := range array {
		re = append(re, v)
	}
	return re
}

//去掉/../获取真实路径
func Realpath(path string) string {
	path_s := strings.Split(path, "/")
	realpath := []string{}
	if len(path_s) == 0 {
		return "error"
	}
	for _, value := range path_s {

		if value == ".." {
			k := len(realpath)
			kk := k - 1
			realpath = append(realpath[:kk], realpath[k:]...)
		} else {
			realpath = append(realpath, value)
		}
	}

	return strings.Join(realpath, "/")
}

//寻找中文的位置
func Chinese_str_index(str string, find string) int {
	rs := []rune(str)
	frs := []rune(find)
	index := -1
one:
	for k, v := range rs {
		if v == frs[0] {
			index = k
			for kk, vv := range frs {
				if k+kk > len(rs) || rs[k+kk] != vv {
					index = -1
					continue one
				}
			}
			return index
		}
	}
	return index
}

//删除最后一个切片，并返回值
func Array_pop(array *[]string) string {
	value := *array
	if value == nil {
		return ""
	}
	length := len(value)
	result := value[length-1]
	value = append(value[:length-1], value[length:]...)
	*array = value
	return result
}

//正则匹配
func Preg_match(regtext string, param ...interface{}) bool {
	text := param[0].(string)
	r, _ := regexp.Compile(regtext)
	result := r.FindStringSubmatch(text)
	if len(param) == 2 {
		*(param[1].(*[]string)) = result
	}
	if len(result) > 0 {
		return true
	}
	if len(param) == 2 {
		*(param[1].(*[]string)) = append(*(param[1].(*[]string)), "")
	}
	return false
}

//正则替换
func Preg_replace(regtext string, text string, src string) string {
	regtext = Substr(regtext, 1, len(regtext)-2)
	r, _ := regexp.Compile(regtext)
	return r.ReplaceAllString(src, text)
}
func test() { fmt.Println() }
func init() {
}

//组合map
func Array_merge(m ...map[string]interface{}) map[string]interface{} {
	if len(m) == 1 {
		return m[0]
	}
	result := make(map[string]interface{})
	for k, v := range m {
		if k > 0 {
			for key, val := range v {
				result[key] = val
			}
		} else {
			result = m[0]
		}
	}
	return result
}

//取字串符某几位
func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}

//返回int64时间戳
func Timestampint() int64 {
	cur := time.Now()
	return cur.UnixNano() / 1000000000
}

//返回时间戳字串符
func Timestamp() string {
	cur := time.Now()
	return strconv.FormatInt(cur.UnixNano()/1000000000, 10)
}

//毫秒级时间戳
func Microtimeint() int64 {
	cur := time.Now()
	return cur.UnixNano() / 1000000
}

//毫秒级时间戳
func Microtime() string {
	cur := time.Now()
	return strconv.FormatInt(cur.UnixNano()/1000000, 10)
}

//获取指定目录下的所有文件，不进入下一级目录搜索，可以匹配后缀过滤。
func ListDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0, 10)
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	PthSep := "/"
	if os.IsPathSeparator('\\') { //前边的判断是否是系统的分隔符
		PthSep = "\\"
	}
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}
	return files, nil
}

/*php函数之返回随机键名
 *第一个参数传入map或者切片
 *第二个参数传入返回数量，默认返回一个string
 */
func Array_rand(param ...interface{}) (result []string) {
	var num int64
	keys := []string{}
	if len(param) < 2 {
		num = 1
	} else {
		num = (int64)(param[1].(int))
	}
	if num < 1 {
		num = 1
	}
	switch param[0].(type) {
	case []interface{}:
		for k, _ := range param[0].([]interface{}) {
			keys = append(keys, strconv.Itoa(k))
		}
	case map[string]interface{}:
		for k, _ := range param[0].(map[string]interface{}) {
			keys = append(keys, k)
		}
	}
	var i int64
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i = 0; i < num; i++ {
		re := keys[r.Intn(len(keys))]
		Unset_ss(&keys, re)
		result = append(result, re)
	}
	return
}

//php unset函数变形之一,这里用于删除切片某个元素
func Unset_ss(src *[]string, dest string) {
	for i, v := range *src {
		if v != dest { //当V=-1时，假定是不需要的数据
			*src = append((*src)[:i], (*src)[i+1:]...)
		}
	}
}

//php array_unique函数之去除[]string重复
func Ss_unique(src *[]string) {
	result := false
end:
	for k1, v1 := range *src {
		for k2, v2 := range *src {
			if k1 == k2 || v2 == "" {
				continue
			}
			if v1 == v2 {
				*src = append((*src)[:k2], (*src)[k2+1:]...)
				result = true
				break end
			}
		}
	}
	if result {
		go Ss_unique(src)
	}
}

//php date函数,部分转换内容
func Date(format string, timestamp string) string {
	t, _ := strconv.ParseInt(timestamp, 10, 64)
	tm := time.Unix(t, 0)
	format = strings.Replace(format, "Y", "2006", 1)
	format = strings.Replace(format, "m", "01", 1)
	format = strings.Replace(format, "d", "02", 1)
	format = strings.Replace(format, "H", "03", 1)
	format = strings.Replace(format, "i", "04", 1)
	format = strings.Replace(format, "s", "05", 1)
	return tm.Format(format)
}

//删除重复的切片元素
func Split_unique(list *[]string) []string {
	var x []string = []string{}
	for _, i := range *list {
		if len(x) == 0 {
			x = append(x, i)
		} else {
			for k, v := range x {
				if i == v {
					break
				}
				if k == len(x)-1 {
					x = append(x, i)
				}
			}
		}
	}
	return x
}

//删除重复的数组值
func Map_unique(m *map[string]string) map[string]string {
	mm := make(map[string]string)
	for k, v := range *m {
		w := true
		for _, vv := range mm {
			if vv == v {
				w = false
				break
			}
		}
		if w {
			mm[k] = v
		}
	}
	return mm
}

//判断切片是否包含某元素
func In_slice(str string, sp *[]string) bool {
	if *sp == nil {
		return false
	}
	for _, v := range *sp {
		if str == v {
			return true
		}
	}
	return false
}

//把字串符转为指定小数点的字串符
func Number_format(s interface{}, d interface{}) string {
	var f float64
	var decimals string
	switch d.(type) {
	case string:
		decimals = d.(string)
	case int:
		decimals = strconv.Itoa(d.(int))
	case int64:
		decimals = strconv.FormatInt(d.(int64), 10)
	default:
		t := reflect.TypeOf(d)
		fmt.Println("Number_format decimals无法识别变量类型", t.Name())
	}

	switch s.(type) {
	case string:
		f, _ = strconv.ParseFloat(s.(string), 64)
	case int:
		f = float64(s.(int))
	case int32:
		f = float64(s.(int32))
	case int64:
		f = float64(s.(int64))
	case float32:
		f = float64(s.(float32))
	case float64:
		f = s.(float64)
	default:
		t := reflect.TypeOf(s)
		fmt.Println("Number_format float无法识别变量类型", t.Name())
	}

	return fmt.Sprintf("%."+decimals+"f", f)
}

func Round(s interface{}, places int) float64 {
	var val float64
	switch s.(type) {
	case string:
		val, _ = strconv.ParseFloat(s.(string), 64)
	case float64:
		val = s.(float64)
	case float32:
		val = float64(s.(float32))
	}
	var t float64
	f := math.Pow10(places)
	x := val * f
	if math.IsInf(x, 0) || math.IsNaN(x) {
		return val
	}
	if x >= 0.0 {
		t = math.Ceil(x)
		if (t - x) > 0.50000000001 {
			t -= 1.0
		}
	} else {
		t = math.Ceil(-x)
		if (t + x) > 0.50000000001 {
			t -= 1.0
		}
		t = -t
	}
	x = t / f

	if !math.IsInf(x, 0) {
		return x
	}

	return t
}

//对string进行html简单转义
func Html_encode(str string) string {
	str = strings.Replace(str, "&", "&amp;", -1)
	str = strings.Replace(str, `"`, "&quot;", -1)
	str = strings.Replace(str, "<", "&lt;", -1)
	str = strings.Replace(str, ">", "&gt;", -1)
	return str
}

//返回随机数
func Rand(start int, end int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(end - start)
	return start + n
}

//js的>>>运算
func Shift_3(s interface{}, r uint8) int {
	var i32 int
	switch s.(type) {
	case int:
		i32 = s.(int)
	case int32:
		i32 = int(s.(int32))
	case int64:
		i32 = int(s.(int64))
	case uint8:
		i32 = int(s.(uint8))
	case uint64:
		i32 = int(s.(uint64))
	default:
		panic("未设置类型")
	}

	right := int(r)
	var n uint32 = uint32(i32)
	temp := []uint32{}
	for true {
		if n/2 >= 1 || n%2 == 1 {
			temp = append(temp, n%2)
			n = n / 2
		} else {
			break
		}
	}
	for i := 0; i < len(temp); i++ {
		if i+right >= len(temp) {
			temp[i] = 0
		} else {
			temp[i] = temp[i+right]
		}

	}
	n = 0
	var stnum, conum float64 = 0, 2
	for i := 0; i < len(temp); i++ {
		n += temp[i] * uint32(math.Pow(conum, stnum))
		stnum++
	}
	return int(n)
}
