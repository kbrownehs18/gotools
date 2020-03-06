package tests

import (
	"testing"

	"github.com/kbrownehs18/gotool/common"
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
