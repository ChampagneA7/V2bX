package limiter

import (
	"github.com/InazumaV/V2bX/api/panel"
	"github.com/InazumaV/V2bX/conf"
	"testing"
)

func TestReviewKeepAliveCountsOnLimitChange(t *testing.T) {
	Init()
	old := panel.UserInfo{Id: 42, Uuid: "a", SpeedLimit: 100, DeviceLimit: 5}
	current := old
	current.DeviceLimit = 1
	l := AddLimiter("review", &conf.LimitConfig{}, []panel.UserInfo{old}, map[int]int{42: 3})
	l.UpdateUser("review", []panel.UserInfo{current}, []panel.UserInfo{old})
	if l.AliveList[42] != 3 {
		t.Fatalf("changing a limit erased cross-node device count: %v", l.AliveList)
	}
	l.UpdateUser("review", nil, []panel.UserInfo{current})
	if _, exists := l.AliveList[42]; exists {
		t.Fatal("removed identity retained its device count")
	}
}
