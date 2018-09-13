package console

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"gitlab.mydadao.com/marketing/message_jobber/responses"
	"os"
	"strings"
)

type Command string

func (this Command) String() string {
	return string(this)
}

const (
	EXIT   Command = "exit"
	QUIT   Command = "quit"
	HELP   Command = "help"
	ENTER  Command = ""
	UNKNOW Command = "unknow"

	ADD      Command = "add"
	CLEAR    Command = "clear"
	START    Command = "start"
	STOP     Command = "stop"
	RESTART  Command = "restart"
	REREAD   Command = "reread"
	REMOVE   Command = "remove"
	UPDATE   Command = "update"
	RELOAD   Command = "reload"
	SHUTDOWN Command = "shutdown"
	STATUS   Command = "status"
	TAIL     Command = "tail"
	VERSION  Command = "version"
)

// *** Unknown syntax: grdszx

var commands = []Command{
	EXIT, QUIT, HELP, ENTER, UNKNOW, ADD, CLEAR, START, STOP, RESTART, REREAD, REMOVE, UPDATE, RELOAD, SHUTDOWN, STATUS, TAIL, VERSION,
}

type cmd struct {
	command Command
	data    []string
}

type Interactive struct {
	ServerUrl string
}

func (this *Interactive) Run(ctx context.Context) {
	scanner := bufio.NewScanner(os.Stdin)
	this.status()

	for scanner.Scan() {
		cmd := this.parse(scanner)
		switch cmd.command {
		case EXIT, QUIT:
			this.Exit()
		case ENTER:
			this.response()
		case HELP:
			this.help()
		case STATUS:
			this.status()
		case STOP:
			this.stop(cmd)
		case START:
			this.start(cmd)
		case REMOVE:
			this.remove(cmd)
		case REREAD:
			this.reread()
		case UPDATE:
			this.update()
		case RESTART:
			this.restart(cmd)
		default:
			this.response("*** Unknown syntax ***")
		}
	}
}

func (this *Interactive) parse(scanner *bufio.Scanner) cmd {
	line := scanner.Text()

	if line == "" {
		return cmd{
			command: ENTER,
			data:    make([]string, 0),
		}
	}

	list := strings.Split(line, " ")
	newList := make([]string, 0)
	for i := 0; i < len(list); i++ {
		if list[i] != "" {
			newList = append(newList, list[i])
		}
	}

	if len(newList) == 0 {
		return cmd{
			command: ENTER,
			data:    make([]string, 0),
		}
	}

	var commd Command = UNKNOW
	for _, c := range commands {
		if c.String() == newList[0] {
			commd = c
			break
		}
	}

	return cmd{
		command: commd,
		data:    newList[1:],
	}
}

func (this *Interactive) response(content ...string) {
	var str string
	if len(content) > 0 {
		str = content[0]
	}

	if str != "" {
		fmt.Println(str)
	}

	fmt.Print("jobber> ")
}

func (this *Interactive) Exit() {
	os.Exit(0)
}

func (this *Interactive) help() {
	this.response(`default commands (type help <topic>):
=====================================
add    clear  fg        open  quit    remove  restart   start   stop  update
avail  exit   maintail  pid   reload  reread  shutdown  status  tail  version`)
}

func (this *Interactive) status() {
	//this.response(this.ServerUrl+"/mq/status")
	res := Get("http://" + this.ServerUrl + "/mq/status")
	//this.response(res)
	//logrus.Infoln(res)
	if res.Success() == false {
		this.response(res.Message)
		return
	}

	data := make([]responses.StatusResponse, 0)
	err := json.Unmarshal(res.Attachment, &data)
	if err != nil {
		this.response(err.Error())
		return
	}

	nameMaxLength := 0
	queueMaxLength := 0
	spaceNum := 4
	for _, jb := range data {
		if len(jb.Name) > nameMaxLength {
			nameMaxLength = len(jb.Name)
		}

		if len(jb.QueueName) > queueMaxLength {
			queueMaxLength = len(jb.QueueName)
		}
	}

	var str string
	for _, jb := range data {
		var (
			nameLength  int
			nameNum     int
			queueLength int
			queueNum    int
		)

		nameLength = len(jb.Name)
		nameNum = spaceNum + nameMaxLength - nameLength

		queueLength = len(jb.QueueName)
		queueNum = spaceNum + queueMaxLength - queueLength

		str = str + fmt.Sprintf(
			"%s%s%s%s%s%s%s",
			jb.Name,
			strings.Repeat(" ", nameNum),
			jb.QueueName,
			strings.Repeat(" ", queueNum),
			jb.Status,
			strings.Repeat(" ", spaceNum),
			jb.StatusTime,
		)
		str = str + "\n"
	}
	str = strings.TrimRight(str, "\n")
	this.response(str)
}

func (this *Interactive) stop(c cmd) {
	if len(c.data) == 0 {
		this.response(`Error: stop requires a process name
stop <name>		Stop a process
stop <gname>:*		Stop all processes in a group
stop <name> <name>	Stop multiple processes or groups
stop all		Stop all processes`)
		return
	}
	name := c.data[0]
	res := Get("http://" + this.ServerUrl + "/mq/stop?name=" + name)
	if res.Success() == false {
		this.response(res.Message)
	} else {
		this.response(res.String())
	}
}

func (this *Interactive) start(c cmd) {
	if len(c.data) == 0 {
		this.response(`Error: start requires a process name
start <name>		Start a process
start <gname>:*		Start all processes in a group
start <name> <name>	Start multiple processes or groups
start all		Start all processes`)
		return
	}
	name := c.data[0]
	res := Get("http://" + this.ServerUrl + "/mq/start?name=" + name)
	if res.Success() == false {
		this.response(res.Message)
	} else {
		this.response(res.String())
	}
}

func (this *Interactive) remove(c cmd) {
	if len(c.data) == 0 {
		this.response(`Error: remove requires a jobber name
remove <name>		Remove a jobber`)
		return
	}

	name := c.data[0]
	res := Get("http://" + this.ServerUrl + "/mq/remove?name=" + name)
	if res.Success() == false {
		this.response(res.Message)
	} else {
		this.response(res.String())
	}
}

func (this *Interactive) reread() {
	res := Get("http://" + this.ServerUrl + "/mq/reread")
	if res.Success() == false {
		this.response(res.Message)
		return
	}

	var data responses.RereadResponse
	err := json.Unmarshal(res.Attachment, &data)
	if err != nil {
		this.response(err.Error())
		return
	}

	var str string

	if len(data.Changes) > 0 {
		str += "Changes: \n\t"
		for _, s := range data.Changes {
			str = str + s + ", "
		}
		str = strings.TrimRight(str, ", ")
		str += "\n"
	}

	if len(data.Removes) > 0 {
		str += "Removes:\n\t"
		for _, s := range data.Removes {
			str = str + s + ", "
		}
		str = strings.TrimRight(str, ", ")
	}

	this.response(str)
}

func (this *Interactive) update() {
	res := Get("http://" + this.ServerUrl + "/mq/update")
	if res.Success() == false {
		this.response(res.Message)
		return
	}

	this.response("Updated all success.")
}

func (this *Interactive) restart(c cmd) {
	if len(c.data) == 0 {
		this.response(`Error: restart requires a jobber name
restart <name>		restart a process
restart all		Restart all jobbers`)
		return
	}
	name := c.data[0]
	res := Get("http://" + this.ServerUrl + "/mq/restart?name=" + name)
	if res.Success() == false {
		this.response(res.Message)
	} else {
		this.response(res.String())
	}
}
