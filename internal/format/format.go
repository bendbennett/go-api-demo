package format

import "time"

// MsecToTime converts msec to time.Time.
// As the first arg to time.Unix is expected to be sec, the supplied
// value needs to be divided by 1000. The second arg to time.Unix is
// nsec, so the msec that remain after converting to sec are
// subtracted and multiplied by 1000000.
func MsecToTime(msec int64) time.Time {
	const (
		divisor    = 1e3
		multiplier = 1e6
	)

	t := time.Time{}

	if msec > 0 {
		sec := msec / divisor
		nsec := (msec - (sec * divisor)) * multiplier
		t = time.Unix(sec, nsec)
	}

	return t
}
