package protoconvertreq

import (
	"fmt"
	"testing"
)

func TestTemplateFns(t *testing.T) {
	ranHex := GetRandomString(16)
	fmt.Println(ranHex)
}
