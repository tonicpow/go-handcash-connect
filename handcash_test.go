package handcash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProfile(t *testing.T) {

	t.Run("missing token", func(t *testing.T) {
		_, err := GetProfile("")
		assert.Error(t, err)
	})

	t.Run("bad token", func(t *testing.T) {
		_, err := GetProfile("fakeToken")
		assert.Error(t, err)
	})
}
