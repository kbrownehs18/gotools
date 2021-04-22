package tests

import (
	"encoding/json"
	"testing"

	"github.com/kbrownehs18/gotools/common"
)

func TestGetLocalIp(t *testing.T) {
	ips, err := common.GetLocalIP()
	if err != nil {
		t.Error(err)
	}

	t.Log(ips)
}

func TestString(t *testing.T) {
	s := " Hello	"
	t.Log(common.Trim(s))

	ss := "Hello world!!!"

	sss := common.Split(ss, " ")
	for _, str := range sss {
		t.Log(str)
	}

	sss = common.SplitBySpaceTab(ss)
	for _, str := range sss {
		t.Log(str)
	}
}

func TestContains(t *testing.T) {
	a := []int8{0, 1, 2, 3, 4, 5, 6, 9}
	var n int8 = 0
	t.Log(common.Contains(a, n))

	m := map[string]bool{
		"name":   true,
		"gender": true,
	}
	x := "apple"
	t.Log(common.Contains(m, x))
	t.Log(common.Contains(m, "name"))
}

func TestHttpRequest(t *testing.T) {
	rtn, err := common.HTTPRequest("https://postman-echo.com/get", common.GET, map[string]string{
		"foo1": "bar1", "foo2": "bar2",
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(rtn)

	rtn, err = common.HTTPRequest("https://postman-echo.com/post", common.POST, map[string]string{
		"foo1": "bar1", "foo2": "bar2",
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(rtn)

	str, err := json.Marshal(map[string]string{
		"foo1": "bar1", "foo2": "bar2",
	})
	if err != nil {
		t.Error(err)
	}
	rtn, err = common.HTTPRequest("https://postman-echo.com/post", common.POST, string(str))
	if err != nil {
		t.Error(err)
	}
	t.Log(rtn)
}

func TestAuthcode(t *testing.T) {
	key := "1234567890"
	var de string
	var err error

	en, err := common.Authcode("scnjl", common.ENCODE, key)
	if err != nil {
		t.Error(err)
	}
	t.Log(en)

	de, err = common.Authcode(en, common.DECODE, key)
	if err != nil {
		t.Error(err)
	}
	t.Log(de)

	// de, err = common.Authcode("", common.DECODE, key)
	// if err != nil {
	// 	t.Error(err)
	// }
	// t.Log(de)
}
