package mail

// Auth ...
type Auth struct {
	Enable   bool
	Username string
	Password string
}

// TLS ...
type TLS struct {
	Enable   bool
	CertFile string
	KeyFile  string
}
