package protoconvertreq

import (
	"fmt"
	"testing"
)

func TestProcess(t *testing.T) {
	Data := &struct {
		ApplicantCert string
	}{
		ApplicantCert: "MIIEBzCCA6ugAwIBAgIHEwMAAADgnTAMBggqgRzPVQGDdQUAMCwxCzAJBgNVBAYTAkNOMR0wGwYDVQQDDBTnqI7liqHmtYvor5VDQTEoU00yKTAeFw0xOTEyMDQwMDAwMDBaFw0zNzAxMDEwMDAwMDBaME0xCzAJBgNVBAYTAkNOMRswGQYDVQQLDBLlm73lrrbnqI7liqHmgLvlsYAxITAfBgNVBAMMGFNNMuetvuWQjeacjeWKoeWZqOa1i+ivlTBZMBMGByqGSM49AgEGCCqBHM9VAYItA0IABLZiwcdS65+7urJAgBOWBv6jNHKmP9TZgVoNC1iRh1lwOGsAYqOJbHg43qour2SL+vNmg04GwZG+1XO7IGCGGgijggKTMIICjzAOBgNVHQ8BAf8EBAMCBPAwDAYDVR0TAQH/BAIwADAfBgNVHSMEGDAWgBRRyMC1kufaJyZskcVtPUgsshBRIzAdBgNVHQ4EFgQUdc1HumqGIYgssgzg5LQvGdbixWowRAYJKoZIhvcNAQkPBDcwNTAOBggqhkiG9w0DAgICAIAwDgYIKoZIhvcNAwQCAgCAMAcGBSsOAwIHMAoGCCqGSIb3DQMHMIHnBgNVHR8Egd8wgdwwa6BpoGeGZWxkYXA6Ly8xOTIuMTY4LjAuMTQwOjIzODkvY249Y3JsMTMwMyxvdT1jcmwxMyxvdT1jcmwsYz1jbj9jZXJ0aWZpY2F0ZVJldm9jYXRpb25MaXN0LCo/YmFzZT9jbj1jcmwxMzAzMG2ga6BphmdsZGFwOi8vY2hpbmF0YXguZ292LmNuOjIzODkvY249Y3JsMTMwMyxvdT1jcmwxMyxvdT1jcmwsYz1jbj9jZXJ0aWZpY2F0ZVJldm9jYXRpb25MaXN0LCo/YmFzZT9jbj1jcmwxMzAzMDkGCCsGAQUFBwEBBC0wKzApBggrBgEFBQcwAYYdaHR0cDovL2NoaW5hdGF4Lmdvdi5jbjoxNjU4OC8wIwYKKwYBBAGpQ2QFCAQVDBMwMy0wMDBkNTk5MTYyOTQxMzk0MCMGCisGAQQBqUNkBQkEFQwTMDMtMDAwZDU5OTE2Mjk0MTM5NDASBgorBgEEAalDZAIBBAQMAjE5MBEGBSpWCwcCBAgMBuaAu+WxgDAWBgUqVgsHAwQNDAswMDAwMDAwMDAwMDAbBgUqVgsHBQQSDBAwMDBkNTk5MTYyOTQxMzk0MB4GCGCGSAGG+EMJBBIMEDAxMDAwMTAwMDAwMzE1NjQwDAYIKoEcz1UBg3UFAANIADBFAiBT5UJMHrXsJ9xMPxXVHDB5ah95lkOvhhKrMoA/1UdnrAIhAPN+gqFDlP1YCYmoDrGxSX6sCYky3XQCpEQyGHgWXJgn",
	}
	process, err := NewProcessByYamlPath("test.yaml", Data)
	if err != nil {
		t.Log(err)
		return
	}

	Res, proxyError := process.ExecAll()
	if proxyError != nil {
		t.Error(proxyError)
		return
	}
	responseInterface := Res.ResponseInterface()
	fmt.Printf("%+v\n", responseInterface)
	//for process.Next() {
	//	exec, processErr := process.Exec()
	//	if processErr != nil {
	//		t.Error(processErr.Error())
	//		return
	//	}
	//
	//	responseInterface := exec.ResponseInterface()
	//	fmt.Printf("%+v\n", responseInterface)
	//}
}

func TestLocalProcess(t *testing.T) {
	process, err := NewProcessByYamlPath("test_local.yaml", nil)
	if err != nil {
		t.Log(err)
		return
	}
	res, proxyError := process.ExecAll()
	if proxyError != nil {
		t.Error(proxyError)
		return
	}
	responseInterface := res.ResponseInterface()
	fmt.Printf("%+v\n", responseInterface)
}

func BenchmarkProcessExec(b *testing.B) {
	process, err := NewProcessByYamlPath("test_local.yaml", nil)
	defer process.Destroy()
	for i := 0; i < b.N; i++ {
		if err != nil {
			b.Error(err)
			return
		}
		_, proxyError := process.ExecAll()
		if proxyError != nil {
			b.Error(proxyError)
			return
		}
		process.Reset()
	}
}

//func TestOther(t *testing.T) {
//	t2 := template.New("default")
//	t2.Funcs(map[string]interface{}{
//		"test": func(a, b int) int {
//			return a + b
//		},
//	})
//
//	t2.Funcs(map[string]interface{}{
//		"test2": func(a, b int) (int, error) {
//			return a * b, nil
//		},
//	})
//
//	t2.Funcs(map[string]interface{}{
//		"test": func(a, b int) (int, error) {
//			return a - b, nil
//		},
//	})
//
//	str := "{{ test 1 2 }}---------{{ test2 1 2 }}"
//	parse, err := t2.Parse(str)
//	if err != nil {
//		panic(err)
//	}
//	buf := &bytes.Buffer{}
//	if err = parse.Execute(buf, nil); err != nil {
//		panic(err)
//	}
//	fmt.Println(buf.String())
//}
