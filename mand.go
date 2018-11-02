package main

import (
	"fmt"
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
	DBHost        string
	DBPort        int
	DBName        string
	DBUser        string
	DBPassword    string
	DBDebug       bool
	EmptyRedirect string
	NotFoundURL   string
}

func main() {
	var mand App

	// Bind variables from file to environment for local development
	_, err := os.Stat("mand.env")
	if !os.IsNotExist(err) {
		err := godotenv.Load("mand.env")
		if err != nil {
			log.Fatalf("unable to read .env file, error: %s", err.Error())
		}
	}

	// Load environment variables for config
	err = envconfig.Process("", &mand.Config)
	if err != nil {
		log.Fatalf("unable to parse environment variables, error: %s", err.Error())
	}

	mand.DB, err = gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", mand.Config.DBHost, mand.Config.DBPort, mand.Config.DBUser, mand.Config.DBPassword, mand.Config.DBName))
	if err != nil {
		log.Fatalf("unable to connect to database, error: %s", err.Error())
	}
	defer mand.DB.Close()

	mand.DB.LogMode(mand.Config.DBDebug)

	// Automigrate for possible struct updates
	mand.DB.AutoMigrate(&Link{})

	r := gin.Default()

	// Load templates
	r.LoadHTMLGlob("./resources/**/*")

	admin := r.Group("/admin")
	{
		admin.GET("/", loginController)
	}

	link := r.Group("/link")
	{
		link.GET("/all", getAllLinkController(mand))
		link.GET("/one/:Name", getLinkController(mand))

		link.POST("/create", createLinkController(mand))
		link.POST("/update/:Name", updateLinkController(mand))

		link.GET("/delete/:Name", deleteLinkController(mand))
	}

	// If it is an undefined route, perform a redirect
	r.NoRoute(redirect(mand))
	r.Run(":8080")
}
