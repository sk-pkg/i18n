package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestI18n(t *testing.T) {
	r := setupRouter()

	// 创建一个测试服务器。
	ts := httptest.NewServer(r)
	defer ts.Close()

	host := "http://127.0.0.1:8080"
	apis := []result{
		{Code: -1, Msg: "系统繁忙", Data: "busy", Trace: trace{
			ID:   "",
			Desc: "busy... ",
		}},
		{Code: 0, Msg: "ok", Data: "ok", Trace: trace{
			ID:   "1213gsgdfd",
			Desc: "",
		}},
		{Code: 500, Msg: "fail", Data: "fail", Trace: trace{
			ID:   "",
			Desc: "",
		}},
		{Code: 400, Msg: "Request parameter error", Data: "params", Trace: trace{
			ID:   "",
			Desc: "",
		}},
		{Code: 1000, Msg: "你好,Seakee!你的账号是:18888888888", Data: "test", Trace: trace{
			ID:   "",
			Desc: "",
		}},
	}

	for _, api := range apis {
		req, err := http.NewRequest("GET", host+"/"+api.Data.(string), nil)
		if err != nil {
			t.Fatal(err)
		}

		if api.Code == 400 {
			// 在请求 Header 中设置语言。
			req.Header.Set("lang", "en-US")
		}

		// 发送请求并获取响应。
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		resp := w.Result()
		resp.Body.Close()

		var res result
		if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, api, res)
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	msg, err := New(WithDebugMode(true))
	if err != nil {
		log.Fatal(err)
	}
	msg.SetLang("zh-CN")
	r.GET("/busy", func(c *gin.Context) {
		msg.JSON(c, -1, "busy", errors.New("busy... "))
	})

	r.GET("/ok", func(c *gin.Context) {
		c.Set("trace_id", "1213gsgdfd")
		msg.JSON(c, 0, "ok", nil)
	})

	r.GET("/fail", func(c *gin.Context) {
		msg.JSON(c, 500, "fail", nil)
	})

	r.GET("/params", func(c *gin.Context) {
		msg.JSON(c, 400, "params", nil)
	})

	r.GET("/test", func(c *gin.Context) {
		msg.JSON(c, 1000, Data{
			Params: []string{"Seakee", "18888888888"},
			Data:   "test",
		}, nil)
	})

	fmt.Println(msg.Trans("en-US", "1000", "Seakee", "18888888888"))
	fmt.Println(msg.Trans("zh-CN", "1000", "Seakee", "18888888888"))

	return r
}
