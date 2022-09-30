package mapper

import (
	"time"
)

var TimeStampFormat = "2006-01-02T15:04:05Z"

func timeStamp(t time.Time) string {
	return t.UTC().Format(TimeStampFormat)
}
