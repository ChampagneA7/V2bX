package node

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/InazumaV/V2bX/api/panel"
	"github.com/InazumaV/V2bX/conf"
	vCore "github.com/InazumaV/V2bX/core"
	"github.com/InazumaV/V2bX/limiter"
)

type reviewCore struct {
	addCalls           int
	currentDeviceLimit int
}

func (*reviewCore) Start() error                                                  { return nil }
func (*reviewCore) Close() error                                                  { return nil }
func (*reviewCore) AddNode(string, *panel.NodeInfo, *conf.Options) error          { return nil }
func (*reviewCore) DelNode(string) error                                          { return nil }
func (*reviewCore) Protocols() []string                                           { return []string{"vless"} }
func (*reviewCore) Type() string                                                  { return "review" }
func (*reviewCore) DelUsers([]panel.UserInfo, string, *panel.NodeInfo) error      { return nil }
func (*reviewCore) GetUserTrafficSlice(string, bool) ([]panel.UserTraffic, error) { return nil, nil }
func (c *reviewCore) AddUsers(p *vCore.AddUsersParams) (int, error) {
	c.addCalls++
	if c.addCalls == 1 {
		return 0, errors.New("simulated core update failure")
	}
	if len(p.Users) > 0 {
		c.currentDeviceLimit = p.Users[0].DeviceLimit
	}
	return len(p.Users), nil
}

func TestReviewRetryUserUpdateAfterCoreFailure(t *testing.T) {
	userRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/server/UniProxy/config":
			w.WriteHeader(http.StatusNotModified)
		case "/api/v1/server/UniProxy/user":
			userRequests++
			if r.Header.Get("If-None-Match") == "\"new-users\"" {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("ETag", "\"new-users\"")
			_, _ = w.Write([]byte("{\"users\":[{\"id\":42,\"uuid\":\"a\",\"device_limit\":1}]}"))
		case "/api/v1/server/UniProxy/alivelist":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{\"alive\":{\"42\":3}}"))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	client, err := panel.New(&conf.ApiConfig{APIHost: server.URL, NodeID: 7, NodeType: "vless", Key: "test"})
	if err != nil {
		t.Fatal(err)
	}
	old := panel.UserInfo{Id: 42, Uuid: "a", DeviceLimit: 5}
	backend := &reviewCore{currentDeviceLimit: 5}
	limiter.Init()
	controller := &Controller{
		server: backend, apiClient: client, tag: "review-poller",
		userList: []panel.UserInfo{old}, info: &panel.NodeInfo{Type: "vless"},
		Options: &conf.Options{},
	}
	controller.limiter = limiter.AddLimiter(controller.tag, &conf.LimitConfig{}, controller.userList, map[int]int{42: 3})
	if err := controller.nodeInfoMonitor(); err != nil {
		t.Fatal(err)
	}
	if err := controller.nodeInfoMonitor(); err != nil {
		t.Fatal(err)
	}
	if userRequests != 2 || backend.currentDeviceLimit != 1 || controller.userList[0].DeviceLimit != 1 {
		t.Fatalf("failed update was lost behind ETag: requests=%d coreCalls=%d coreLimit=%d users=%+v", userRequests, backend.addCalls, backend.currentDeviceLimit, controller.userList)
	}
	if controller.limiter.AliveList[42] != 3 {
		t.Fatal("retry erased live device counts")
	}
}
