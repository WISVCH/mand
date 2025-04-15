package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type App struct {
	DB     *gorm.DB
	Config Config
}

type Config struct {
	DBHost             string
	DBPort             int
	DBName             string
	DBUser             string
	DBPassword         string
	DBOptions          string
	DBConnectionString string

	DBDebug bool

	ConnectURL      string
	ConnectClientID string
	ClientSecret    string
	RedirectURL     string
	AllowedGroup    string

	EmptyRedirect string
	NotFoundURL   string
}

func main() {
	// Load config
	mand := loadEnv()

	// Connects to database & automigrates structs
	connectDB(mand)
    defer func() {
        if err := mand.DB.Close(); err != nil {
            log.Errorf("error closing database: %v", err)
        }
    }()

	router := getHandler(mand)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("run exited with error: %s", err)
	}
}

func loadEnv() *App {
	var conf App

	// Bind variables from file to environment for local development
	_, err := os.Stat("mand.env")
	if !os.IsNotExist(err) {
		err := godotenv.Load("mand.env")
		if err != nil {
			log.Fatalf("unable to read .env file, error: %s", err)
		}
	}

	// Load environment variables for config
	err = envconfig.Process("", &conf.Config)
	if err != nil {
		log.Fatalf("unable to parse environment variables, error: %s", err)
	}

	return &conf
}

func connectDB(mand *App) {
	// Use the config variables, otherwise use the connectionString
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s %s", mand.Config.DBHost, mand.Config.DBPort, mand.Config.DBUser, mand.Config.DBPassword, mand.Config.DBName, mand.Config.DBOptions)
	if mand.Config.DBConnectionString != "" {
		connectionString = mand.Config.DBConnectionString
	}

	// Connect to database
	var err error
	mand.DB, err = gorm.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("unable to connect to database, error: %s", err)
	}

	mand.DB.LogMode(mand.Config.DBDebug)

	// Automigrate for possible struct updates
	mand.DB.AutoMigrate(&Link{})
}

func getHandler(mand *App) *gin.Engine {
	// Same as the Default() instance without the logger
	r := gin.New()
	r.Use(gin.Recovery())

	// Set up health check endpoint
	r.GET("/healthz", func(c *gin.Context) {
		if err := mand.DB.DB().Ping(); err != nil {
			log.Printf("database ping failed: %s", err)
			c.String(http.StatusInternalServerError, "database ping failed")
		} else {
			c.String(http.StatusOK, "ok")
		}
	})

	// Static file serving
	r.StaticFS("/admin", http.Dir("web"))

	link := r.Group("/link")
	{
		link.GET("/", getAllLinkController(mand))
		link.POST("/", createLinkController(mand))
		link.PATCH("/:Name", updateLinkController(mand))
		link.DELETE("/:Name", deleteLinkController(mand))
	}

	// If it is an undefined route, perform a redirect
	r.NoRoute(redirect(mand))

	return r
}
