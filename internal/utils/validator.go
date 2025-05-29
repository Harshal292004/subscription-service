package utils

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

func ValidateStruct(s interface{}) error {
	var validate = validator.New()
	return validate.Struct(s)
}

func ValidateSession(tokenStr string) (uint, error) {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user_id not found in token claims")
	}

	return uint(userIDFloat), nil
}
