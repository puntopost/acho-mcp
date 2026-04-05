package cli

import "time"

const (
	DateYMD      = "2006-01-02"
	DateYMD_HM   = "2006-01-02 15:04"
	DateYMD_HMS  = "2006-01-02 15:04:05"
	DateYMD_HMST = "2006-01-02T15:04:05"
)

// ParseDate tries multiple date formats and returns the first successful parse.
func ParseDate(s string) (time.Time, bool) {
	for _, layout := range []string{DateYMD_HMST, DateYMD_HMS, DateYMD_HM, DateYMD} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
