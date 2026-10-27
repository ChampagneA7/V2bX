package sing

import (
	"github.com/InazumaV/V2bX/api/panel"
	"github.com/InazumaV/V2bX/common/counter"
	"testing"
)

func TestReviewIdentityRemovalDoesNotBreakOtherNodes(t *testing.T) {
	backend := &Sing{
		users:                     &UserMap{uidMap: make(map[string]int)},
		hookServer:                &HookServer{},
		nodeReportMinTrafficBytes: make(map[string]int64),
	}
	backend.rememberUserID("node-a", panel.UserInfo{Id: 42, Uuid: "shared-credential"})
	backend.rememberUserID("node-b", panel.UserInfo{Id: 90001, Uuid: "shared-credential"})
	trafficA, trafficB := counter.NewTrafficCounter(), counter.NewTrafficCounter()
	trafficA.Tx("shared-credential", 12)
	trafficB.Tx("shared-credential", 34)
	backend.hookServer.counter.Store("node-a", trafficA)
	backend.hookServer.counter.Store("node-b", trafficB)
	backend.forgetUserID("node-a", "shared-credential")
	removed, err := backend.GetUserTrafficSlice("node-a", false)
	if err != nil || len(removed) != 0 {
		t.Fatalf("removed identity reported: %v %v", removed, err)
	}
	preserved, err := backend.GetUserTrafficSlice("node-b", false)
	if err != nil || len(preserved) != 1 || preserved[0].UID != 90001 || preserved[0].Upload != 34 {
		t.Fatalf("removing one node lost another node's traffic identity: %v %v", preserved, err)
	}
}
