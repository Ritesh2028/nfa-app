package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func ErrorResponse(c *gin.Context, message string, code int) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

func SuccessResponse(c *gin.Context, message string, code int) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

const secretKey = "blueinvent" // Make sure this is accessible and secured

// GenerateJWT creates a new JWT for the given email.
func GenerateJWT(email string) (string, error) {
	// Set up claims, including "exp" for expiration (24 hours from creation)
	claims := jwt.MapClaims{
		"email": email,                                 // Consistent lowercase key
		"exp":   time.Now().Add(24 * time.Hour).Unix(), // Token expiry set to 24 hours
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with the secret key
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err // Return any error encountered during signing
	}

	return signedToken, nil
}

// ValidateJWT parses and validates a JWT string.
func ValidateJWT(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing error: %w", err)
	}

	// Check token validity explicitly
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

func ValidatePassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
