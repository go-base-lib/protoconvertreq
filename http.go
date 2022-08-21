package protoconvertreq

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/dop251/goja"
	"github.com/go-base-lib/logs"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

type httpReq struct {
	requestInfo *RequestInfos
	client      *http.Client
	flag        string
	jsVm        *goja.Runtime
}

func NewHttpReq(requestInfo *RequestInfos, jsVm *goja.Runtime) (*httpReq, error) {

	if requestInfo.Url == "" {
		return nil, errors.New("请求路径不能为空")
	}

	if !strings.HasPrefix(requestInfo.Url, "http://") {
		requestInfo.Url = "http://" + requestInfo.Url
	}

	requestInfo.Method = strings.ToUpper(requestInfo.Method)

	return &httpReq{
		requestInfo: requestInfo,
		client:      http.DefaultClient,
		flag:        "http",
		jsVm:        jsVm,
	}, nil
}

func (this *httpReq) Send(templateParams *ProcessData) (ResHandleInterface, *ReqProxyError) {
	logs.Debugf("将要转发%s请求... \n", this.flag)
	url := this.requestInfo.Url
	if u, err := ParseTemplateStrByData(this.requestInfo.Url, templateParams); err != nil {
		logs.Debugln("解析请求路径失败 => ", err.Error())
		return nil, NewReqProxyError(err, "PARSER_URL", "解析请求路径失败")
	} else {
		url = u
	}
	var body *bytes.Buffer
	var bodyStr string
	if this.requestInfo.ReqBody != "" {
		bodyBytes, err := ParseTemplateStr2Bytes(this.requestInfo.ReqBody, templateParams)
		if err != nil {
			logs.Debugln("解析请求参数体失败 => ", err.Error())
			return nil, NewReqProxyError(err, "PARSER_BODY", "解析参数体失败")
		}
		body = bytes.NewBuffer(bodyBytes)
		bodyStr = string(bodyBytes)
	}

	request, err := http.NewRequest(this.requestInfo.Method, url, body)
	if err != nil {
		logs.Debugln("创建请求代理对象失败 => ", err.Error())
		return nil, NewReqProxyError(err, "CREATE_REQUEST", "创建请求对象失败")
	}

	for k, v := range this.requestInfo.Headers {
		kStr, err := ParseTemplateStrByData(k, templateParams)
		if err != nil {
			logs.Debugln("解析header key失败 => ", err.Error())
			return nil, NewReqProxyError(err, "PARSER_HEADER_KEY", "解析头名称失败")
		}

		kVal, err := ParseTemplateStrByData(v, templateParams)
		if err != nil {
			logs.Debugln("解析header val失败 => ", err.Error())
			return nil, NewReqProxyError(err, "PARSER_HEADER_VAL", "解析头内容失败")
		}

		request.Header.Add(kStr, kVal)
	}

	res, err := this.client.Do(request)
	if err != nil {
		logs.Debugln("发送请求失败 => ", err.Error())
		return nil, NewReqProxyError(err, "SEND_ERROR", "发送请求到目标服务失败")
	}
	defer res.Body.Close()

	resData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, NewReqProxyError(err, "RECEIVE_ERR_GET_RES_DATA", "获取错误的响应结果势必")
	}

	if templateParams.Req == nil {
		templateParams.Req = &ProcessReqData{}
	}

	if templateParams.Res == nil {
		templateParams.Res = &ProcessResData{}
	}

	templateParams.Req.Data = bodyStr
	templateParams.Req.Headers = this.requestInfo.Headers

	templateParams.Res.Data = string(resData)
	templateParams.Res.StatusCode = res.StatusCode

	logs.Debugf(`%s请求路径: %s
%s请求方式: %s
%s请求头: %v
%s请求参数体: %s
%s响应状态码: %d
%s响应数据: %s`,
		this.flag, url,
		this.flag, this.requestInfo.Method,
		this.flag, request.Header,
		this.flag, bodyStr,
		this.flag, res.StatusCode,
		this.flag, resData)

	if res.StatusCode < 200 || res.StatusCode > 299 {

		if this.requestInfo.ResponseParser != nil && this.requestInfo.ResponseParser.ErrFn != "" {
			logs.Debugln("发生异常, 发现异常恢复脚本，将要执行，脚本内容: \n", this.requestInfo.ResponseParser.ErrFn)
			restoreVal, err := this.jsVm.RunString(this.requestInfo.ResponseParser.ErrFn)
			if err != nil {
				logs.Debugln("异常恢复脚本恢复异常失败, 新的异常将向上层抛出 => ", err.Error())
				err = ConvertGojaException(err)
				return nil, NewReqProxyWithCode(err, "RECEIVE_ERR_RESTORE_EXEC")
			}

			exportType := restoreVal.ExportType().Kind()
			switch exportType {
			case reflect.String:
				str := restoreVal.String()
				templateParams.Res.Data = str
				resData = []byte(str)
			case reflect.Slice, reflect.Map:
				marshal, err := json.Marshal(restoreVal.Export())
				if err != nil {
					return nil, NewReqProxyError(err, "CONVERT_ERR_RESTORE_DATA", "解析从异常中获取的数据失败")
				}
				resData = marshal
				templateParams.Res.Data = string(marshal)
			default:
				str := restoreVal.ToString().String()
				resData = []byte(str)
				templateParams.Res.Data = str
			}
			logs.Debugln("异常恢复成功, 恢复之后的响应数据 => ", string(resData))
		} else {
			logs.Debugln("未发现异常恢复脚本, 异常将向上层抛出")
			return nil, NewReqProxyError(err, "RECEIVE", "对端服务响应信息不正确")
		}
	}

	return NewCommonDataHandle(resData, templateParams, this.requestInfo.ResponseParser, this.jsVm)
}
