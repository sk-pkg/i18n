package main

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
		msg.JSON(c, 1000, i18n.Data{
			Params: []string{"Seakee", "18888888888"},
			Data:   "test",
		}, nil)
	})

	return r
}
