package auth

import (
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/env"
	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserId string        `json:"user_id"`
	Roles  []domain.Role `json:"roles"`
	jwt.RegisteredClaims
}

var secretKey []byte

const (
	secretDefaultValue  = "your_jwt_secret"
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
)

func init() {
	secretKey = []byte(env.GetString("JWT_SECRET", secretDefaultValue))
}

func GenerateToken(userId string, roles []domain.Role, expiresAt time.Time) (string, error) {
	claims := UserClaims{
		UserId: userId,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func GetClaimsFromHeader(req *http.Request) (*UserClaims, error) {
	token := req.Header.Get(authorizationHeader)
	if token == "" || !strings.HasPrefix(token, bearerPrefix) {
		return nil, domain.ErrInvalidUserCredentials
	}

	token = strings.TrimPrefix(token, bearerPrefix)
	return GetClaims(token)
}

func GetClaims(token string) (*UserClaims, error) {
	claims := &UserClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})
	return claims, err
}

func HasRequiredRoles(userRoles []domain.Role, rolesNeeded []domain.Role) bool {
	for _, role := range userRoles {
		if slices.Contains(rolesNeeded, role) {
			return true
		}
	}
	return false
}
