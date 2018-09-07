package errno

type Errno struct {
	Code    int
	Message string
}

func (err *Errno) Error() string {
	return err.Message
}

func (this *Errno) Add(msg string) *Errno {
	this.Message += " " + msg
	return this
}
