package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"queue-management-system/internal/forms"
	"queue-management-system/internal/middleware"
	"queue-management-system/internal/models"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/nosurf"
)

// Application holds application dependencies
type Application struct {
	queue     *models.InMemoryQueueModel
	users     *models.InMemoryUserModel
	sessions  *models.InMemorySessionModel
	templates map[string]*template.Template
	upgrader  websocket.Upgrader
}

// TemplateData holds data for template rendering
type TemplateData struct {
	CSRFToken       string
	Flash           string
	Form            *forms.Form
	User            *models.User
	QueueEntries    []*models.QueueEntry
	QueueStats      *models.QueueStats
	AdminStats      *AdminStats
	AllStats        []*models.QueueStats
	ServiceTypes    []models.ServiceType
	CurrentEntry    *models.QueueEntry
	IsAuthenticated bool
}

// AdminStats holds aggregated statistics for admin dashboard
type AdminStats struct {
	TotalServed     int          `json:"total_served"`
	CurrentWaiting  int          `json:"current_waiting"`
	TotalPostponed  int          `json:"total_postponed"`
	AverageWaitTime float64      `json:"average_wait_time"`
	ServiceA        ServiceStats `json:"service_a"`
	ServiceB        ServiceStats `json:"service_b"`
	ServiceC        ServiceStats `json:"service_c"`
}

// ServiceStats holds statistics for a specific service
type ServiceStats struct {
	Served      int     `json:"served"`
	Waiting     int     `json:"waiting"`
	Postponed   int     `json:"postponed"`
	AverageWait float64 `json:"average_wait"`
}

func main() {
	app := &Application{
		queue:     models.NewInMemoryQueueModel(),
		users:     models.NewInMemoryUserModel(),
		sessions:  models.NewInMemorySessionModel(),
		templates: make(map[string]*template.Template),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
			HandshakeTimeout:  30 * time.Second,
			WriteBufferSize:   4096, // Increased buffer size
			ReadBufferSize:    4096, // Increased buffer size
			EnableCompression: true, // Enable compression for better performance
			Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
				log.Printf("WebSocket upgrade failed: %v", reason)
			},
		},
	}

	// Initialize templates
	err := app.initTemplates()
	if err != nil {
		log.Fatal(err)
	}

	// Set up routes
	router := app.routes()

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	fmt.Printf("Starting server on :%s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// routes sets up the application routes
func (app *Application) routes() http.Handler {
	router := httprouter.New()

	// Static files
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	// WebSocket for real-time updates
	router.HandlerFunc(http.MethodGet, "/ws", app.websocketHandler)

	// Public routes
	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodPost, "/", app.joinQueue)
	router.HandlerFunc(http.MethodGet, "/queue/status/:number", app.queueStatus)
	router.HandlerFunc(http.MethodGet, "/display", app.displayBoard)

	// Authentication routes
	router.HandlerFunc(http.MethodGet, "/user/login", app.loginForm)
	router.HandlerFunc(http.MethodPost, "/user/login", app.loginUser)
	router.HandlerFunc(http.MethodPost, "/user/logout", app.logoutUser)

	// Operator routes (protected)
	router.Handler(http.MethodGet, "/operator",
		middleware.RequireAuthentication(
			middleware.RequireRole(app.users, "operator")(http.HandlerFunc(app.operatorDashboard))))

	router.Handler(http.MethodPost, "/operator/call-next",
		middleware.RequireAuthentication(
			middleware.RequireRole(app.users, "operator")(http.HandlerFunc(app.callNext))))

	router.Handler(http.MethodPost, "/operator/postpone",
		middleware.RequireAuthentication(
			middleware.RequireRole(app.users, "operator")(http.HandlerFunc(app.postponeQueue))))

	router.Handler(http.MethodPost, "/operator/complete",
		middleware.RequireAuthentication(
			middleware.RequireRole(app.users, "operator")(http.HandlerFunc(app.completeService))))

	router.Handler(http.MethodGet, "/operator/postponed",
		middleware.RequireAuthentication(
			middleware.RequireRole(app.users, "operator")(http.HandlerFunc(app.postponedQueues))))

	router.Handler(http.MethodPost, "/operator/call-postponed",
		middleware.RequireAuthentication(
			middleware.RequireRole(app.users, "operator")(http.HandlerFunc(app.callPostponed))))

	// Admin routes (protected)
	router.Handler(http.MethodGet, "/admin",
		middleware.RequireAuthentication(
			middleware.RequireRole(app.users, "admin")(http.HandlerFunc(app.adminDashboard))))

	router.Handler(http.MethodGet, "/admin/stats",
		middleware.RequireAuthentication(
			middleware.RequireRole(app.users, "admin")(http.HandlerFunc(app.adminStats))))

	router.Handler(http.MethodGet, "/admin/users",
		middleware.RequireAuthentication(
			middleware.RequireRole(app.users, "admin")(http.HandlerFunc(app.manageUsers))))

	router.Handler(http.MethodPost, "/admin/users",
		middleware.RequireAuthentication(
			middleware.RequireRole(app.users, "admin")(http.HandlerFunc(app.createUser))))

	// Apply global middleware
	return middleware.RecoverPanic(
		middleware.LogRequest(
			middleware.SecurityHeaders(
				middleware.NoSurf(
					middleware.Authenticate(app.sessions)(router)))))
}

// templateFunctions returns a map of custom template functions
func templateFunctions() template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"formatTime": func(t time.Time) string {
			return t.Format("15:04:05")
		},
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"formatDateTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"timeAgo": func(t time.Time) string {
			duration := time.Since(t)
			if duration < time.Minute {
				return "just now"
			} else if duration < time.Hour {
				minutes := int(duration.Minutes())
				return fmt.Sprintf("%d min ago", minutes)
			} else if duration < 24*time.Hour {
				hours := int(duration.Hours())
				return fmt.Sprintf("%d hours ago", hours)
			} else {
				days := int(duration.Hours() / 24)
				return fmt.Sprintf("%d days ago", days)
			}
		},
		"isZero": func(t time.Time) bool {
			return t.IsZero()
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"ne": func(a, b interface{}) bool {
			return a != b
		},
		"lt": func(a, b int) bool {
			return a < b
		},
		"gt": func(a, b int) bool {
			return a > b
		},
		"le": func(a, b int) bool {
			return a <= b
		},
		"ge": func(a, b int) bool {
			return a >= b
		},
	}
}

// initTemplates initializes HTML templates
func (app *Application) initTemplates() error {
	pages, err := filepath.Glob("./ui/html/*.page.tmpl")
	if err != nil {
		return err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(templateFunctions()).ParseFiles(page)
		if err != nil {
			return err
		}

		ts, err = ts.ParseGlob("./ui/html/*.layout.tmpl")
		if err != nil {
			return err
		}

		// Check if partial templates exist before parsing
		partials, err := filepath.Glob("./ui/html/*.partial.tmpl")
		if err != nil {
			return err
		}

		// Only parse partials if they exist
		if len(partials) > 0 {
			ts, err = ts.ParseGlob("./ui/html/*.partial.tmpl")
			if err != nil {
				return err
			}
		}

		app.templates[name] = ts
	}

	return nil
}

// render renders a template with data
func (app *Application) render(w http.ResponseWriter, r *http.Request, name string, td *TemplateData) {
	ts, ok := app.templates[name]
	if !ok {
		http.Error(w, fmt.Sprintf("Template %s does not exist", name), 500)
		return
	}

	if td == nil {
		td = &TemplateData{}
	}

	// Add CSRF token
	td.CSRFToken = nosurf.Token(r)

	// Add authentication status
	td.IsAuthenticated = app.isAuthenticated(r)

	// Add current user
	if td.IsAuthenticated {
		userID := r.Context().Value("userID").(int)
		user, err := app.users.Get(userID)
		if err == nil {
			td.User = user
		}
	}

	// Add service types
	td.ServiceTypes = []models.ServiceType{
		models.ServiceTypeA,
		models.ServiceTypeB,
		models.ServiceTypeC,
	}

	// Execute template to buffer first to catch errors
	buf := new(bytes.Buffer)
	err := ts.Execute(buf, td)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Write the buffer to the response writer
	w.Write(buf.Bytes())
}

// isAuthenticated checks if the current request is from an authenticated user
func (app *Application) isAuthenticated(r *http.Request) bool {
	return r.Context().Value("userID") != nil
}

// getCurrentUser gets the current authenticated user
func (app *Application) getCurrentUser(r *http.Request) (*models.User, error) {
	userID := r.Context().Value("userID")
	if userID == nil {
		return nil, models.ErrNoRecord
	}

	return app.users.Get(userID.(int))
}

// parseForm parses the request form
func (app *Application) parseForm(r *http.Request) (*forms.Form, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	return forms.New(r.PostForm), nil
}

// WebSocket handler for real-time updates
func (app *Application) websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := app.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Set up connection cleanup
	defer func() {
		log.Printf("WebSocket connection closing")
		conn.Close()
	}()

	// Set initial connection timeouts
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// Set up pong handler
	conn.SetPongHandler(func(string) error {
		log.Printf("Received pong from client")
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Set up close handler
	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("WebSocket connection closed with code: %d, text: %s", code, text)
		return nil
	})

	// Send initial data
	if err := app.sendQueueUpdate(conn); err != nil {
		log.Printf("WebSocket initial send error: %v", err)
		return
	}

	// Create channels for coordination
	done := make(chan struct{})
	writeErrors := make(chan error, 1)

	// Goroutine to handle reading messages (mainly for pong responses)
	go func() {
		defer close(done)
		for {
			messageType, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				} else {
					log.Printf("WebSocket connection closed normally")
				}
				return
			}

			// Handle different message types if needed
			if messageType == websocket.PongMessage {
				log.Printf("Received pong message")
			}
		}
	}()

	// Goroutine to handle periodic updates and pings
	go func() {
		updateTicker := time.NewTicker(5 * time.Second)
		pingTicker := time.NewTicker(30 * time.Second)
		defer updateTicker.Stop()
		defer pingTicker.Stop()

		for {
			select {
			case <-done:
				return
			case <-updateTicker.C:
				if err := app.sendQueueUpdate(conn); err != nil {
					select {
					case writeErrors <- err:
					default:
					}
					return
				}
			case <-pingTicker.C:
				log.Printf("Sending ping to client")
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					select {
					case writeErrors <- err:
					default:
					}
					return
				}
			}
		}
	}()

	// Wait for either the read goroutine to finish or a write error
	select {
	case <-done:
		log.Printf("WebSocket read goroutine finished")
	case err := <-writeErrors:
		log.Printf("WebSocket write error: %v", err)
	}
}

// sendQueueUpdate sends queue updates via WebSocket
func (app *Application) sendQueueUpdate(conn *websocket.Conn) error {
	// Set write deadline before any write operation
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// Get current queue status for all services
	update := make(map[string]interface{})

	for _, serviceType := range []models.ServiceType{
		models.ServiceTypeA,
		models.ServiceTypeB,
		models.ServiceTypeC,
	} {
		active, _ := app.queue.GetByStatus(models.StatusActive, &serviceType)
		processing, _ := app.queue.GetByStatus(models.StatusProcessing, &serviceType)

		serviceData := map[string]interface{}{
			"active": len(active),
		}

		// Add active queue entries for operator dashboard
		if len(active) > 0 {
			activeEntries := make([]map[string]interface{}, 0, len(active))
			for _, entry := range active {
				activeEntries = append(activeEntries, map[string]interface{}{
					"ID":          entry.ID,
					"QueueNumber": entry.QueueNumber,
					"ServiceType": string(entry.ServiceType),
					"PhoneNumber": entry.PhoneNumber,
					"CreatedAt":   entry.CreatedAt.Format("15:04"),
					"Status":      string(entry.Status),
				})
			}
			serviceData["activeEntries"] = activeEntries
		} else {
			serviceData["activeEntries"] = []interface{}{}
		}

		// Add processing data in the format expected by the client
		if len(processing) > 0 {
			serviceData["processing"] = []map[string]interface{}{
				{
					"QueueNumber":  processing[0].QueueNumber,
					"queue_number": processing[0].QueueNumber, // Alternative field name for compatibility
					"Number":       processing[0].QueueNumber, // Another alternative
					"ServiceType":  string(processing[0].ServiceType),
					"CalledAt":     processing[0].CalledAt,
				},
			}
		} else {
			serviceData["processing"] = []interface{}{}
		}

		update[string(serviceType)] = serviceData
	}

	// Add timestamp for debugging
	update["timestamp"] = time.Now().Unix()
	update["server_time"] = time.Now().Format("15:04:05")

	log.Printf("Sending WebSocket update: %+v", update)

	err := conn.WriteJSON(update)
	if err != nil {
		// Log the specific error type
		if websocket.IsCloseError(err,
			websocket.CloseGoingAway,
			websocket.CloseAbnormalClosure,
			websocket.CloseNormalClosure) {
			log.Printf("WebSocket closed during write: %v", err)
		} else {
			log.Printf("WebSocket write error: %v", err)
		}
		return err
	}

	return nil
}
