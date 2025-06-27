package main

import (
	"net/http"
	"strconv"
	"time"

	"queue-management-system/internal/forms"
	"queue-management-system/internal/models"

	"github.com/julienschmidt/httprouter"
)

// home displays the main page for clients to join the queue
func (app *Application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.tmpl", nil)
}

// joinQueue handles client joining the queue
func (app *Application) joinQueue(w http.ResponseWriter, r *http.Request) {
	form, err := app.parseForm(r)
	if err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	form.Required("service_type", "phone_number")
	form.PermittedValues("service_type", "A", "B", "C")
	form.MatchesPattern("phone_number", forms.PhoneRX)

	if !form.Valid() {
		app.render(w, r, "home.page.tmpl", &TemplateData{Form: form})
		return
	}

	serviceType := models.ServiceType(form.Get("service_type"))
	phoneNumber := form.Get("phone_number")

	entry, err := app.queue.Insert(serviceType, phoneNumber)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	app.render(w, r, "queue-success.page.tmpl", &TemplateData{
		CurrentEntry: entry,
	})
}

// queueStatus shows the status of a specific queue number
func (app *Application) queueStatus(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	queueNumber := params.ByName("number")

	entry, err := app.queue.GetByQueueNumber(queueNumber)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Get position in queue
	position := 0
	if entry.Status == models.StatusActive {
		activeEntries, _ := app.queue.GetByStatus(models.StatusActive, &entry.ServiceType)
		for i, activeEntry := range activeEntries {
			if activeEntry.ID == entry.ID {
				position = i + 1
				break
			}
		}
	}

	app.render(w, r, "queue-status.page.tmpl", &TemplateData{
		CurrentEntry: entry,
		Form: &forms.Form{
			Values: map[string][]string{
				"position": {strconv.Itoa(position)},
			},
		},
	})
}

// displayBoard shows the public display board
func (app *Application) displayBoard(w http.ResponseWriter, r *http.Request) {
	// Get currently processing entries for all services
	var processingEntries []*models.QueueEntry

	for _, serviceType := range []models.ServiceType{
		models.ServiceTypeA,
		models.ServiceTypeB,
		models.ServiceTypeC,
	} {
		entries, _ := app.queue.GetByStatus(models.StatusProcessing, &serviceType)
		processingEntries = append(processingEntries, entries...)
	}

	app.render(w, r, "display-board.page.tmpl", &TemplateData{
		QueueEntries: processingEntries,
	})
}

// loginForm displays the login form
func (app *Application) loginForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.tmpl", nil)
}

// loginUser handles user authentication
func (app *Application) loginUser(w http.ResponseWriter, r *http.Request) {
	form, err := app.parseForm(r)
	if err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	form.Required("email", "password")
	form.MatchesPattern("email", forms.EmailRX)

	if !form.Valid() {
		app.render(w, r, "login.page.tmpl", &TemplateData{Form: form})
		return
	}

	userID, err := app.users.Authenticate(form.Get("email"), form.Get("password"))
	if err != nil {
		form.Errors.Add("generic", "Email or Password is incorrect")
		app.render(w, r, "login.page.tmpl", &TemplateData{Form: form})
		return
	}

	// Create session
	token, err := app.sessions.Insert(userID, time.Now().Add(12*time.Hour))
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Set session cookie
	cookie := &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
	}
	http.SetCookie(w, cookie)

	// Redirect based on user role
	user, _ := app.users.Get(userID)
	if user.Role == "admin" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/operator", http.StatusSeeOther)
	}
}

// logoutUser handles user logout
func (app *Application) logoutUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		app.sessions.Delete(cookie.Value)
	}

	// Clear session cookie
	cookie = &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
