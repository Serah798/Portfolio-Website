package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Contact represents the data collected from the contact form
type Contact struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Country     string `json:"country"`
	JobTitle    string `json:"job_title"`
	SubmittedAt string
}

// Function to handle requests to the root ("/") URL
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the HTML template from the templates folder
	tmplPath := filepath.Join("templates", "index.html") // Using relative path to templates folder
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	// Send the response by rendering the template
	tmpl.Execute(w, nil)
}

// Function to display the contact form (HTML embedded in Go code)
func contactHandler(w http.ResponseWriter, r *http.Request) {
	// Directly write the HTML for the contact form
	fmt.Fprintf(w, `
	<h1>Contact Me</h1>
	<form method="POST" action="/submit-contact">
		<label for="name">Your Name:</label><br>
		<input type="text" id="name" name="name" required><br><br>

		<label for="email">Your Email:</label><br>
		<input type="email" id="email" name="email" required><br><br>

		<label for="phone">Your Phone:</label><br>
		<input type="tel" id="phone" name="phone"><br><br>

		<label for="country">Country:</label><br>
		<input type="text" id="country" name="country"><br><br>

		<label for="job_title">Job Title:</label><br>
		<input type="text" id="job_title" name="job_title"><br><br>

		<input type="submit" value="Submit">
	</form>
	`)
}

// Function to handle inserting contact form data into the database
func submitContactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form data", http.StatusBadRequest)
			log.Printf("Error parsing form data: %v\n", err)
			return
		}

		contact := Contact{
			Name:     r.FormValue("name"),
			Email:    r.FormValue("email"),
			Phone:    r.FormValue("phone"),
			Country:  r.FormValue("country"),
			JobTitle: r.FormValue("job_title"),
			SubmittedAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		log.Printf("Inserting into database: %+v\n", contact)

		// Open a connection to the database
		db, err := sql.Open("mysql", "root:serah36@tcp(127.0.0.1:3306)/portfolio")
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			log.Printf("Database connection error: %v\n", err)
			return 
		}
		defer db.Close()

		// Insert the contact data into the database
		_, err = db.Exec("INSERT INTO contact_form (name, email, phone, country, job_title, submitted_at) VALUES (?, ?, ?, ?, ?, ?)",
			contact.Name, contact.Email, contact.Phone, contact.Country, contact.JobTitle, contact.SubmittedAt)
		if err != nil {
			log.Printf("Database insertion error: %v\n", err)
			http.Error(w, "Database insertion error", http.StatusInternalServerError)
			
			return
		}

		// Send a JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	} else {
		// Show the contact form if it's not a POST request
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// Function to view all contact form submissions from the database
func viewContactsHandler(w http.ResponseWriter, r *http.Request) {
	// Open a connection to the database
	db, err := sql.Open("mysql", "root:serah36@tcp(127.0.0.1:3306)/portfolio")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query all contacts
	rows, err := db.Query("SELECT name, email, phone, country, job_title, submitted_at FROM contact_form")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Display the contacts
	var contacts []Contact
	for rows.Next() {
		var contact Contact
		if err := rows.Scan(&contact.Name, &contact.Email, &contact.Phone, &contact.Country, &contact.JobTitle, &contact.SubmittedAt); err != nil {
			log.Fatal(err)
		}
		contacts = append(contacts, contact)
	}

	// Check for errors after iterating over rows
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// Render the contacts in the response
	tmpl, err := template.New("contacts").Parse(`
	<h1>Contact Submissions</h1>
	<ul>
		{{range .}}
			<li><strong>{{.Name}}</strong> ({{.Email}}) - {{.Phone}} - {{.Country}} - {{.JobTitle}} - Submitted At: {{.SubmittedAt}}</li>
		{{end}}
	</ul>`)
	if err != nil {
		log.Fatal(err)
	}
	tmpl.Execute(w, contacts)
}

// Function to test the database connection
func testDBConnectionHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "root:serah36@tcp(127.0.0.1:3306)/portfolio")
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Printf("Database connection error: %v\n", err)
		return
	}
	defer db.Close()
	
	err = db.Ping()
	if err != nil {
		http.Error(w, "Database ping error", http.StatusInternalServerError)
		log.Printf("Database ping error: %v\n", err)
		return
	}
	
	fmt.Fprintf(w, "Database connection successful")
}

func main() {
	// Serve static files (like script.js) from the current directory
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))

	// Set up the routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/contact", contactHandler) // Route for the contact form
	http.HandleFunc("/submit-contact", submitContactHandler) // Route for form submission
	http.HandleFunc("/view-contacts", viewContactsHandler) // Route to view contact submissions
	http.HandleFunc("/test-db-connection", testDBConnectionHandler) // Route to test database connection

	// Start the server on port 8080
	fmt.Println("Server is running on http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
