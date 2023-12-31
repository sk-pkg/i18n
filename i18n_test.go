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
	go r.Run(":8080")

	host := "http://127.0.0.1:8080"
	apis := []result{
		{Code: -1, Msg: "系统繁忙", Data: "busy", Trace: "busy... "},
		{Code: 0, Msg: "ok", Data: "ok", Trace: ""},
		{Code: 500, Msg: "fail", Data: "fail", Trace: ""},
		{Code: 400, Msg: "请求参数错误", Data: "params", Trace: ""},
		{Code: 1000, Msg: "你好,Seakee!你的账号是:18888888888", Data: "test", Trace: ""},
	}

	for _, api := range apis {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", host+"/"+api.Data.(string), nil)
		r.ServeHTTP(w, req)

		rs := result{}
		err := json.Unmarshal(w.Body.Bytes(), &rs)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, api, rs)
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
