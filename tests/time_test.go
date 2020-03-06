package tests

import (
	"testing"

	"github.com/kbrownehs18/gotool/common"
)

func TestStrToTime(t *testing.T) {
	tt, err := common.StrToTime("20191231")
	if err != nil {
		t.Error(err)
	}
	t.Log(tt)
}
