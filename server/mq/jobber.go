package mq

import (
	"context"
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"sync"
	"sync/atomic"
	"time"
)

type jobberOptions struct {
	Name     string
	Queue    string
	Exchange struct {
		Name  string
		Etype string `yaml:"type"`
	}
	BindKey    string `yaml:"bindkey"`
	Consumer   string
	WorkerNum  int    `yaml:"workernum"`
	TargetUrl  string `yaml:"url"`
	configFile struct {
		filePath     string
		lastModified time.Time
	}
}

/**
 * NewJobber，创建一个 Jobber，但不启动
 * options 配置
 * fileName 配置文件名称
 * lastModified 配置文件的最后修改日期
 */
func NewJobber(options jobberOptions) *Jobber {
	return &Jobber{
		name:          options.Name,
		options:       options,
		workers:       hashmap.New(),
		stopTime:      time.Now(),
		closeNotifies: make([]chan bool, 0),
		status:        0,
	}
}

type Jobber struct {
	name          string
	channel       *amqp.Channel
	options       jobberOptions
	workers       *hashmap.Map
	ctx           context.Context
	cancle        context.CancelFunc
	once          sync.Once
	status        int32
	closeNotifies []chan bool
	startTime     time.Time
	stopTime      time.Time
}

// Start 启动 Jobber
func (this *Jobber) Start() {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case error:
				//atomic.StoreInt32(&this.status, -1)
				//logrus.Errorf("Jobber: %s exits with error: %s", this.name, err.(error).Error())
				//default:
				//	logrus.Infof("Jobber: %s exits", this.name)
			}
		} else {
			logrus.Infof("Jobber: %s exits", this.name)
		}
	}()

	if atomic.LoadInt32(&this.status) == 1 {
		logrus.Errorf("Jobber: %s is running.", this.name)
		return
	}

	ctx, cancle := context.WithCancel(context.Background())
	this.ctx = ctx
	this.cancle = cancle

	var err error
	this.channel, err = Connection.getChannel()
	if err != nil {
		logrus.Errorf("Jobber: %s start failed with error: %s", this.name, err.Error())
		return
	}

	_, err = this.channel.QueueDeclare(
		this.options.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logrus.Errorf("Jobber: %s declare queue %s failed with error: %s", this.name, this.options.Queue, err.Error())
		return
	}
	//logrus.Infof("Jobber %s declare queue %s success", this.name, this.options.Queue)

	err = this.channel.ExchangeDeclare(
		this.options.Exchange.Name,
		this.options.Exchange.Etype,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logrus.Errorf("Jobber: %s declare exchange %s failed with error: %s", this.name, this.options.Exchange.Name, err.Error())
		return
	}

	err = this.channel.QueueBind(
		this.options.Queue,
		this.options.BindKey,
		this.options.Exchange.Name,
		false,
		nil,
	)
	if err != nil {
		logrus.Errorf("Jobber: %s bind queue %s to exchange %s failed with error: %s", this.name, this.options.Queue, this.options.Exchange.Name, err.Error())
		return
	}

	err = this.channel.Qos(this.options.WorkerNum, 0, false)
	if err != nil {
		logrus.Errorf("Jobber: %s set basic qos failed with error: %s", this.name, err.Error())
		return
	}

	msg, err := this.channel.Consume(
		this.options.Queue,
		this.options.Consumer,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logrus.Errorf("Jobber: %s consume failed with error: %s", this.name, err.Error())
		return
	}

	atomic.StoreInt32(&this.status, 1)
	this.startTime = time.Now()
	logrus.Infof("Jobber: %s started success.", this.name)

	wg := new(sync.WaitGroup)
	for i := 0; i < this.options.WorkerNum; i++ {
		w := &worker{
			jobber: this,
			msg:    msg,
			id:     i,
			wg:     wg,
			locker: new(sync.Mutex),
		}
		this.workers.Put(i, w)
		wg.Add(1)
		go w.Run()
	}
	wg.Wait()

	logrus.Infoln("到这了111")
	atomic.StoreInt32(&this.status, 0)
	this.stopTime = time.Now()
	logrus.Infoln("到这了22")

	logrus.Infoln("到这了333")
	if len(this.closeNotifies) > 0 {
		for _, c := range this.closeNotifies {
			close(c)
		}
		this.closeNotifies = []chan bool{}
	}

	this.channel.Close()
	logrus.Infoln("到这了444")
}

// Stop 停止，并阻塞等待停止完成
func (this *Jobber) Stop(c chan bool) chan bool {
	s := atomic.LoadInt32(&this.status)
	if s != 1 {
		close(c)
	} else {
		this.closeNotifies = append(this.closeNotifies, c)
		this.cancle()
	}

	return c
}

// GetStatus 获取当前状态
func (this *Jobber) GetStatus() (int32, string) {
	s := atomic.LoadInt32(&this.status)
	var str string
	switch s {
	case 0:
		str = "STOPPED"
	case 1:
		str = "RUNNING"
	case -1:
		str = "FATAL"
	}

	return s, str
}

// GetName 获取当前 Jobber.Name
func (this *Jobber) GetName() string {
	return this.name
}

// GetQueueName 获取监听的队列名称
func (this *Jobber) GetQueueName() string {
	return this.options.Queue
}

// GetWorkers 获取所有的 workers
func (this *Jobber) GetWorkers() []*worker {
	wks := make([]*worker, 0)
	vals := this.workers.Values()
	for _, v := range vals {
		wks = append(wks, v.(*worker))
	}

	return wks
}

// GetStartTime 获取开始日期
func (this *Jobber) GetStartTime() time.Time {
	return this.startTime
}

// GetStopTime 获取停止日期
func (this *Jobber) GetStopTime() time.Time {
	return this.stopTime
}
