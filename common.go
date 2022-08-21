package protoconvertreq

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/dop251/goja"
	"github.com/go-base-lib/logs"
	"reflect"
	"strings"
)

type commonDataHandle struct {
	data         []byte
	resInterface interface{}
}

func NewCommonDataHandle(data []byte, templateData *ProcessData, responseParser *ResponseParser, jsVm *goja.Runtime) (*commonDataHandle, *ReqProxyError) {
	commonHandle := &commonDataHandle{
		data: data,
	}
	if responseParser == nil {
		return commonHandle, nil
	}

	if responseParser.Simple != "" {
		simpleList := strings.Split(responseParser.Simple, "|")
		for _, simpleStr := range simpleList {
			simpleStr = strings.TrimSpace(simpleStr)
			simpleStr = strings.ToLower(simpleStr)
			var err *ReqProxyError
			switch simpleStr {
			case "json":
				logs.Debugln("执行simple json解析器")
				err = commonHandle.json()
			case "base64":
				logs.Debugln("执行simple base64解析器")
				err = commonHandle.base64()
			case "hex":
				logs.Debugln("执行simple hex解析器")
				err = commonHandle.hex()
			}
			if err != nil {
				logs.Debugln("执行simple解析器发生异常, 错误将向上层抛出 => ", err.Error())
				return nil, err
			}
		}
	}

	if commonHandle.resInterface != nil && templateData.Res != nil {
		templateData.Res.Obj = commonHandle.resInterface
	}

	if len(responseParser.Fns) > 0 {
		var res goja.Value
		for _, fnStr := range responseParser.Fns {
			logs.Debugln("执行正常数据处理脚本, 脚本内容: \n", fnStr)
			runRes, err := jsVm.RunString(fnStr)
			if err != nil {
				logs.Debugln("正常数据处理脚本执行错误，将向上层抛出异常 => ", err.Error())
				err = ConvertGojaException(err)
				return nil, NewReqProxyError(err, "EXEC_CODE", err.Error())
			}

			res = runRes
			resData := res.Export()
			templateData.Prev = resData
			templateData.History = append(templateData.History, resData)
			//if err = jsVm.Set("_"+strconv.Itoa(i), resData); err != nil {
			//	return nil, NewReqProxyError(err, "PARSER_EXEC_VARS", "解析器变量设置失败")
			//}
			//if err = jsVm.Set("_prev", resData); err != nil {
			//	return nil, NewReqProxyError(err, "PARSER_EXEC_VARS", "解析器变量设置失败")
			//}
		}
		if res == nil {
			commonHandle.data = nil
			commonHandle.resInterface = nil
		} else {
			commonHandle.resInterface = res.Export()
			exportType := res.ExportType()
			t := exportType.Kind()
			switch t {
			case reflect.String:
				commonHandle.data = []byte(res.String())
			case reflect.Slice, reflect.Map:
				marshal, e := json.Marshal(commonHandle.resInterface)
				if e != nil {
					return nil, NewReqProxyError(e, "PARSER_EXEC_RES", "解析器结果解析失败")
				}
				commonHandle.data = marshal
			default:
				commonHandle.data = []byte(res.ToString().String())
			}
		}
	}

	templateData.Res.Data = string(commonHandle.data)
	templateData.Res.Obj = commonHandle.resInterface
	logs.Debugln("本次处理之后的响应结果: \n", templateData.Res.Data)
	return commonHandle, nil

}

func (this *commonDataHandle) ResponseText() string {
	return string(this.data)
}

func (this *commonDataHandle) ResponseJson(v interface{}) *ReqProxyError {
	if err := json.Unmarshal(this.data, v); err != nil {
		return NewReqProxyError(err, "CONVERT_DATA", "转换数据到json格式失败")
	}
	return nil
}

func (this *commonDataHandle) ResponseBytes() []byte {
	return this.data
}

func (this *commonDataHandle) ResponseInterface() interface{} {
	if this.resInterface != nil {
		return this.resInterface
	}
	proxyError := this.json()
	if proxyError != nil {
		this.resInterface = this.data
	}
	return this.resInterface
}

func (this *commonDataHandle) ResponseJsonInterface() (interface{}, *ReqProxyError) {
	if this.resInterface != nil {
		return this.resInterface, nil
	}
	if proxyError := this.json(); proxyError != nil {
		return nil, proxyError
	}
	return this.resInterface, nil
}

func (this *commonDataHandle) base64() *ReqProxyError {
	decodeString, err := base64.StdEncoding.DecodeString(string(this.data))
	if err != nil {
		return NewReqProxyError(err, "DECODE_BASE64", "base64解码失败")
	}

	this.data = decodeString
	return nil
}

func (this *commonDataHandle) hex() *ReqProxyError {
	decodeString, err := hex.DecodeString(string(this.data))
	if err != nil {
		return NewReqProxyError(err, "DECODE_HEX", "hex解码失败")
	}
	this.data = decodeString
	return nil
}

func (this *commonDataHandle) json() *ReqProxyError {
	v := new(interface{})
	if err := json.Unmarshal(this.data, v); err != nil {
		return NewReqProxyError(err, "DECODE_JSON", "转换json结构失败")
	}
	this.resInterface = *v
	return nil
}
