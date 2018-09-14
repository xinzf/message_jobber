package mq

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"sync"
	"sync/atomic"
	"time"
)

type connection struct {
	status int32
	conn   *amqp.Connection
	ctx    context.Context
	cancle context.CancelFunc
	once   sync.Once
}

func (this *connection) connect() error {
	logrus.Info("Try to connect rabbitMQ server.")

	dial := func(addr string) (*amqp.Connection, error) {
		u := fmt.Sprintf(
			"amqp://%s:%s@%s%s",
			Options.User,
			Options.Pswd,
			addr,
			Options.Vhost,
		)
		conn, err := amqp.Dial(u)
		return conn, err
	}

	var conn *amqp.Connection
	var err error
	for _, host := range Options.Brokers {
		conn, err = dial(host)
		if err == nil {
			atomic.StoreInt32(&this.status, 1)
			this.conn = conn
			break
		}
	}

	if err != nil {
		return err
	}

	atomic.StoreInt32(&this.status, 1)
	this.conn = conn

	vals := Jobbers.jobbers.Values()
	for _, val := range vals {
		jb := val.(*Jobber)
		status, _ := jb.GetStatus()
		if status == -1 {
			Jobbers.Start(jb.name)
		}
	}

	logrus.Info("Connect rabbitMQ server success.")
	return nil
}

func (this *connection) run(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case error:
				logrus.Panicln("RabbitMQ connection has disconnected with error: ", err.(error).Error())
			}
		}
	}()

	ctx, cancle := context.WithCancel(ctx)
	this.ctx = ctx
	this.cancle = cancle

	for this.status == 0 {
		this.connect()
		time.Sleep(2 * time.Second)
	}

	notify := make(chan *amqp.Error)
RETRY:
	for {
		select {
		case <-this.ctx.Done():
			this.close()
		case err, flag := <-this.conn.NotifyClose(notify):
			atomic.StoreInt32(&this.status, 0)

			Jobbers.StopAll()

			if !flag {
				logrus.Errorf("RabbitMQ connection has went away")
				return
			}

			if err != nil {
				logrus.Errorln("Closed by notify from RabbitMQ with error: ", err)
				break RETRY
			} else {
				logrus.Infoln("Closed by notify from RabbitMQ")
				return
			}
		}
	}

	go this.run(this.ctx)
}

func (this *connection) getChannel() (*amqp.Channel, error) {
	if this.status == 0 {
		return nil, errors.New("RabbitMQ has not connected.")
	}
	channel, err := this.conn.Channel()
	return channel, err
}

func (this *connection) close() {
	this.conn.Close()
}
