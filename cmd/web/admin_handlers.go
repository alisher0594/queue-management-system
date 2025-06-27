package main

import (
	"net/http"
	"time"

	"queue-management-system/internal/forms"
	"queue-management-system/internal/models"
)

// adminDashboard displays the admin dashboard
func (app *Application) adminDashboard(w http.ResponseWriter, r *http.Request) {
	// Get statistics for all service types
	var allStats []*models.QueueStats
	today := time.Now()

	for _, serviceType := range []models.ServiceType{
		models.ServiceTypeA,
		models.ServiceTypeB,
		models.ServiceTypeC,
	} {
		stats, _ := app.queue.GetStats(serviceType, today)
		allStats = append(allStats, stats)
	}

	// Get all entries for today
	allEntries, _ := app.queue.GetAll(today, nil)

	app.render(w, r, "admin-dashboard.page.tmpl", &TemplateData{
		QueueEntries: allEntries,
		AllStats:     allStats,
		Form: &forms.Form{
			Values: map[string][]string{
				"stats": {""}, // This will be populated by template logic
			},
		},
	})
}

// adminStats shows detailed statistics
func (app *Application) adminStats(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for date filtering
	dateStr := r.URL.Query().Get("date")
	serviceTypeStr := r.URL.Query().Get("service_type")

	// Default to today if no date specified
	date := time.Now()
	if dateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = parsedDate
		}
	}

	var serviceType *models.ServiceType
	if serviceTypeStr != "" {
		st := models.ServiceType(serviceTypeStr)
		serviceType = &st
	}

	// Get entries for the specified date and service type
	entries, _ := app.queue.GetAll(date, serviceType)

	// Get statistics
	var allStats []*models.QueueStats
	servicesToCheck := []models.ServiceType{
		models.ServiceTypeA,
		models.ServiceTypeB,
		models.ServiceTypeC,
	}

	if serviceType != nil {
		servicesToCheck = []models.ServiceType{*serviceType}
	}

	for _, st := range servicesToCheck {
		stats, _ := app.queue.GetStats(st, date)
		allStats = append(allStats, stats)
	}

	// Create aggregated admin stats
	adminStats := &AdminStats{}
	for _, stats := range allStats {
		adminStats.TotalServed += stats.TotalServiced
		adminStats.CurrentWaiting += stats.TotalActive
		adminStats.TotalPostponed += stats.TotalPostponed
		adminStats.AverageWaitTime += stats.AverageWaitTime

		// Assign to specific service
		switch stats.ServiceType {
		case models.ServiceTypeA:
			adminStats.ServiceA = ServiceStats{
				Served:      stats.TotalServiced,
				Waiting:     stats.TotalActive,
				Postponed:   stats.TotalPostponed,
				AverageWait: stats.AverageWaitTime,
			}
		case models.ServiceTypeB:
			adminStats.ServiceB = ServiceStats{
				Served:      stats.TotalServiced,
				Waiting:     stats.TotalActive,
				Postponed:   stats.TotalPostponed,
				AverageWait: stats.AverageWaitTime,
			}
		case models.ServiceTypeC:
			adminStats.ServiceC = ServiceStats{
				Served:      stats.TotalServiced,
				Waiting:     stats.TotalActive,
				Postponed:   stats.TotalPostponed,
				AverageWait: stats.AverageWaitTime,
			}
		}
	}

	// Average the wait time
	if len(allStats) > 0 {
		adminStats.AverageWaitTime /= float64(len(allStats))
	}

	app.render(w, r, "admin-stats.page.tmpl", &TemplateData{
		QueueEntries: entries,
		AdminStats:   adminStats,
		AllStats:     allStats,
		Form: &forms.Form{
			Values: map[string][]string{
				"date":         {date.Format("2006-01-02")},
				"service_type": {serviceTypeStr},
			},
		},
	})
}

// manageUsers shows user management page
func (app *Application) manageUsers(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "manage-users.page.tmpl", nil)
}

// createUser creates a new user (operator)
func (app *Application) createUser(w http.ResponseWriter, r *http.Request) {
	form, err := app.parseForm(r)
	if err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	form.Required("name", "email", "password", "role", "service_type")
	form.MinLength("name", 1)
	form.MaxLength("name", 100)
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 6)
	form.PermittedValues("role", "admin", "operator")
	form.PermittedValues("service_type", "A", "B", "C")

	if !form.Valid() {
		app.render(w, r, "manage-users.page.tmpl", &TemplateData{Form: form})
		return
	}

	err = app.users.Insert(
		form.Get("name"),
		form.Get("email"),
		form.Get("password"),
		form.Get("role"),
		models.ServiceType(form.Get("service_type")),
	)

	if err != nil {
		if err == models.ErrDuplicateEmail {
			form.Errors.Add("email", "Email address is already in use")
			app.render(w, r, "manage-users.page.tmpl", &TemplateData{Form: form})
		} else {
			http.Error(w, "Internal Server Error", 500)
		}
		return
	}

	// Redirect back to user management with success message
	app.render(w, r, "manage-users.page.tmpl", &TemplateData{
		Flash: "User created successfully",
	})
}
