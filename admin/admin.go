package main

import (
	"io"
	"text/template"

	"github.com/labstack/echo/v4"
)

func main() {
	//初始化echo实例
	e := echo.New()
	//初始化模版引擎
	t := &Template{
		//模版引擎支持提前编译模版, 这里对views目录下以html结尾的模版文件进行预编译处理
		//预编译处理的目的是为了优化后期渲染模版文件的速度
		templates: template.Must(template.ParseGlob("html/*.html")),
	}
	//向echo实例注册模版引擎
	e.Renderer = t
	//初始化路由和控制器函数
	e.GET("/index", Index)
	e.Start(":80")
}

//自定义的模版引擎struct
type Template struct {
	templates *template.Template
}

//实现接口，Render函数
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	//调用模版引擎渲染模版
	return t.templates.ExecuteTemplate(w, name, data)
}

// Index
func Index(c echo.Context) error {
	params := make(map[string]interface{})
	params["list"] = []map[string]interface{}{map[string]interface{}{
		"ip":    "127.0.0.1",
		"uuid":  "9527",
		"count": "100",
	}, map[string]interface{}{
		"ip":    "127.0.0.2",
		"uuid":  "9528",
		"count": "101",
	}}
	return c.Render(200, "index.html", params)
}
