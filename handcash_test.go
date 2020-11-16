package handcash

import "testing"

func TestGetProfile(t *testing.T) {

	_, err := GetProfile("fakeToken")
	if err != nil {
		t.Logf("Got err %xs", err)
	}
}
