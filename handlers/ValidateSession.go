package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"nfa-app/models"
	"strings"

	"github.com/gin-gonic/gin"
)

func ValidateSession(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestData struct {
			SessionData string `json:"sessionData" binding:"required"`
		}

		err := c.ShouldBindJSON(&requestData)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
			return
		}

		session, err := getSessionBySessionID(db, requestData.SessionData)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			return
		}

		if !strings.EqualFold(session.SessionID, requestData.SessionData) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		}

		var host_name string
		err = db.QueryRow("SELECT role_id FROM users WHERE email = $1", session.HostName).Scan(&host_name)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session", "details:": err.Error()})
			return
		}

		var role_name string
		err = db.QueryRow("SELECT role_name FROM roles WHERE role_id = $1", host_name).Scan(&role_name)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid role", "details:": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "Session validated",
			"session_id": session.SessionID,
			"host_name":  session.HostName,
			"role_name":  role_name,
		})
	}
}

func getSessionBySessionID(db *sql.DB, sessionID string) (*models.Session, error) {
	query := `SELECT session_id, user_id, host_name, timestp FROM session WHERE session_id = $1`

	var session models.Session

	err := db.QueryRow(query, sessionID).Scan(&session.SessionID, &session.UserID, &session.HostName, &session.Timestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	return &session, nil
}
