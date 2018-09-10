package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gitlab.mydadao.com/marketing/message_jobber/responses"
	"gitlab.mydadao.com/marketing/message_jobber/server/mq"
	"gitlab.mydadao.com/marketing/message_jobber/server/pkg/errno"
	"gitlab.mydadao.com/marketing/wechat/src/utils"
)

type Mq struct {
	Base
}

func (this *Mq) Start(c *gin.Context) {
	name := c.DefaultQuery("name", "")
	if name == "" {
		this.Failed(c, errno.ParamsErr.Add("name"))
		return
	}

	var err error
	if name == "all" {
		err = mq.Jobbers.StartAll()
	} else {
		err = mq.Jobbers.Start(name)
	}

	if err != nil {
		this.Failed(c, errno.InternalServerError.Add(err.Error()))
		return
	}

	if name == "all" {
		this.Success(c, "All started.")
	} else {
		this.Success(c, fmt.Sprintf("%s started.", name))
	}

}

func (this *Mq) Stop(c *gin.Context) {
	name := c.DefaultQuery("name", "")
	if name == "" {
		this.Failed(c, errno.ParamsErr.Add("name"))
		return
	}

	var err error
	if name == "all" {
		err = mq.Jobbers.StopAll()
	} else {
		err = mq.Jobbers.Stop(name)
	}

	if err != nil {
		this.Failed(c, errno.InternalServerError.Add(err.Error()))
		return
	}

	if name == "all" {
		this.Success(c, "All stoped.")
	} else {
		this.Success(c, fmt.Sprintf("%s stopped", name))
	}
}

func (this *Mq) Status(c *gin.Context) {
	jbs := mq.Jobbers.List()

	logrus.Infoln(len(jbs))
	list := make([]responses.StatusResponse, 0, len(jbs))

	for _, jb := range jbs {
		status, statusStr := jb.GetStatus()

		var t string
		if status == 1 {
			t = utils.TimeFormat(jb.GetStartTime())
		} else {
			t = utils.TimeFormat(jb.GetStopTime())
		}

		list = append(list, responses.StatusResponse{
			Name:       jb.GetName(),
			QueueName:  jb.GetQueueName(),
			Status:     statusStr,
			StatusTime: t,
		})
	}

	this.Success(c, list)
}

func (this *Mq) Reread(c *gin.Context) {
	changes, removes, err := mq.Jobbers.Reread()
	if err != nil {
		this.Failed(c, errno.InternalServerError.Add(err.Error()))
		return
	}

	this.Success(c, gin.H{
		"changes": changes,
		"removes": removes,
	})
}

func (this *Mq) Update(c *gin.Context) {
	mq.Jobbers.Update()
	this.Success(c, "")
}

func (this *Mq) Remove(c *gin.Context) {
	name := c.DefaultQuery("name", "")
	if name == "" {
		this.Failed(c, errno.ParamsErr.Add("name"))
		return
	}

	err := mq.Jobbers.Remove(name)
	if err != nil {
		this.Failed(c, errno.InternalServerError.Add(err.Error()))
		return
	}

	this.Success(c, fmt.Sprintf("Jobber %s removed", name))
}

func (this *Mq) Restart(c *gin.Context) {
	name := c.DefaultQuery("name", "")
	if name == "" {
		this.Failed(c, errno.ParamsErr.Add("name"))
		return
	}

	var err error
	if name == "all" {
		err = mq.Jobbers.RestartAll()
	} else {
		err = mq.Jobbers.Restart(name)
	}

	if err != nil {
		this.Failed(c, errno.InternalServerError.Add(err.Error()))
		return
	}

	if name == "all" {
		this.Success(c, "All restarted.")
	} else {
		this.Success(c, fmt.Sprintf("%s restarted", name))
	}
}