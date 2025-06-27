package main

import (
	"net/http"
	"time"

	"queue-management-system/internal/models"
)

// operatorDashboard displays the operator dashboard
func (app *Application) operatorDashboard(w http.ResponseWriter, r *http.Request) {
	user, err := app.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get queue entries for operator's service type
	activeEntries, _ := app.queue.GetByStatus(models.StatusActive, &user.ServiceType)
	processingEntries, _ := app.queue.GetByStatus(models.StatusProcessing, &user.ServiceType)

	// Get current processing entry for this operator
	var currentEntry *models.QueueEntry
	for _, entry := range processingEntries {
		if entry.OperatorID != nil && *entry.OperatorID == user.ID {
			currentEntry = entry
			break
		}
	}

	// Get statistics
	stats, _ := app.queue.GetStats(user.ServiceType, time.Now())

	app.render(w, r, "operator-dashboard.page.tmpl", &TemplateData{
		QueueEntries: activeEntries,
		CurrentEntry: currentEntry,
		QueueStats:   stats,
	})
}

// callNext calls the next queue entry
func (app *Application) callNext(w http.ResponseWriter, r *http.Request) {
	user, err := app.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if operator is already serving someone
	processingEntries, _ := app.queue.GetByStatus(models.StatusProcessing, &user.ServiceType)
	for _, entry := range processingEntries {
		if entry.OperatorID != nil && *entry.OperatorID == user.ID {
			http.Error(w, "You are already serving a client", 400)
			return
		}
	}

	// Get next active entry
	nextEntry, err := app.queue.GetNextActive(user.ServiceType)
	if err != nil {
		http.Error(w, "No active queue entries", 404)
		return
	}

	// Update status to processing
	err = app.queue.UpdateStatus(nextEntry.ID, models.StatusProcessing, &user.ID)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	http.Redirect(w, r, "/operator", http.StatusSeeOther)
}

// postponeQueue postpones the current queue entry
func (app *Application) postponeQueue(w http.ResponseWriter, r *http.Request) {
	user, err := app.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	form, err := app.parseForm(r)
	if err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	entryID, err := form.GetInt("entry_id")
	if err != nil {
		http.Error(w, "Invalid entry ID", 400)
		return
	}

	// Get the entry
	entry, err := app.queue.Get(entryID)
	if err != nil {
		http.Error(w, "Queue entry not found", 404)
		return
	}

	// Check postpone limit (maximum 3 times)
	if entry.PostponeCount >= 3 {
		http.Error(w, "Maximum postpone limit reached", 400)
		return
	}

	// Check if entry belongs to this operator
	if entry.OperatorID == nil || *entry.OperatorID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Update status to postponed
	err = app.queue.UpdateStatus(entryID, models.StatusPostponed, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	http.Redirect(w, r, "/operator", http.StatusSeeOther)
}

// completeService marks the current service as complete
func (app *Application) completeService(w http.ResponseWriter, r *http.Request) {
	user, err := app.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	form, err := app.parseForm(r)
	if err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	entryID, err := form.GetInt("entry_id")
	if err != nil {
		http.Error(w, "Invalid entry ID", 400)
		return
	}

	// Get the entry
	entry, err := app.queue.Get(entryID)
	if err != nil {
		http.Error(w, "Queue entry not found", 404)
		return
	}

	// Check if entry belongs to this operator
	if entry.OperatorID == nil || *entry.OperatorID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Update status to serviced
	err = app.queue.UpdateStatus(entryID, models.StatusServiced, &user.ID)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	http.Redirect(w, r, "/operator", http.StatusSeeOther)
}

// postponedQueues shows postponed queue entries
func (app *Application) postponedQueues(w http.ResponseWriter, r *http.Request) {
	user, err := app.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get postponed entries for operator's service type
	postponedEntries, _ := app.queue.GetByStatus(models.StatusPostponed, &user.ServiceType)

	app.render(w, r, "postponed-queues.page.tmpl", &TemplateData{
		QueueEntries: postponedEntries,
	})
}

// callPostponed calls a specific postponed queue entry
func (app *Application) callPostponed(w http.ResponseWriter, r *http.Request) {
	user, err := app.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	form, err := app.parseForm(r)
	if err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	entryID, err := form.GetInt("entry_id")
	if err != nil {
		http.Error(w, "Invalid entry ID", 400)
		return
	}

	// Check if operator is already serving someone
	processingEntries, _ := app.queue.GetByStatus(models.StatusProcessing, &user.ServiceType)
	for _, entry := range processingEntries {
		if entry.OperatorID != nil && *entry.OperatorID == user.ID {
			http.Error(w, "You are already serving a client", 400)
			return
		}
	}

	// Get the postponed entry
	entry, err := app.queue.Get(entryID)
	if err != nil {
		http.Error(w, "Queue entry not found", 404)
		return
	}

	// Verify it's postponed and for the correct service type
	if entry.Status != models.StatusPostponed || entry.ServiceType != user.ServiceType {
		http.Error(w, "Invalid queue entry", 400)
		return
	}

	// Update status to processing
	err = app.queue.UpdateStatus(entryID, models.StatusProcessing, &user.ID)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	http.Redirect(w, r, "/operator", http.StatusSeeOther)
}
