// +build windows

package notification

import (
	"github.com/Henry-Tsukino/wbg-albiondata-client/log"
	toast "gopkg.in/toast.v1"
)

func Push(msg string) {
	note := toast.Notification{AppID: "WBG Client", Title: "WBG Client", Message: msg}

	err := note.Push()
	if err != nil {
		log.Error(err)
	}
}
