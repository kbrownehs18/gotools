package tests

import (
	"testing"

	"github.com/kbrownehs18/gotools/common"
)

func TestGetLocalIp(t *testing.T) {
	ips, err := common.GetLocalIp()
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
