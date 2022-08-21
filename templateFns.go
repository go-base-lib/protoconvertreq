package protoconvertreq

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/rand"
	"text/template"
	"time"
)

var templateFnMap template.FuncMap = map[string]interface{}{
	"randHex":      GetRandomString,
	"jsonTo":       DecodeJson,
	"toHex":        EncodeHex,
	"HexTo":        DecodeHex,
	"toBytes":      ToBytes,
}

func ToBytes(str string) []byte {
	return []byte(str)
}

// GetRandomString 获取指定长度的随机字符串
func GetRandomString(l int) string {
	str := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	//str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []rune(str)
	result := make([]rune, l, l)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result[i] = bytes[r.Intn(len(bytes))]
		//result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func EncodeHex(str string) string {
	return hex.EncodeToString([]byte(str))
}

func DecodeHex(str string) (string, error) {
	d, err := hex.DecodeString(str)
	if err != nil {
		return "", err
	}
	return string(d), err
}

func EncodeBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func DecodeBase64(str string) (string, error) {
	d, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	return string(d), err
}

func EncodeJson(v interface{}) (string, error) {
	marshal, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

func DecodeJson(str string) (interface{}, error) {
	v := new(interface{})
	if err := json.Unmarshal([]byte(str), v); err != nil {
		return nil, errors.New("数据转换为json结构失败 => " + err.Error())
	}
	return *v, nil
}

func init() {
	templateParser.Funcs(templateFnMap)
}
