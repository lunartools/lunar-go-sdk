package lunartools

import "lunar-go-sdk/src/client"

type Client = client.Client

var NewClient = client.NewClient

func String(s string) *string {
	return &s
}

func Int(i int) *int {
	return &i
}

func Float64(f float64) *float64 {
	return &f
}

func Bool(b bool) *bool {
	return &b
}
