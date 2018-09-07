package router

import (
	"gitlab.mydadao.com/marketing/message_jobber/server/handlers"
	"gitlab.mydadao.com/marketing/message_jobber/server/router/middleware"

	"github.com/gin-gonic/gin"
)

func Load(g *gin.Engine, mw ...gin.HandlerFunc) *gin.Engine {

	// 防止 Panic 把进程干死
	g.Use(gin.Recovery())
	g.Use(middleware.Logger())

	// 自定义的中间件
	g.Use(mw...)

	// 默认404
	g.NoRoute(func(context *gin.Context) {
		context.JSON(404, gin.H{
			"code": 404,
			"msg":  "请求地址有误，请核实",
			"data": gin.H{},
		})
	})

	g.GET("/", new(handlers.Home).Index)

	mq := g.Group("/mq")
	{
		mqHandler := new(handlers.Mq)
		mq.GET("/status", mqHandler.Status)
		mq.GET("/stop", mqHandler.Stop)
		mq.GET("/start", mqHandler.Start)
		mq.GET("/remove", mqHandler.Remove)
		mq.GET("/reread", mqHandler.Reread)
		mq.GET("/update", mqHandler.Update)
	}
	return g
}
