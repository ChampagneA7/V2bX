package panel

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

func TestReviewRejectsIncompleteOrAmbiguousLists(t *testing.T) {
	for name, body := range map[string]string{
		"truncated after empty array": "{\"users\":[]",
		"truncated extra member":      "{\"users\":[],\"error\":",
		"trailing garbage":            "{\"users\":[]} garbage",
		"duplicate users member":      "{\"users\":[],\"users\":[{\"id\":42,\"uuid\":\"a\"}]}",
		"duplicate identity":          "{\"users\":[{\"id\":42,\"uuid\":\"a\"},{\"id\":42,\"uuid\":\"b\"}]}",
		"duplicate uuid":              "{\"users\":[{\"id\":42,\"uuid\":\"a\"},{\"id\":43,\"uuid\":\"a\"}]}",
		"zero identity":               "{\"users\":[{\"id\":0,\"uuid\":\"a\"}]}",
		"missing uuid":                "{\"users\":[{\"id\":42}]}",
	} {
		t.Run(name, func(t *testing.T) {
			client := compatibilityClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("ETag", "\"bad-list\"")
				_, _ = w.Write([]byte(body))
			})
			client.userEtag = "\"last-valid-list\""
			if users, err := client.GetUserList(); err == nil {
				t.Fatalf("unsafe response accepted: users=%+v", users)
			}
			if client.userEtag != "\"last-valid-list\"" {
				t.Fatal("invalid response changed cached ETag")
			}
		})
	}
}

func TestReviewUsesOnlyTopLevelUsers(t *testing.T) {
	client := compatibilityClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{\"metadata\":{\"users\":[]},\"users\":[{\"id\":42,\"uuid\":\"a\"}]}"))
	})
	users, err := client.GetUserList()
	if err != nil || len(users) != 1 || users[0].Id != 42 {
		t.Fatalf("nested metadata was mistaken for actual users: users=%+v err=%v", users, err)
	}
}

func TestReviewRejectsTrailingMessagePack(t *testing.T) {
	body, err := msgpack.Marshal(map[string]any{"users": []map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	body = append(body, 0xff)
	client := compatibilityClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/x-msgpack")
		_, _ = w.Write(body)
	})
	if users, err := client.GetUserList(); err == nil {
		t.Fatalf("concatenated MessagePack accepted: users=%+v", users)
	}
}

func TestReviewRejectsDuplicateMessagePackUsers(t *testing.T) {
	var body bytes.Buffer
	encoder := msgpack.NewEncoder(&body)
	if err := encoder.EncodeMapLen(2); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := encoder.EncodeString("users"); err != nil {
			t.Fatal(err)
		}
		if err := encoder.Encode([]UserInfo{}); err != nil {
			t.Fatal(err)
		}
	}
	client := compatibilityClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/x-msgpack")
		_, _ = w.Write(body.Bytes())
	})
	if _, err := client.GetUserList(); err == nil {
		t.Fatal("duplicate MessagePack users accepted")
	}
}
