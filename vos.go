package protoconvertreq

import (
	"errors"
	"github.com/dop251/goja"
)

type ProcessInfo struct {
	Process []*RequestContainer `yaml:"process,omitempty"`
}

type RequestContainer struct {
	Http  *RequestInfos `yaml:"http,omitempty"`
	Https *RequestInfos `yaml:"https,omitempty"`
}

func (this *RequestContainer) Get() (string, *RequestInfos, error) {
	if this.Http != nil {
		return "http", this.Http, nil
	}
	if this.Https != nil {
		return "https", this.Https, nil
	}

	return "", nil, errors.New("未被允许的请求数据")
}

type RequestInfos struct {
	Desc           string            `yaml:"desc,omitempty"`
	Url            string            `yaml:"url,omitempty"`
	CaCertPem      string            `yaml:"CaCertPem,omitempty"`
	ClientCertPem  string            `yaml:"clientCertPem,omitempty"`
	ClientKeyPem   string            `yaml:"clientKeyPem,omitempty"`
	SkipVerify     bool              `yaml:"skipVerify,omitempty"`
	Method         string            `yaml:"method,omitempty"`
	Headers        map[string]string `yaml:"headers,omitempty"`
	ReqBody        string            `yaml:"reqBody,omitempty"`
	ResponseParser *ResponseParser   `yaml:"responseParser,omitempty"`
	ArgsCheck      string            `yaml:"argsCheck,omitempty"`
}

type ResponseParser struct {
	Simple string   `yaml:"simple,omitempty"`
	Fns    []string `yaml:"fns,omitempty"`
	ErrFn  string   `yaml:"errFn,omitempty"`
}

type ProcessData struct {
	// Out 外置参数
	Out interface{}
	// Global 全局参数
	Global interface{}
	// Http http封装
	Http    *ProcessReqAndResContainerData
	Https   *ProcessReqAndResContainerData
	Req     *ProcessReqData
	Res     *ProcessResData
	Prev    interface{}
	History []interface{}
}

type ProcessReqAndResContainerData struct {
	Prev *ProcessReqAndResData
	List []*ProcessReqAndResData
}

type ProcessReqAndResData struct {
	Req *ProcessReqData
	Res *ProcessResData
}

type ProcessReqData struct {
	Data    string
	Headers map[string]string
}

type ProcessResData struct {
	Data       string
	StatusCode int
	Obj        interface{}
}

type jsVmInfo struct {
	jsVm *goja.Runtime
	data *ProcessData
}
