// Filename: cmd/api/main.go
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"water.biling.system.driane.perez.net/internal/data"
	"water.biling.system.driane.perez.net/internal/jsonlog"
)

// The Application version number
const version = "1.0.0"

// The Configuration setting

type config struct {
	port int
	env  string // Development , staging, Production, etc.
	db   struct {
		//are gotten by flags
		dsn               string
		maxOpenConnection int
		maxIdleConnection int
		maxIdleTime       string
	}
	limiter struct {
		rps     float64 // requests/second
		burst   int
		enabled bool
	}
}

// Dependency Injection
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func main() {

	var cfg config
	//read in the flags that are needed to populate our config
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development | stagging | production )")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("WATER_DB_DSN"), "PostgresSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConnection, "db-max-open-connection", 25, "PostgreSQL max open connection")
	flag.IntVar(&cfg.db.maxIdleConnection, "db-max-idle-connection", 25, "PostgreSQL max idle connection")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max idle time")
	// These are flags for the rate limiter
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	
	flag.Parse()
	//create a logger
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	//create the connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	//close the connection to the db
	defer db.Close()
	//Log the seccessful connection pool
	logger.PrintInfo("database connection pool established", nil)
	//create an instance of our application struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}
	//Create our new servermux
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)
	//create our Http server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(), // using the routes function from routes,go
		ErrorLog: 	  log.New(logger,"",0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	//start our server
	logger.PrintInfo("starting server",map[string]string{
		"addr":srv.Addr,
		"env":cfg.env,
	})
	err = srv.ListenAndServe()
	logger.PrintFatal(err,nil)

}

// openDB() function returns a pointer *sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConnection)
	db.SetMaxIdleConns(cfg.db.maxIdleConnection)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	//create a context with a 5 second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
