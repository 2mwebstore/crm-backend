package utils

import "time"

// PhnomPenhLocation returns the Asia/Phnom_Penh location, falling back to a
// fixed UTC+7 offset if the timezone database isn't available on the host.
func PhnomPenhLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Phnom_Penh")
	if err != nil {
		loc = time.FixedZone("ICT", 7*60*60)
	}
	return loc
}

// NowInPhnomPenh returns the current time in Phnom Penh (Asia/Phnom_Penh, UTC+7).
func NowInPhnomPenh() *time.Time {
	now := time.Now().In(PhnomPenhLocation())
	return &now
}
