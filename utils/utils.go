package utils

import "time"

const isoFormat = "2006-01-02T15:04:05.999Z07:00"

// ISOTimestamp generates a timestamp in the same format as Javascript Date().toISOString()
func ISOTimestamp() string {
	return time.Now().UTC().Format(isoFormat)
}
