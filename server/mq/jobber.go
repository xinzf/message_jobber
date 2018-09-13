package mq

import (
	"bytes"
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"io/ioutil"
	"net/http"
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
	LogPath    string `yaml:"logpath"`
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
		stopTime:      time.Now(),
		closeNotifies: make([]chan bool, 0),
		status:        0,
		logger:        NewLogger(options.LogPath),
	}
}

type Jobber struct {
	name          string
	channel       *amqp.Channel
	options       jobberOptions
	ctx           context.Context
	cancle        context.CancelFunc
	once          sync.Once
	status        int32
	closeNotifies []chan bool
	startTime     time.Time
	stopTime      time.Time
	workers       chan int
	logger        *logger
}

func (this *Jobber) preparStart() (msg <-chan amqp.Delivery, err error) {
	ctx, cancle := context.WithCancel(context.Background())
	this.ctx = ctx
	this.cancle = cancle

	// 获取一个 channel
	this.channel, err = Connection.getChannel()
	if err != nil {
		return
	}

	// 创建队列
	_, err = this.channel.QueueDeclare(
		this.options.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return
	}

	// 创建路由
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
		return
	}

	// 绑定队列到路由
	err = this.channel.QueueBind(
		this.options.Queue,
		this.options.BindKey,
		this.options.Exchange.Name,
		false,
		nil,
	)
	if err != nil {
		return
	}

	// 设置 QOS
	err = this.channel.Qos(this.options.WorkerNum, 0, false)
	if err != nil {
		return
	}

	// 订阅队列
	msg, err = this.channel.Consume(
		this.options.Queue,
		this.options.Consumer,
		false,
		false,
		false,
		false,
		nil,
	)

	// 初始化工作线程池，线程池容量等于 mq.prefetchCount
	this.workers = make(chan int, this.options.WorkerNum)
	for i := 0; i < this.options.WorkerNum; i++ {
		this.workers <- i
	}

	// 初始化 jobber stop 的阻塞通知池
	this.closeNotifies = make([]chan bool, 0)

	// 设置状态和开始时间
	atomic.StoreInt32(&this.status, 1)
	this.startTime = time.Now()

	return
}

// Start 启动 Jobber
func (this *Jobber) Start() {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case error:
				logrus.Errorln(err)
			}
		} else {
			logrus.Infof("Jobber: %s exits", this.name)
		}
	}()

	var (
		msg <-chan amqp.Delivery
		err error
	)

	if atomic.LoadInt32(&this.status) == 1 {
		this.logger.Warnln("Jobber is always running,can't start again.")
		return
	}

	if msg, err = this.preparStart(); err != nil {
		this.logger.Errorln(err)
		return
	}

	this.logger.Infoln("Jobber started successful.")

	notify := make(chan *amqp.Error)
	var runErr error
BREAK:
	for {
		select {
		case <-this.ctx.Done():
			break BREAK
		case runErr = <-this.channel.NotifyClose(notify):
			break BREAK
		case delivery, ok := <-msg:
			if !ok {
				runErr = errors.New("delivery channel has closed")
				break BREAK
			}

			i, ok := <-this.workers
			if !ok {
				runErr = errors.New("workers channel has closed")
				break BREAK
			}

			if atomic.LoadInt32(&this.status) != 1 {
				break BREAK
			}

			go this.do(delivery, i)
		}
	}

	// 等待所有工作线程退出
	for i := 0; i < this.options.WorkerNum; i++ {
		<-this.workers
	}
	close(this.workers)

	// 根据运行中的错误情况判定，程序是正常退出还是异常退出
	if runErr != nil {
		atomic.StoreInt32(&this.status, -1)
	} else {
		atomic.StoreInt32(&this.status, 0)
	}
	this.stopTime = time.Now()

	// 通知所有需要得知当前 Jobber 退出情况的监听者
	if len(this.closeNotifies) > 0 {
		for _, c := range this.closeNotifies {
			close(c)
		}
	}

	if runErr != nil {
		this.logger.Errorln("Jobber exits with error: ", runErr.Error())
	} else {
		this.logger.Infoln("Jobber exits.")
	}
	this.channel.Close()
}

func (this *Jobber) do(msg amqp.Delivery, i int) {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case error:
				this.logger.With("workerId", i).Errorln("Jobber do request has some error: ", err.(error).Error())
			}
		}

		msg.Ack(false)
		this.workers <- i
		//this.logger.With("workerId",i).Info("Do request end")
	}()

	post := func(postData []byte, u string) ([]byte, int, error) {
		httpCode := 0
		client := &http.Client{}
		rawData := bytes.NewBuffer(postData)
		req, err := http.NewRequest("POST", u, rawData)
		if err != nil {
			return []byte(""), httpCode, err
		}
		req.Header.Set("Content-type", "application/json")

		response, err := client.Do(req)
		if err != nil {
			return []byte(""), httpCode, err
		}

		this.logger.Warnln("do request")
		if response.StatusCode != 200 {
			httpCode = response.StatusCode
			return []byte(""), httpCode, nil
		}
		defer response.Body.Close()

		body, _ := ioutil.ReadAll(response.Body)

		return body, response.StatusCode, nil
	}

	rsp, httpcode, err := post(msg.Body, this.options.TargetUrl)

	if err != nil {
		this.logger.WithFields(map[string]interface{}{
			"delivery":  string(msg.Body[:]),
			"http_code": httpcode,
			"workerId":  i,
			"response":  string(rsp[:]),
			"url":       this.options.TargetUrl,
		}).Warnln("end request")
	} else {
		this.logger.WithFields(map[string]interface{}{
			"delivery":  string(msg.Body[:]),
			"http_code": httpcode,
			"workerId":  i,
			"response":  string(rsp[:]),
			"url":       this.options.TargetUrl,
		}).Infoln("end request")
	}
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
//func (this *Jobber) GetWorkers() []*worker {
//	wks := make([]*worker, 0)
//	vals := this.workers.Values()
//	for _, v := range vals {
//		wks = append(wks, v.(*worker))
//	}
//
//	return wks
//}

// GetStartTime 获取开始日期
func (this *Jobber) GetStartTime() time.Time {
	return this.startTime
}

// GetStopTime 获取停止日期
func (this *Jobber) GetStopTime() time.Time {
	return this.stopTime
}
