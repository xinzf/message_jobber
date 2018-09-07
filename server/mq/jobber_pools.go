package mq

import (
	"errors"
	"fmt"
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type jobberPools struct {
	jobbers *hashmap.Map // 正在运行中的 jobber
	changed *hashmap.Map
	removed []string
}

func (this *jobberPools) init() error {
	ops, err := this.read()
	if err != nil {
		return err
	}

	for _, o := range ops {
		this.jobbers.Put(o.Name, NewJobber(o))
	}

	return nil
}

func (this *jobberPools) read() (ops map[string]jobberOptions, err error) {
	includePath := viper.GetString("include")
	if includePath == "" {
		err = errors.New("Jobber config path is empty.")
		return
	}

	match, err := filepath.Glob(includePath)
	if err != nil {
		return
	}

	ops = make(map[string]jobberOptions)
	for _, v := range match {
		op, err := this.parseConfig(v)
		if err != nil {
			logrus.Errorf("Parse config: %s failed with error: %s", v, err.Error())
			continue
		}

		fileInfo, _ := os.Stat(v)
		op.configFile.filePath = v
		op.configFile.lastModified = fileInfo.ModTime()

		ops[op.Name] = op
	}

	return
}

func (this *jobberPools) parseConfig(fileName string) (options jobberOptions, err error) {
	var b []byte
	b, err = ioutil.ReadFile(fileName)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(b, &options)
	if err != nil {
		return
	}

	return options, err
}

func (this *jobberPools) Start(name string) error {
	var jb *Jobber

	temp, found := this.jobbers.Get(name)
	if !found {
		return errors.New(fmt.Sprintf("Not found jobber %s", name))
	}

	jb = temp.(*Jobber)

	status, _ := jb.GetStatus()
	if status == 1 {
		return errors.New(fmt.Sprintf("Jobber %s has started.", name))
	}

	go jb.Start()
	return nil
}

func (this *jobberPools) Stop(name string) error {
	temp, found := this.jobbers.Get(name)
	if !found {
		return errors.New(fmt.Sprintf("Not found jobber %s", name))
	}

	c := make(chan bool)
	jb := temp.(*Jobber)
	<-jb.Stop(c)

	return nil
}

func (this *jobberPools) Restart(name string) error {
	if err := this.Stop(name); err != nil {
		return err
	}

	return this.Start(name)
}

func (this *jobberPools) Reload(name string) {

}

func (this *jobberPools) StartAll() error {
	vals := this.jobbers.Keys()
	for _, v := range vals {
		name := v.(string)
		this.Start(name)
	}

	return nil
}

func (this *jobberPools) StopAll() error {
	temp := this.jobbers.Values()
	for _, v := range temp {
		j := v.(*Jobber)
		c := make(chan bool)
		<-j.Stop(c)
	}
	return nil
}

func (this *jobberPools) RestartAll() error {
	keys := this.jobbers.Keys()
	for _, v := range keys {
		name := v.(string)
		this.Restart(name)
	}
	return nil
}

func (this *jobberPools) List() []*Jobber {
	jbs := make([]*Jobber, 0)
	vals := this.jobbers.Values()
	for _, val := range vals {
		jbs = append(jbs, val.(*Jobber))
	}
	return jbs
}

func (this *jobberPools) Remove(name string) error {
	temp, found := this.jobbers.Get(name)
	if !found {
		return errors.New(fmt.Sprintf("Not found jobber %s", name))
	}

	jb := temp.(*Jobber)
	c := make(chan bool)
	<-jb.Stop(c)
	this.jobbers.Remove(name)
	return nil
}

func (this *jobberPools) Reread() (changeNames []string, removes []string, err error) {
	var ops map[string]jobberOptions
	ops, err = this.read()
	if err != nil {
		return
	}

	changeNames = make([]string, 0)
	removes = make([]string, 0)

	this.changed.Clear()
	this.removed = []string{}

	for _, op := range ops {
		temp, found := this.jobbers.Get(op.Name)
		if !found {
			this.changed.Put(op.Name, op)
			continue
		}

		jb := temp.(*Jobber)
		if op.configFile.lastModified.Unix() > jb.options.configFile.lastModified.Unix() {
			this.changed.Put(op.Name, op)
		}
	}

	keys := this.jobbers.Keys()
	for _, v := range keys {
		name := v.(string)
		if _, found := ops[name]; !found {
			this.removed = append(this.removed, name)
		}
	}

	temp := this.changed.Keys()
	for _, v := range temp {
		changeNames = append(changeNames, v.(string))
	}

	removes = this.removed

	return
}

// @todo 待测
func (this *jobberPools) Update() error {
	for _, v := range this.removed {
		this.Remove(v)
	}

	keys := this.changed.Keys()
	for _, k := range keys {
		name := k.(string)
		temp, found := this.jobbers.Get(name)
		if !found {
			temp, _ = this.changed.Get(name)
			op := temp.(jobberOptions)
			this.jobbers.Put(name, NewJobber(op))
			this.Start(name)
		} else {
			jb := temp.(*Jobber)
			temp, _ := this.changed.Get(name)
			op := temp.(jobberOptions)
			jb.options = op
		}
	}
	return nil
}
