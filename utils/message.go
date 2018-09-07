package utils

type Message struct {
	Data []byte
	Size int
}

func (this *Message) String() string {
	return string(this.Data[0:this.Size])
}

func (this *Message) Byte() []byte {
	return this.Data[0:this.Size]
}

func (this *Message) Len() int {
	return this.Size
}
