package models

import (
	"time"

	_ "github.com/lib/pq"
)

type User struct {
	ID             int       `json:"id"`
	Email          string    `json:"email"`
	Password       string    `json:"password"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	FirstAccess    time.Time `json:"first_access,omitempty"`
	LastAccess     time.Time `json:"last_access,omitempty"`
	ProfilePic     string    `json:"profile_picture"`
	Address        string    `json:"address"`
	PhoneNo        string    `json:"phone_no"`
	RoleID         int       `json:"role_id"`
	RoleName       string    `json:"role_name"` // role_name fetched dynamically
	DepartmentID   int       `json:"department_id"`
	DepartmentName string    `json:"department_name"` // department_name fetched dynamically
}

type Role struct {
	RoleID   int    `json:"role_id"`
	RoleName string `json:"role_name"`
}

type Permission struct {
	PermissionID   int    `json:"permission_id"`
	PermissionName string `json:"permission_name"`
}

type Session struct {
	UserID    int       `json:"user_id"`
	SessionID string    `json:"session_id"`
	HostName  string    `json:"host_name"`
	IPAddress string    `json:"ip_address"`
	Timestamp time.Time `json:"timestp"`
	ExpiresAt time.Time `json:"expires_at"`
}

type RolePermission struct {
	RoleID       int `json:"role_id"`
	PermissionID int `json:"permission_id"`
}

type Setting struct {
	UserID                int  `json:"user_id"`
	AllowMultipleSessions bool `json:"allow_multiple_sessions"`
}

type Department struct {
	DepartmentID   int    `json:"department_id"`
	DepartmentName string `json:"department_name"`
}

type Area struct {
	AreaID   int    `json:"area_id"`
	AreaName string `json:"area_name"`
}

type Project struct {
	ProjectID   int    `json:"project_id"`
	ProjectName string `json:"project_name"`
	AreaID      int    `json:"area_id"`
}

type Tower struct {
	TowerID   int    `json:"tower_id"`
	TowerName string `json:"tower_name"`
	ProjectID int    `json:"project_id"`
}

type Hierarchy struct {
	HierarchyID  int `json:"hierarchy_id"`
	UserID       int `json:"user_id"`
	DepartmentID int `json:"department_id"`
	Order        int `json:"order_value"`
}

type NFA struct {
	NFAID           int               `json:"nfa_id"`
	ProjectID       int               `json:"project_id"`
	TowerID         int               `json:"tower_id"`
	AreaID          int               `json:"area_id"`
	DepartmentID    int               `json:"department_id"`
	Priority        string            `json:"priority"`
	Subject         string            `json:"subject"`
	Description     string            `json:"description"`
	Reference       string            `json:"reference"`
	Recommender     int               `json:"recommender"`
	LastRecommender int               `json:"last_recommender"`
	InitiatorID     int               `json:"initiator_id"`
	Approvals       []NFAApprovalList `json:"approvals"`
	Files           []NFAFile         `json:"files"`
	Status          string            `json:"status"`
}

type NFAFile struct {
	ID    int    `json:"id"`
	NFAID int    `json:"nfa_id"`
	Path  string `json:"file_path"`
	Name  string `json:"file_name"`
}

type NFAApprovalList struct {
	ID            int       `json:"id"`
	NFAID         int       `json:"nfa_id"`
	ApproverID    int       `json:"approver_id"`
	Order         int       `json:"order_value"`
	Status        string    `json:"status"`
	Comments      string    `json:"comments"`
	ApproverName  string    `json:"approver_name"`
	StartedDate   time.Time `json:"started_at"`
	CompletedDate time.Time `json:"completed_at"`
}
