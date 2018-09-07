package main

import (
	"github.com/gin-gonic/gin"
	"gitlab.mydadao.com/marketing/message_jobber/config"
	"gitlab.mydadao.com/marketing/message_jobber/server/router"
	"net/http"

	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gitlab.mydadao.com/marketing/message_jobber/server/mq"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var cfg = pflag.StringP("config", "c", "", "config file path.")

	pflag.Parse()

	// init config
	if err := config.Init(*cfg); err != nil {
		panic(err)
	}

	ctx, _ := context.WithCancel(context.Background())
	mq.Init(ctx)

	g := gin.New()
	gin.SetMode(viper.GetString("server.runmode"))

	middlewares := []gin.HandlerFunc{}
	router.Load(g, middlewares...)

	serve(g)
}

func serve(g *gin.Engine) {
	log.Infof("http server listen on %s", viper.GetString("server.addr"))
	err := http.ListenAndServe(viper.GetString("server.addr"), g).Error()
	if err != "" {
		log.Panic(err)
	}
}
