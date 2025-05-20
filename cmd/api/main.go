package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"sync"
	"time"

	db "github.com/RagoDevs/poult-be/db/sqlc"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gopkg.in/go-playground/validator.v9"
	"github.com/RagoDevs/poult-be/mail"
)


type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	frontend_url string
	emails     string
	version    string
}

type envelope map[string]interface{}

type application struct {
	config    config
	wg        sync.WaitGroup
	store     db.Store
	validator *validator.Validate
	mailer    *mail.Mailer
}

func init() {

	_ = godotenv.Load()

	var programLevel = new(slog.LevelVar)

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))
}

func main() {
	var cfg config
	var ml mail.Mailer

	flag.IntVar(&cfg.port, "port", 5055, "API server port")
	flag.StringVar(&cfg.env, "env", os.Getenv("ENV_STAGE"), "Environment (development|Staging|production")
	flag.StringVar(&cfg.version, "version", os.Getenv("VERSION"), "API version")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max ilde connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection  connections")
	flag.StringVar(&cfg.emails, "admin-email", os.Getenv("ADMIN_EMAIL"), "admin email")
	flag.StringVar(&cfg.frontend_url, "frontend-url", os.Getenv("FRONTEND_URL"), "frontend url ")
	flag.StringVar(&ml.Host, "MAIL HOST", os.Getenv("EMAIL_HOST"), "MAIL HOST")
	flag.StringVar(&ml.Port, "MAIL PORT", os.Getenv("EMAIL_PORT"), "MAIL PORT")
	flag.StringVar(&ml.User, "MAIL USER ", os.Getenv("EMAIL_USER"), "MAIL USER")
	flag.StringVar(&ml.Pwd, "MAIL PASSWORD", os.Getenv("EMAIL_PASS"), "MAIL PWD")

	flag.Parse()

	dbConn, err := openDB(cfg)
	if err != nil {
		slog.Error("error opening db", "error", err)
		os.Exit(1)
	}

	defer dbConn.Close()

	slog.Info("database connection established")

	app := &application{
		config:    cfg,
		store:     db.NewStore(dbConn),
		validator: validator.New(),
		mailer:    mail.NewMailer(ml.User, ml.Pwd, ml.Host, ml.Port),
	}

	err = app.serve()
	if err != nil {
		slog.Error("error starting server", "error", err)
		os.Exit(1)
	}

}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		return nil, err
	}

	return db, nil
}
