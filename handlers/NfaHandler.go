package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"nfa-app/models"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func UploadNFAFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := c.Request.ParseMultipartForm(0)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
			return
		}

		// Get the uploaded file
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
			return
		}
		defer file.Close()

		// Define upload path (modify as needed)
		uploadDir := "./uploads/nfa/"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}

		// Save file
		filePath := uploadDir + header.Filename
		outFile, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
			return
		}

		// Return file path
		c.JSON(http.StatusOK, gin.H{
			"message":   "File uploaded successfully",
			"file_path": "/uploads/nfa/" + header.Filename,
		})
	}
}

func GetNFAByProjectID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		fetchNFAByField(db, c, "SELECT * FROM nfa WHERE project_id = $1", c.Param("project_id"))
	}
}

func GetNFAByDepartmentID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		fetchNFAByField(db, c, "SELECT * FROM nfa WHERE department_id = $1", c.Param("department_id"))
	}
}

func GetNFAByAreaID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		fetchNFAByField(db, c, "SELECT * FROM nfa WHERE area_id = $1", c.Param("area_id"))
	}
}

func GetNFAByTowerID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		fetchNFAByField(db, c, "SELECT * FROM nfa WHERE tower_id = $1", c.Param("tower_id"))
	}
}

func GetNFAByPriority(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		fetchNFAByField(db, c, "SELECT * FROM nfa WHERE priority = $1", c.Param("priority"))
	}
}

func GetNFAByRecommender(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session ID from header
		sessionID := c.GetHeader("Authorization")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "session-id header is required",
				"details": "Authorization header is missing"})
			return
		}

		// Get user_id from session
		var recommenderID int
		err := db.QueryRow("SELECT user_id FROM session WHERE session_id = $1", sessionID).Scan(&recommenderID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Invalid session",
					"details": "No session found with provided session ID"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Database error",
					"details": fmt.Sprintf("Error fetching session: %v", err)})
			}
			return
		}

		// Query to fetch NFAs with all related information
		query := `
            SELECT 
                n.nfa_id, n.project_id, n.tower_id, n.area_id, n.department_id, 
                n.status,
                COALESCE(n.priority, '') as priority, 
                COALESCE(n.subject, '') as subject, 
                COALESCE(n.description, '') as description, 
                COALESCE(n.reference, '') as reference, 
                n.recommender, n.last_recommender, n.initiator_id,
                COALESCE(initiator.name, '') as initiator_name,
                COALESCE(recommender.name, '') as recommender_name,
                COALESCE(last_recommender.name, '') as last_recommender_name,
                COALESCE(p.project_name, '') as project_name,
                COALESCE(t.tower_name, '') as tower_name,
                COALESCE(a.area_name, '') as area_name,
                COALESCE(d.department_name, '') as department_name
            FROM nfa n
            LEFT JOIN users initiator ON n.initiator_id = initiator.id
            LEFT JOIN users recommender ON n.recommender = recommender.id
            LEFT JOIN users last_recommender ON n.last_recommender = last_recommender.id
            LEFT JOIN projects p ON n.project_id = p.project_id
            LEFT JOIN towers t ON n.tower_id = t.tower_id
            LEFT JOIN areas a ON n.area_id = a.area_id
            LEFT JOIN departments d ON n.department_id = d.department_id
            WHERE n.recommender = $1
            ORDER BY n.nfa_id DESC`

		rows, err := db.Query(query, recommenderID)
		if err != nil {
			log.Printf("Database query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":       "Database query failed",
				"details":     fmt.Sprintf("Failed to fetch NFAs: %v", err),
				"query_error": err.Error()})
			return
		}
		defer rows.Close()

		type NFAWithNames struct {
			models.NFA
			InitiatorName       string `json:"initiator_name"`
			RecommenderName     string `json:"recommender_name"`
			LastRecommenderName string `json:"last_recommender_name"`
			ProjectName         string `json:"project_name"`
			TowerName           string `json:"tower_name"`
			AreaName            string `json:"area_name"`
			DepartmentName      string `json:"department_name"`
		}

		var nfas []NFAWithNames

		for rows.Next() {
			var nfa NFAWithNames

			err := rows.Scan(
				&nfa.NFAID, &nfa.ProjectID, &nfa.TowerID, &nfa.AreaID,
				&nfa.DepartmentID, &nfa.Status, &nfa.Priority, &nfa.Subject,
				&nfa.Description, &nfa.Reference, &nfa.Recommender,
				&nfa.LastRecommender, &nfa.InitiatorID, &nfa.InitiatorName,
				&nfa.RecommenderName, &nfa.LastRecommenderName, &nfa.ProjectName,
				&nfa.TowerName, &nfa.AreaName, &nfa.DepartmentName,
			)
			if err != nil {
				log.Printf("Row scan error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":      "Data scan failed",
					"details":    fmt.Sprintf("Failed to scan NFA data: %v", err),
					"scan_error": err.Error()})
				return
			}

			if err := fetchApprovalsAndFiles(db, &nfa.NFA); err != nil {
				log.Printf("Approvals and files fetch error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":       "Related data fetch failed",
					"details":     fmt.Sprintf("Failed to fetch approvals and files: %v", err),
					"fetch_error": err.Error()})
				return
			}

			nfas = append(nfas, nfa)
		}

		if err = rows.Err(); err != nil {
			log.Printf("Row iteration error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":           "Data iteration failed",
				"details":         fmt.Sprintf("Error during row iteration: %v", err),
				"iteration_error": err.Error()})
			return
		}

		if len(nfas) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"message":        "No NFAs found for this recommender",
				"nfas":           []NFAWithNames{},
				"recommender_id": recommenderID})
			return
		}

		c.JSON(http.StatusOK, nfas)
	}
}

func GetAllNFA(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		fetchNFAByField(db, c, "SELECT * FROM nfa")
	}
}

func fetchNFAByField(db *sql.DB, c *gin.Context, query string, args ...interface{}) {
	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch NFAs"})
		return
	}
	defer rows.Close()

	var nfas []models.NFA

	for rows.Next() {
		var nfa models.NFA
		if err := rows.Scan(&nfa.NFAID, &nfa.ProjectID, &nfa.TowerID, &nfa.AreaID, &nfa.DepartmentID, &nfa.Priority, &nfa.Subject, &nfa.Description, &nfa.Reference, &nfa.Recommender, &nfa.LastRecommender); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan NFAs"})
			return
		}

		if err := fetchApprovalsAndFiles(db, &nfa); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		nfas = append(nfas, nfa)
	}

	c.JSON(http.StatusOK, gin.H{"nfas": nfas})
}

func UpdateNFA(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get NFA ID from the request parameters
		nfaID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid NFA ID"})
			return
		}

		// Define the request structure
		var request struct {
			ProjectID       int                      `json:"project_id"`
			TowerID         int                      `json:"tower_id"`
			AreaID          int                      `json:"area_id"`
			DepartmentID    int                      `json:"department_id"`
			Priority        string                   `json:"priority"`
			Subject         string                   `json:"subject"`
			Description     string                   `json:"description"`
			Reference       string                   `json:"reference"`
			Recommender     int                      `json:"recommender"`
			LastRecommender int                      `json:"last_recommender"`
			ApprovalList    []models.NFAApprovalList `json:"approval_list"`
			Files           []models.NFAFile         `json:"files"`
		}

		// Bind the JSON request
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
			return
		}

		// Update the NFA record
		updateQuery := `UPDATE nfa SET 
            project_id = $1, tower_id = $2, area_id = $3, department_id = $4, 
            priority = $5, subject = $6, description = $7, reference = $8, 
            recommender = $9, last_recommender = $10
            WHERE nfa_id = $11`

		_, err = db.Exec(updateQuery, request.ProjectID, request.TowerID, request.AreaID, request.DepartmentID,
			request.Priority, request.Subject, request.Description, request.Reference, request.Recommender,
			request.LastRecommender, nfaID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update NFA"})
			return
		}

		// Delete old approvals and insert updated approval list
		_, err = db.Exec("DELETE FROM nfa_approval_list WHERE nfa_id = $1", nfaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear old approval list"})
			return
		}

		for i := range request.ApprovalList {
			request.ApprovalList[i].NFAID = nfaID
			approvalQuery := `INSERT INTO nfa_approval_list (nfa_id, approver_id, "order_value") VALUES ($1, $2, $3) RETURNING id`
			err := db.QueryRow(approvalQuery, request.ApprovalList[i].NFAID, request.ApprovalList[i].ApproverID, request.ApprovalList[i].Order).Scan(&request.ApprovalList[i].ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert approval list"})
				return
			}
		}

		// Delete old files and insert updated files
		_, err = db.Exec("DELETE FROM nfa_files WHERE nfa_id = $1", nfaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear old file records"})
			return
		}

		for i := range request.Files {
			request.Files[i].NFAID = nfaID
			fileQuery := `INSERT INTO nfa_files (nfa_id, file_name, file_path) VALUES ($1, $2, $3) RETURNING id`
			err := db.QueryRow(fileQuery, request.Files[i].NFAID, request.Files[i].Name, request.Files[i].Path).Scan(&request.Files[i].ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert file records"})
				return
			}
		}

		// Success response
		c.JSON(http.StatusOK, gin.H{
			"message":       "NFA updated successfully",
			"nfa_id":        nfaID,
			"approval_list": request.ApprovalList,
			"files":         request.Files,
		})
	}
}

func DeleteNFA(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get NFA ID from request parameters
		nfaID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid NFA ID"})
			return
		}

		// Begin a transaction to ensure atomicity
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}

		// Delete approval list associated with the NFA
		_, err = tx.Exec("DELETE FROM nfa_approval_list WHERE nfa_id = $1", nfaID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete approval list"})
			return
		}

		// Delete files associated with the NFA
		_, err = tx.Exec("DELETE FROM nfa_files WHERE nfa_id = $1", nfaID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file records"})
			return
		}

		// Delete the NFA record itself
		_, err = tx.Exec("DELETE FROM nfa WHERE nfa_id = $1", nfaID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete NFA"})
			return
		}

		// Commit the transaction
		err = tx.Commit()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		// Success response
		c.JSON(http.StatusOK, gin.H{
			"message": "NFA deleted successfully",
			"nfa_id":  nfaID,
		})
	}
}

func GetNFAApprovalList(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Convert nfa_id from string to integer
		nfaID, err := strconv.Atoi(c.Param("nfa_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid nfa_id"})
			return
		}

		// Modified query to include approver names from users table
		query := `
            SELECT 
                al.id, 
                al.nfa_id, 
                al.approver_id, 
                al.order_value, 
                al.status, 
                COALESCE(al.comments, '') as comments,
                COALESCE(u.name, '') as approver_name
            FROM nfa_approval_list al
            LEFT JOIN users u ON al.approver_id = u.id
            WHERE al.nfa_id = $1 AND al.status != 'Complete'
            ORDER BY al.order_value ASC`

		rows, err := db.Query(query, nfaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
			return
		}
		defer rows.Close()

		type ApprovalWithName struct {
			models.NFAApprovalList
			ApproverName string `json:"approver_name"`
		}

		var approvals []ApprovalWithName
		for rows.Next() {
			var approval ApprovalWithName
			if err := rows.Scan(
				&approval.ID,
				&approval.NFAID,
				&approval.ApproverID,
				&approval.Order,
				&approval.Status,
				&approval.Comments,
				&approval.ApproverName,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning data"})
				return
			}
			approvals = append(approvals, approval)
		}

		if len(approvals) == 0 {
			c.JSON(http.StatusOK, []ApprovalWithName{})
			return
		}

		c.JSON(http.StatusOK, approvals)
	}
}
func AddApprover(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newApprover models.NFAApprovalList

		// Bind JSON request body
		if err := c.ShouldBindJSON(&newApprover); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Validate required fields
		if newApprover.NFAID == 0 || newApprover.ApproverID == 0 || newApprover.Order == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
			return
		}

		tx, err := db.Begin() // Start a transaction
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}

		// Step 1: Check if there's a "Pending" approver
		var pendingOrder int
		err = tx.QueryRow(`SELECT order_value FROM nfa_approval_list WHERE nfa_id = $1 AND status = 'Pending'`, newApprover.NFAID).Scan(&pendingOrder)
		if err != nil && err != sql.ErrNoRows {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check pending approver"})
			return
		}

		// Step 2: Shift existing approvers
		updateQuery := `UPDATE nfa_approval_list 
		                SET order_value = order_value + 1 
		                WHERE nfa_id = $1 AND order_value >= $2`
		_, err = tx.Exec(updateQuery, newApprover.NFAID, newApprover.Order)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
			return
		}

		// Step 3: Determine the status of the new approver
		status := "Waiting" // Default for new approvers
		if pendingOrder == 0 && newApprover.Order == 1 {
			status = "Pending" // If no pending approver exists, first one is Pending
		}

		// Step 4: Insert the new approver
		insertQuery := `INSERT INTO nfa_approval_list (nfa_id, approver_id, order_value, status, comments) 
		                VALUES ($1, $2, $3, $4, $5)`
		_, err = tx.Exec(insertQuery, newApprover.NFAID, newApprover.ApproverID, newApprover.Order, status, newApprover.Comments)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert new approver"})
			return
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Approver added successfully"})
	}
}

func ApproveNFA(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		nfaID, err := strconv.Atoi(c.Param("nfa_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid nfa_id"})
			return
		}

		approverID, err := strconv.Atoi(c.Param("approver_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid approver_id"})
			return
		}

		// Start transaction
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}

		// Step 1: Check if the approver is the currently active one (`Pending`)
		var currentOrder int
		checkApprover := `SELECT order_value FROM nfa_approval_list 
		                  WHERE nfa_id = $1 AND approver_id = $2 AND status = 'Pending'`
		err = tx.QueryRow(checkApprover, nfaID, approverID).Scan(&currentOrder)
		if err == sql.ErrNoRows {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Approver is not next in line or already approved"})
			return
		} else if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify approver status"})
			return
		}

		// Step 2: Mark the current approver as "Complete"
		updateCurrent := `UPDATE nfa_approval_list 
		                  SET status = 'Complete' 
		                  WHERE nfa_id = $1 AND approver_id = $2`
		_, err = tx.Exec(updateCurrent, nfaID, approverID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update current approver"})
			return
		}

		// Step 3: Find and activate the next approver in order
		var nextApproverID int
		findNext := `SELECT approver_id FROM nfa_approval_list 
		             WHERE nfa_id = $1 AND order_value > $2 AND status = 'Pending' 
		             ORDER BY order_value ASC LIMIT 1`
		err = tx.QueryRow(findNext, nfaID, currentOrder).Scan(&nextApproverID)

		if err == sql.ErrNoRows {
			// No next approver; final approval complete
			tx.Commit()
			c.JSON(http.StatusOK, gin.H{"message": "Final approval completed. No more approvers remaining."})
			return
		} else if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch next approver"})
			return
		}

		// Step 4: Activate the next approver (set status to "Pending")
		updateNext := `UPDATE nfa_approval_list SET status = 'Pending' WHERE nfa_id = $1 AND approver_id = $2`
		_, err = tx.Exec(updateNext, nfaID, nextApproverID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate next approver"})
			return
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Approval completed. Next approver activated."})
	}
}

func RemoveApprover(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse approver ID and NFA ID from request parameters
		approverID, err := strconv.Atoi(c.Param("approver_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid approver_id"})
			return
		}

		nfaID, err := strconv.Atoi(c.Param("nfa_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid nfa_id"})
			return
		}

		tx, err := db.Begin() // Start a transaction
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}

		// Step 1: Get the order of the approver to be deleted
		var deletedOrder int
		err = tx.QueryRow(`SELECT order_value FROM nfa_approval_list WHERE nfa_id = $1 AND approver_id = $2`, nfaID, approverID).Scan(&deletedOrder)
		if err == sql.ErrNoRows {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Approver not found"})
			return
		} else if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get approver order"})
			return
		}

		// Step 2: Delete the approver
		deleteQuery := `DELETE FROM nfa_approval_list WHERE nfa_id = $1 AND approver_id = $2`
		_, err = tx.Exec(deleteQuery, nfaID, approverID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete approver"})
			return
		}

		// Step 3: Shift orders for remaining approvers
		updateQuery := `UPDATE nfa_approval_list 
		                SET order_value = order_value - 1 
		                WHERE nfa_id = $1 AND order_value > $2`
		_, err = tx.Exec(updateQuery, nfaID, deletedOrder)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to shift order of remaining approvers"})
			return
		}

		// Step 4: Check if a "Pending" approver exists
		var pendingExists bool
		err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM nfa_approval_list WHERE nfa_id = $1 AND status = 'Pending')`, nfaID).Scan(&pendingExists)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check pending approver"})
			return
		}

		// Step 5: If no "Pending" approver exists, set the next in line to "Pending"
		if !pendingExists {
			updateNextQuery := `UPDATE nfa_approval_list 
			                    SET status = 'Pending' 
			                    WHERE nfa_id = $1 AND order_value = (SELECT MIN(order_value) FROM nfa_approval_list WHERE nfa_id = $1 AND status = 'Waiting')`
			_, err = tx.Exec(updateNextQuery, nfaID)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update next approver status"})
				return
			}
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Approver removed successfully"})
	}
}

func GetNFAByInitiator(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Session ID validation (unchanged)
		sessionID := c.GetHeader("Authorization")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "session-id header is required",
				"details": "Authorization header is missing"})
			return
		}

		// Get user_id from session (unchanged)
		var initiatorID int
		err := db.QueryRow("SELECT user_id FROM session WHERE session_id = $1", sessionID).Scan(&initiatorID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Invalid session",
					"details": "No session found with provided session ID"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Database error",
					"details": fmt.Sprintf("Error fetching session: %v", err)})
			}
			return
		}

		// Modified query to include all required names
		query := `
            SELECT 
                n.nfa_id, n.project_id, n.tower_id, n.area_id, n.department_id, 
                n.status,
                COALESCE(n.priority, '') as priority, 
                COALESCE(n.subject, '') as subject, 
                COALESCE(n.description, '') as description, 
                COALESCE(n.reference, '') as reference, 
                n.recommender, n.last_recommender, n.initiator_id,
                COALESCE(initiator.name, '') as initiator_name,
                COALESCE(recommender.name, '') as recommender_name,
                COALESCE(last_recommender.name, '') as last_recommender_name,
                COALESCE(p.project_name, '') as project_name,
                COALESCE(t.tower_name, '') as tower_name,
                COALESCE(a.area_name, '') as area_name,
                COALESCE(d.department_name, '') as department_name
            FROM nfa n
            LEFT JOIN users initiator ON n.initiator_id = initiator.id
            LEFT JOIN users recommender ON n.recommender = recommender.id
            LEFT JOIN users last_recommender ON n.last_recommender = last_recommender.id
            LEFT JOIN projects p ON n.project_id = p.project_id
            LEFT JOIN towers t ON n.tower_id = t.tower_id
            LEFT JOIN areas a ON n.area_id = a.area_id
            LEFT JOIN departments d ON n.department_id = d.department_id
            WHERE n.initiator_id = $1
            ORDER BY n.nfa_id DESC`

		rows, err := db.Query(query, initiatorID)
		if err != nil {
			log.Printf("Database query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":       "Database query failed",
				"details":     fmt.Sprintf("Failed to fetch NFAs: %v", err),
				"query_error": err.Error()})
			return
		}
		defer rows.Close()

		// Define the complete structure once
		type NFAWithNames struct {
			models.NFA
			InitiatorName       string `json:"initiator_name"`
			RecommenderName     string `json:"recommender_name"`
			LastRecommenderName string `json:"last_recommender_name"`
			ProjectName         string `json:"project_name"`
			TowerName           string `json:"tower_name"`
			AreaName            string `json:"area_name"`
			DepartmentName      string `json:"department_name"`
		}

		var nfas []NFAWithNames

		for rows.Next() {
			var nfa NFAWithNames

			err := rows.Scan(
				&nfa.NFAID, &nfa.ProjectID, &nfa.TowerID, &nfa.AreaID,
				&nfa.DepartmentID, &nfa.Status, &nfa.Priority, &nfa.Subject,
				&nfa.Description, &nfa.Reference, &nfa.Recommender,
				&nfa.LastRecommender, &nfa.InitiatorID, &nfa.InitiatorName,
				&nfa.RecommenderName, &nfa.LastRecommenderName, &nfa.ProjectName,
				&nfa.TowerName, &nfa.AreaName, &nfa.DepartmentName,
			)
			if err != nil {
				log.Printf("Row scan error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":      "Data scan failed",
					"details":    fmt.Sprintf("Failed to scan NFA data: %v", err),
					"scan_error": err.Error()})
				return
			}

			if err := fetchApprovalsAndFiles(db, &nfa.NFA); err != nil {
				log.Printf("Approvals and files fetch error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":       "Related data fetch failed",
					"details":     fmt.Sprintf("Failed to fetch approvals and files: %v", err),
					"fetch_error": err.Error()})
				return
			}

			nfas = append(nfas, nfa)
		}

		if err = rows.Err(); err != nil {
			log.Printf("Row iteration error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":           "Data iteration failed",
				"details":         fmt.Sprintf("Error during row iteration: %v", err),
				"iteration_error": err.Error()})
			return
		}

		if len(nfas) == 0 {
			c.JSON(http.StatusNoContent, gin.H{
				"message": "No pending approvals found",
			})
			return
		}

		c.JSON(http.StatusOK, nfas)
	}
}

// Add this new function to fetch approvals with approver names
func fetchApprovalsAndFiles(db *sql.DB, nfa *models.NFA) error {
	// Fetch approval list with approver names, using COALESCE with proper timestamp handling
	approvalQuery := `
        SELECT 
            al.id, 
            al.nfa_id, 
            al.approver_id, 
            al.order_value,
            COALESCE(al.status, 'Waiting') as status,
            COALESCE(al.comments, '') as comments,
            COALESCE(u.name, '') as approver_name,
            al.started_at,
            al.updated_at
        FROM nfa_approval_list al
        LEFT JOIN users u ON al.approver_id = u.id
        WHERE al.nfa_id = $1
        ORDER BY al.order_value`

	approvalRows, err := db.Query(approvalQuery, nfa.NFAID)
	if err != nil {
		return fmt.Errorf("query error: %v", err)
	}
	defer approvalRows.Close()

	var approvals []struct {
		models.NFAApprovalList
		ApproverName string `json:"approver_name"`
	}

	for approvalRows.Next() {
		var approval struct {
			models.NFAApprovalList
			ApproverName string `json:"approver_name"`
		}

		// Use sql.NullTime for nullable timestamp fields
		var startedAt, updatedAt sql.NullTime

		if err := approvalRows.Scan(
			&approval.ID,
			&approval.NFAID,
			&approval.ApproverID,
			&approval.Order,
			&approval.Status,
			&approval.Comments,
			&approval.ApproverName,
			&startedAt,
			&updatedAt,
		); err != nil {
			return fmt.Errorf("scan error: %v", err)
		}

		// Convert sql.NullTime to time.Time
		if startedAt.Valid {
			approval.StartedDate = startedAt.Time
		}
		if updatedAt.Valid {
			approval.CompletedDate = updatedAt.Time
		}

		approvals = append(approvals, approval)
	}

	// Fetch files with COALESCE for nullable fields
	filesQuery := `
        SELECT 
            id, 
            nfa_id, 
            COALESCE(file_name, '') as file_name, 
            COALESCE(file_path, '') as file_path 
        FROM nfa_files 
        WHERE nfa_id = $1`

	fileRows, err := db.Query(filesQuery, nfa.NFAID)
	if err != nil {
		return fmt.Errorf("file query error: %v", err)
	}
	defer fileRows.Close()

	var files []models.NFAFile
	for fileRows.Next() {
		var file models.NFAFile
		if err := fileRows.Scan(&file.ID, &file.NFAID, &file.Name, &file.Path); err != nil {
			return fmt.Errorf("file scan error: %v", err)
		}
		files = append(files, file)
	}

	nfa.Files = files
	nfa.Approvals = make([]models.NFAApprovalList, len(approvals))
	for i, approval := range approvals {
		nfa.Approvals[i] = approval.NFAApprovalList
		nfa.Approvals[i].ApproverName = approval.ApproverName
	}

	return nil
}

func ApproveOrRejectNFA(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session ID from header
		sessionID := c.GetHeader("Authorization")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session-id header is required"})
			return
		}

		// Get user_id from session
		var userID int
		err := db.QueryRow("SELECT user_id FROM session WHERE session_id = $1", sessionID).Scan(&userID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching session: " + err.Error()})
			}
			return
		}

		// Request structure
		var request struct {
			NFAID   int    `json:"nfa_id"`
			Action  string `json:"action"` // "approve" or "reject"
			Comment string `json:"comment"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Validate request data
		if request.NFAID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid NFA ID"})
			return
		}

		if request.Action != "approve" && request.Action != "reject" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Action must be either 'approve' or 'reject'"})
			return
		}

		// Start transaction
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback()

		// First, check if the NFA exists
		var nfaExists bool
		err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM nfa WHERE nfa_id = $1)`, request.NFAID).Scan(&nfaExists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if !nfaExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "NFA not found"})
			return
		}

		// Check if user is a recommender for this NFA
		var isRecommender bool
		err = tx.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM nfa 
                WHERE nfa_id = $1 
                AND recommender = $2 
                AND status = 'Pending'
            )`, request.NFAID, userID).Scan(&isRecommender)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// If not recommender, check if user is current approver
		var isApprover bool
		var currentOrder int
		if !isRecommender {
			err = tx.QueryRow(`
                SELECT order_value 
                FROM nfa_approval_list 
                WHERE nfa_id = $1 
                AND approver_id = $2 
                AND started_at IS NOT NULL 
                AND updated_at IS NULL
                AND status = 'Pending'`,
				request.NFAID, userID).Scan(&currentOrder)

			isApprover = err != sql.ErrNoRows
			if err != nil && err != sql.ErrNoRows {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}
		}

		// If neither recommender nor current approver, return error
		if !isRecommender && !isApprover {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Not authorized to perform this action",
				"details": "You must be either the current recommender or the current approver"})
			return
		}

		var actionErr error
		// In ApproveOrRejectNFA function, update the function call:
		if isRecommender {
			actionErr = processRecommenderAction(tx, request.NFAID, request.Action, request.Comment)
		} else {
			actionErr = processApproverAction(tx, request.NFAID, currentOrder, request.Action, request.Comment, userID)
		}

		if actionErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to process action",
				"details": actionErr.Error()})
			return
		}

		if err = tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to commit transaction",
				"details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Action processed successfully",
			"role":    map[bool]string{true: "recommender", false: "approver"}[isRecommender],
			"nfa_id":  request.NFAID,
			"action":  request.Action,
		})
	}
}

func processRecommenderAction(tx *sql.Tx, nfaID int, action, comment string) error {
	if action == "approve" {
		// First check if there are any approvers
		var hasApprovers bool
		err := tx.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM nfa_approval_list 
                WHERE nfa_id = $1 AND order_value = 1
            )`, nfaID).Scan(&hasApprovers)
		if err != nil {
			return fmt.Errorf("failed to check approvers: %v", err)
		}

		if !hasApprovers {
			// If no approvers, mark NFA as completed
			_, err = tx.Exec(`
                UPDATE nfa 
                SET status = 'Completed' 
                WHERE nfa_id = $1`,
				nfaID)
			if err != nil {
				return fmt.Errorf("failed to update NFA status: %v", err)
			}
		} else {
			// First update NFA status
			_, err = tx.Exec(`
                UPDATE nfa 
                SET status = 'Initiated' 
                WHERE nfa_id = $1`,
				nfaID)
			if err != nil {
				return fmt.Errorf("failed to update NFA status: %v", err)
			}

			// Then update approval list
			_, err = tx.Exec(`
                UPDATE nfa_approval_list 
                SET started_at = CURRENT_TIMESTAMP,
                    status = 'Pending'
                WHERE nfa_id = $1 AND order_value = 1`,
				nfaID)
			if err != nil {
				return fmt.Errorf("failed to update approval status: %v", err)
			}
		}

	} else if action == "reject" {
		_, err := tx.Exec(`
            UPDATE nfa 
            SET status = 'Rejected',
                comments = NULLIF($1, '')
            WHERE nfa_id = $2`,
			comment, nfaID)
		if err != nil {
			return fmt.Errorf("failed to update status: %v", err)
		}
	}
	return nil
}

func processApproverAction(tx *sql.Tx, nfaID int, currentOrder int, action, comment string, userID int) error {
	if action == "approve" {
		// Update current approver's status
		_, err := tx.Exec(`
            UPDATE nfa_approval_list 
            SET status = 'Approved',
                comments = NULLIF($1, ''),
                updated_at = CURRENT_TIMESTAMP 
            WHERE nfa_id = $2 
            AND approver_id = $3 
            AND order_value = $4`,
			comment, nfaID, userID, currentOrder)
		if err != nil {
			return fmt.Errorf("failed to update approval: %v", err)
		}

		// Check if there's a next approver
		var nextApproverExists bool
		err = tx.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM nfa_approval_list 
                WHERE nfa_id = $1 AND order_value = $2
            )`, nfaID, currentOrder+1).Scan(&nextApproverExists)
		if err != nil {
			return fmt.Errorf("failed to check next approver: %v", err)
		}

		if nextApproverExists {
			// Start next approver's timer
			_, err = tx.Exec(`
                UPDATE nfa_approval_list 
                SET started_at = CURRENT_TIMESTAMP,
                    status = 'Pending'
                WHERE nfa_id = $1 AND order_value = $2`,
				nfaID, currentOrder+1)
			if err != nil {
				return fmt.Errorf("failed to start next approval: %v", err)
			}
		} else {
			// If no next approver, mark NFA as completed
			_, err = tx.Exec(`
                UPDATE nfa 
                SET status = 'Completed' 
                WHERE nfa_id = $1`,
				nfaID)
			if err != nil {
				return fmt.Errorf("failed to complete NFA: %v", err)
			}
		}

	} else if action == "reject" {
		// Update approval status
		_, err := tx.Exec(`
            UPDATE nfa_approval_list 
            SET status = 'Rejected',
                comments = NULLIF($1, ''),
                updated_at = CURRENT_TIMESTAMP 
            WHERE nfa_id = $2 
            AND approver_id = $3 
            AND order_value = $4`,
			comment, nfaID, userID, currentOrder)
		if err != nil {
			return fmt.Errorf("failed to update approval status: %v", err)
		}

		// Update NFA status
		_, err = tx.Exec(`
            UPDATE nfa 
            SET status = 'Rejected_By_Approver' 
            WHERE nfa_id = $1`,
			nfaID)
		if err != nil {
			return fmt.Errorf("failed to update NFA status: %v", err)
		}
	}
	return nil
}

func CreateNFA(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Get session ID from header
		sessionID := c.GetHeader("Authorization")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session-id header is required"})
			return
		}

		// Get user_id from session
		var initiatorID int
		err := db.QueryRow("SELECT user_id FROM session WHERE session_id = $1", sessionID).Scan(&initiatorID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching session: " + err.Error()})
			}
			return
		}
		// Define the request structure
		var request struct {
			ProjectID       int                      `json:"project_id"`
			TowerID         int                      `json:"tower_id"`
			AreaID          int                      `json:"area_id"`
			DepartmentID    int                      `json:"department_id"`
			Priority        string                   `json:"priority"`
			Subject         string                   `json:"subject"`
			Description     string                   `json:"description"`
			Reference       string                   `json:"reference"`
			Status          string                   `json:"status"`
			Recommender     int                      `json:"recommender"`
			LastRecommender int                      `json:"last_recommender"`
			ApprovalList    []models.NFAApprovalList `json:"approval_list"`
			Files           []models.NFAFile         `json:"files"`
		}

		// Bind the JSON request
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
			return
		}

		// Insert NFA details and get NFA ID
		var nfaID int
		query := `INSERT INTO nfa 
            (project_id, tower_id, area_id, department_id, priority, subject, description, reference, recommender, last_recommender, initiator_id, status) 
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'Pending') RETURNING nfa_id`

		err = db.QueryRow(query, request.ProjectID, request.TowerID, request.AreaID, request.DepartmentID, request.Priority,
			request.Subject, request.Description, request.Reference, request.Recommender, request.LastRecommender, initiatorID).Scan(&nfaID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Insert into NFA approval list and store nfa_id
		for i := range request.ApprovalList {
			request.ApprovalList[i].NFAID = nfaID
			approvalQuery := `INSERT INTO nfa_approval_list (nfa_id, approver_id, "order_value") VALUES ($1, $2, $3) RETURNING id`
			err := db.QueryRow(approvalQuery, request.ApprovalList[i].NFAID, request.ApprovalList[i].ApproverID, request.ApprovalList[i].Order).Scan(&request.ApprovalList[i].ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert approval list"})
				return
			}
		}

		// Insert files and store nfa_id
		for i := range request.Files {
			request.Files[i].NFAID = nfaID
			fileQuery := `INSERT INTO nfa_files (nfa_id, file_name, file_path) VALUES ($1, $2, $3) RETURNING id`
			err := db.QueryRow(fileQuery, request.Files[i].NFAID, request.Files[i].Name, request.Files[i].Path).Scan(&request.Files[i].ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert file records"})
				return
			}
		}

		// Success response
		c.JSON(http.StatusCreated, gin.H{
			"message":       "NFA created successfully",
			"nfa_id":        nfaID,
			"initiator_id":  initiatorID,
			"approval_list": request.ApprovalList,
			"files":         request.Files,
		})
	}
}

func GetPendingApprovals(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session ID from header
		sessionID := c.GetHeader("Authorization")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "session-id header is required",
				"details": "Authorization header is missing"})
			return
		}

		// Get user_id from session
		var userID int
		err := db.QueryRow("SELECT user_id FROM session WHERE session_id = $1", sessionID).Scan(&userID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Invalid session",
					"details": "No session found with provided session ID"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Database error",
					"details": fmt.Sprintf("Error fetching session: %v", err)})
			}
			return
		}

		// Query to fetch pending NFAs for approval with all related information
		query := `
            SELECT 
                n.nfa_id, 
                n.project_id, 
                n.tower_id, 
                n.area_id, 
                n.department_id,
                n.status,
                COALESCE(n.priority, '') as priority,
                COALESCE(n.subject, '') as subject,
                COALESCE(n.description, '') as description,
                COALESCE(n.reference, '') as reference,
                n.recommender,
                n.last_recommender,
                n.initiator_id,
                COALESCE(initiator.name, '') as initiator_name,
                COALESCE(recommender.name, '') as recommender_name,
                COALESCE(last_recommender.name, '') as last_recommender_name,
                COALESCE(p.project_name, '') as project_name,
                COALESCE(t.tower_name, '') as tower_name,
                COALESCE(a.area_name, '') as area_name,
                COALESCE(d.department_name, '') as department_name,
                al.order_value,
                al.started_at
            FROM nfa n
            INNER JOIN nfa_approval_list al ON n.nfa_id = al.nfa_id
            LEFT JOIN users initiator ON n.initiator_id = initiator.id
            LEFT JOIN users recommender ON n.recommender = recommender.id
            LEFT JOIN users last_recommender ON n.last_recommender = last_recommender.id
            LEFT JOIN projects p ON n.project_id = p.project_id
            LEFT JOIN towers t ON n.tower_id = t.tower_id
            LEFT JOIN areas a ON n.area_id = a.area_id
            LEFT JOIN departments d ON n.department_id = d.department_id
            WHERE al.approver_id = $1
            AND al.status = 'Pending'
            AND al.started_at IS NOT NULL
            AND al.updated_at IS NULL
            ORDER BY al.started_at ASC`

		rows, err := db.Query(query, userID)
		if err != nil {
			log.Printf("Database query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":       "Database query failed",
				"details":     fmt.Sprintf("Failed to fetch pending approvals: %v", err),
				"query_error": err.Error()})
			return
		}
		defer rows.Close()

		// Define the response structure
		type NFAWithNames struct {
			models.NFA
			InitiatorName       string    `json:"initiator_name"`
			RecommenderName     string    `json:"recommender_name"`
			LastRecommenderName string    `json:"last_recommender_name"`
			ProjectName         string    `json:"project_name"`
			TowerName           string    `json:"tower_name"`
			AreaName            string    `json:"area_name"`
			DepartmentName      string    `json:"department_name"`
			OrderValue          int       `json:"order_value"`
			StartedAt           time.Time `json:"started_at"`
		}

		var pendingNFAs []NFAWithNames

		for rows.Next() {
			var nfa NFAWithNames

			err := rows.Scan(
				&nfa.NFAID,
				&nfa.ProjectID,
				&nfa.TowerID,
				&nfa.AreaID,
				&nfa.DepartmentID,
				&nfa.Status,
				&nfa.Priority,
				&nfa.Subject,
				&nfa.Description,
				&nfa.Reference,
				&nfa.Recommender,
				&nfa.LastRecommender,
				&nfa.InitiatorID,
				&nfa.InitiatorName,
				&nfa.RecommenderName,
				&nfa.LastRecommenderName,
				&nfa.ProjectName,
				&nfa.TowerName,
				&nfa.AreaName,
				&nfa.DepartmentName,
				&nfa.OrderValue,
				&nfa.StartedAt,
			)
			if err != nil {
				log.Printf("Row scan error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":      "Data scan failed",
					"details":    fmt.Sprintf("Failed to scan NFA data: %v", err),
					"scan_error": err.Error()})
				return
			}

			// Fetch approvals and files for each NFA
			if err := fetchApprovalsAndFiles(db, &nfa.NFA); err != nil {
				log.Printf("Approvals and files fetch error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":       "Related data fetch failed",
					"details":     fmt.Sprintf("Failed to fetch approvals and files: %v", err),
					"fetch_error": err.Error()})
				return
			}

			pendingNFAs = append(pendingNFAs, nfa)
		}

		if err = rows.Err(); err != nil {
			log.Printf("Row iteration error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":           "Data iteration failed",
				"details":         fmt.Sprintf("Error during row iteration: %v", err),
				"iteration_error": err.Error()})
			return
		}

		// Return empty array if no pending approvals
		if len(pendingNFAs) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"message": "No pending approvals found",
				"nfas":    []NFAWithNames{},
				"user_id": userID})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Pending approvals retrieved successfully",
			"nfas":    pendingNFAs,
			"count":   len(pendingNFAs),
			"user_id": userID,
		})
	}
}

func GetNFAByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get NFA ID from URL parameter/
		nfaID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid NFA ID",
				"details": "NFA ID must be a number"})
			return
		}

		// Updated query removing created_at and updated_at
		query := `
            SELECT 
                n.nfa_id, 
                n.project_id, 
                n.tower_id, 
                n.area_id, 
                n.department_id,
                n.status,
                COALESCE(n.priority, '') as priority,
                COALESCE(n.subject, '') as subject,
                COALESCE(n.description, '') as description,
                COALESCE(n.reference, '') as reference,
                n.recommender,
                n.last_recommender,
                n.initiator_id,
                COALESCE(initiator.name, '') as initiator_name,
                COALESCE(initiator.email, '') as initiator_email,
                COALESCE(recommender.name, '') as recommender_name,
                COALESCE(recommender.email, '') as recommender_email,
                COALESCE(last_recommender.name, '') as last_recommender_name,
                COALESCE(last_recommender.email, '') as last_recommender_email,
                COALESCE(p.project_name, '') as project_name,
                COALESCE(t.tower_name, '') as tower_name,
                COALESCE(a.area_name, '') as area_name,
                COALESCE(d.department_name, '') as department_name
            FROM nfa n
            LEFT JOIN users initiator ON n.initiator_id = initiator.id
            LEFT JOIN users recommender ON n.recommender = recommender.id
            LEFT JOIN users last_recommender ON n.last_recommender = last_recommender.id
            LEFT JOIN projects p ON n.project_id = p.project_id
            LEFT JOIN towers t ON n.tower_id = t.tower_id
            LEFT JOIN areas a ON n.area_id = a.area_id
            LEFT JOIN departments d ON n.department_id = d.department_id
            WHERE n.nfa_id = $1`

		// Updated response structure removing timestamps
		type NFADetailResponse struct {
			models.NFA
			InitiatorName        string `json:"initiator_name"`
			InitiatorEmail       string `json:"initiator_email"`
			RecommenderName      string `json:"recommender_name"`
			RecommenderEmail     string `json:"recommender_email"`
			LastRecommenderName  string `json:"last_recommender_name"`
			LastRecommenderEmail string `json:"last_recommender_email"`
			ProjectName          string `json:"project_name"`
			TowerName            string `json:"tower_name"`
			AreaName             string `json:"area_name"`
			DepartmentName       string `json:"department_name"`
		}

		var nfaDetail NFADetailResponse

		// Updated scan removing timestamps
		err = db.QueryRow(query, nfaID).Scan(
			&nfaDetail.NFAID,
			&nfaDetail.ProjectID,
			&nfaDetail.TowerID,
			&nfaDetail.AreaID,
			&nfaDetail.DepartmentID,
			&nfaDetail.Status,
			&nfaDetail.Priority,
			&nfaDetail.Subject,
			&nfaDetail.Description,
			&nfaDetail.Reference,
			&nfaDetail.Recommender,
			&nfaDetail.LastRecommender,
			&nfaDetail.InitiatorID,
			&nfaDetail.InitiatorName,
			&nfaDetail.InitiatorEmail,
			&nfaDetail.RecommenderName,
			&nfaDetail.RecommenderEmail,
			&nfaDetail.LastRecommenderName,
			&nfaDetail.LastRecommenderEmail,
			&nfaDetail.ProjectName,
			&nfaDetail.TowerName,
			&nfaDetail.AreaName,
			&nfaDetail.DepartmentName,
		)

		// Rest of the code remains the same...
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "NFA not found",
					"details": fmt.Sprintf("No NFA found with ID: %d", nfaID)})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Database error",
				"details": fmt.Sprintf("Error fetching NFA details: %v", err)})
			return
		}

		// Fetch approvals with approver details
		approvalQuery := `
            SELECT 
                al.id,
                al.nfa_id,
                al.approver_id,
                al.order_value,
                COALESCE(al.status, '') as status,
                COALESCE(al.comments, '') as comments,
                COALESCE(u.name, '') as approver_name,
                COALESCE(u.email, '') as approver_email,
                al.started_at,
                al.updated_at
            FROM nfa_approval_list al
            LEFT JOIN users u ON al.approver_id = u.id
            WHERE al.nfa_id = $1
            ORDER BY al.order_value`

		approvalRows, err := db.Query(approvalQuery, nfaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Database error",
				"details": fmt.Sprintf("Error fetching approval details: %v", err)})
			return
		}
		defer approvalRows.Close()

		type ApprovalDetail struct {
			models.NFAApprovalList
			ApproverName  string `json:"approver_name"`
			ApproverEmail string `json:"approver_email"`
		}

		var approvals []ApprovalDetail

		for approvalRows.Next() {
			var approval ApprovalDetail
			var startedAt, completedAt sql.NullTime

			err := approvalRows.Scan(
				&approval.ID,
				&approval.NFAID,
				&approval.ApproverID,
				&approval.Order,
				&approval.Status,
				&approval.Comments,
				&approval.ApproverName,
				&approval.ApproverEmail,
				&startedAt,
				&completedAt,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Data scan failed",
					"details": fmt.Sprintf("Failed to scan approval data: %v", err)})
				return
			}

			if startedAt.Valid {
				approval.StartedDate = startedAt.Time
			}
			if completedAt.Valid {
				approval.CompletedDate = completedAt.Time
			}

			approvals = append(approvals, approval)
		}

		// Fetch files
		filesQuery := `
            SELECT 
                id,
                nfa_id,
                COALESCE(file_name, '') as file_name,
                COALESCE(file_path, '') as file_path
            FROM nfa_files
            WHERE nfa_id = $1`

		fileRows, err := db.Query(filesQuery, nfaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Database error",
				"details": fmt.Sprintf("Error fetching file details: %v", err)})
			return
		}
		defer fileRows.Close()

		var files []models.NFAFile
		for fileRows.Next() {
			var file models.NFAFile
			if err := fileRows.Scan(&file.ID, &file.NFAID, &file.Name, &file.Path); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Data scan failed",
					"details": fmt.Sprintf("Failed to scan file data: %v", err)})
				return
			}
			files = append(files, file)
		}

		c.JSON(http.StatusOK, gin.H{
			"details":   nfaDetail,
			"approvals": approvals,
			"files":     files,
		})
	}
}
