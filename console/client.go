package console

import "net/http"
import (
	"encoding/json"
	"io/ioutil"
)

type Response struct {
	Code       int             `json:"msg_code"`
	Message    string          `json:"message"`
	Attachment json.RawMessage `json:"attachment"`
}

func (this *Response) Success() bool {
	if this.Code != 200 {
		return false
	}
	return true
}

func (this *Response) String() string {
	if this.Attachment == nil {
		return ""
	}

	return string(this.Attachment)
}

func Get(uri string) (res *Response) {
	res = new(Response)
	resp, err := http.Get(uri)
	if err != nil {
		res.Code = -1
		res.Message = err.Error()
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		res.Code = -1
		res.Message = err.Error()
		return
	}

	err = json.Unmarshal(body, res)
	if err != nil {
		res.Code = -1
		res.Message = err.Error()
		return
	}

	return
}
