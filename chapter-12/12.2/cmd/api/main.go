package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"time"
	// Import the pq driver so that it can register itsel with the database/sql
	// package. Note that we alias this import to the blank identifier, to stop the GO
	// compiler complaining that the package isn't being used.

	_ "github.com/lib/pq"
	"greelight.techkunstler.com/internal/data"
)

const version = "1.0.0"

// Add a db struct field to hold the configuration settings for our database connection pool
// For now this only holds the DSN, (Data Source Name) which we will read in from a
// command line flag.

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	// Add a new limiter struct containing fields for the requests-per-second and burst
	// values, and a boolean field which we can use to enable/disable rate limiting
	// altotether.

	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
}

func main() {

	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API Server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	// Read the DSN value from the db-dsn command-line flag into the config struct. We
	// default to using our development DSN if no flag is provided.

	/* 	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://greenlight:pa55word@localhost/greenlight", "PostgreSQL DSN") */
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgresSQL DSN")

	// Read the connection pool settings from command-line flags into theconfig struct.
	// Notice that thedefault values we're using are the ones we discussed above?
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "Postgress max open connection")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Postgress max idle connection")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreeSQL max idel timeout")

	// Create command line flags to read the setting values into the config struct.
	// notice that we use true as default for the 'enabled' setting?

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum request per second.")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Call the openDB() helper function (see below) to create the connection pool,
	// passing in the config struct. If this returns an error, we log it and exit the application immediately.

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Defer a call to db.Close() so that connection pool is closed before the main() function exits.

	defer db.Close()

	// Alos log a message to say that the connection pool has been successfully established.

	logger.Info("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	err = app.server()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// The openDB() function retusn a sql.DB connection pool.

func openDB(cfg config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using the DSN from the config
	// struct.

	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections in the pool. Note that
	// passing a value less than or equal to 0 will mean there is no limit

	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	// Create a context with a 5-second timeout deadline.

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	// Use PingContext() to establish a new connection to the database,
	// passing in the context we created above as a parameter. If the connection couldn't be
	// established successfully within the 5 second deadline, then this will return
	// an error. If we get thiserror, or any other, we close the connection pool and return the error

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Return the sql.DB connection pool

	return db, nil
}
