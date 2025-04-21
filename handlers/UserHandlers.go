package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/smtp"
	"nfa-app/models"
	"nfa-app/storage"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GetUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		if idStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required in the URL"})
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		user, err := getUserByID(db, id)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func getUserByID(db *sql.DB, id int) (models.User, error) {
	var user models.User
	var firstAccess, lastAccess sql.NullTime
	var profilePicture sql.NullString

	query := `
		SELECT 
			u.id, u.email, u.password, u.name,  
			u.created_at, u.updated_at, u.first_access, u.last_access, 
			u.profile_picture, u.address,
			 u.phone_no, u.role_id, u.department_id, r.role_name, d.department_name
		FROM 
			users u
		JOIN roles r ON u.role_id = r.role_id
		JOIN departments d ON u.department_id = d.department_id
		WHERE u.id = $1`
	err := db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.CreatedAt, &user.UpdatedAt, &firstAccess, &lastAccess, &profilePicture, &user.Address, &user.PhoneNo, &user.RoleID, &user.DepartmentID, &user.RoleName, &user.DepartmentName)

	if err != nil {
		return user, err
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

	// Handle sql.NullString for ProfilePicture
	if profilePicture.Valid {
		user.ProfilePic = profilePicture.String
	} else {
		user.ProfilePic = "" // Set to empty string if NULL
	}

	return user, nil
}

func GetUsersByRoleName(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleName := c.Param("role")
		if roleName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role name is required in the URL"})
			return
		}

		users, err := getUsersByRoleName(db, roleName) // Fetch multiple users
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users: " + err.Error()})
			return
		}

		// Return an empty array if no users found
		if len(users) == 0 {
			c.JSON(http.StatusOK, []interface{}{})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

func getUsersByRoleName(db *sql.DB, roleName string) ([]models.User, error) {
	query := `
		SELECT 
			u.id, u.email, u.password, u.name,  
			u.created_at, u.updated_at, u.first_access, u.last_access, 
			u.profile_picture, u.address, u.phone_no, 
			u.role_id, u.department_id, r.role_name, d.department_name
		FROM 
			users u
		JOIN roles r ON u.role_id = r.role_id
		JOIN departments d ON u.department_id = d.department_id
		WHERE r.role_name = $1`

	rows, err := db.Query(query, roleName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User
		var firstAccess, lastAccess sql.NullTime
		var profilePicture sql.NullString

		err := rows.Scan(
			&user.ID, &user.Email, &user.Password, &user.Name,
			&user.CreatedAt, &user.UpdatedAt, &firstAccess, &lastAccess,
			&profilePicture, &user.Address, &user.PhoneNo,
			&user.RoleID, &user.DepartmentID, &user.RoleName, &user.DepartmentName)

		if err != nil {
			return nil, err
		}

		// Handle sql.NullTime for FirstAccess and LastAccess
		if firstAccess.Valid {
			user.FirstAccess = firstAccess.Time
		} else {
			user.FirstAccess = time.Time{} // Zero value of time.Time
		}

		if lastAccess.Valid {
			user.LastAccess = lastAccess.Time
		} else {
			user.LastAccess = time.Time{} // Zero value of time.Time
		}

		// Handle sql.NullString for ProfilePicture
		if profilePicture.Valid {
			user.ProfilePic = profilePicture.String
		} else {
			user.ProfilePic = "" // Set to empty string if NULL
		}

		users = append(users, user)
	}

	return users, nil
}

func GetAllUsers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db := storage.GetDB()

		rows, err := db.Query(`
		SELECT 
			u.id, u.email, u.password, u.name,  
			u.created_at, u.updated_at, u.first_access, u.last_access, 
			u.profile_picture, u.address,
			 u.phone_no, u.role_id, u.department_id, r.role_name, d.department_name
		FROM 
			users u
		JOIN roles r ON u.role_id = r.role_id
		JOIN departments d ON u.department_id = d.department_id`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users: " + err.Error()})
			return
		}
		defer rows.Close()

		var users []models.User
		for rows.Next() {
			var user models.User
			var firstAccess, lastAccess sql.NullTime
			var profilePicture sql.NullString

			err := rows.Scan(
				&user.ID, &user.Email, &user.Password, &user.Name, &user.CreatedAt, &user.UpdatedAt, &firstAccess, &lastAccess, &profilePicture, &user.Address, &user.PhoneNo, &user.RoleID, &user.DepartmentID, &user.RoleName, &user.DepartmentName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan user: " + err.Error()})
				return
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

			// Handle sql.NullString for ProfilePicture
			if profilePicture.Valid {
				user.ProfilePic = profilePicture.String
			} else {
				user.ProfilePic = "" // Set to empty string if NULL
			}

			users = append(users, user)
		}

		if err = rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Row iteration error: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

func GetUserFromSession(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the session ID from the Authorization header
		sessionID := c.GetHeader("Authorization")

		// Ensure the session ID is not empty
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session ID in Authorization header"})
			return
		}

		fmt.Println("Session ID:", sessionID) // Debugging

		// Retrieve the UserID from the Session table
		var userID int
		sessionQuery := `SELECT user_id FROM session WHERE session_id = $1`
		err := db.QueryRow(sessionQuery, sessionID).Scan(&userID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		// Fetch the user details
		var user models.User
		var firstAccess, lastAccess sql.NullTime
		var profilePicture sql.NullString // To handle employee_id being NULL
		var departmentID sql.NullInt64

		userQuery := `
		SELECT 
			u.id, u.email, u.password, u.name,  
			u.created_at, u.updated_at, u.first_access, u.last_access, 
			u.profile_picture, u.address,
			u.phone_no, u.role_id, u.department_id, r.role_name
		FROM users u
		JOIN roles r ON u.role_id = r.role_id
		WHERE u.id = $1`

		err = db.QueryRow(userQuery, userID).Scan(
			&user.ID, &user.Email, &user.Password, &user.Name, &user.CreatedAt, &user.UpdatedAt, &firstAccess, &lastAccess, &profilePicture, &user.Address, &user.PhoneNo, &user.RoleID, &departmentID, &user.RoleName)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		fmt.Println("Retrieved User ID:", userID) // Debugging

		// Handle sql.NullTime for FirstAccess and LastAccess
		user.FirstAccess = firstAccess.Time
		if !firstAccess.Valid {
			user.FirstAccess = time.Time{} // Zero value of time.Time
		}

		user.LastAccess = lastAccess.Time
		if !lastAccess.Valid {
			user.LastAccess = time.Time{} // Zero value of time.Time
		}

		// Handle sql.NullString for ProfilePicture
		if profilePicture.Valid {
			user.ProfilePic = profilePicture.String
		} else {
			user.ProfilePic = "" // Set to empty string if NULL
		}

		// Return the user details
		c.JSON(http.StatusOK, user)
	}
}

func CreateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate session ID
		sessionID := c.GetHeader("Authorization")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session-id header is required"})
			return
		}

		// Fetch user_id from the session table
		var userID int
		err := db.QueryRow("SELECT user_id FROM session WHERE session_id = $1", sessionID).Scan(&userID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session. Session ID not found."})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching session: " + err.Error()})
			}
			return
		}

		// Fetch the user's role_id
		var roleID int
		err = db.QueryRow("SELECT role_id FROM users WHERE id = $1", userID).Scan(&roleID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		var currentUserRole string
		err = db.QueryRow("SELECT role_name FROM roles where role_id = $1", roleID).Scan(&currentUserRole)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch role name: " + err.Error()})
			return
		}

		// Parse request body for the new user
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if user already exists (by email or employee ID)
		var existingUserID int
		err = db.QueryRow(
			"SELECT id FROM users WHERE email = $1",
			user.Email,
		).Scan(&existingUserID)

		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}

		if err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
			return
		}

		// Set timestamps
		user.CreatedAt = time.Now()
		user.UpdatedAt = user.CreatedAt
		user.FirstAccess = time.Now()
		user.LastAccess = user.CreatedAt

		// Insert new user into the database
		sqlStatement := `
			INSERT INTO users (email, password, name, created_at, updated_at, first_access, last_access, profile_picture, address, phone_no, role_id, department_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING id`

		err = db.QueryRow(
			sqlStatement,
			user.Email, user.Password, user.Name,
			time.Now(), time.Now(), time.Now(), time.Now(), user.ProfilePic,
			user.Address, user.PhoneNo, user.RoleID, user.DepartmentID,
		).Scan(&user.ID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
			return
		}

		var roleName string
		err = db.QueryRow("SELECT role_name FROM roles WHERE role_id = $1", user.RoleID).Scan(&roleName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch role name: " + err.Error()})
			return
		}

		// Send confirmation email (optional, handle errors gracefully)
		if err = SendEmail(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User created, but failed to send email: " + err.Error()})
			return
		}

		// // Create notification for the creator
		// creatorMessage := fmt.Sprintf("You have successfully created a member with role '%s' in the project.", roleName)
		// _, err = CreateNotification(db, userID, creatorMessage)
		// if err != nil {
		// 	log.Println("Failed to create notification for creator:", err)
		// }

		// // Create notification for the new user
		// newUserMessage := fmt.Sprintf("You are assigned as '%s' in the project by '%s'.", roleName, currentUserRole)
		// _, err = CreateNotification(db, user.ID, newUserMessage)
		// if err != nil {
		// 	log.Println("Failed to create notification for new user:", err)
		// }

		// Success response
		c.JSON(http.StatusCreated, gin.H{
			"message": "User created successfully",
			"user_id": user.ID,
		})
	}
}

func UpdateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract session ID from Authorization header
		sessionID := c.GetHeader("Authorization")
		if sessionID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing session ID in Authorization header"})
			return
		}

		// Fetch the role of the current user (from session)
		var currentUserRole string
		err := db.QueryRow("SELECT u.role FROM session s JOIN users u ON s.user_id = u.id WHERE s.session_id = $1", sessionID).Scan(&currentUserRole)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session or user not found"})
			return
		}

		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse user ID from the URL
		userIDStr := c.Param("id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Check if the user exists
		var existingUserID int
		err = db.QueryRow("SELECT id FROM users WHERE id = $1", userID).Scan(&existingUserID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var updates []string
		var fields []interface{}
		placeholderIndex := 1

		// Role can only be updated by Super Admin
		if user.RoleID != 0 {
			if currentUserRole != "Super Admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "Only Super Admin can modify user roles"})
				return
			}
			updates = append(updates, fmt.Sprintf("role = $%d", placeholderIndex))
			fields = append(fields, user.RoleID)
			placeholderIndex++
		}

		if user.Email != "" {
			updates = append(updates, fmt.Sprintf("email = $%d", placeholderIndex))
			fields = append(fields, user.Email)
			placeholderIndex++
		}
		if user.Password != "" {
			updates = append(updates, fmt.Sprintf("password = $%d", placeholderIndex))
			fields = append(fields, user.Password)
			placeholderIndex++
		}
		if user.ProfilePic != "" {
			updates = append(updates, fmt.Sprintf("profile_picture = $%d", placeholderIndex))
			fields = append(fields, user.ProfilePic)
			placeholderIndex++
		}
		if user.PhoneNo != "" {
			updates = append(updates, fmt.Sprintf("phone_no = $%d", placeholderIndex))
			fields = append(fields, user.PhoneNo)
			placeholderIndex++
		}
		if user.Address != "" {
			updates = append(updates, fmt.Sprintf("address = $%d", placeholderIndex))
			fields = append(fields, user.Address)
			placeholderIndex++
		}

		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
			return
		}

		sqlStatement := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(updates, ", "), placeholderIndex)
		fields = append(fields, userID)

		_, err = db.Exec(sqlStatement, fields...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
	}
}

func SendEmail(user models.User) error {
	auth := smtp.PlainAuth(
		"",
		"om.s@blueinvent.com",
		"gloycbfukxdyeczj",
		"smtp.gmail.com",
	)

	from := "vasug7409@gmail.com"
	to := []string{user.Email}
	subject := "Welcome to Our Platform!"

	role := "User"
	body := fmt.Sprintf("Hello %s,\n\nYour account has been created successfully.\n\nHere are your credentials:\n\nPassword: %s\nRole: %s\n\nPlease change your password after logging in for the first time.\n\nBest Regards,\nYour Company",
		user.Name, user.Password, role)

	msg := []byte("From: " + from + "\r\n" +
		"To: " + user.Email + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		body + "\r\n")

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		from,
		to,
		msg,
	)

	return err
}

func DeleteUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		err = deleteUserFromDB(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("User with ID %d successfully deleted", id)})
	}
}

func deleteUserFromDB(id int) error {
	db := storage.GetDB()
	query := "DELETE FROM users WHERE id = $1"
	_, err := db.Exec(query, id)
	return err
}
