package protoconvertreq

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"strings"
)

type httpsReq struct {
	*httpReq
}

func NewHttpsReq(requestInfo *RequestInfos) (*httpsReq, error) {
	if requestInfo.Url == "" {
		return nil, errors.New("请求路径不能为空")
	}
	if !strings.HasPrefix(requestInfo.Url, "https://") {
		requestInfo.Url = "https://" + requestInfo.Url
	}

	requestInfo.Method = strings.ToUpper(requestInfo.Method)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: requestInfo.SkipVerify,
		},
	}

	if requestInfo.ClientKeyPem != "" && requestInfo.ClientCertPem != "" && requestInfo.CaCertPem != "" {

		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM([]byte(requestInfo.CaCertPem))
		cliCrt, err := tls.X509KeyPair([]byte(requestInfo.ClientCertPem), []byte(requestInfo.ClientKeyPem))
		if err != nil {
			return nil, errors.New("加载HTTPS客户端证书以及私钥失败 => " + err.Error())
		}

		tr.TLSClientConfig.RootCAs = pool
		tr.TLSClientConfig.Certificates = []tls.Certificate{cliCrt}

	}

	return &httpsReq{
		httpReq: &httpReq{
			requestInfo: requestInfo,
			client:      &http.Client{Transport: tr},
			flag:        "http",
		},
	}, nil
}
