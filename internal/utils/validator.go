package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

func ValidateStruct(s interface{}) error {
	log.Printf("[ValidateStruct] Validating struct: %T", s)
	var validate = validator.New()
	err := validate.Struct(s)
	if err != nil {
		log.Printf("[ValidateStruct] Validation failed: %v", err)
	} else {
		log.Println("[ValidateStruct] Validation successful")
	}
	return err
}

func ValidateSession(tokenStr string) (uint, error) {
	log.Printf("[ValidateSession] === Starting token validation ===")
	log.Printf("[ValidateSession] Token length: %d", len(tokenStr))

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		log.Println("[ValidateSession] JWT_SECRET is empty")
		return 0, fmt.Errorf("JWT_SECRET not configured")
	}
	log.Printf("[ValidateSession] JWT_SECRET found, length: %d", len(jwtSecret))

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		log.Printf("[ValidateSession] Token method: %v", token.Method)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("[ValidateSession] Unexpected signing method: %v", token.Method)
			return nil, fmt.Errorf("unexpected signing method")
		}
		log.Println("[ValidateSession] Signing method validated")
		return jwtSecret, nil
	})

	if err != nil {
		log.Printf("[ValidateSession] Token parsing failed: %v", err)
		log.Println("[ValidateSession] === Returning error ===")
		return 0, err
	}

	if !token.Valid {
		log.Println("[ValidateSession] Token is not valid")
		log.Println("[ValidateSession] === Returning error ===")
		return 0, fmt.Errorf("invalid token")
	}

	log.Println("[ValidateSession] Token is valid, extracting claims")

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("[ValidateSession] Failed to extract claims")
		log.Println("[ValidateSession] === Returning error ===")
		return 0, fmt.Errorf("invalid token claims")
	}

	log.Printf("[ValidateSession] Claims extracted: %+v", claims)

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		log.Printf("[ValidateSession] user_id not found or wrong type in claims. Available claims: %+v", claims)
		log.Println("[ValidateSession] === Returning error ===")
		return 0, fmt.Errorf("user_id not found in token claims")
	}

	userID := uint(userIDFloat)
	log.Printf("[ValidateSession] Successfully extracted user_id: %d", userID)
	log.Println("[ValidateSession] === Returning success ===")
	return userID, nil
}
