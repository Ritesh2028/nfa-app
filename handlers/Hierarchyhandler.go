package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"nfa-app/models"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateHierarchy(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Define request structure
		var request struct {
			DepartmentID int `json:"department_id"`
			Users        []struct {
				UserID int `json:"user_id"`
			} `json:"users"`
		}

		// Parse JSON
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Delete existing hierarchy for the department
		deleteQuery := `DELETE FROM hierarchy WHERE department_id = $1`
		_, err := db.Exec(deleteQuery, request.DepartmentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing hierarchy"})
			return
		}

		// Insert new hierarchy with auto-incremented order_value
		query := `INSERT INTO hierarchy (user_id, department_id, order_value) VALUES`
		values := []interface{}{}
		placeholders := []string{}

		for i, user := range request.Users {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
			values = append(values, user.UserID, request.DepartmentID, i+1) // Order starts from 1
		}

		query += strings.Join(placeholders, ", ")
		_, err = db.Exec(query, values...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Hierarchy updated successfully"})
	}
}

func GetHierarchies(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT hierarchy_id, user_id, department_id, order_value FROM hierarchy ORDER BY order_value ASC")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var hierarchies []models.Hierarchy
		for rows.Next() {
			var hierarchy models.Hierarchy
			if err := rows.Scan(&hierarchy.HierarchyID, &hierarchy.UserID, &hierarchy.DepartmentID, &hierarchy.Order); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			hierarchies = append(hierarchies, hierarchy)
		}
		c.JSON(http.StatusOK, hierarchies)
	}
}

func GetHierarchyByDepartmentID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		departmentID := c.Param("department_id")
		rows, err := db.Query("SELECT hierarchy_id, user_id, department_id, order_value FROM hierarchy WHERE department_id=$1 ORDER BY order_value ASC", departmentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var hierarchies []models.Hierarchy
		for rows.Next() {
			var hierarchy models.Hierarchy
			if err := rows.Scan(&hierarchy.HierarchyID, &hierarchy.UserID, &hierarchy.DepartmentID, &hierarchy.Order); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			hierarchies = append(hierarchies, hierarchy)
		}
		c.JSON(http.StatusOK, hierarchies)
	}
}

func UpdateHierarchy(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var hierarchy models.Hierarchy
		if err := c.ShouldBindJSON(&hierarchy); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err := db.Exec("UPDATE hierarchy SET user_id=$1, department_id=$2, order_value=$3 WHERE hierarchy_id=$4", hierarchy.UserID, hierarchy.DepartmentID, hierarchy.Order, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Hierarchy updated"})
	}
}

func DeleteHierarchy(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM hierarchy WHERE hierarchy_id=$1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Hierarchy deleted"})
	}
}
