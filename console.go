package main

import (
	"context"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gitlab.mydadao.com/marketing/message_jobber/config"
	"gitlab.mydadao.com/marketing/message_jobber/console"
)

func main() {
	var cfg = pflag.StringP("config", "c", "", "config file path.")

	pflag.Parse()

	if err := config.Init(*cfg); err != nil {
		panic(err)
	}

	inter := &console.Interactive{
		ServerUrl: viper.GetString("server.addr"),
	}
	inter.Run(context.TODO())
}
