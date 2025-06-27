# Queue Management System

A comprehensive web application built with Go for managing customer queues with real-time updates.

## Features

- **Customer Interface**: Join queue by selecting service type and providing contact information
- **Operator Panel**: Call next customer, postpone, or complete service for specific service types
- **Admin Panel**: User management, statistics, and overall system oversight
- **Display Board**: Real-time queue status display for public viewing
- **Real-time Updates**: WebSocket integration for live status updates across all interfaces

## Architecture

- **Backend**: Go with standard libraries + httprouter for routing
- **Frontend**: HTML templates with Bootstrap 5 styling
- **Real-time**: WebSocket connections for live updates
- **Storage**: In-memory data storage (thread-safe with mutex protection)
- **Security**: CSRF protection, bcrypt password hashing, session management

## Service Types

The system supports three service types:
- **Service A**: General services
- **Service B**: Premium services  
- **Service C**: Consultation services

## Queue Statuses

- **Active**: Customer is waiting in queue
- **Processing**: Customer is currently being served
- **Serviced**: Service completed successfully
- **Postponed**: Customer postponed (max 2 postponements allowed)

## Default Users

The system comes with pre-configured users:

### Admin User
- **Username**: admin
- **Password**: admin123
- **Role**: admin
- **Access**: Full system access, user management, statistics

### Operator Users
- **Username**: operator_a
- **Password**: operator123
- **Role**: operator
- **Service**: Service A

- **Username**: operator_b  
- **Password**: operator123
- **Role**: operator
- **Service**: Service B

- **Username**: operator_c
- **Password**: operator123
- **Role**: operator
- **Service**: Service C

## Installation & Setup

### Prerequisites
- Go 1.19 or higher
- Git

### Installation

1. Clone or extract the project:
```bash
cd /path/to/your/projects
# If using git: git clone <repository-url>
cd queue-v1.2
```

2. Install dependencies:
```bash
go mod download
```

3. Run the application:
```bash
go run ./cmd/web
```

4. Access the application:
- **Home Page**: http://localhost:4000
- **Display Board**: http://localhost:4000/display
- **Login**: http://localhost:4000/user/login

## Usage Guide

### For Customers

1. Visit http://localhost:4000
2. Fill in your details (name, phone, email)
3. Select your service type (A, B, or C)
4. Click "Join Queue" to get your queue number
5. Check your status at the provided link or on the display board

### For Operators

1. Login at http://localhost:4000/user/login with operator credentials
2. Access your dashboard to see customers waiting for your service type
3. Use "Call Next" to call the next customer
4. Use "Postpone" if a customer needs to wait (max 2 times)
5. Use "Complete Service" when done serving a customer
6. View postponed customers and call them back when ready

### For Administrators

1. Login with admin credentials
2. Access admin dashboard for system overview
3. View detailed statistics at /admin/stats
4. Manage users (create new operators) at /admin/users
5. Monitor overall system performance

### Display Board

- Visit http://localhost:4000/display for a public display
- Shows currently serving numbers for all service types
- Auto-refreshes every 5 seconds
- Can be displayed on monitors in waiting areas

## API Endpoints

### Public Routes
- `GET /` - Home page (customer interface)
- `POST /` - Join queue
- `GET /queue/status/:number` - Check queue status
- `GET /display` - Display board
- `GET /user/login` - Login form
- `POST /user/login` - Process login
- `POST /user/logout` - Logout

### Operator Routes (Authentication Required)
- `GET /operator` - Operator dashboard
- `POST /operator/call-next` - Call next customer
- `POST /operator/postpone` - Postpone current customer
- `POST /operator/complete` - Complete service
- `GET /operator/postponed` - View postponed customers
- `POST /operator/call-postponed` - Call postponed customer

### Admin Routes (Admin Role Required)
- `GET /admin` - Admin dashboard
- `GET /admin/stats` - Detailed statistics
- `GET /admin/users` - User management
- `POST /admin/users` - Create new user

### Real-time Updates
- `GET /ws` - WebSocket endpoint for real-time updates

## Configuration

### Server Configuration
- **Port**: 4000 (can be changed in `cmd/web/main.go`)
- **Static Files**: Served from `./ui/static/`
- **Templates**: Located in `./ui/html/`

### Security Features
- CSRF protection on all forms
- Session-based authentication
- Role-based access control
- Secure headers (XSS protection, frame options)
- Password hashing with bcrypt

## File Structure

```
queue-v1.2/
├── cmd/web/                    # Application entry point
│   ├── main.go                # Main application and routing
│   ├── handlers.go            # Public route handlers
│   ├── operator_handlers.go   # Operator route handlers
│   └── admin_handlers.go      # Admin route handlers
├── internal/
│   ├── forms/                 # Form validation
│   ├── middleware/            # HTTP middleware
│   └── models/                # Data models and storage
├── ui/
│   ├── html/                  # HTML templates
│   └── static/               # CSS, JS, images
│       ├── css/main.css      # Custom styles
│       └── js/main.js        # Custom JavaScript
├── go.mod                     # Go module definition
└── README.md                  # This file
```

## Development

### Building
```bash
go build ./cmd/web
```

### Running Tests
```bash
go test ./...
```

### Adding New Features

1. **Models**: Add new data structures in `internal/models/`
2. **Handlers**: Add new route handlers in appropriate files
3. **Templates**: Create new templates in `ui/html/`
4. **Middleware**: Add new middleware in `internal/middleware/`
5. **Static Assets**: Add CSS/JS in `ui/static/`

## Troubleshooting

### Common Issues

1. **Port Already in Use**
   - Change the port in `main.go` or kill the process using port 4000

2. **Template Errors**
   - Ensure all templates have proper `{{define}}` blocks
   - Check template syntax for typos

3. **WebSocket Connection Failed**
   - Check browser console for errors
   - Ensure server is running and accessible

4. **Database/Storage Issues**
   - The system uses in-memory storage, so data is lost on restart
   - For production, implement persistent storage

### Logging

The application logs requests and errors to stdout. For production, consider:
- Redirecting logs to files
- Using structured logging
- Implementing log rotation

## Future Enhancements

- Persistent database storage (PostgreSQL, MySQL)
- REST API for mobile applications
- Email/SMS notifications
- Queue time predictions
- Multi-language support
- Appointment scheduling
- Integration with external systems

## License

This project is provided as-is for educational and commercial use.

## Support

For issues and questions, please check the troubleshooting section or review the code comments for implementation details.
