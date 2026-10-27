package node

import (
	"reflect"
	"testing"

	"github.com/InazumaV/V2bX/api/panel"
)

func TestCompareUserList(t *testing.T) {
	original := panel.UserInfo{Id: 42, Uuid: "subscription-a", SpeedLimit: 100, DeviceLimit: 2}
	tests := []struct {
		name                   string
		current                []panel.UserInfo
		wantDeleted, wantAdded []panel.UserInfo
	}{
		{name: "unchanged legacy user", current: []panel.UserInfo{original}},
		{name: "explicit empty list revokes user", current: []panel.UserInfo{}, wantDeleted: []panel.UserInfo{original}},
		{name: "new independent subscription", current: []panel.UserInfo{original, {Id: 90002, Uuid: "subscription-b", SpeedLimit: 200, DeviceLimit: 5}}, wantAdded: []panel.UserInfo{{Id: 90002, Uuid: "subscription-b", SpeedLimit: 200, DeviceLimit: 5}}},
	}
	for _, field := range []string{"identity", "speed", "device", "uuid"} {
		changed := original
		switch field {
		case "identity":
			changed.Id = 90001
		case "speed":
			changed.SpeedLimit = 200
		case "device":
			changed.DeviceLimit = 5
		case "uuid":
			changed.Uuid = "replacement-uuid"
		}
		tests = append(tests, struct {
			name                   string
			current                []panel.UserInfo
			wantDeleted, wantAdded []panel.UserInfo
		}{field + " change replaces user", []panel.UserInfo{changed}, []panel.UserInfo{original}, []panel.UserInfo{changed}})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleted, added := compareUserList([]panel.UserInfo{original}, tt.current)
			if !reflect.DeepEqual(deleted, tt.wantDeleted) || !reflect.DeepEqual(added, tt.wantAdded) {
				t.Fatalf("deleted=%+v added=%+v; want deleted=%+v added=%+v", deleted, added, tt.wantDeleted, tt.wantAdded)
			}
		})
	}
}
