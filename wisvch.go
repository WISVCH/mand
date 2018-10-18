package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

// TODO: Build open connect login
// TODO: Build redirect
// TODO: Build admin pages
// TODO: Add environment variables for configuration and file settings for local development
// TODO: Change functions to return a handlerfunc which takes the db
// TODO: Move link routes to link file
// TODO: Move login logic
// TODO: Return error on non-existing link update

// Link type containing information for the redirect entry
type Link struct {
	gorm.Model
	Name     string `gorm:"unique;not null" json:"name" binding:"required"`
	Redirect string `gorm:"not null" json:"redirect" binding:"required"`
}

var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open("postgres", "host=localhost port=5432 user=postgres dbname=wisvch password=postgres sslmode=disable")
	if err != nil {
		log.Fatalf("unable to connect to database, error: %s", err.Error())
	}
	defer db.Close()

	// Automigrate for possible struct updates
	db.AutoMigrate(&Link{})

	r := gin.Default()
	admin := r.Group("/admin")
	{
		admin.GET("/", login)
	}

	link := r.Group("/link")
	{
		link.GET("", getAllLink)
		link.POST("", createLink)
		link.PATCH("", updateLink)
		link.DELETE("/:ID", deleteLink)
	}

	r.NoRoute(redirect)
	r.Run(":8080")
}

func redirect(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Rest of the pages.",
		"route":   c.Request.RequestURI,
	})
}

func login(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "admin login here.",
	})
}

func getAllLink(c *gin.Context) {
	var links []*Link

	err := db.Order("name").
		Find(&links).
		Error
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Errorf("unable to retrieve all links, error: %s", err.Error())
		return
	}

	c.JSON(http.StatusOK, links)
}

// create action for a link, required fields are name and redirect
func createLink(c *gin.Context) {
	// Get link body
	link, err := getLinkFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create link
	err = db.Create(&link).
		Error
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.WithFields(log.Fields{
			"name":     link.Name,
			"redirect": link.Redirect,
		}).Errorf("unable to create link, error: %s", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "link created.",
		"ID":       link.ID,
		"name":     link.Name,
		"redirect": link.Redirect,
	})
}

// update action for a link
func updateLink(c *gin.Context) {
	// Get link body
	link, err := getLinkFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update link in DB
	err = db.Model(&Link{}).
		Where("name = ?", link.Name).
		Update(Link{Redirect: link.Redirect}).
		Error
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.WithFields(log.Fields{
			"name":     link.Name,
			"redirect": link.Redirect,
		}).Errorf("unable to update link, error: %s", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "link updated.",
	})
}

func deleteLink(c *gin.Context) {
	// Get path param ID
	id, err := strconv.Atoi(c.Param("ID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Delete link on ID
	err = db.Where("id = ?", id).
		Delete(&Link{}).
		Error
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Errorf("unable to delete link with id=%d, error: %s", id, err.Error())
		return
	}


	c.JSON(http.StatusOK, gin.H{
		"message": "link deleted.",
	})
}

func getLinkFromContext(c *gin.Context) (Link, error) {
	var link Link
	err := c.ShouldBindJSON(&link)
	return link, err
}
