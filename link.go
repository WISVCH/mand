package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// Link type containing information for the redirect entry
type Link struct {
	gorm.Model
	Name     string `gorm:"unique;not null" form:"name" binding:"required"`
	Redirect string `gorm:"not null" form:"redirect" binding:"required"`
}

// get all links, in alphabetical order
func getAllLinkController(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		links, err := a.getAllLink(c.GetHeader("X-Search"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"errorMessage": fmt.Sprintf("Unable to retrieve links, error: %s", err.Error()),
			})
			log.Errorf("unable to retrieve all links, error: %s", err.Error())
			return
		}

		c.JSON(http.StatusOK, links)
	}
}

// get link
func getLinkController(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("Name")

		var link Link
		err := a.DB.Model(&Link{}).
			Where("name = ?", name).
			Find(&link).
			Error
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"errorMessage": fmt.Sprintf("Unable to retrieve link, error: %s", err.Error()),
			})
			log.Errorf("unable to retrieve link: %s, error: %s", name, err.Error())
			return
		}

		c.JSON(http.StatusOK, link)
	}
}

// create a link, required fields are name and redirect
func createLinkController(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get link body
		link, err := getLinkFromContext(c)
		if err != nil {
			log.Errorf("unable to parse request, error: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"errorMessage": fmt.Sprintf("Unable to parse request, error: %s", err.Error()),
			})
			return
		}

		// Create link
		err = a.createLink(link)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"errorMessage": fmt.Sprintf("Unable to create link, error: %s", err.Error()),
			})
			log.WithFields(log.Fields{
				"name":     link.Name,
				"redirect": link.Redirect,
			}).Errorf("unable to create link, error: %s", err.Error())
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}

// update a link, required fields are name and redirect
func updateLinkController(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get path param Name
		name := c.Param("Name")

		// Get link body
		link, err := getLinkFromContext(c)
		if err != nil {
			log.Errorf("unable to parse request, error: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"errorMessage": fmt.Sprintf("Unable to parse request, error: %s", err.Error()),
			})
			return
		}

		// Update link
		err = a.updateLink(name, link)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"errorMessage": fmt.Sprintf("Unable to update link, error: %s", err.Error()),
			})
			log.WithFields(log.Fields{
				"name":     link.Name,
				"redirect": link.Redirect,
			}).Errorf("unable to update link, error: %s", err.Error())
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}

// delete a link with path parameter Name
func deleteLinkController(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get path param Name
		name := c.Param("Name")

		// Delete link on Name
		err := a.DB.Model(&Link{}).
			Where("name = ?", name).
			Unscoped().
			Delete(&Link{}).
			Error
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"errorMessage": fmt.Sprintf("Unable to delete link, error: %s", err.Error()),
			})
			log.Errorf("unable to delete link with name=%s, error: %s", name, err.Error())
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}
