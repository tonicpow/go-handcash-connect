package handcash

/*
// BalanceResponse is the balance response
type BalanceResponse struct {
	SpendableSatoshiBalance uint64       `json:"spendableSatoshiBalance"`
	SpendableFiatBalance    float64      `json:"spendableFiatBalance"`
	CurrencyCode            CurrencyCode `json:"currencyCode"`
}*/

/*
// GetBalance gets the user's balance from the handcash connect API
func GetBalance(authToken string) (balanceResponse *BalanceResponse, err error) {

	// Make sure we have an auth token
	if len(authToken) == 0 {
		return nil, fmt.Errorf("missing auth token")
	}

	// TODO: The actual request

	balanceResponse = new(BalanceResponse)

	err = json.Unmarshal([]byte("dummy"), balanceResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance %w", err)
	}

	return
}
*/
