package utils

import "time"

var CambodiaTZ *time.Location

func init() {
	CambodiaTZ, _ = time.LoadLocation("Asia/Phnom_Penh")
}

func Now() time.Time {
	return time.Now().In(CambodiaTZ)
}
