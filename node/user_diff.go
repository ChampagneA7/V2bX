package node

import "github.com/InazumaV/V2bX/api/panel"

// Id is an opaque panel-issued identity, which can represent a subscription.
// Any changed identity or limit must replace the old core user and limiter entry.
func compareUserList(old, current []panel.UserInfo) (deleted, added []panel.UserInfo) {
	oldByUUID := make(map[string]panel.UserInfo, len(old))
	for _, user := range old {
		oldByUUID[user.Uuid] = user
	}
	for _, user := range current {
		if previous, exists := oldByUUID[user.Uuid]; exists && previous == user {
			delete(oldByUUID, user.Uuid)
		} else {
			added = append(added, user)
		}
	}
	for _, user := range oldByUUID {
		deleted = append(deleted, user)
	}
	return deleted, added
}
