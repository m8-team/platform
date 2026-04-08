package database

type Config struct {
	DSN string
}

type Postgres struct {
	Config Config
}

func NewPostgres(cfg Config) *Postgres {
	return &Postgres{Config: cfg}
}
