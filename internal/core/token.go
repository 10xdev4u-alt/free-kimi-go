package core

type TokenInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserId       string `json:"userId"`
	RefreshTime  int64  `json:"-"`
}
