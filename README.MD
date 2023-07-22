# I18n

I18n是适用于gin框架的一个国际化处理包。

## 安装

```shell
go get -u "github.com/sk-pkg/i18n"
```

## 快速开始
使用开始前需要定义好项目国际化语言包，如：
`zh-CN.json`
```json
{
  "-1": "系统繁忙",
  "0": "ok",
  "500": "fail",
  "400": "请求参数错误"
}
```
`en-US.json`
```json
{
  "-1": "System is busy",
  "0": "ok",
  "500": "fail",
  "400": "Request parameter error"
}
```
开始使用
```go
import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sk-pkg/i18n"
	"log"
)

func main() {
	r := httpServer()
	r.Run(":8888")
}

func httpServer() *gin.Engine {
	r := gin.Default()
	msg, err := i18n.New(i18n.WithDebugMode(true))
	if err != nil {
		log.Fatal(err)
	}

	r.GET("/busy", func(c *gin.Context) {
		msg.XML(c, -1, "busy", errors.New("busy... "))
	})

	r.GET("/ok", func(c *gin.Context) {
		msg.JSON(c, 0, "success", nil)
	})

	r.GET("/fail", func(c *gin.Context) {
		msg.JSONP(c, 500, "fail", nil)
	})

	r.GET("/params", func(c *gin.Context) {
		msg.YAML(c, 400, "params", nil)
	})

    r.GET("/test", func(c *gin.Context) {
    	// 带参数及输出结果的国际化
        msg.JSON(c, 1000, i18n.Data{
            Params: []string{"Seakee", "18888888888"},
            Data:   "test",
        }, nil)
    })

	// 翻译
    fmt.Println(msg.Trans("en-US", "1000", "Seakee", "18888888888"))
    fmt.Println(msg.Trans("zh-CN", "1000", "Seakee", "18888888888"))

	return r
}
```
Json返回如下
```json
{
    "code": 0,
    "msg": "ok",
    "trace": "",
    "data": "success"
}
```
## 可配置项
1.LangDir国际化语言配置目录，此项为必填
```go
i18n.WithLangDir("./lang")
```
2.DefaultLang默认语言，默认为`zh-CN`，此项可选
```go
i18n.WithDefaultLang("zh-CN")
```
也可以在运行中通过
```go
i18n.SetLang("en-US")
```
来改变默认语言

3.EnvKey环境变量key，用于获取运行环境，默认为`RUN_MODE`，此项可选
```go
i18n.WithEnvKey("RUN_MODE")
```
4.DebugMode调试模式，用来设置返回结果中是否显示错误`trace`的详细信息，默认值为false，此项可选
```go
i18n.WithDebugMode(true)
```
## 可用方法
1.JSON 以`application/json`输出
```go
i18n.JSON(c *gin.Context, status int, data interface{}, err error)
```
2.JSONP 以`application/javascript`输出
```go
i18n.JSONP(c *gin.Context, status int, data interface{}, err error)
```
3.AsciiJSON 以`application/json`输出
```go
i18n.AsciiJSON(c *gin.Context, status int, data interface{}, err error)
```
4.PureJSON 以`application/json`输出
```go
i18n.PureJSON(c *gin.Context, status int, data interface{}, err error)
```
5.XML 以`application/xml`输出
```go
i18n.XML(c *gin.Context, status int, data interface{}, err error)
```
6.YAML 以`yaml`输出
```go
i18n.YAML(c *gin.Context, status int, data interface{}, err error)
```
7.SetLang 修改默认语言
```go
i18n.SetLang("en-US")
```
8.Trans 输出翻译后的结果
```go
i18n.(lang string, code string, params ...string)
```