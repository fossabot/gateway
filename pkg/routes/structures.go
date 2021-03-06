package routes

// URL define routes of services
type URL struct {
	Method string
	Path   string
}

// Server define services active servers
type Server struct {
	Server string
	Up     bool
}

// Service define structure of service loaded to this gateway
type Service struct {
	Name        string
	Path        string
	Server      []Server
	Version     int
	Description string
	Urls        []URL
}
