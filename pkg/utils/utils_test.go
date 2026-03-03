package utils

import "testing"

func TestGetLocalIP(t *testing.T) {
	ip, err := GetLocalIP()
	if err != nil {
		t.Errorf("Error: %v\n", err)
	}
	t.Logf("ip: %s\n", ip.String())
}


