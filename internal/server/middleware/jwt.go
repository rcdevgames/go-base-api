package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const jwtClaimsKey contextKey = "jwtClaims"

// JWT ensures incoming requests provide a valid JWT bearer token before reaching protected handlers.
func JWT(secret string, maxAge time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secret == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := parseJWT(r.Header.Get("Authorization"), secret, maxAge)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), jwtClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func parseJWT(authHeader, secret string, maxAge time.Duration) (jwt.MapClaims, error) {
	if authHeader == "" {
		return nil, errors.New("missing auth header")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, errors.New("invalid authorization header")
	}

	token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	if err := validateClaims(claims, maxAge); err != nil {
		return nil, err
	}
	return claims, nil
}

// JWTClaimsFromContext extracts parsed JWT claims from the request context.
func JWTClaimsFromContext(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(jwtClaimsKey).(jwt.MapClaims)
	return claims, ok
}

func validateClaims(claims jwt.MapClaims, maxAge time.Duration) error {
	if expVal, ok := claims["exp"]; ok {
		exp, err := parseNumericDate(expVal)
		if err != nil {
			return err
		}
		if time.Now().After(exp) {
			return errors.New("token expired")
		}
		return nil
	}
	if maxAge <= 0 {
		return nil
	}
	iatVal, ok := claims["iat"]
	if !ok {
		return errors.New("token missing expiration information")
	}
	iat, err := parseNumericDate(iatVal)
	if err != nil {
		return err
	}
	if time.Since(iat) > maxAge {
		return errors.New("token expired")
	}
	return nil
}

func parseNumericDate(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case float64:
		return time.Unix(int64(v), 0), nil
	case float32:
		return time.Unix(int64(v), 0), nil
	case int64:
		return time.Unix(v, 0), nil
	case int:
		return time.Unix(int64(v), 0), nil
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(int64(f), 0), nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(int64(f), 0), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported numeric date type %T", value)
	}
}
