package utils

import (
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

func ValidateStruct(s interface{}) error {
	var validate = validator.New()
	return validate.Struct(s)
}

func ValidateSession(tokenStr string) (uint, error) {
	var jwt_secret = os.Getenv("JWT_SECRET")

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return jwt_secret, nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}
	claims := token.Claims.(jwt.MapClaims)
	userId := uint(claims["user_id"].(float64))
	return userId, nil
}
