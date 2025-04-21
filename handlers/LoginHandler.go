package handlers

import (
	"database/sql"
	"net/http"
	"nfa-app/models"
	"nfa-app/storage"
	"nfa-app/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func LoginHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for the token in the Authorization header
		token := c.GetHeader("Authorization")

		if token != "" {
			parsedToken, err := utils.ValidateJWT(token)
			if err != nil || !parsedToken.Valid {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token", "details": err.Error()})
				return
			}

			// Ensure parsedToken.Claims can be type-asserted to jwt.MapClaims
			claims, ok := parsedToken.Claims.(jwt.MapClaims)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims structure"})
				return
			}

			// Check if the "Email" field exists and is of type string
			email, ok := claims["email"].(string)
			if !ok || email == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Email claim missing or invalid"})
				return
			}

			// Retrieve the user based on the email claim
			user, err := storage.GetUserByEmail(db, email)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found", "details": err.Error()})
				return
			}

			// Token is valid and user is found
			c.JSON(http.StatusOK, gin.H{
				"message": "User successfully logged in via token",
				"token":   token,
				"user": gin.H{
					"id":    user.ID,
					"email": user.Email,
				},
			})
			return
		}

		// No token; proceed with email and password login
		var loginData struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
			IP       string `json:"ip" binding:"required"` // Include IP in the JSON payload
		}

		if err := c.ShouldBindJSON(&loginData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// Retrieve user by email
		user, err := storage.GetUserByEmail(db, loginData.Email)
		if err != nil || user.Password != loginData.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials", "details": err.Error()})
			return
		}

		// Generate a new JWT token
		newToken, err := utils.GenerateJWT(user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token", "details": err.Error()})
			return
		}

		// // Fetch the "multiple sessions" setting
		// allowMultipleSessions := false
		// err = db.QueryRow("SELECT allow_multiple_sessions FROM settings LIMIT 1").Scan(&allowMultipleSessions)
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings", "details": err.Error()})
		// 	return
		// }

		// Create and save a new session
		session := &models.Session{
			UserID:    user.ID,
			SessionID: newToken,
			HostName:  user.Email,
			IPAddress: loginData.IP,
			Timestamp: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		if err := storage.SaveSession(db, session); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "New token generated successfully",
			"token":   newToken,
		})
	}
}

func GetSessionHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			utils.ErrorResponse(c, "No token provided", http.StatusUnauthorized)
			return
		}

		parsedToken, err := utils.ValidateJWT(token)
		if err != nil {
			utils.ErrorResponse(c, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims := parsedToken.Claims.(jwt.MapClaims)
		exp, ok := claims["exp"].(float64)
		if !ok || time.Now().Unix() > int64(exp) {
			utils.ErrorResponse(c, "Token expired", http.StatusUnauthorized)
			return
		}

		email, ok := claims["Email"].(string)
		if !ok {
			utils.ErrorResponse(c, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		user, err := storage.GetUserByEmail(db, email)
		if err != nil {
			utils.ErrorResponse(c, "User not found", http.StatusUnauthorized)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User is logged in", "user": user})
	}
}

func DeleteSessionHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")

		userIDInt, err := strconv.Atoi(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		if err := storage.DeleteSession(db, userIDInt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Session deleted, user logged out"})
	}
}
