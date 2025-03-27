package main

import (
	"context"
	"expvar"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"misc.sahilsasane.net/internal/data"
	"misc.sahilsasane.net/internal/jsonlog"
	"misc.sahilsasane.net/internal/llm"
)

var (
	version   string
	buildTime string
)

type config struct {
	env  string
	port int
	db   struct {
		uri         string
		database    string
		maxPoolSize uint64
		minPoolSize uint64
		maxIdleTime string
	}
	// limiter struct {
	// 	rps     float64
	// 	burst   int
	// 	enabled bool
	// }
	// cors struct {
	// 	trustedOrigins []string
	// }
	jwt struct {
		secret string
	}
	apiKey struct {
		gemini string
	}
}

type application struct {
	config         config
	logger         *jsonlog.Logger
	models         data.Models
	wg             sync.WaitGroup
	activeSessions map[string]*llm.ChatSession
	sessionMutex   sync.RWMutex
	geminiClient   *llm.GeminiClient
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment")
	flag.StringVar(&cfg.db.uri, "mongo-uri", "", "Mongo Uri")
	flag.StringVar(&cfg.db.database, "db-name", "url", "Database name")
	flag.StringVar(&cfg.apiKey.gemini, "gemini-api-key", "", "Gemini Api key")
	flag.Uint64Var(&cfg.db.maxPoolSize, "db-max-pool-size", 100, "Max pool size")
	flag.Uint64Var(&cfg.db.minPoolSize, "db-min-pool-size", 10, "Min pool size")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "Max idle time")

	flag.StringVar(&cfg.jwt.secret, "jwt-secret", "", "JWT secret")

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := OpenDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer func() {
		if err := db.Disconnect(context.Background()); err != nil {
			logger.PrintFatal(err, nil)
		}
	}()

	logger.PrintInfo("database connection pool established", nil)

	expvar.NewString("version").Set(version)

	app := &application{
		config:         cfg,
		logger:         logger,
		models:         data.NewModels(db, cfg.db.database),
		activeSessions: make(map[string]*llm.ChatSession),
		geminiClient:   llm.NewGeminiClient(cfg.apiKey.gemini),
	}

	err = app.serve()
	if err != nil {
		log.Fatal(err)
	}
}

func OpenDB(cfg config) (*mongo.Client, error) {
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	clientOptions := options.Client().
		ApplyURI(cfg.db.uri).
		SetMaxPoolSize(cfg.db.maxPoolSize).
		SetMinPoolSize(cfg.db.minPoolSize).
		SetMaxConnIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}
