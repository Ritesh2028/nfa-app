package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"nfa-app/handlers"
	"nfa-app/storage"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
)

func CORSConfig() cors.Config {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{
		"Content-Type", "Content-Length", "Accept-Encoding", "X-XSRF-TOKEN",
		"Accept", "Origin", "X-Requested-With", "Authorization", "User-Agent",
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"}
	return corsConfig
}

// IsAdminByRoleID checks if a user is an admin based on their role ID.
func IsAdminByRoleID(db *sql.DB, roleID int) (bool, error) {
	var roleName string
	err := db.QueryRow("SELECT role_name FROM roles WHERE role_id = $1", roleID).Scan(&roleName)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return strings.EqualFold(roleName, "superadmin") || strings.EqualFold(roleName, "admin"), nil
}

// HasProjectPermission checks if a user has a specific permission in a project.
func HasProjectPermission(db *sql.DB, userID int, projectID int, permissionID int) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM project_roles pr
		JOIN role_permissions rp ON pr.role_id = rp.role_id
		WHERE pr.project_id = $1
		  AND rp.permission_id = $2`
	err := db.QueryRow(query, projectID, permissionID).Scan(&count)
	return count > 0, err
}

// RBACMiddleware ensures users have appropriate permissions based on roles and project access.
func RBACMiddleware(db *sql.DB, requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.GetHeader("Authorization")
		if sessionID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session ID required"})
			c.Abort()
			return
		}

		// Fetch user by session ID
		user, err := storage.GetUserBySessionID(db, sessionID)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session", "err": err.Error()})
			c.Abort()
			return
		}

		// Fetch the role_id from the users table by user_id
		var roleID int
		err = db.QueryRow("SELECT role_id FROM users WHERE id = $1", user.ID).Scan(&roleID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to retrieve role ID"})
			c.Abort()
			return
		}

		// Check if the user is an admin by RoleID
		isAdmin, err := IsAdminByRoleID(db, roleID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to check admin role"})
			c.Abort()
			return
		}

		// Allow admins unrestricted access
		if isAdmin {
			c.Set("user", user)
			c.Next()
			return
		}

		// Check if the user has the required permission
		hasPermission, err := HasProjectPermission(db, user.ID, 1, 1)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to check project permission"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		// Store user info for use in handlers
		c.Set("user", user)
		c.Next()
	}
}

func main() {
	db := storage.InitDB()
	defer db.Close()

	// Setup cron job to run cleanup every hour
	c := cron.New()
	c.AddFunc("@hourly", func() {
		if err := storage.CleanupExpiredSessions(db); err != nil {
			log.Printf("Error cleaning up sessions: %v", err)
		}
	})
	c.Start()

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20

	r.Use(cors.New(CORSConfig()))

	r.POST("/api/login", handlers.LoginHandler(db))
	r.POST("/api/validate-session", handlers.ValidateSession(db))
	r.GET("/api/session/:user_id", handlers.GetSessionHandler(db))
	r.DELETE("/api/session/:user_id", handlers.DeleteSessionHandler(db))

	userRoutes := r.Group("/api/user")
	{
		userRoutes.POST("/create", handlers.CreateUser(db))
		userRoutes.PUT("/update/:id", handlers.UpdateUser(db))
		userRoutes.GET("/fetch/:id", handlers.GetUser(db))
		userRoutes.GET("/", handlers.GetAllUsers(db))
		userRoutes.DELETE("/delete/:id", handlers.DeleteUser(db))
	}
	r.GET("/api/get_user", handlers.GetUserFromSession(db))

	r.GET("/api/user/:role", handlers.GetUsersByRoleName(db))

	departmentRoutes := r.Group("/api/department")
	{
		departmentRoutes.POST("/create", handlers.CreateDepartment(db))
		departmentRoutes.PUT("/update/:id", handlers.UpdateDepartment(db))
		departmentRoutes.GET("/", handlers.GetAllDepartments(db))
		departmentRoutes.DELETE("/delete/:id", handlers.DeleteDepartment(db))
	}

	areaRoutes := r.Group("/api/area")
	{
		areaRoutes.POST("/create", handlers.CreateArea(db))
		areaRoutes.PUT("/update/:id", handlers.UpdateArea(db))
		areaRoutes.GET("/", handlers.GetAllAreas(db))
		areaRoutes.DELETE("/delete/:id", handlers.DeleteArea(db))
	}

	projectRoutes := r.Group("/api/project")
	{
		projectRoutes.POST("/create", handlers.CreateProject(db))
		projectRoutes.PUT("/update/:id", handlers.UpdateProject(db))
		projectRoutes.GET("/", handlers.GetAllProjects(db))
		projectRoutes.DELETE("/delete/:id", handlers.DeleteProject(db))
		projectRoutes.GET("/:area_id", handlers.GetProjectsByAreaID(db))
	}

	towerRoutes := r.Group("/api/tower")
	{
		towerRoutes.POST("/create", handlers.CreateTower(db))
		towerRoutes.PUT("/update/:id", handlers.UpdateTower(db))
		towerRoutes.GET("/:project_id", handlers.GetTowersByProjectID(db))
		towerRoutes.DELETE("/delete/:id", handlers.DeleteTower(db))
	}

	roleRoutes := r.Group("/api/roles")
	{
		roleRoutes.POST("/create", handlers.CreateRole(db))
		roleRoutes.GET("/", handlers.GetRoles(db))
		roleRoutes.PUT("/update/:id", handlers.UpdateRole(db))
		roleRoutes.DELETE("/delete/:id", handlers.DeleteRole(db))
	}

	permissionRoutes := r.Group("/api/permissions")
	{
		permissionRoutes.POST("/create", handlers.CreatePermission(db))
		permissionRoutes.GET("/", handlers.GetPermissions(db))
		permissionRoutes.PUT("/update/:id", handlers.UpdatePermission(db))
		permissionRoutes.DELETE("/delete/:id", handlers.DeletePermission(db))
	}

	rolePermissionRoutes := r.Group("/api/role_permissions")
	{
		rolePermissionRoutes.POST("/create", handlers.CreateRolePermission(db))
		rolePermissionRoutes.GET("/", handlers.GetRolePermissions(db))
		rolePermissionRoutes.GET("/:id", handlers.GetRolePermissionByRoleID(db))
		rolePermissionRoutes.PUT("/update/:id", handlers.UpdateRolePermission(db))
		rolePermissionRoutes.DELETE("/delete/:id", handlers.DeleteRolePermission(db))
	}

	settingRoutes := r.Group("/api/settings")
	{
		settingRoutes.POST("/create", handlers.CreateSettingHandler(db))
		settingRoutes.GET("/", handlers.GetSettingHandler(db))
	}

	hierarchyRoutes := r.Group("/api/hierarchies")
	{
		hierarchyRoutes.POST("/crreate", handlers.CreateHierarchy(db))
		hierarchyRoutes.GET("/", handlers.GetHierarchies(db))
		hierarchyRoutes.GET("/:department_id", handlers.GetHierarchyByDepartmentID(db))
		hierarchyRoutes.PUT("/update/:id", handlers.UpdateHierarchy(db))
		hierarchyRoutes.DELETE("/delete/:id", handlers.DeleteHierarchy(db))
	}

	nfaRoutes := r.Group("/api/nfa")
	{
		nfaRoutes.GET("/:id", handlers.GetNFAByID(db))
		nfaRoutes.GET("/project/:project_id", handlers.GetNFAByProjectID(db))
		nfaRoutes.GET("/department/:department_id", handlers.GetNFAByDepartmentID(db))
		nfaRoutes.GET("/area/:area_id", handlers.GetNFAByAreaID(db))
		nfaRoutes.GET("/tower/:tower_id", handlers.GetNFAByTowerID(db))
		nfaRoutes.GET("/priority/:priority", handlers.GetNFAByPriority(db))
		nfaRoutes.GET("/recommender", handlers.GetNFAByRecommender(db))
		nfaRoutes.GET("/all", handlers.GetAllNFA(db))
		nfaRoutes.GET("/initiator", handlers.GetNFAByInitiator(db))

		nfaRoutes.POST("/create", handlers.CreateNFA(db))
		nfaRoutes.PUT("/update/:id", handlers.UpdateNFA(db))
		nfaRoutes.DELETE("/delete/:id", handlers.DeleteNFA(db))
	}

	r.PUT("/api/reject_approve", handlers.ApproveOrRejectNFA(db))
	r.GET("/api/pending_approvals", handlers.GetPendingApprovals(db))
	r.GET("/api/fetch/nfa_data/:nfa_id", handlers.GetNFAApprovalList(db))
	r.POST("/api/add_approver", handlers.AddApprover(db))
	r.PUT("/api/approve/:nfa_id/:approver_id", handlers.ApproveNFA(db))
	r.DELETE("/api/approvers/:nfa_id/:approver_id", handlers.RemoveApprover(db))

	r.POST("/api/upload", handlers.UploadFiles)
	r.GET("/api/get_file", handlers.ServeNFAFile)

	// Add PDF generation route
	r.GET("/api/pdf/generate/:nfa_id", handlers.GenerateNFAPDF(db))

	if err := r.Run(":9000"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
