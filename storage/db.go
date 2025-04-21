package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"nfa-app/models"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() *sql.DB {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		user, password, dbname, host, port)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	return db
}

func GetDB() *sql.DB {
	return db
}

func SaveSession(db *sql.DB, session *models.Session) error {
	// If multiple sessions are NOT allowed, delete existing sessions for this user
	deleteQuery := `DELETE FROM session WHERE user_id = $1`
	_, err := db.Exec(deleteQuery, session.UserID)
	if err != nil {
		return err
	}

	// Create the new session after deleting old ones
	insertQuery := `INSERT INTO session (user_id, session_id, host_name, ip_address, timestp, expires_at)
                    VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = db.Exec(insertQuery, session.UserID, session.SessionID, session.HostName, session.IPAddress, session.Timestamp, session.ExpiresAt)
	return err
}

func GetSession(db *sql.DB, userID int) (*models.Session, error) {
	var session models.Session
	query := `SELECT user_id, session_id, host_name, timestp FROM session WHERE user_id = $1`
	err := db.QueryRow(query, userID).Scan(&session.UserID, &session.SessionID, &session.HostName, &session.Timestamp)
	return &session, err
}

func DeleteSession(db *sql.DB, userID int) error {
	query := `DELETE FROM session WHERE user_id = $1`
	_, err := db.Exec(query, userID)
	return err
}

func GetUserByEmail(db *sql.DB, email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password FROM users WHERE email = $1`

	err := db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to query user: %v", err)
	}

	return &user, nil
}

func UpdateSessionToken(db *sql.DB, userID int, token string, email string) error {
	query := `UPDATE session SET session_id = $1, host_name = $2, timestamp = $3 WHERE user_id = $4`
	_, err := db.Exec(query, token, email, time.Now(), userID)
	return err
}

func CleanupExpiredSessions(db *sql.DB) error {
	threshold := time.Now().Add(-24 * time.Hour)
	_, err := db.Exec("DELETE FROM session WHERE expires_at < $1", threshold)
	return err
}

func LogChange(db *sql.DB, userID, entityType, entityID, changeType, oldValue, newValue string) error {
	query := `INSERT INTO user_changes (user_id, entity_type, entity_id, change_type, old_value, new_value) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(query, userID, entityType, entityID, changeType, oldValue, newValue)
	return err
}

// GetUserBySessionID retrieves a User by the given session ID from the database.
func GetUserBySessionID(db *sql.DB, sessionID string) (*models.User, error) {
	// Assuming there is a table named 'session' where the session ID is stored along with user ID
	// and a table named 'users' where user details are stored.
	query := `
		SELECT u.id, u.email, u.name,
			   u.created_at, u.updated_at, u.first_access, u.last_access,
			   u.profile_picture, u.is_admin, u.address, u.city, 
			   u.state, u.country, u.zip_code, u.phone_no, r.role_name, d.department_name
		FROM session s
		JOIN users u ON s.user_id = u.id
		JOIN roles r ON u.role_id = r.role_id
		JOIN departments d ON u.department_id = d.department_id
		WHERE s.session_id = $1
	`

	var user models.User
	var firstAccess, lastAccess sql.NullTime

	err := db.QueryRow(query, sessionID).Scan(
		&user.ID, &user.Email, &user.Name,
		&user.CreatedAt, &user.UpdatedAt,
		&firstAccess, &lastAccess, &user.ProfilePic,
		&user.Address,
		&user.PhoneNo, &user.RoleName, &user.DepartmentName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found for the given session ID")
		}
		return nil, err
	}

	// Handle sql.NullTime for FirstAccess and LastAccess
	user.FirstAccess = firstAccess.Time
	if !firstAccess.Valid {
		user.FirstAccess = time.Time{} // Zero value of time.Time
	}

	user.LastAccess = lastAccess.Time
	if !lastAccess.Valid {
		user.LastAccess = time.Time{} // Zero value of time.Time
	}

	return &user, nil
}

// GetPermissionID fetches the permission ID by its name from the database
func GetPermissionID(db *sql.DB, permissionName string) (int, error) {
	var permissionID int
	query := `SELECT permission_id FROM permissions WHERE permission_name = $1`
	err := db.QueryRow(query, permissionName).Scan(&permissionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("permission not found")
		}
		return 0, err
	}
	return permissionID, nil
}

func GetRoleIDByUserID(db *sql.DB, userID int) (int, error) {
	var roleID int
	query := `SELECT role_id FROM user_roles WHERE user_id = $1` // Adjust table name if needed.
	err := db.QueryRow(query, userID).Scan(&roleID)
	return roleID, err
}
