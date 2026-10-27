package sing

import (
	"github.com/InazumaV/V2bX/api/panel"
	"github.com/InazumaV/V2bX/common/format"
)

// Callers hold users.mapLock. A UUID can be present on several local nodes.
func (b *Sing) rememberUserID(tag string, user panel.UserInfo) {
	b.users.uidMap[format.UserTag(tag, user.Uuid)] = user.Id
}
func (b *Sing) forgetUserID(tag, uuid string) {
	delete(b.users.uidMap, format.UserTag(tag, uuid))
}
func (b *Sing) userID(tag, uuid string) int {
	return b.users.uidMap[format.UserTag(tag, uuid)]
}
