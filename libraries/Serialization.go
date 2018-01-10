package libraries

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ugorji/go/codec"
	"gopkg.in/vmihailenco/msgpack.v2" //此代码反序列化速度快，序列化稍慢
	"io"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

func S2B(s *string) []byte {
	return *(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(s))))
}

func B2S(buf []byte) string {
	return *(*string)(unsafe.Pointer(&buf))
}

func mspack_unpack(s interface{}) interface{} {
	if s == nil {
		return nil
	}
	var v interface{}

	/*var (
	    v interface{} // value to decode/encode into
	    mh codec.MsgpackHandle
	)*/
	var err error
	switch s.(type) {
	case string:
		if s.(string) == "" || (len([]rune(s.(string))) == 1 && []rune(s.(string))[0] == 63) {
			return nil
		}
		err = msgpack.Unmarshal([]byte(s.(string)), &v)
		//err = codec.NewDecoder(strings.NewReader(s.(string)), &mh).Decode(&v)
	case []uint8:
		if len(s.([]uint8)) == 0 {
			return nil
			return nil
		}
		err = msgpack.Unmarshal(s.([]uint8), &v)
		//err = codec.NewDecoder(bytes.NewBuffer(s.([]uint8)), &mh).Decode(&v)
	}

	if err != nil {
		fmt.Println("失败内容", s)
		panic("msgpack反序列失败")
	}
	return v
}
func Msgpack_unpack(s interface{}) interface{} {
	return Initresult(mspack_unpack(s))
}

//slice_string
func Msgpack_unpack_ss(s interface{}) []string {
	v := mspack_unpack(s)
	if v == nil {
		return nil
	}
	tmp := make([]string, len(v.([]interface{})))
	for _k, _v := range v.([]interface{}) {
		tmp[_k] = initstring(_v)
	}
	return tmp
}

//slice_map_string
func Msgpack_unpack_smps(s interface{}) []map[string]string {
	v := mspack_unpack(s)
	if v == nil {
		return nil
	}
	tmp := make([]map[string]string, len(v.([]interface{})))
	for _k, _v := range v.([]interface{}) {
		t := make(map[string]string)
		for kk, vv := range _v.(map[interface{}]interface{}) {
			k1 := initkey(kk)
			t[k1] = initstring(vv)
		}
		tmp[_k] = t
	}
	return tmp
}

//slice_map_interface
func Msgpack_unpack_smpi(s interface{}) []map[string]interface{} {
	v := mspack_unpack(s)
	if v == nil {
		return nil
	}
	tmp := make([]map[string]interface{}, len(v.([]interface{})))
	for _k, _v := range v.([]interface{}) {
		t := make(map[string]interface{})
		for kk, vv := range _v.(map[interface{}]interface{}) {
			k1 := initkey(kk)
			t[k1] = vv
		}
		tmp[_k] = t
	}
	return tmp
}

//返回map[string]map[string]string
func Msgpack_unpack_mpsmps(s interface{}) map[string]map[string]string {
	v := mspack_unpack(s)
	if v == nil {
		return nil
	}
	tmp := make(map[string]map[string]string)
	for _k, _v := range v.(map[interface{}]interface{}) {
		k1 := initkey(_k)
		tmp[k1] = make(map[string]string)
		for kk, vv := range _v.(map[interface{}]interface{}) {
			k2 := initkey(kk)
			tmp[k1][k2] = initstring(vv)
		}
	}
	return tmp
}

//返回map[string]map[string]interface
func Msgpack_unpack_mpsmpi(s interface{}) map[string]map[string]interface{} {
	v := mspack_unpack(s)
	if v == nil {
		return nil
	}
	tmp := make(map[string]map[string]interface{})
	for _k, _v := range v.(map[interface{}]interface{}) {
		k1 := initkey(_k)
		tmp[k1] = make(map[string]interface{})
		for kk, vv := range _v.(map[interface{}]interface{}) {
			k2 := initkey(kk)
			tmp[k1][k2] = vv
		}
	}
	return tmp
}

//返回map[string]string
func Msgpack_unpack_mps(s interface{}) map[string]string {
	v := mspack_unpack(s)
	if v == nil {
		return nil
	}
	tmp := make(map[string]string)
	for _k, _v := range v.(map[interface{}]interface{}) {
		key := initkey(_k)
		tmp[key] = initstring(_v)
	}
	return tmp

}

//返回map[string]interface{}
func Msgpack_unpack_mpi(s interface{}) map[string]interface{} {
	v := mspack_unpack(s)
	if v == nil {
		return nil
	}
	tmp := make(map[string]interface{})
	for _k, _v := range v.(map[interface{}]interface{}) {
		key := initkey(_k)
		tmp[key] = Initresult(_v)
	}
	return tmp
}

//返回map[string]string
func Json_unpack_mps(s interface{}) map[string]string {
	v := json_unpack(s)
	if v == nil {
		return nil
	}
	tmp := make(map[string]string)
	for _k, _v := range v.(map[interface{}]interface{}) {
		key := initkey(_k)
		tmp[key] = initstring(_v)
	}
	return tmp

}

//返回map[string]interface{}
func Json_unpack_mpi(s interface{}) map[string]interface{} {
	v := json_unpack(s)
	if v == nil {
		return nil
	}
	tmp := make(map[string]interface{})
	for _k, _v := range v.(map[interface{}]interface{}) {
		key := initkey(_k)
		tmp[key] = Initresult(_v)
	}
	return tmp
}

//返回interface
func Json_unpack(s interface{}) interface{} {
	return Initresult(json_unpack(s))
}
func json_unpack(s interface{}) interface{} {
	if s == nil {
		return nil
	}
	var (
		v  interface{} // value to decode/encode into
		mh codec.JsonHandle
	)
	var err error
	switch s.(type) {
	case string:
		err = codec.NewDecoder(strings.NewReader(s.(string)), &mh).Decode(&v)
	case []uint8:
		err = codec.NewDecoder(bytes.NewBuffer(s.([]uint8)), &mh).Decode(&v)
	}
	if err != nil {
		fmt.Println("json反序列失败", err)
		return nil
	} else {
		return v
	}

}

//反序列化，返回map[string]interface{}类型，如果反序列化结果不是map则返回err
func Unserialize_map(s interface{}, untype string) map[string]interface{} {
	var (
		v      interface{} // value to decode/encode into
		jh     codec.JsonHandle
		mh     codec.MsgpackHandle
		reader io.Reader
		err    error
		src    string
	)
	switch s.(type) {
	case string:
		reader = strings.NewReader(s.(string))
		src = s.(string)
	case []uint8:
		reader = bytes.NewBuffer(s.([]uint8))
		src = B2S(s.([]uint8))
	}
	switch untype {
	case "json":
		err = codec.NewDecoder(reader, &jh).Decode(&v)
	case "msgpack":
		err = codec.NewDecoder(reader, &mh).Decode(&v)
	default:
		return nil
	}
	if err != nil {
		fmt.Println("map反序列失败", err)
		return map[string]interface{}{"errcode": "err", "errmsg": "反序列化失败"}
	} else {
		result := make(map[string]interface{})
		for k, value := range v.(map[interface{}]interface{}) {
			key := initkey(k)
			switch value.(type) {
			case []byte:
				result[key] = B2S(value.([]byte))
			case map[interface{}]interface{}:
				result[key] = Initresult(value)
			case []interface{}:
				result[key] = Initresult(value)
			case string:
				result[key] = value.(string)
			case uint64:
				result[key] = value.(uint64)
			}
		}
		return result
	}

}

func Msgpack_pack(s interface{}) string {

	var (
		b  []byte
		mh codec.MsgpackHandle
	)
	enc := codec.NewEncoderBytes(&b, &mh)
	err := enc.Encode(s)
	if err != nil {
		fmt.Println("msgpack序列化失败", err)
		return ""
	} else {
		return B2S(b)
	}
	return ""
}

func Json_pack(s interface{}) string {
	var (
		b  []byte
		mh codec.JsonHandle
	)
	enc := codec.NewEncoderBytes(&b, &mh)
	err := enc.Encode(s)
	if err != nil {
		fmt.Println("json序列化失败", err)
		return ""
	} else {
		return B2S(b)
	}
	return ""
}
func Json_pack_b(s interface{}) []byte {
	var (
		b  []byte
		mh codec.JsonHandle
	)
	enc := codec.NewEncoderBytes(&b, &mh)
	err := enc.Encode(s)
	if err != nil {
		fmt.Println("json序列化失败", err)
		return []byte{}
	} else {
		return b
	}
	return []byte{}
}

func Msgpack_pack_b(s interface{}) []byte {
	var (
		b  []byte
		mh codec.MsgpackHandle
	)
	enc := codec.NewEncoderBytes(&b, &mh)
	err := enc.Encode(s)
	if err != nil {
		fmt.Println("msgpack序列化失败", err)
		return []byte{}
	} else {
		return b
	}
	return []byte{}
}
func Initresult(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	var result = make(map[string]interface{})
	switch v.(type) {
	case map[interface{}]interface{}:
		return initmap(v.(map[interface{}]interface{}))
	case []interface{}:
		return initslice(v.([]interface{}))
	default:
		return v
	}
	return result
}
func initmap(v map[interface{}]interface{}) interface{} {

	result := make(map[string]interface{})
	for k, value := range v {
		key := initkey(k)
		switch value.(type) {
		case []byte:
			result[key] = B2S(value.([]byte))
		case map[interface{}]interface{}:
			result[key] = Initresult(value)
		case []interface{}:
			result[key] = Initresult(value)
		case string:
			result[key] = value.(string)
		case uint64:
			result[key] = value.(uint64)
		case nil:
			result[key] = nil
		default:
			t := reflect.TypeOf(value)
			fmt.Println("序列化initmap未设置类型", t.Name())
		}
	}
	return result
}
func initslice(v []interface{}) interface{} {
	result := make([]interface{}, len(v))
	for i := 0; i < len(v); i++ {
		switch v[i].(type) {
		case []byte:
			result[i] = B2S(v[i].([]byte))
		case map[interface{}]interface{}:
			result[i] = Initresult(v[i])
		case []interface{}:
			result[i] = Initresult(v[i])
		case string:
			result[i] = v[i].(string)
		case uint64:
			result[i] = v[i].(uint64)
		default:
			t := reflect.TypeOf(v[i])
			fmt.Println("序列化initslice未设置类型", t.Name())
		}
	}
	return result
}
func initstring(v interface{}) (result string) {
	switch v.(type) {
	case string:
		result = v.(string)
	case uint64:
		result = fmt.Sprintf("%d", v)
	case uint32:
		result = fmt.Sprintf("%d", v)
	case uint16:
		result = fmt.Sprintf("%d", v)
	case uint8:
		result = fmt.Sprintf("%d", v)
	case int64:
		result = fmt.Sprintf("%d", v)
	case int32:
		result = fmt.Sprintf("%d", v)
	case int:
		result = fmt.Sprintf("%d", v)
	case float32:
		result = Number_format(v, 10)
	case float64:
		//精度10位小数
		result = Number_format(v, 10)
	case nil:
		result = ""
	default:
		t := reflect.TypeOf(v)
		fmt.Println("Number_format float无法识别变量类型", t.Name())
	}
	return
}
func initkey(k interface{}) (key string) {
	switch k.(type) {
	case string:
		key = k.(string)
	case uint64:
		key = strconv.FormatUint(k.(uint64), 10)
	case int:
		key = strconv.Itoa(k.(int))
	case int64:
		key = strconv.FormatInt(k.(int64), 10)
	default:
		t := reflect.TypeOf(k)
		fmt.Println("序列化未设置key类型", t.Name())
	}
	return
}

func init() {

}
