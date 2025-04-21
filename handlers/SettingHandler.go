package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"nfa-app/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Roles -------------------------------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------------------------------------

// CreateRole
// CreateRole
// func CreateRole(db *sql.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var role models.Role
// 		if err := c.ShouldBindJSON(&role); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		_, err := db.Exec("INSERT INTO roles (role_name, project_id) VALUES ($1, $2)", role.RoleName, role.ProjectID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		c.JSON(http.StatusCreated, gin.H{"message": "Role created"})
// 	}
// }

func CreateRole(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var role models.Role
		if err := c.ShouldBindJSON(&role); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec("INSERT INTO roles (role_name) VALUES ($1)", role.RoleName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Role created"})
	}
}

// GetRoles
func GetRoles(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT role_id, role_name FROM roles")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var roles []models.Role
		for rows.Next() {
			var role models.Role
			if err := rows.Scan(&role.RoleID, &role.RoleName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			roles = append(roles, role)
		}
		c.JSON(http.StatusOK, roles)
	}
}

// UpdateRole
func UpdateRole(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var role models.Role
		if err := c.ShouldBindJSON(&role); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec("UPDATE roles SET role_name=$1 WHERE role_id=$2", role.RoleName, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Role updated"})
	}
}

// DeleteRole
func DeleteRole(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		_, err := db.Exec("DELETE FROM roles WHERE role_id=$1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Role deleted"})
	}
}

// func GetAllRoleByProjectID(db *sql.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		projectIDStr := c.Param("project_id")              // Extract ProjectID from the URL parameter
// 		fmt.Println("Extracted project_id:", projectIDStr) // Log to check the parameter

// 		if projectIDStr == "" {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
// 			return
// 		}

// 		// Validate that the project_id is a valid integer
// 		projectID, err := strconv.Atoi(projectIDStr) // Convert the string to integer
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project_id"})
// 			return
// 		}

// 		// Query roles based on project_id
// 		rows, err := db.Query("SELECT role_id, role_name, project_id FROM roles WHERE project_id = $1", projectID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		defer rows.Close()

// 		var roles []models.Role
// 		for rows.Next() {
// 			var role models.Role
// 			if err := rows.Scan(&role.RoleID, &role.RoleName, &role.ProjectID); err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 				return
// 			}
// 			roles = append(roles, role)
// 		}

// 		// Return the roles or a not found message
// 		if len(roles) == 0 {
// 			c.JSON(http.StatusOK, gin.H{"message": "No roles found for the given project"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, roles)
// 	}
// }

// Permissions ----------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------------------

// CreatePermission
func CreatePermission(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var perm models.Permission
		if err := c.ShouldBindJSON(&perm); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec("INSERT INTO permissions (permission_name) VALUES ($1)", perm.PermissionName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Permission created"})
	}
}

// GetPermissions
func GetPermissions(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT permission_id, permission_name FROM permissions")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var permissions []models.Permission
		for rows.Next() {
			var perm models.Permission
			if err := rows.Scan(&perm.PermissionID, &perm.PermissionName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			permissions = append(permissions, perm)
		}
		c.JSON(http.StatusOK, permissions)
	}
}

// UpdatePermission
func UpdatePermission(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var perm models.Permission
		if err := c.ShouldBindJSON(&perm); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec("UPDATE permissions SET permission_name=$1 WHERE permission_id=$2", perm.PermissionName, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Permission updated"})
	}
}

// DeletePermission
func DeletePermission(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		_, err := db.Exec("DELETE FROM permissions WHERE permission_id=$1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Permission deleted"})
	}
}

// Role Permissions ---------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------------------------------------
//--------------------------------------------------------------------------------------------------------------------------

// CreateRolePermission
func CreateRolePermission(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var bulkInput []struct {
			RoleID      int   `json:"role_id"`
			Permissions []int `json:"permissions"`
		}

		if err := c.ShouldBindJSON(&bulkInput); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
			return
		}

		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}

		stmt, err := tx.Prepare("INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)")
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare statement"})
			return
		}
		defer stmt.Close()

		for _, item := range bulkInput {
			for _, permID := range item.Permissions {
				var exists bool
				err := db.QueryRow(`SELECT EXISTS (
					SELECT 1 FROM role_permissions 
					WHERE role_id = $1 AND permission_id = $2
				)`, item.RoleID, permID).Scan(&exists)
				if err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking existing permissions"})
					return
				}

				if !exists {
					if _, err := stmt.Exec(item.RoleID, permID); err != nil {
						tx.Rollback()
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert permission"})
						return
					}
				}
			}
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Role permissions created successfully"})
	}
}

// GetRolePermissions
func GetRolePermissions(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT role_id, permission_id FROM role_permissions")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var rolePerms []models.RolePermission
		for rows.Next() {
			var rolePerm models.RolePermission
			if err := rows.Scan(&rolePerm.RoleID, &rolePerm.PermissionID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			rolePerms = append(rolePerms, rolePerm)
		}
		c.JSON(http.StatusOK, rolePerms)
	}
}

func GetRolePermissionByRoleID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID := c.Param("role_id")
		if roleID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role ID is required"})
			return
		}

		query := `
			SELECT p.permission_id, p.permission_name
			FROM role_permissions rp
			JOIN permissions p ON rp.permission_id = p.permission_id
			WHERE rp.role_id = $1
		`
		rows, err := db.Query(query, roleID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch permissions"})
			return
		}
		defer rows.Close()

		var permissions []models.Permission

		for rows.Next() {
			var perm models.Permission
			if err := rows.Scan(&perm.PermissionID, &perm.PermissionName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read permissions"})
				return
			}
			permissions = append(permissions, perm)
		}

		if len(permissions) == 0 {
			c.JSON(http.StatusOK, gin.H{"message": "No permissions found for this role"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"role_id":     roleID,
			"permissions": permissions,
		})
	}
}

func UpdateRolePermission(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var bulkInput []struct {
			RoleID      int   `json:"role_id"`
			Permissions []int `json:"permissions"`
		}

		// Bind JSON input
		if err := c.ShouldBindJSON(&bulkInput); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
			return
		}

		// Start a new database transaction
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}

		// Prepare DELETE statement to remove existing role permissions
		stmtDelete, err := tx.Prepare("DELETE FROM role_permissions WHERE role_id = $1")
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare delete statement"})
			return
		}
		defer stmtDelete.Close()

		// Prepare INSERT statement to add new permissions
		stmtInsert, err := tx.Prepare("INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)")
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare insert statement"})
			return
		}
		defer stmtInsert.Close()

		// Loop through the input to delete and insert permissions for each role
		for _, item := range bulkInput {
			// Skip superadmin role (role_id = 1)
			if item.RoleID == 1 {
				continue
			}

			// Delete existing permissions for non-superadmin roles
			if _, err := stmtDelete.Exec(item.RoleID); err != nil {
				tx.Rollback()
				fmt.Printf("Error executing DELETE for role_id=%d: %v\n", item.RoleID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing permissions"})
				return
			}

			// Insert new permissions
			for _, permID := range item.Permissions {
				if _, err := stmtInsert.Exec(item.RoleID, permID); err != nil {
					tx.Rollback()
					fmt.Printf("Error executing INSERT for role_id=%d, permID=%d: %v\n", item.RoleID, permID, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert new permissions"})
					return
				}
			}
		}

		// Commit the transaction to finalize the changes
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
			return
		}

		// Return success message
		c.JSON(http.StatusOK, gin.H{"message": "Role permissions updated successfully"})
	}
}

// DeleteRolePermission
func DeleteRolePermission(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		_, err := db.Exec("DELETE FROM role_permissions WHERE role_id=$1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Role permission deleted"})
	}
}

// -------------------------------------------------------------------------------------------------
// USER SETTING ------------------------------------------------------------------------------------

func CreateSettingHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Define the input struct for JSON binding
		var setting models.Setting

		// Bind the JSON request to the `Setting` struct
		if err := c.ShouldBindJSON(&setting); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid input data",
			})
			return
		}

		// Validate `user_id`
		if setting.UserID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user_id",
			})
			return
		}

		// Insert or update the setting in the database
		query := `
            INSERT INTO settings (user_id, allow_multiple_sessions)
            VALUES ($1, $2)
            ON CONFLICT (user_id) DO UPDATE 
            SET allow_multiple_sessions = EXCLUDED.allow_multiple_sessions
        `

		_, err := db.Exec(query, setting.UserID, setting.AllowMultipleSessions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to create or update setting: %v", err),
			})
			return
		}

		// Return success response
		c.JSON(http.StatusCreated, gin.H{
			"message": "Setting updated successfully",
			"setting": setting,
		})
	}
}

func GetSettingHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract user_id from the query parameters
		userIDStr := c.Param("user_id")

		// Validate the presence of user_id
		if userIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}

		// Convert user_id to an integer
		userID, err := strconv.Atoi(userIDStr)
		if err != nil || userID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
			return
		}

		// Fetch the setting for the given user_id
		query := `SELECT user_id, allow_multiple_sessions FROM settings WHERE user_id = $1`
		var setting models.Setting

		err = db.QueryRow(query, userID).Scan(&setting.UserID, &setting.AllowMultipleSessions)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "No settings found for the given user_id"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch settings: %v", err)})
			}
			return
		}

		// Return the fetched setting
		c.JSON(http.StatusOK, gin.H{"setting": setting})
	}
}

// Department Handler -----------------------------------------------------------------------------------
func CreateDepartment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var department models.Department
		if err := c.ShouldBindJSON(&department); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec("INSERT INTO departments (department_name) VALUES ($1)", department.DepartmentName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Role created"})
	}
}

func GetAllDepartments(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT department_id, department_name FROM departments")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var departments []models.Department
		for rows.Next() {
			var department models.Department
			if err := rows.Scan(&department.DepartmentID, &department.DepartmentName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			departments = append(departments, department)
		}
		c.JSON(http.StatusOK, departments)
	}
}

func UpdateDepartment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var department models.Department
		if err := c.ShouldBindJSON(&department); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err := db.Exec("UPDATE departments SET department_name=$1 WHERE department_id=$2", department.DepartmentName, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Department updated"})
	}
}

func DeleteDepartment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM departments WHERE department_id=$1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Department deleted"})
	}
}

func CreateArea(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var area models.Area
		if err := c.ShouldBindJSON(&area); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err := db.Exec("INSERT INTO areas (area_name) VALUES ($1)", area.AreaName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Area created"})
	}
}

func GetAllAreas(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT area_id, area_name FROM areas")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var areas []models.Area
		for rows.Next() {
			var area models.Area
			if err := rows.Scan(&area.AreaID, &area.AreaName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			areas = append(areas, area)
		}
		c.JSON(http.StatusOK, areas)
	}
}

func UpdateArea(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var area models.Area
		if err := c.ShouldBindJSON(&area); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err := db.Exec("UPDATE areas SET area_name=$1 WHERE area_id=$2", area.AreaName, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Area updated"})
	}
}

func DeleteArea(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM areas WHERE area_id=$1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Area deleted"})
	}
}

func CreateProject(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var project models.Project
		if err := c.ShouldBindJSON(&project); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err := db.Exec("INSERT INTO projects (project_name, area_id) VALUES ($1, $2)", project.ProjectName, project.AreaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Project created"})
	}
}

func GetAllProjects(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT project_id, project_name FROM projects")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var projects []models.Project
		for rows.Next() {
			var project models.Project
			if err := rows.Scan(&project.ProjectID, &project.ProjectName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			projects = append(projects, project)
		}
		c.JSON(http.StatusOK, projects)
	}
}

func GetProjectsByAreaID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		areaID := c.Param("area_id") // Get area_id from URL parameter

		rows, err := db.Query("SELECT project_id, project_name FROM projects WHERE area_id = $1", areaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var projects []models.Project
		for rows.Next() {
			var project models.Project
			if err := rows.Scan(&project.ProjectID, &project.ProjectName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			projects = append(projects, project)
		}

		// Always return an array, even if empty
		c.JSON(http.StatusOK, projects)
	}
}

func UpdateProject(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var project models.Project
		if err := c.ShouldBindJSON(&project); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err := db.Exec("UPDATE projects SET project_name=$1 WHERE project_id=$2", project.ProjectName, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Project updated"})
	}
}

func DeleteProject(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM projects WHERE project_id=$1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Project deleted"})
	}
}

func CreateTower(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tower models.Tower
		if err := c.ShouldBindJSON(&tower); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err := db.Exec("INSERT INTO towers (tower_name, project_id) VALUES ($1, $2)", tower.TowerName, tower.ProjectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Tower created"})
	}
}

func GetTowersByProjectID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("project_id")
		rows, err := db.Query("SELECT tower_id, tower_name, project_id FROM towers WHERE project_id=$1", projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var towers []models.Tower
		for rows.Next() {
			var tower models.Tower
			if err := rows.Scan(&tower.TowerID, &tower.TowerName, &tower.ProjectID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			towers = append(towers, tower)
		}
		c.JSON(http.StatusOK, towers)
	}
}

func UpdateTower(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var tower models.Tower
		if err := c.ShouldBindJSON(&tower); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err := db.Exec("UPDATE towers SET tower_name=$1, project_id=$2 WHERE tower_id=$3", tower.TowerName, tower.ProjectID, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Tower updated"})
	}
}

func DeleteTower(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM towers WHERE tower_id=$1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Tower deleted"})
	}
}
