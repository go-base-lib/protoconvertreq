package protoconvertreq

import (
	"bytes"
	"errors"
	"github.com/dop251/goja"
	"text/template"
)

var templateParser = template.New("default")

func ParseTemplateStrByData(str string, data interface{}) (string, error) {
	parse, err := templateParser.Parse(str)
	if err != nil {
		return "", err
	}

	buffer := &bytes.Buffer{}
	if err = parse.Execute(buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func ParseTemplateStr2Bytes(str string, data interface{}) ([]byte, error) {
	parse, err := templateParser.Parse(str)
	if err != nil {
		return nil, err
	}

	buffer := &bytes.Buffer{}
	if err = parse.Execute(buffer, data); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func ConvertGojaException(err error) error {
	if err == nil {
		return err
	}
	exception, ok := err.(*goja.Exception)
	if !ok {
		return err
	}
	return errors.New(exception.Value().String())
}
