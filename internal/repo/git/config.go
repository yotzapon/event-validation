package git

type Config struct {
	Url  string
	Auth struct {
		Username string // yes, this can be anything except an empty string
		Password string
	}
	RemoteName  string
	Destination string
}
