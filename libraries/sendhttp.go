package libraries

import (
	"github.com/astaxie/beego/httplib"
	"net/http"
	//"strings"
	"errors"
	"fmt"
	"time"
)

type Httplib struct {
	Enable_cookie bool
	Cookies       []*http.Cookie
	Req           *(httplib.BeegoHTTPRequest)
}

func Newhttp(cookie bool) *Httplib {
	return &Httplib{Enable_cookie: cookie}
}

func (this *Httplib) GET(url string, cookies ...[]*http.Cookie) *Httplib {
	this.Req = httplib.Get(url)
	this.Req.SetTimeout(1*time.Second, 5*time.Second)
	if len(cookies) == 1 {
		for _, v := range cookies[0] {
			this.Req.SetCookie(v)
		}
	}
	if this.Enable_cookie {
		this.Req.SetEnableCookie(this.Enable_cookie)
		//for _,v := range this.Cookies{
		//this.Req.SetCookie(v)
		//}
	}
	return this
}
func (this *Httplib) POST(url string, param map[string]string, cookies ...[]*http.Cookie) *Httplib {
	this.Req = httplib.Post(url)
	this.Req.SetTimeout(5*time.Second, 100*time.Second)
	for k, v := range param {
		this.Req.Param(k, v)
	}
	if len(cookies) == 1 {
		for _, v := range cookies[0] {
			this.Req.SetCookie(v)
		}
	}
	if this.Enable_cookie {
		this.Req.SetEnableCookie(this.Enable_cookie)
		//for _,v := range this.Cookies{
		//this.Req.SetCookie(v)
		//}
	}
	return this
}
func (this *Httplib) String() (result string, err error) {
	result, err = this.Req.String()
	if this.Enable_cookie && err == nil {
		res, _ := this.Req.Response()
		cookies := res.Cookies()
		this.Cookies = cookies
	}
	return
}
func (this *Httplib) No_302() {
	this.Req.SetCheckRedirect(func(req *http.Request, via []*http.Request) error {
		this.Cookies = req.Response.Cookies()
		if len(via) >= 0 {
			return errors.New("stopped after 0 redirects")
		}
		return nil
	})
}

func (this *Httplib) T() {
	fmt.Println(this.Req.Response())
}

func (this *Httplib) Rebiuldcookie(cookies_name []string) {
	newcookie := []*http.Cookie{}
	for _, name := range cookies_name {
		tmp := new(http.Cookie)
		for k, _ := range this.Cookies {
			var v = this.Cookies[len(this.Cookies)-1-k]
			if v.Name == name && v.Value != "" {
				tmp = v
			}
		}
		if tmp.Name == "" {
			tmp.Name = name
			tmp.Value = ""
			tmp.Path = "/"
		}
		newcookie = append(newcookie, tmp)
	}
	this.Cookies = newcookie
}
func (this *Httplib) ToFile(filename string) (err error) {
	err = this.Req.ToFile(filename)
	if this.Enable_cookie && err == nil {
		res, _ := this.Req.Response()
		cookies := res.Cookies()
		this.Cookies = cookies
	}
	return
}
func (this *Httplib) ToJSON() (result interface{}, err error) {
	err = this.Req.ToJSON(result)
	if this.Enable_cookie && err == nil {
		res, _ := this.Req.Response()
		cookies := res.Cookies()
		this.Cookies = cookies
	}
	return
}

func (this *Httplib) Getcookie(name string) string {
	for _, v := range this.Cookies {
		if v.Name == name {
			return v.Value
		}
	}
	return ""
}

func (this *Httplib) GetDOMAIN(name string) string {
	for _, v := range this.Cookies {
		if v.Name == name {
			return v.Domain
		}
	}
	return ""
}
