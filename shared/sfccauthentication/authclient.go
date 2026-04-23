package sfccauthentication

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/raysaurav/GraphCommerceGateway/shared/constant"
	"github.com/raysaurav/GraphCommerceGateway/shared/gin"
	"github.com/raysaurav/GraphCommerceGateway/shared/httpclient"
	"github.com/raysaurav/GraphCommerceGateway/shared/sfccauthentication/model"
)

type AuthClient interface {
	TokenFetchStrategy(ctx context.Context) (*model.TokenInfo, error)
}

type AuthWrapper struct {
	httpClient httpclient.HttpClientInterface
	authConfig *AuthConfig
}

func NewAuthClient(httpClient httpclient.HttpClientInterface, authConfig *AuthConfig) AuthClient {
	return &AuthWrapper{
		httpClient: httpClient,
		authConfig: authConfig,
	}
}

func (a *AuthWrapper) TokenFetchStrategy(ctx context.Context) (*model.TokenInfo, error) {
	authorization := GetAuthorizationHeaderFromContext(ctx)

	if authorization == "" {
		// a.logger.Warn().Msg("Authorization token missing — routing to guest flow")
		return a.FetchTokenInfoForGuest(ctx)
	}
	// Check login scope (outside switch)
	loginScope, ok := ctx.Value("login_scope").(string)
	isGuestScope := ok && loginScope == "guest"
	isRegisteredUser := loginScope != "guest"
	// Decision logic
	switch {
	case isGuestScope:
		// a.logger.Warn().Msg("Login context indicates guest, routing to guest flow")
		return a.FetchTokenInfoForGuest(ctx)
	case isRegisteredUser:
		// a.logger.Debug().Msg("Auth0 token detected, routing to registered flow")
		return a.FetchTokenInfoForRegisteredUser(ctx)
	default:
		// a.logger.Error().Msg("Token is not from a valid identity provider")
		return nil, errors.New("unauthorized: invalid identity provider in token issuer")
	}
}

func (a *AuthWrapper) FetchTokenInfoForGuest(ctx context.Context) (*model.TokenInfo, error) {
	authKey := GetKeyFromContext(ctx)
	redisGuestKey := constant.SFCC_GUEST_TOKEN_NS + authKey
	if redisGuestKey == constant.SFCC_GUEST_TOKEN_NS {
		// a.logger.Warn().Msg("App-Session-Id not found in context")
		return nil, errors.New("App Session Id not found")
	}
	return a.FetchOrRefreshTokenFromOAuth(ctx, redisGuestKey, authKey)
}

// FetchOrRefreshToken is a common function for guest/Akamai APIs tokens.
func (a *AuthWrapper) FetchOrRefreshTokenFromOAuth(ctx context.Context, redisKey string, authKey string) (*model.TokenInfo, error) {

	// TODO: Implement Redis lookup for existing token
	// var authObject *model.SFCCAuthResponse
	// For now, generate a new token
	authObject, err := a.generateSFCCAuthToken(ctx, authKey)
	if err != nil {
		return nil, err
	}

	if a.isAccessTokenExpired(authObject.AccessToken) {
		refreshAuthObject, authError := a.GenerateSFCCAccessTokenFromRefreshToken(ctx, redisKey, authObject.RefreshToken, authObject.AccessToken)
		if authError != nil {
			return nil, authError
		}

		return a.constructTokenInfo(ctx, authKey, refreshAuthObject), nil
	} else {
		return a.constructTokenInfo(ctx, authKey, authObject), nil
	}
}

// Check if access token is expired extracting the expiry and evaluating it with unix time
func (a *AuthWrapper) isAccessTokenExpired(tokenString string) bool {
	token, _ := jwt.Parse(tokenString, nil)

	// Check if the token is valid and extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Get the expiration time from the claims
		if exp, ok := claims["exp"].(float64); ok {
			expirationTime := time.Unix(int64(exp), 0)
			return time.Now().After(expirationTime)
		}
	}
	return false
}

// GenerateSFCCAccessTokenFromRefreshToken generates a new SFCC access token using a refresh token for guest user
func (a *AuthWrapper) GenerateSFCCAccessTokenFromRefreshToken(ctx context.Context, key string, refreshToken string, accessToken string) (*model.SFCCAuthResponse, error) {

	// Construct the base URL by replacing %s with shortCode
	// a.logger.Info().Msgf(" *** GenerateSFCCAccessTokenFromRefreshToken *** ")
	endpoint, headers, formData := a.prepareSFCCAccessTokenFromRefreshToken(accessToken, refreshToken)
	resp, err := a.httpClient.Post(ctx, endpoint, headers, nil, nil, formData.Encode())

	if err != nil {
		// a.logger.Error().Err(err).Msg("Error while calling the SFCC Auth API")
		return nil, err
	}

	if resp.IsError() {

		// Handle 409 Conflict: fallback to Redis
		if resp.StatusCode() == 409 {
			// a.logger.Debug().Msgf(" *** Encounterd 409 : GenerateSFCCAccessTokenFromRefreshToken *** ")
			return nil, errors.New("409 conflict")
		}
		return a.generateSFCCAuthToken(ctx, key)
	}

	var authResponse model.SFCCAuthResponse
	if err := UnmarshalJSON(resp.Body(), &authResponse); err != nil {
		return nil, err
	}
	return &authResponse, nil
}

func (a *AuthWrapper) prepareSFCCAccessTokenFromRefreshToken(token string, refreshToken string) (string, map[string]string, url.Values) {
	baseURL := fmt.Sprintf(a.authConfig.BaseURL, a.authConfig.ShortCode)
	if IsGuest(token) {
		endpoint := fmt.Sprintf("%s/shopper/auth/v1/organizations/%s/oauth2/token", baseURL, a.authConfig.OrganizationID)
		headers := map[string]string{
			constant.CONSTANT_TYPE_NAME: constant.CONTENT_TYPE,
			constant.AUTHORIZATION:      basicAuth(a.authConfig.Username, a.authConfig.Password),
		}
		formData := url.Values{
			constant.GRANT_TYPE: {"refresh_token"},
			"refresh_token":     {refreshToken}}

		return endpoint, headers, formData
	} else {

		endpoint := fmt.Sprintf("%s/shopper/auth/v1/organizations/%s/oauth2/trusted-system/token", baseURL, a.authConfig.OrganizationID)
		// Prepare headers
		headers := map[string]string{
			constant.CONSTANT_TYPE_NAME: constant.CONTENT_TYPE,
			constant.AUTHORIZATION:      basicAuth(a.authConfig.Username, a.authConfig.Password),
		}

		// Extract usid from context
		UsidContextKey := extractUSID(token)

		// Prepare form data
		formData := url.Values{
			"grant_type": {"client_credentials"},
			"hint":       {a.authConfig.Hint},
			"login_id":   {extractLoginIdForRegisteredCustomer(token)},
			"idp_origin": {a.authConfig.IDPOrigin},
			"client_id":  {a.authConfig.Username},
			"channel_id": {a.authConfig.ChannelID},
			"usid":       {extractUSID(token)},
		}
		// Conditionally add usid to form data
		if UsidContextKey != "" {
			// a.logger.Debug().Msgf("Adding USID to formData %s", UsidContextKey)
			formData.Set("usid", UsidContextKey)
		}
		return endpoint, headers, formData
	}

}

// isGuest checks if the provided JWT token represents a guest user.
// It parses the JWT token to extract claims and checks specific claim values.
// Returns true if the user is a guest, false otherwise.
func IsGuest(tokenString string) bool {
	// Parse the JWT token without validation
	token, _ := jwt.Parse(tokenString, nil)
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if isb, ok := claims["isb"].(string); ok {
			if extractParts(isb, "upn") == "Guest" {
				return true
			}
		}
	}
	return false

}

// extractUSID extracts the USID from the JWT token's 'sub' claim.
// It returns the USID as a string if found, or an empty string if not found.
func extractUSID(tokenString string) string {
	// Parse the JWT token without validation
	token, _ := jwt.Parse(tokenString, nil)
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if sub, ok := claims["sub"].(string); ok {
			return extractParts(sub, "usid")

		}
	}
	return ""
}

func extractLoginId(tokenString string) string {
	// Parse the JWT token without validation
	token, _ := jwt.Parse(tokenString, nil)
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if isb, ok := claims["isb"].(string); ok {
			if extractParts(isb, "upn") != "Guest" {
				loginId := extractParts(isb, "upn")
				return loginId
			}
		}
	}
	return ""
}

// extractCustomerID extracts the customer ID from the JWT token's 'isb' claim.
// It returns the customer ID as a string if found, or an empty string if not found.
func extractCustomerID(tokenString string) string {
	// Parse the JWT token without validation
	token, _ := jwt.Parse(tokenString, nil)

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if isb, ok := claims["isb"].(string); ok {
			if IsGuest(tokenString) {
				return extractParts(isb, "gcid")
			} else {
				return extractParts(isb, "rcid")
			}
		}
	}
	return ""
}

func extractParts(isb string, segment string) string {
	// Split the string by "::"
	parts := strings.Split(isb, "::")

	// Initialize a variable to hold the upn value
	var isbPartsValue string = ""

	// Iterate through the parts to find the upn value
	for _, part := range parts {
		if strings.HasPrefix(part, segment+":") {
			// Extract the upn value
			isbPartsValue = strings.TrimPrefix(part, segment+":")
			break
		}
	}
	return isbPartsValue
}

// extractLoginIdForRegisteredCustomer extracts the login ID for a registered customer from a JWT token.
func extractLoginIdForRegisteredCustomer(tokenString string) string {
	// Parse the JWT token without validation
	token, _ := jwt.Parse(tokenString, nil)

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if isb, ok := claims["isb"].(string); ok {
			return extractParts(isb, "upn")
		}
	}
	return ""
}

func (a *AuthWrapper) generateSFCCAuthToken(ctx context.Context, key string) (*model.SFCCAuthResponse, error) {

	// Construct the base URL by replacing %s with shortCode
	baseURL := fmt.Sprintf(a.authConfig.BaseURL, a.authConfig.ShortCode)

	endpoint := fmt.Sprintf("%s/shopper/auth/v1/organizations/%s/oauth2/token", baseURL, a.authConfig.OrganizationID)

	headers := map[string]string{
		constant.CONSTANT_TYPE_NAME: constant.CONTENT_TYPE,
		constant.AUTHORIZATION:      basicAuth(a.authConfig.Username, a.authConfig.Password),
	}

	formData := url.Values{
		constant.GRANT_TYPE:      {a.authConfig.GrantType},
		constant.REDIRECT_URI:    {a.authConfig.RedirectURI},
		constant.CHANNEL_ID_NAME: {a.authConfig.SiteId},
	}
	resp, err := a.httpClient.Post(ctx, endpoint, headers, nil, nil, formData.Encode())

	if err != nil {
		// a.logger.Error().Err(err).Msg("Error while calling the SFCC Auth API")
		return nil, err
	}

	if resp.IsError() {
		// Handle 409 Conflict: fallback to Redis
		if resp.StatusCode() == 409 {
			// a.logger.Debug().Msgf(" *** Encounterd 409 :generateSFCCAuthTokenAndCache *** ")
			return nil, errors.New("409 conflict")
		}
		return nil, errors.New(strconv.Itoa(resp.StatusCode()))
	}

	// Log the response

	var authResponse model.SFCCAuthResponse
	if err := UnmarshalJSON(resp.Body(), &authResponse); err != nil {
		return nil, err
	}

	return &authResponse, nil
}

func (a *AuthWrapper) constructTokenInfo(ctx context.Context, key string, sfccauthResponse *model.SFCCAuthResponse) *model.TokenInfo {

	/*** As part of Migration from USID to App Session ID , USID storing in Context not required
	err := SetUSIDtoContext(ctx, key)
	if err != nil {
		a.logger.Error().Msgf("Error setting up x-custom-usid response header value: %s", key)
	}
	***/

	customerIdErr := SetCustomerIdToContext(ctx, sfccauthResponse.CustomerID)
	if customerIdErr != nil {
		// a.logger.Error().Msgf("Error setting up x-customer-id response header value: %s", sfccauthResponse.CustomerID)
	}
	/*** As part of App solely depends on USID , accessToken storing in context was removed
	accessTokenErr := SetAccessTokenToContext(ctx, sfccauthResponse.AccessToken)
	if accessTokenErr != nil {
		a.logger.Error().Msgf("Error setting up x-access-token response header value: %s", sfccauthResponse.AccessToken)
	}
	***/

	return &model.TokenInfo{
		UUID:         key,
		AccessToken:  sfccauthResponse.AccessToken,
		ExpiresIn:    sfccauthResponse.ExpiresIn,
		RefreshToken: sfccauthResponse.RefreshToken,
		CustomerID:   sfccauthResponse.CustomerID,
		Usid:         sfccauthResponse.Usid,
		LoginId:      extractLoginId(sfccauthResponse.AccessToken),
	}
}

// parseJWTClaims parses JWT token (without verifying signature) and returns claims map
func parseJWTClaims(tokenString string) (jwt.MapClaims, error) {
	tokenString = stripBearerPrefix(tokenString)

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWT authorization token format")
	}

	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse authorization token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid JWT claims format")
	}

	return claims, nil
}

// ExtractEmailFromJWT extracts email from JWT
func ExtractEmailFromJWT(tokenString string) (string, error) {
	claims, err := parseJWTClaims(tokenString)
	if err != nil {
		return constant.EMPTY_STRING, err
	}

	if email, exists := claims["email"].(string); exists {
		return email, nil
	}

	if userData, exists := claims["user_data"].(map[string]interface{}); exists {
		if email, ok := userData["email"].(string); ok {
			return email, nil
		}
	}

	return constant.EMPTY_STRING, errors.New("email not found in token payload")
}

// FetchTokenInfoForRegisteredUser validates the access token and fetches token information.
// It handles different scenarios such as missing USID, expired tokens, and guest user checks.
func (a *AuthWrapper) FetchTokenInfoForRegisteredUser(ctx context.Context) (*model.TokenInfo, error) {
	authorization := GetAuthorizationHeaderFromContext(ctx)

	oktaToken := stripBearerPrefix(authorization)
	email, _ := ExtractEmailFromJWT(authorization)
	authKey := GetKeyFromContext(ctx)
	redisKey := constant.SFCC_REGISTERED_CUSTOMER_TOKEN_NS + GetSHA256Hash(email)

	// Get session ID from context
	if sessionID, ok := ctx.Value(gin.XAppSessionID_Key).(string); ok && sessionID != "" {
		redisKey = redisKey + ":" + sessionID
	}
	if oktaToken == "" || !a.isAccessTokenExpired(oktaToken) {
		// create a new token
		authResponse, authError := a.generateSFCCTrustedTokenAndCache(ctx, nil, redisKey, email)
		if authError != nil {
			return nil, authError
		}

		return a.constructTokenInfo(ctx, authKey, &model.SFCCAuthResponse{
			AccessToken: authResponse.AccessToken,
			CustomerID:  authResponse.CustomerID,
			Usid:        authResponse.Usid,
		}), nil
	}

	// TODO: Implement Redis lookup for existing token
	// For now, generate a new token when expired
	if a.isAccessTokenExpired(oktaToken) {
		// refresh token
		authResponse, authError := a.generateSFCCTrustedTokenAndCache(ctx, &oktaToken, redisKey, email)
		if authError != nil {
			return nil, authError
		}

		return a.constructTokenInfo(ctx, authKey, &model.SFCCAuthResponse{
			AccessToken: authResponse.AccessToken,
			CustomerID:  authResponse.CustomerID,
			Usid:        authResponse.Usid,
		}), nil

	} else {
		// Token is still valid, but we need to get it from somewhere
		// TODO: Implement Redis lookup here
		// For now, generate a new token
		authResponse, authError := a.generateSFCCTrustedTokenAndCache(ctx, &oktaToken, redisKey, email)
		if authError != nil {
			return nil, authError
		}

		return a.constructTokenInfo(ctx, authKey, &model.SFCCAuthResponse{
			AccessToken: authResponse.AccessToken,
			CustomerID:  authResponse.CustomerID,
			Usid:        authResponse.Usid,
		}), nil

	}
}

func (a *AuthWrapper) generateSFCCTrustedTokenAndCache(ctx context.Context, token *string, redisKey string, loginId string) (*model.SFCCAuthResponse, error) {
	// Construct the base URL by replacing %s with shortCode

	baseURL := fmt.Sprintf(a.authConfig.BaseURL, a.authConfig.ShortCode)

	endpoint := fmt.Sprintf("%s/shopper/auth/v1/organizations/%s/oauth2/trusted-system/token", baseURL, a.authConfig.OrganizationID)

	headers := map[string]string{
		constant.CONSTANT_TYPE_NAME: constant.CONTENT_TYPE,
		constant.AUTHORIZATION:      basicAuth(a.authConfig.Username, a.authConfig.Password),
	}

	formData := url.Values{
		"grant_type": {"client_credentials"},
		"hint":       {a.authConfig.Hint},
		"login_id":   {loginId},
		"idp_origin": {a.authConfig.IDPOrigin},
		"client_id":  {a.authConfig.Username},
		"channel_id": {a.authConfig.ChannelID},
	}

	if token != nil {
		formData.Set("usid", extractUSID(*token))
	}

	resp, err := a.httpClient.Post(ctx, endpoint, headers, nil, nil, formData.Encode())
	if err != nil {
		// a.logger.Error().Err(err).Msg("Error while calling the SFCC Auth API")
		return nil, err
	}

	if resp.IsError() {

		// Handle 409 Conflict: fallback to Redis
		if resp.StatusCode() == 409 {
			// a.logger.Debug().Msgf(" *** Encounterd 409 : generateSFCCTrustedTokenAndCache *** ")
			return nil, errors.New("409 conflict")
		}

		return nil, errors.New("SFCC API request returned an error: " + resp.Status())
	}

	var authResponse model.SFCCAuthResponse
	if err := UnmarshalJSON(resp.Body(), &authResponse); err != nil {
		return nil, err
	}
	return &authResponse, nil
}
