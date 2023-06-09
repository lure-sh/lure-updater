package config

type Config struct {
	Git     Git     `toml:"git"`
	Webhook Webhook `toml:"webhook"`
}

type Git struct {
	RepoDir     string      `toml:"repoDir"`
	RepoURL     string      `toml:"repoURL"`
	Commit      Commit      `toml:"commit"`
	Credentials Credentials `toml:"credentials"`
}

type Credentials struct {
	Username string
	Password string
}

type Commit struct {
	Name  string `toml:"name"`
	Email string `toml:"email"`
}

type Webhook struct {
	PasswordHash string `toml:"pwd_hash"`
}
