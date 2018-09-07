package main

import (
	"context"
	"github.com/spf13/pflag"
	"gitlab.mydadao.com/marketing/message_jobber/config"
	"gitlab.mydadao.com/marketing/message_jobber/server/mq"
	"time"
)

var (
	cfg = pflag.StringP("config", "c", "", "config file path.")
)

func main() {
	pflag.Parse()

	if err := config.Init(*cfg); err != nil {
		panic(err)
	}

	ctx, _ := context.WithCancel(context.Background())
	mq.Init(ctx)

	time.Sleep(time.Hour)
}

//func main() {
//mq.Options.Brokers = []string{
//	"127.0.0.1:5672",
//	"127.0.0.1:5673",
//	"127.0.0.1:5674",
//}
//mq.Options.User = "guest"
//mq.Options.Pswd = "guest"
//mq.Options.Vhost = "/"
//
//ctx, _ := context.WithCancel(context.Background())
//mq.Init(ctx)

//time.Sleep(3 * time.Second)
//cancle()
//time.Sleep(time.Hour)

//}
