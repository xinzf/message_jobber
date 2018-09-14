package mq

import (
	"context"
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/spf13/viper"
)

const (
	FANOUT = "fanout"
	DIRECT = "direct"
)

var (
	Jobbers = &jobberPools{
		jobbers: hashmap.New(),
		changed: hashmap.New(),
	}

	Options = struct {
		Brokers []string
		User    string
		Pswd    string
		Vhost   string
	}{}

	Connection = new(connection)
)

func Init(ctx context.Context) error {
	Options.Brokers = viper.GetStringSlice("server.rabbitmq.brokers")
	Options.User = viper.GetString("server.rabbitmq.user")
	Options.Pswd = viper.GetString("server.rabbitmq.pswd")
	Options.Vhost = viper.GetString("server.rabbitmq.vhost")
	if err := Jobbers.init(); err != nil {
		return err
	}
	go Connection.run(ctx)

	return nil
}
