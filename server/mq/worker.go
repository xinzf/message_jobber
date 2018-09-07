package mq

import (
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"sync"
)

type worker struct {
	jobber *Jobber
	msg    <-chan amqp.Delivery
	id     int
	wg     *sync.WaitGroup
	locker *sync.Mutex
}

func (this *worker) Run() {

	defer func() {
		logrus.Infoln(this.id, " exits")
		this.wg.Done()
	}()

	notify := make(chan *amqp.Error)

BREAK:
	for {
		select {
		case <-this.jobber.ctx.Done():
			break BREAK
		case <-this.jobber.channel.NotifyClose(notify):
			break BREAK
		case delivery, ok := <-this.msg:
			if ok == false {
				break BREAK
			}

			if err := this.exec(delivery); err != nil {
				logrus.Errorln("Worker %s[%d] exec error: ", err.Error())
			}
		}
	}

	this.locker.Lock()
	this.locker.Unlock()

	return
}

func (this *worker) exec(d amqp.Delivery) error {
	this.locker.Lock()

	defer this.locker.Unlock()
	defer d.Ack(false)

	logrus.Infof("%s[%d]: %s", this.jobber.name, this.id, this.jobber.options.TargetUrl)

	return nil
}

func (this *worker) Stop() {
	//this.cancle()
}
