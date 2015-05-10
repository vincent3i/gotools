package convertor

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

// 最重要的封装类之一
// Golang是强静态语言,无法动态添加/删除属性, 而元数据(map[string]interface{})允许包含用户自定义的key
// 所以只能使用Mapper这类封装部分常用Getter
type Mapper map[string]interface{}

type SiteConfig struct {
	Title      string
	Tagline    string
	Author     Mapper
	Navigation []string
	//Urls map[string]interface{} // for user custom
}
type TopConfig struct {
	Theme          string
	Production_url string
	Posts          PostConfig
	Pages          PageConfig
	Paginator      PaginatorConfig
}
type PostConfig struct {
	Permalink     string
	Summary_lines int
	Latest        int
	Layout        string
	Exclude       string
}
type PageConfig struct {
	Permalink string
	Layout    string
	Exclude   string
}
type PaginatorConfig struct {
	Namespace string
	Per_page  int
	Root_page string
	Layout    string
}
type PostBean struct {
	Id         string
	Title      string
	Date       string
	Layout     string
	Permalink  string
	Categories []string
	Tags       []string
	Url        string
	_Date      time.Time
	_Meta      map[string]interface{}
}
type PageBean struct {
	Id         string
	Title      string
	Date       time.Time
	Layout     string
	Permalink  string
	Categories []string
	Tags       []string
	Url        string
	_Meta      map[string]interface{}
}

func Test_Map2Struct(t *testing.T) {
	m := map[string]interface{}{"permalink": "/:title/:year", "latest": 10}
	post := &PostConfig{}
	Map2Struct(m, reflect.ValueOf(post))
	fmt.Println(*post)
	if post.Permalink != m["permalink"].(string) {
		t.Fail()
	}
	if post.Latest != m["latest"].(int) {
		t.Fail()
	}
}

func Test_Map2Struct2(t *testing.T) {
	m := map[string]interface{}{"Theme": "facebook", "pages": map[string]interface{}{"permalink": "/wendal"}}
	top := &TopConfig{}
	Map2Struct(m, reflect.ValueOf(top))
	fmt.Println(*top)
	if top.Theme != "facebook" {
		t.Fail()
	}
	if top.Pages.Permalink != "/wendal" {
		t.Error("top.Pages.Permalink error")
	}
}
func Test_Map2Struct3(t *testing.T) {
	m := map[string]interface{}{"title": "wendal", "navigation": []string{"admin.html", "user.html"}, "author": map[string]interface{}{"name": "wendal"}}
	site := &SiteConfig{}
	Map2Struct(m, reflect.ValueOf(site))
	fmt.Println(*site)
	if site.Title != "wendal" {
		t.Fail()
	}
	if site.Navigation[0] != "admin.html" {
		t.Fail()
	}
	if site.Navigation[1] != "user.html" {
		t.Fail()
	}

}
