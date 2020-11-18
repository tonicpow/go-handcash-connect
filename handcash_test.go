package handcash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProfile(t *testing.T) {

	t.Run("missing token", func(t *testing.T) {

		profile, err := GetProfile("79f8dd768817ca86bcc06d468a67e64e4a349658c9d8af70e2251dfc2553de91")

		// assert.Error(t, err)
		t.Errorf("%s %s", profile, err)
	})

	t.Run("bad token", func(t *testing.T) {
		_, err := GetProfile("fakeToken")
		assert.Error(t, err)
	})
}
