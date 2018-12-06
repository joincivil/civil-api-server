package auth

// Token is a decoded JWT auth token
type Token struct {
	Sub     string
	IsAdmin bool
}

// LoginResponse is sent when a User successfully logs in
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	UID          string `json:"uid"`
}
