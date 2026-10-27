package panel

import (
	"fmt"
	"io"
	"strings"

	"encoding/json/v2"

	"github.com/vmihailenco/msgpack/v5"
)

type OnlineUser struct {
	UID int
	IP  string
}

// Id is opaque: legacy account ID or a stable subscription node_user_id.
// Keep the wire field names unchanged for old panels and old nodes.
type UserInfo struct {
	Id          int    `json:"id" msgpack:"id"`
	Uuid        string `json:"uuid" msgpack:"uuid"`
	SpeedLimit  int    `json:"speed_limit" msgpack:"speed_limit"`
	DeviceLimit int    `json:"device_limit" msgpack:"device_limit"`
}

type UserListBody struct {
	Users []UserInfo `json:"users" msgpack:"users"`
}

type AliveMap struct {
	Alive map[int]int `json:"alive"`
}

// InvalidateUserList makes the next poll retry a list that the core could not apply.
func (c *Client) InvalidateUserList() { c.userEtag = "" }

// GetUserList will pull user from v2board
func (c *Client) GetUserList() ([]UserInfo, error) {
	const path = "/api/v1/server/UniProxy/user"
	r, err := c.client.R().
		SetHeader("If-None-Match", c.userEtag).
		SetHeader("X-Response-Format", "msgpack").
		SetDoNotParseResponse(true).
		Get(path)
	if r == nil || r.RawResponse == nil {
		return nil, fmt.Errorf("received nil response or raw response")
	}
	defer r.RawResponse.Body.Close()

	if r.StatusCode() == 304 {
		return nil, nil
	}

	if err = c.checkResponse(r, path, err); err != nil {
		return nil, err
	}
	userlist := &UserListBody{}
	if strings.Contains(r.Header().Get("Content-Type"), "application/x-msgpack") {
		decoder := msgpack.NewDecoder(r.RawResponse.Body)
		length, err := decoder.DecodeMapLen()
		if err != nil {
			return nil, fmt.Errorf("decode user list error: %w", err)
		}
		seenUsers := false
		for i := 0; i < length; i++ {
			key, err := decoder.DecodeString()
			if err != nil {
				return nil, fmt.Errorf("decode user list error: %w", err)
			}
			if key == "users" {
				if seenUsers {
					return nil, fmt.Errorf("decode user list error: duplicate users member")
				}
				seenUsers = true
				err = decoder.Decode(&userlist.Users)
			} else {
				err = decoder.Skip()
			}
			if err != nil {
				return nil, fmt.Errorf("decode user list error: %w", err)
			}
		}
		if _, err := decoder.PeekCode(); err != io.EOF {
			return nil, fmt.Errorf("decode user list error: trailing MessagePack data")
		}
	} else {
		if err := json.UnmarshalRead(r.RawResponse.Body, userlist); err != nil {
			return nil, fmt.Errorf("decode user list error: %w", err)
		}
	}

	if userlist.Users == nil {
		return nil, fmt.Errorf("decode user list error: expected users array")
	}
	seenIDs := make(map[int]struct{}, len(userlist.Users))
	seenUUIDs := make(map[string]struct{}, len(userlist.Users))
	for _, user := range userlist.Users {
		if user.Id <= 0 || user.Uuid == "" {
			return nil, fmt.Errorf("decode user list error: missing or invalid identity")
		}
		if _, exists := seenIDs[user.Id]; exists {
			return nil, fmt.Errorf("decode user list error: duplicate identity %d", user.Id)
		}
		if _, exists := seenUUIDs[user.Uuid]; exists {
			return nil, fmt.Errorf("decode user list error: duplicate credential")
		}
		seenIDs[user.Id] = struct{}{}
		seenUUIDs[user.Uuid] = struct{}{}
	}
	c.userEtag = r.Header().Get("ETag")
	return userlist.Users, nil
}

// GetUserAlive will fetch the alive_ip count for users
func (c *Client) GetUserAlive() (map[int]int, error) {
	c.AliveMap = &AliveMap{}
	const path = "/api/v1/server/UniProxy/alivelist"
	r, err := c.client.R().
		ForceContentType("application/json").
		Get(path)
	if err != nil || r == nil || r.RawResponse == nil || r.StatusCode() >= 399 {
		c.AliveMap.Alive = make(map[int]int)
		return c.AliveMap.Alive, nil
	}
	defer r.RawResponse.Body.Close()
	if err := json.Unmarshal(r.Body(), c.AliveMap); err != nil {
		fmt.Printf("unmarshal user alive list error: %s", err)
		c.AliveMap.Alive = make(map[int]int)
	}

	return c.AliveMap.Alive, nil
}

type UserTraffic struct {
	UID      int
	Upload   int64
	Download int64
}

// ReportUserTraffic reports the user traffic
func (c *Client) ReportUserTraffic(userTraffic []UserTraffic) error {
	data := make(map[int][]int64, len(userTraffic))
	for i := range userTraffic {
		data[userTraffic[i].UID] = []int64{userTraffic[i].Upload, userTraffic[i].Download}
	}
	const path = "/api/v1/server/UniProxy/push"
	r, err := c.client.R().
		SetBody(data).
		ForceContentType("application/json").
		Post(path)
	err = c.checkResponse(r, path, err)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ReportNodeOnlineUsers(data *map[int][]string) error {
	const path = "/api/v1/server/UniProxy/alive"
	r, err := c.client.R().
		SetBody(data).
		ForceContentType("application/json").
		Post(path)
	err = c.checkResponse(r, path, err)

	if err != nil {
		return err
	}

	return nil
}
