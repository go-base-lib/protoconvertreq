package protoconvertreq

type ReqProxyInterface interface {
	Send(templateParams *ProcessData) (ResHandleInterface, *ReqProxyError)
}

type ResHandleInterface interface {
	ResponseText() string
	ResponseJson(v interface{}) *ReqProxyError
	ResponseBytes() []byte
	ResponseJsonInterface() (interface{}, *ReqProxyError)
	ResponseInterface() interface{}
}

type ReqProxyError struct {
	code string
	msg  string
	err  error
}

func (this *ReqProxyError) Code() string {
	return this.code
}

func (this *ReqProxyError) Msg() string {
	return this.msg
}

func (this *ReqProxyError) Err() error {
	return this.err
}

func (this *ReqProxyError) Error() string {
	return this.Code() + ": " + this.Msg()
}

func NewReqProxyError(err error, code string, msg string) *ReqProxyError {
	return &ReqProxyError{
		err:  err,
		code: code,
		msg:  msg,
	}
}

func NewReqProxyWithCode(err error, code string) *ReqProxyError {
	return NewReqProxyError(err, code, err.Error())
}
