package handlers

import (
	"bytes"
	"database/sql"
	"io"
	"net/http"
	"nfa-app/models"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

// downloadImage downloads an image from URL and saves it to a temporary file
func downloadImage(url string) (string, error) {
	// Create temp directory if it doesn't exist
	tempDir := "temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", err
	}

	// Create temporary file
	tmpFile := filepath.Join(tempDir, "logo.png")

	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(tmpFile)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return tmpFile, nil
}

// Helper function to clean HTML tags and format text
func cleanHTML(html string) string {
	// Remove <p> tags
	html = strings.ReplaceAll(html, "<p>", "")
	html = strings.ReplaceAll(html, "</p>", "\n")

	// Replace <li> tags with bullet points
	html = strings.ReplaceAll(html, "<li>", "• ")
	html = strings.ReplaceAll(html, "</li>", "\n")

	// Remove <ul> tags
	html = strings.ReplaceAll(html, "<ul>", "")
	html = strings.ReplaceAll(html, "</ul>", "\n")

	// Remove any remaining HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	html = re.ReplaceAllString(html, "")

	// Trim extra whitespace
	html = strings.TrimSpace(html)

	return html
}

func GenerateNFAPDF(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		nfaIDStr := c.Param("nfa_id")
		nfaID, err := strconv.Atoi(nfaIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid NFA ID"})
			return
		}

		var nfa models.NFA
		err = db.QueryRow(`
			SELECT nfa_id, project_id, tower_id, area_id, department_id, 
			       priority, subject, description, reference, recommender, last_recommender
			FROM nfa WHERE nfa_id = $1`, nfaID).Scan(
			&nfa.NFAID, &nfa.ProjectID, &nfa.TowerID, &nfa.AreaID, &nfa.DepartmentID,
			&nfa.Priority, &nfa.Subject, &nfa.Description, &nfa.Reference, &nfa.Recommender, &nfa.LastRecommender)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "NFA not found"})
			return
		}

		var (
			projectName     = getName(db, "SELECT project_name FROM projects WHERE project_id = $1", nfa.ProjectID)
			departmentName  = getName(db, "SELECT department_name FROM departments WHERE department_id = $1", nfa.DepartmentID)
			areaName        = getName(db, "SELECT area_name FROM areas WHERE area_id = $1", nfa.AreaID)
			towerName       = getName(db, "SELECT tower_name FROM towers WHERE tower_id = $1", nfa.TowerID)
			recommenderName = getName(db, "SELECT name FROM users WHERE id = $1", nfa.Recommender)
		)

		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.SetMargins(20, 20, 20)
		pdf.AddPage()

		// Set footer callback
		pdf.SetFooterFunc(func() {
			pdf.SetY(-15) // Position at 15 mm from bottom
			pdf.SetFont("Arial", "I", 8)
			pdf.SetTextColor(128, 128, 128) // Gray color
			pdf.CellFormat(0, 10, "This is a system generated Approved NFA, does not require signature.", "", 0, "C", false, 0, "")
		})

		// Download and add logo
		logoURL := "https://nfa.blueinvent.com/api/get_file?file=1744967687116863871-image_2025_02_18T15_14_58_472Z.png"
		logoPath, err := downloadImage(logoURL)
		if err == nil {
			pdf.Image(logoPath, 150, 10, 40, 0, false, "", 0, "")
			// Clean up the temporary file
			defer os.Remove(logoPath)
		} else {
			// Fallback to text if image download fails
			pdf.SetFont("Arial", "B", 16)
			pdf.SetXY(150, 10)
			pdf.Cell(40, 10, "JAYPEE")
		}

		// NFA Number with better spacing
		pdf.SetFont("Arial", "B", 12)
		pdf.SetXY(20, 20)
		pdf.Cell(40, 10, "NFA No. "+strconv.Itoa(nfa.NFAID))

		// Title centered with better spacing and dark blue color
		pdf.SetFont("Arial", "B", 14)
		pdf.SetTextColor(0, 0, 139) // Dark blue color
		pdf.SetY(35)
		pdf.CellFormat(170, 10, "Note For Approval", "", 0, "C", false, 0, "")
		pdf.Ln(15)
		pdf.SetTextColor(0, 0, 0) // Reset to black color

		// Header section with improved alignment and spacing
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(25, 8, "Area:-")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(50, 8, areaName)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(30, 8, "Project:-")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(65, 8, projectName)
		pdf.Ln(10)

		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(25, 8, "Tower:-")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(50, 8, towerName)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(30, 8, "Department:-")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(65, 8, departmentName)
		pdf.Ln(10)

		// Reference section in original text format
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(25, 8, "Reference:-")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(50, 8, nfa.Reference)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(30, 8, "Priority:-")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(65, 8, nfa.Priority)
		pdf.Ln(12)

		// Initiator with improved spacing
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(25, 8, "Initiator:-")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 8, recommenderName)
		pdf.Ln(12)

		// Subject on same line
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(25, 8, "Subject:-")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(50, 8, nfa.Subject)
		pdf.Ln(12)

		// Description with HTML handling
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(25, 8, "Description:-")
		pdf.SetFont("Arial", "", 10)
		pdf.Ln(8)
		pdf.SetX(25)

		if nfa.Description == "" {
			pdf.MultiCell(0, 6, "No description provided", "", "L", false)
		} else {
			// Clean and format the HTML description
			cleanedDescription := cleanHTML(nfa.Description)

			// Split into lines and handle bullet points
			lines := strings.Split(cleanedDescription, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				// If line starts with bullet point, add proper indentation
				if strings.HasPrefix(line, "•") {
					pdf.SetX(25)
					pdf.MultiCell(0, 6, line, "", "L", false)
				} else {
					pdf.SetX(25)
					pdf.MultiCell(0, 6, line, "", "L", false)
				}
			}
		}
		pdf.Ln(10)

		// Approval Summary section with improved spacing
		pdf.SetFont("Arial", "B", 14)
		pdf.CellFormat(170, 10, "NFA Approval Summary", "", 0, "C", false, 0, "")
		pdf.Ln(12)

		// Table headers with better alignment
		pdf.SetFont("Arial", "B", 10)
		pdf.SetFillColor(240, 240, 240)
		pdf.SetDrawColor(128, 128, 128)
		headers := []string{"S. No.", "Particular", "Name & Desig.", "Received", "Approved"}
		widths := []float64{15, 35, 40, 40, 40}

		for i, h := range headers {
			pdf.CellFormat(widths[i], 8, h, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		// Table content with consistent formatting
		pdf.SetFont("Arial", "", 10)
		rows, err := db.Query(`
			SELECT nal.order_value, u.name, nal.status, nal.started_at, nal.updated_at 
			FROM nfa_approval_list nal 
			JOIN users u ON nal.approver_id = u.id 
			WHERE nal.nfa_id = $1 
			ORDER BY nal.order_value`, nfaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch approval list: " + err.Error()})
			return
		}
		defer rows.Close()

		orderNo := 1
		for rows.Next() {
			var order int
			var name, status string
			var createdAt, updatedAt time.Time

			if err := rows.Scan(&order, &name, &status, &createdAt, &updatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan approval data: " + err.Error()})
				return
			}

			particular := "Initiator"
			if orderNo == 2 {
				particular = "Recommender"
			} else if orderNo > 2 {
				particular = "Approver"
			}

			data := []string{
				strconv.Itoa(orderNo),
				particular,
				name,
				createdAt.Format("02-01-2006 15:04"),
				updatedAt.Format("02-01-2006 15:04"),
			}

			for i, txt := range data {
				pdf.CellFormat(widths[i], 8, txt, "1", 0, "C", false, 0, "")
			}
			pdf.Ln(-1)
			orderNo++
		}

		if err = rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating approval list: " + err.Error()})
			return
		}

		var buf bytes.Buffer
		err = pdf.Output(&buf)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF: " + err.Error()})
			return
		}

		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename=DOC-"+time.Now().Format("20060102")+"-WA"+strconv.Itoa(nfaID)+".pdf")
		c.Header("Content-Type", "application/pdf")
		c.Header("Expires", "0")
		c.Header("Cache-Control", "must-revalidate")
		c.Header("Pragma", "public")

		c.Data(http.StatusOK, "application/pdf", buf.Bytes())
	}
}

// Helper function to safely fetch names with fallback
func getName(db *sql.DB, query string, id int) string {
	var name string
	err := db.QueryRow(query, id).Scan(&name)
	if err != nil {
		return "Unknown"
	}
	return name
}
