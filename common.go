package main

import "time"

func pString(s string) *string {
	return &s
}

func pInt(i int) *int {
	return &i
}

func pTime(t time.Time) *time.Time {
	return &t
}
