package mail

// AssetFilePath is file location path
const AssetFilePath = "./public/assets"

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
