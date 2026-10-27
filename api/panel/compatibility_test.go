package panel

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/InazumaV/V2bX/conf"
	"github.com/vmihailenco/msgpack/v5"
)

func compatibilityClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := New(&conf.ApiConfig{APIHost: server.URL, Key: "test-token", NodeType: "vless", NodeID: 7})
	if err != nil {
		t.Fatal(err)
	}
	client.client.SetRetryCount(0)
	return client
}

func TestCompatibilityUserLists(t *testing.T) {
	tests := []struct {
		name, format string
		users        []map[string]any
	}{
		{"old panel JSON", "json", []map[string]any{{"id": 42, "uuid": "legacy-uuid", "speed_limit": 100, "device_limit": 2}}},
		{"new panel JSON", "json", []map[string]any{
			{"id": 90001, "uuid": "subscription-a", "speed_limit": 100, "device_limit": 2, "user_id": 42, "subscription_id": 501},
			{"id": 90002, "uuid": "subscription-b", "speed_limit": 200, "device_limit": 5, "user_id": 42, "subscription_id": 502},
		}},
		{"old panel MessagePack", "msgpack", []map[string]any{{"id": 42, "uuid": "legacy-uuid", "speed_limit": 100, "device_limit": 2}}},
		{"new panel MessagePack", "msgpack", []map[string]any{
			{"id": 90001, "uuid": "subscription-a", "speed_limit": 100, "device_limit": 2, "subscription_id": 501},
			{"id": 90002, "uuid": "subscription-b", "speed_limit": 200, "device_limit": 5, "subscription_id": 502},
		}},
		{"empty JSON", "json", []map[string]any{}},
		{"empty MessagePack", "msgpack", []map[string]any{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := compatibilityClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/server/UniProxy/user" || r.Method != http.MethodGet {
					t.Errorf("unexpected endpoint %s %s", r.Method, r.URL.Path)
				}
				if r.URL.Query().Get("node_id") != "7" || r.URL.Query().Get("token") != "test-token" {
					t.Error("legacy authentication changed")
				}
				payload := map[string]any{"users": tt.users}
				if tt.format == "msgpack" {
					w.Header().Set("Content-Type", "application/x-msgpack")
					if err := msgpack.NewEncoder(w).Encode(payload); err != nil {
						t.Error(err)
					}
				} else {
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(payload); err != nil {
						t.Error(err)
					}
				}
			})
			users, err := client.GetUserList()
			if err != nil {
				t.Fatal(err)
			}
			if users == nil || len(users) != len(tt.users) {
				t.Fatalf("users=%+v; explicit empty responses must not mean unchanged", users)
			}
			for i, expected := range tt.users {
				if users[i].Id != expected["id"] || users[i].Uuid != expected["uuid"] || users[i].SpeedLimit != expected["speed_limit"] || users[i].DeviceLimit != expected["device_limit"] {
					t.Fatalf("unexpected identity or limits: %+v", users[i])
				}
			}
		})
	}
}

func TestCompatibilityETagAndMalformedList(t *testing.T) {
	requests := 0
	client := compatibilityClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		switch requests {
		case 1:
			w.Header().Set("ETag", "\"users-v1\"")
			_, _ = w.Write([]byte("{\"users\":[{\"id\":42,\"uuid\":\"legacy\"}]}"))
		case 2:
			if r.Header.Get("If-None-Match") != "\"users-v1\"" {
				t.Error("ETag was not retained")
			}
			w.WriteHeader(http.StatusNotModified)
		default:
			w.Header().Set("Content-Type", "application/x-msgpack")
			w.Header().Set("ETag", "\"invalid\"")
			_ = msgpack.NewEncoder(w).Encode(map[string]any{"unexpected": true})
		}
	})
	if _, err := client.GetUserList(); err != nil {
		t.Fatal(err)
	}
	users, err := client.GetUserList()
	if err != nil || users != nil {
		t.Fatalf("304 should mean unchanged: %+v %v", users, err)
	}
	if _, err := client.GetUserList(); err == nil {
		t.Fatal("malformed list must not revoke all subscriptions")
	}
	if client.userEtag != "\"users-v1\"" {
		t.Fatal("invalid payload changed ETag")
	}
}

func TestCompatibilityReportIdentities(t *testing.T) {
	requests := 0
	client := compatibilityClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodPost || r.URL.Query().Get("node_id") != "7" {
			t.Error("legacy report contract changed")
		}
		switch r.URL.Path {
		case "/api/v1/server/UniProxy/push":
			var got map[string][]int64
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Error(err)
			}
			want := map[string][]int64{"90001": {10, 20}, "90002": {30, 40}}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("traffic identities changed: %v", got)
			}
		case "/api/v1/server/UniProxy/alive":
			var got map[string][]string
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Error(err)
			}
			want := map[string][]string{"90001": {"198.51.100.1"}, "90002": {"198.51.100.2"}}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("device identities changed: %v", got)
			}
		default:
			t.Errorf("unexpected endpoint %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"data\":true}"))
	})
	if err := client.ReportUserTraffic([]UserTraffic{{UID: 90001, Upload: 10, Download: 20}, {UID: 90002, Upload: 30, Download: 40}}); err != nil {
		t.Fatal(err)
	}
	devices := map[int][]string{90001: {"198.51.100.1"}, 90002: {"198.51.100.2"}}
	if err := client.ReportNodeOnlineUsers(&devices); err != nil {
		t.Fatal(err)
	}
	if requests != 2 {
		t.Fatalf("unexpected report count %d", requests)
	}
}

func TestCompatibilityAliveCountsAndErrors(t *testing.T) {
	client := compatibilityClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/server/UniProxy/alivelist" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{\"alive\":{\"90001\":1,\"90002\":2}}"))
			return
		}
		http.Error(w, "panel unavailable", http.StatusServiceUnavailable)
	})
	counts, err := client.GetUserAlive()
	if err != nil || !reflect.DeepEqual(counts, map[int]int{90001: 1, 90002: 2}) {
		t.Fatalf("counts=%v err=%v", counts, err)
	}
	devices := map[int][]string{90001: {"198.51.100.1"}}
	if err := client.ReportNodeOnlineUsers(&devices); err == nil {
		t.Fatal("online report failure was hidden")
	}
}
