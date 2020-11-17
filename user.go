package handcash

// User are the user fields returned by the public and private profile endpoints
type User struct {
	PublicProfile  PublicProfile  `json:"publicProfile"`
	PrivateProfile PrivateProfile `json:"privateProfile"`
}

// PublicProfile is the public profile
type PublicProfile struct {
	AvatarURL         string `json:"avatarUrl"`
	DisplayName       string `json:"displayName"`
	Handle            string `json:"handle"`
	Paymail           string `json:"paymail"`
	ID                string `json:"id"`
	LocalCurrencyCode string `json:"localCurrencyCode"`
}

// TODO: Spendable balance was made its own request
// SpendableBalance  int64  `json:"spendableBalance"`

// PrivateProfile is the private profile
type PrivateProfile struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}
