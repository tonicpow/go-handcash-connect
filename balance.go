package handcash

import "fmt"

// Balance is the balance response
type Balance struct {
}

// GetBalance gets the user's balance from the handcash connect API
func GetBalance(authToken string) (balance *Balance, err error) {

	// Make sure we have an auth token
	if len(authToken) == 0 {
		return nil, fmt.Errorf("missing auth token")
	}
	return
}
