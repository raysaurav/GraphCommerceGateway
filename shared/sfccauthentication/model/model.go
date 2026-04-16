package model

type SFCCAuthResponse struct {
	AccessToken string `json:"access_token"`
	// IDToken               string `json:"id_token"`
	RefreshToken          string `json:"refresh_token"`
	ExpiresIn             int    `json:"expires_in"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
	TokenType             string `json:"token_type"`
	Usid                  string `json:"usid"`
	CustomerID            string `json:"customer_id"`
	EncUserID             string `json:"enc_user_id"`
	IDPAccessToken        string `json:"idp_access_token"`
	IDPRefreshToken       string `json:"idp_refresh_token"`
}

type TokenInfo struct {
	UUID         string `json:"uuid"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // expiration time in seconds
	CustomerID   string `json:"customer_id"`
	// we are having UUID field that holds USID data but we would be needing one specific identifier when okta integrates
	Usid    string `json:"usid"`
	LoginId string `json:"login_id"`
}

type SFCCAuthErrorResponse struct {
	StatusCode string `json:"status_code"`
	Message    string `json:"message"`
}
