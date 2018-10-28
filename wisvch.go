package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"os"
)

// DONE: Build redirect
// DONE: Build admin pages
// DONE: Change to forms as create input
// DONE: Change to forms as update input
// DONE: Make teplate for create
// DONE: Make teplate for update
// TODO: Build open connect login
// DONE: Add environment variables for configuration and file settings for local development
// DONE: Change functions to return a handlerfunc which takes the db
// DONE: Move link routes to link.go
// DONE: Move login logic to login.go
// DONE: Return error on non-existing link update

type App struct {
	DB     *gorm.DB
	Config Config
}

type Config struct {
	Host          string
	Port          int
	DBName        string
	DBUser        string
	DBPassword    string
	DBDebug       bool
	EmptyRedirect string
	NotFoundURL   string
}

func main() {
	var wisvch App

	// Bind variables from file to environment for local development
	_, err := os.Stat("wisvch.env")
	if !os.IsNotExist(err) {
		err := godotenv.Load("wisvch.env")
		if err != nil {
			log.Fatalf("unable to read .env file, error: %s", err.Error())
		}
	}

	// Load environment variables for config
	err = envconfig.Process("", &wisvch.Config)
	if err != nil {
		log.Fatalf("unable to parse environment variables, error: %s", err.Error())
	}

	wisvch.DB, err = gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", wisvch.Config.Host, wisvch.Config.Port, wisvch.Config.DBUser, wisvch.Config.DBPassword, wisvch.Config.DBName))
	if err != nil {
		log.Fatalf("unable to connect to database, error: %s", err.Error())
	}
	defer wisvch.DB.Close()

	wisvch.DB.LogMode(wisvch.Config.DBDebug)

	// Automigrate for possible struct updates
	wisvch.DB.AutoMigrate(&Link{})

	r := gin.Default()

	// Load templates
	r.LoadHTMLGlob("./resources/**/*")

	admin := r.Group("/admin")
	{
		admin.GET("/", loginController)
	}

	link := r.Group("/link")
	{
		link.GET("/all", getAllLinkController(wisvch))
		link.GET("/one/:Name", getLinkController(wisvch))

		link.POST("/create", createLinkController(wisvch))
		link.POST("/update/:Name", updateLinkController(wisvch))

		link.GET("/delete/:Name", deleteLinkController(wisvch))
	}

	// If it is an undefined route, perform a redirect
	r.NoRoute(redirect(wisvch))
	r.Run(":8080")
}
