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
// TODO: Build admin pages
// TODO: Build open connect login
// TODO: Add environment variables for configuration and file settings for local development
// DONE: Change functions to return a handlerfunc which takes the db
// DONE: Move link routes to link.go
// DONE: Move login logic to login.go
// DONE: Return error on non-existing link update

// Link type containing information for the redirect entry
type Link struct {
	gorm.Model
	Name     string `gorm:"unique;not null" json:"name" binding:"required"`
	Redirect string `gorm:"not null" json:"redirect" binding:"required"`
}

type App struct {
	DB     *gorm.DB
	Config Config
}

type Config struct {
	Host        string
	Port        int
	DBName      string
	DBUser      string
	DBPassword  string
	NotFoundURL string
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

	// Automigrate for possible struct updates
	wisvch.DB.AutoMigrate(&Link{})

	r := gin.Default()
	admin := r.Group("/admin")
	{
		admin.GET("/", login)
	}

	link := r.Group("/link")
	{
		link.GET("", getAllLink(wisvch))
		link.POST("", createLink(wisvch))
		link.PATCH("", updateLink(wisvch))
		link.DELETE("/:ID", deleteLink(wisvch))
	}

	// If it is an undefined route, perform a redirect
	r.NoRoute(redirect(wisvch))
	r.Run(":8080")
}
