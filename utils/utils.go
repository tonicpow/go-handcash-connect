package utils

import "time"

// ISOTimestamp generates a timestamp in the same format as Javascript Date().toISOString()
func ISOTimestamp() string {
	JavascriptISOString := "2006-01-02T15:04:05.999Z07:00"
	return time.Now().UTC().Format(JavascriptISOString)
}
