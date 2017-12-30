package libraries

import (
	"math/rand"
	"strconv"
	"time"
	//"fmt"
)

/*获取令牌值
 *第一个参数传入token
 *第二个参数用于计算的hash
 *第三个参数长度
 *生成随机令牌只需要传入token，会把随机hash校验值存入token
 */
func GetToken(param ...string) string {
	switch len(param) {
	case 0:
		return ""
	case 1:
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		param = append(param, strconv.Itoa(r.Intn(10000000)), "16")
		cache := map[string]interface{}{"hash": param[1]}
		Hset(param[0], cache, "token")
	case 2:
		param = append(param, "16")
	}
	timestamp := time.Now().Unix()
	length, err := strconv.Atoi(param[2])
	if err != nil || length < 1 {
		length = 16
	}
	token := Substr(Newhash().Hash(strconv.FormatInt(timestamp/60*60, 10), param[1]), 0, length)
	return token
}
func init() {

}
