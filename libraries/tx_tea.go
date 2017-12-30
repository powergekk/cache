package libraries
import (
	"strings"
	"fmt"
	"strconv"
)
var(
f []uint8
m2 = 0
m []uint8
m1 []uint8
v = 0
y = 0
w  []uint8
k  []int
b bool = true
)
const big = 4294967295 
func E()int{
	return Rand(0,4294967295)
}
func e()int{
	return 3294967295
	//return Rand(0,4294967295)
}
func a(t []uint8)[]uint8{
	m = make([]uint8,8)
	m1 = make([]uint8,8)
	v = 0 
	y = 0
	b = true
	m2 = 0
    i := len(t)
	n := 0
    m2 = (i + 10) % 8
	if(m2 != 0){
		m2 = 8 - m2
	}
	w = make([]uint8,i + m2 + 10)
	m[0] = uint8(255 & (248 & e() | m2))
    for o := 1; o <= m2; o++ {
		m[o] = uint8(255 & e())
	}
    m2++
    for n = 1; n <= 2; {
		if(m2 < 8){
			m[m2] = uint8(255 & e())
			n++
			m2++
		}else if(8 == m2){
			c()
		}
	}
	o := 0
    for i > 0 {
		if(m2 < 8){
			m[m2] = t[o] 
			i--
			m2++
			o++
		}
		if(8 == m2){c()}
	}
	
	n = 1
    for n <= 7 {
		if(m2 < 8){
			m[m2] = 0
			n++
			m2++
		}
		if(8 == m2){
			c()
		}
	}
    return w
}

func c() {
	for t := 0; t < 8; t++ {
		if(b){
			m[t] ^= m1[t]
		}else{
			m[t] ^= w[y + t]
		} 
	}
	e := u(m)
	for t := 0; t < 8; t++ {
		w[v + t] = e[t] ^ m1[t]
		m1[t] = m[t]
	}
	y = v
	v += 8
	m2 = 0
	b = false
}
func u(t []uint8)[]uint8{
	o := i(t, 0, 4)
	p := i(t, 4, 4)
	r := i(f, 0, 4)
	s := i(f, 4, 4)
	a := i(f, 8, 4)
	l := i(f, 12, 4)
	var c1  = 0
	for e := 16; e > 0;e-- {
		c1 += 2654435769
		c1 = Shift_3(4294967295 & c1 , 0)
		o += int(int32(p <<4)+int32(r) ^ int32(p+c1) ^ (int32(Shift_3(p , 5))+int32(s)))
		o = Shift_3((4294967295 & o), 0)
		p += int(int32(o << 4) + int32(a) ^int32( o + c1) ^ (int32(Shift_3(o , 5)) + int32(l)))
		p = Shift_3((4294967295 & p), 0)
	}
	u := make([]uint8,8)
	n(&u, 0, o)
	n(&u, 4, p)
	return  u
}
func n(t *[]uint8, e int, i int) {
	(*t)[e + 3] = uint8(i >> 0 & 255)
	(*t)[e + 2] = uint8(i >> 8 & 255)
	(*t)[e + 1] = uint8(i >> 16 & 255)
	(*t)[e + 0] = uint8(i >> 24 & 255)
}

func i(t []uint8, e int, i int)int{
	if(i==0 || i > 4) {
		i = 4
	}
	n := 0
	for o := e; o < e + i; o++ {
		n <<= 8
		n |= int(t[o])
	}
	return Shift_3(4294967295 & n, 0)
}
func h(t string, e...bool)[]uint8{
	count := strings.Count(t,"")-1
    i := []uint8{}
	a := false
	if(len(e)==1){
		a = e[0]
	}
    if (a){
		for n := 0; n < count; n++ {
			i = append(i,255 & []byte(t)[n])
		}
    }else{
		for n := 0; n < count; n += 2 {
			m,_ := strconv.ParseInt(Substr(t,n, 2),16,64)  
			i = append(i,uint8(m))
		}
	}
    return i
}
func r(t string)string{

	if (t==""){
		return ""
	}
	
	i := []byte(t)
	return o1(i)
}
func o1(t []byte)string{
	if (len(t)==0){
		return ""
	}

	e := ""
	for i := 0; i < len(t); i++ {
		var n = fmt.Sprintf("%x",t[i])
		if(strings.Count(n,"")<3){
			n = "0" +n
		}
		e += n
	}
	return e
}

type TEA struct {
	
}
func(this *TEA)StrToBytes(t string)string{
	return r(t)
}
func(this *TEA)Encrypt(t string)string{
	return o1(a(h(t)))
}
func(this *TEA)Initkey(t string,e...bool){
	e1:= false
	if(len(e)==1){
		e1 = e[0]
	}
	f = h(t, e1)

}
func init(){
	
}
/*






func p(t) {
for (var e = "", i = 0; i < t.length; i += 2) e += String.fromCharCode(parseInt(t.substr(i, 2), 16));
return e
}







func l(t) {
var e = 0,
i = new Array(8),
n = t.length;
if (k = t, n % 8 != 0 || n < 16) return null;
if (m1 = g(t), m2 = 7 & m1[0], (e = n - m2 - 10) < 0) return null;
for (var o = 0; o < i.length; o++) i[o] = 0;
w = new Array(e), y = 0, v = 8, m2++;
for (var p = 1; p <= 2;)
if (m2 < 8 && (m2++, p++), 8 == m2 && (i = t, !d())) return null;
for (var o = 0; 0 != e;)
if (m2 < 8 && (w[o] = 255 & (i[y + m2] ^ m1[m2]), o++, e--, m2++), 8 == m2 && (i = t, y = v - 8, !d())) return null;
for (p = 1; p < 8; p++) {
if (m2 < 8) {
if (0 != (i[y + m2] ^ m1[m2])) return null;
m2++
}
if (8 == m2 && (i = t, y = v, !d())) return null
}
return w
}



func u(t) {
for (var e = 16, o = i(t, 0, 4), p = i(t, 4, 4), r = i(f, 0, 4), s = i(f, 4, 4), a = i(f, 8, 4), l = i(f, 12, 4), c = 0; e-- > 0;) c += 2654435769, c = (4294967295 & c) >>> 0, o += (p << 4) + r ^ p + c ^ (p >>> 5) + s, o = (4294967295 & o) >>> 0, p += (o << 4) + a ^ o + c ^ (o >>> 5) + l, p = (4294967295 & p) >>> 0;
var u = new Array(8);
return n(u, 0, o), n(u, 4, p), u
}

func g(t) {
for (var e = 16, o = i(t, 0, 4), p = i(t, 4, 4), r = i(f, 0, 4), s = i(f, 4, 4), a = i(f, 8, 4), l = i(f, 12, 4), c = 3816266640; e-- > 0;) p -= (o << 4) + a ^ o + c ^ (o >>> 5) + l, p = (4294967295 & p) >>> 0, o -= (p << 4) + r ^ p + c ^ (p >>> 5) + s, o = (4294967295 & o) >>> 0, c -= 2654435769, c = (4294967295 & c) >>> 0;
var u = new Array(8);
return n(u, 0, o), n(u, 4, p), u
}

func d() {
for (var t = (k.length, 0); t < 8; t++) m1[t] ^= k[v + t];
return m1 = g(m1), v += 8, m2 = 0, !0
}


*/