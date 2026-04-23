package sfccauthentication

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/raysaurav/GraphCommerceGateway/shared/gin"
)

/*** Migrating from USID to App Session ID
func SetUSIDtoContext(ctx context.Context, usid string) error {
	gc, err := gin.GinContextFromContext(ctx)
	if err != nil {
		return err
	}
	// 1. Set USID in HTTP Response Header (optional, for client/frontend)
	gc.Writer.Header().Set(string(gin.UsidKey), usid)
	// 2. Set USID inside gin.Context internal key-value store
	gc.Set(string(gin.UsidKey), usid)
	return err
}
***/

/**func GetKeyFromContext(ctx context.Context) string {
	zl := logger.GetLogger()
	gc, err := gin.GinContextFromContext(ctx)
	if err != nil {
		return ""
	}

	// 1. Try to get App-Session-Id from gin.Context key-value store
	if appSessionIdValue, exists := gc.Get(string(gin.XAppSessionID_Key)); exists {
		if appSessionId, ok := appSessionIdValue.(string); ok && appSessionId != "" {
			return appSessionId
		}
	}

	// 2. If not found, fallback to standard Go context
	key, ok := ctx.Value(gin.XAppSessionID_Key).(string)
	zl.Debug().Msgf("ctx", ctx.Value(gin.XAppSessionID_Key))
	if !ok || key == "" {
		return ""
	}
	return key
}**/

func GetKeyFromContext(ctx context.Context) string {

	if gc, err := gin.GinContextFromContext(ctx); err == nil {
		if appSessionIdValue, exists := gc.Get(string(gin.XAppSessionID_Key)); exists {
			if appSessionId, ok := appSessionIdValue.(string); ok && appSessionId != "" {
				return appSessionId
			}
		}
	}

	if key, ok := ctx.Value(gin.XAppSessionID_Key).(string); ok && key != "" {
		// zl.Debug().Str("appSessionId", key).Msg("Extracted App-Session-Id from context")
		return key
	}

	// zl.Warn().Msg("App-Session-Id not found in context")
	return ""
}

func basicAuth(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

// UnmarshalJSON is a generic function to unmarshal JSON data into any type.
func UnmarshalJSON(data []byte, result interface{}) error {
	if err := json.Unmarshal(data, result); err != nil {
		// logger.Error().Err(err).Msgf("Error unmarshalling data into %T", result)
		return err
	}
	return nil
}

func SetAccessTokenToContext(ctx context.Context, accessToken string) error {
	gc, err := gin.GinContextFromContext(ctx)
	if err != nil {
		return err
	}
	gc.Writer.Header().Set(string(gin.AccessToken), accessToken)
	return err
}

func GetAccessTokenFromContext(ctx context.Context) string {
	key, ok := ctx.Value(gin.AccessToken).(string)
	if !ok || key == "" {
		return ""
	}
	return key
}
func SetCustomerIdToContext(ctx context.Context, customerId string) error {
	gc, err := gin.GinContextFromContext(ctx)
	if err != nil {
		return err
	}
	gc.Writer.Header().Set(string(gin.CustomerID), customerId)
	return err
}

func GetCustomerIdFromContext(ctx context.Context) string {
	key, ok := ctx.Value(gin.CustomerID).(string)
	if !ok || key == "" {
		return ""
	}
	return key
}

func GetAuthorizationHeaderFromContext(ctx context.Context) string {
	authHeader, ok := ctx.Value(gin.Authorization).(string)
	if !ok || authHeader == "" {
		return ""
	}

	token := stripBearerPrefix(authHeader)

	// Make token blank as its a client token
	if IssuerIdentifier(token, "sts.windows.net") {
		return ""
	}

	return authHeader
}

// GetSHA256Hash returns the SHA-256 hash of the input string in hexadecimal format.
func GetSHA256Hash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// Strip bearer prefix if any from the token
func stripBearerPrefix(tokenString string) string {
	const bearerPrefix = "Bearer "
	if len(tokenString) >= len(bearerPrefix) &&
		strings.EqualFold(tokenString[:len(bearerPrefix)], bearerPrefix) {
		return strings.TrimSpace(tokenString[len(bearerPrefix):])
	}
	return tokenString
}

// IssuerIdentifier checks if the JWT token's issuer contains the given issuerContext
func IssuerIdentifier(tokenString, issuerContext string) bool {

	// Parse the JWT without verifying the signature
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		// zl.Error().Msgf("Failed to parse token: %v\n", err)
		return false
	}
	// zl.Debug().Msgf("Token parsed successfully")

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		// zl.Error().Msgf("Failed to cast token claims to MapClaims")
		return false
	}

	iss, ok := claims["iss"].(string)
	if !ok {
		// zl.Error().Msgf("Issuer (iss) claim not found or not a string")
		return false
	}

	// zl.Debug().Msgf("Issuer found: %s\n", iss)

	// Check against each issuer fragment
	for _, context := range strings.Split(issuerContext, ",") {
		context = strings.TrimSpace(context)
		if context != "" && strings.Contains(iss, context) {
			// zl.Debug().Msgf("Issuer matched context: %s\n", context)
			return true
		}
	}

	// zl.Debug().Msgf("No matching issuer context found")
	return false
}
