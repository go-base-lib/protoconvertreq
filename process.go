package protoconvertreq

import (
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja/unistring"
	"github.com/go-base-lib/logs"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"sync"
	"text/template"
)

var processLock = &sync.Mutex{}

var JsVmPoolSize = 16

var jsVmPool chan *jsVmInfo

//
//func init() {
//	InitJsVmPool(32)
//}

func InitJsVmPool() {
	processLock.Lock()
	defer processLock.Unlock()
	if jsVmPool != nil {
		return
	}
	jsVmPool = make(chan *jsVmInfo, JsVmPoolSize)
	for i := 0; i < JsVmPoolSize; i++ {
		runtime := goja.New()
		data := &ProcessData{}
		_ = runtime.Set("$", data)
		jsVmPool <- &jsVmInfo{
			jsVm: runtime,
			data: data,
		}
	}
	logs.Debugln("js虚拟机池初始化完成, 数量 => ", JsVmPoolSize)
}

var allowProcessMap = map[string]func(info *RequestInfos, jsVm *goja.Runtime) (ReqProxyInterface, error){
	"http": func(info *RequestInfos, jsVm *goja.Runtime) (ReqProxyInterface, error) {
		req, err := NewHttpReq(info, jsVm)
		if err != nil {
			return nil, err
		}
		return req, nil
	},
}

func RegistryTemplateFns(fns template.FuncMap) {
	templateParser.Funcs(fns)
}

type ProcessExec struct {
	processInfos *ProcessInfo
	nowIndex     int
	processLen   int
	jsInfo       *jsVmInfo
	outData      interface{}
	//data         *ProcessData
	//jsVm         *goja.Runtime
	sync.Mutex
}

func NewProcessByYamlPath(p string, data interface{}) (*ProcessExec, error) {
	file, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, errors.New("打开yaml文件失败")
	}
	return NewProcessByYamlStr(string(file), data)
}

func NewProcessByYamlStr(yamlStr string, data interface{}) (*ProcessExec, error) {
	info := &ProcessInfo{}
	if err := yaml.Unmarshal([]byte(yamlStr), info); err != nil {
		return nil, errors.New("解析yaml文件格式失败")
	}
	return NewProcess(info, data), nil
}

func NewProcess(info *ProcessInfo, data interface{}) *ProcessExec {
	if jsVmPool == nil {
		InitJsVmPool()
	}
	requestInfoLen := len(info.Process)
	//processJsInfo := <-jsVmPool
	//processJsInfo.data.Out = data

	//d := &ProcessData{
	//	Out: data,
	//}
	//jsVm := goja.New()
	//_ = jsVm.Set("$", d)
	tmpExec := &ProcessExec{
		processInfos: info,
		nowIndex:     -1,
		processLen:   requestInfoLen,
		//jsInfo:       processJsInfo,
		outData: data,
		//data:         d,
		//jsVm:         jsVm,
	}
	tmpExec.init()
	return tmpExec
}

func (this *ProcessExec) init() {
	processJsInfo := <-jsVmPool
	fromString := unistring.NewFromString("keyAuthInfo")
	fmt.Println(fromString)
	processJsInfo.data.Out = this.outData
	processJsInfo.data.Http = &ProcessReqAndResContainerData{
		List: make([]*ProcessReqAndResData, 0, this.processLen),
	}
	processJsInfo.data.Https = &ProcessReqAndResContainerData{
		List: make([]*ProcessReqAndResData, 0, this.processLen),
	}
	processJsInfo.data.History = make([]interface{}, 0, 8)
	processJsInfo.data.Req = &ProcessReqData{}
	processJsInfo.data.Res = &ProcessResData{}
	this.jsInfo = processJsInfo
}

func (this *ProcessExec) Reset() {
	this.Lock()
	defer this.Unlock()
	this.nowIndex = -1
	if this.jsInfo == nil {
		this.init()
		return
	}
	this.jsInfo.data.Https.List = this.jsInfo.data.Https.List[:0]
	this.jsInfo.data.Http.List = this.jsInfo.data.Https.List[:0]
	this.jsInfo.data.History = this.jsInfo.data.History[:0]
}

func (this *ProcessExec) Destroy() {
	this.Lock()
	defer this.Unlock()

	if this.jsInfo == nil {
		return
	}
	jsVmPool <- this.jsInfo
	this.jsInfo = nil
}

func (this *ProcessExec) Next() bool {
	this.Lock()
	defer this.Unlock()
	this.nowIndex += 1
	if this.nowIndex >= this.processLen {
		return false
	}
	return true
}

func (this *ProcessExec) ExecAll() (ResHandleInterface, *ReqProxyError) {
	var res ResHandleInterface
	for this.Next() {
		execRes, proxyError := this.Exec()
		if proxyError != nil {
			return nil, proxyError
		}
		res = execRes
	}
	return res, nil
}

func (this *ProcessExec) Exec() (ResHandleInterface, *ReqProxyError) {
	this.Lock()
	defer this.Unlock()

	if this.nowIndex >= this.processLen {
		return nil, NewReqProxyError(io.EOF, "EOF", "没有更多数据")
	}

	if this.nowIndex < 0 {
		return nil, NewReqProxyError(nil, "CALL_NEXT", "请先调用Next方法")
	}

	reqContainer := this.processInfos.Process[this.nowIndex]
	name, requestInfos, err := reqContainer.Get()
	if err != nil {
		return nil, NewReqProxyError(err, "NO_ALLOW", err.Error())
	}

	if fn, ok := allowProcessMap[name]; !ok {
		return nil, NewReqProxyError(nil, "NO_ALLOW", "未被允许的请求方式")
	} else {
		if requestInfos.ArgsCheck != "" {
			logs.Debugln("请求之前检测参数, 检测脚本: \n", requestInfos.ArgsCheck)
			_, e := this.jsInfo.jsVm.RunString(requestInfos.ArgsCheck)
			if e != nil {
				logs.Debugln("检测脚本未能正确通过, 报错 => ", e.Error())
				return nil, NewReqProxyWithCode(ConvertGojaException(e), "ARGS_REQUEST_CHECK")
			}
		}

		//return fn(requestInfos)
		proxyInterface, proxyErr := fn(requestInfos, this.jsInfo.jsVm)
		if proxyErr != nil {
			return nil, NewReqProxyError(proxyErr, "INIT_EXEC_CALL", proxyErr.Error())
		}
		res, proxyError := proxyInterface.Send(this.jsInfo.data)
		if proxyError != nil {
			return nil, proxyError
		}

		//prevData := map[string]interface{}{
		//	"Req": this.Data["Req"],
		//	"Res": this.Data["Res"],
		//}

		var processContainer *ProcessReqAndResContainerData
		switch name {
		case "http":
			processContainer = this.jsInfo.data.Http
			if processContainer == nil {
				processContainer = &ProcessReqAndResContainerData{
					List: make([]*ProcessReqAndResData, 0, this.processLen),
				}
				this.jsInfo.data.Http = processContainer
			}
		case "https":
			processContainer = this.jsInfo.data.Https
			if processContainer == nil {
				processContainer = &ProcessReqAndResContainerData{
					List: make([]*ProcessReqAndResData, 0, this.processLen),
				}
				this.jsInfo.data.Https = processContainer
			}
		default:
			return nil, NewReqProxyWithCode(errors.New("未被支持的协议类型"), "NO_ALLOW_PROTO")
		}
		prevData := &ProcessReqAndResData{
			Req: this.jsInfo.data.Req,
			Res: this.jsInfo.data.Res,
		}
		processContainer.List = append(processContainer.List, prevData)
		processContainer.Prev = prevData
		return res, proxyError
	}
}
