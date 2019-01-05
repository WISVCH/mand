package main

import (
	"fmt"
	"net/http"

	"regexp"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var linkNamePattern *regexp.Regexp = regexp.MustCompile("[a-zA-Z-0-9]+")
var searchPattern *regexp.Regexp = regexp.MustCompile("[a-zA-Z-0-9]*")

// Link type containing information for the redirect entry
type Link struct {
	Name     string `gorm:"unique;not null" form:"name" binding:"required"`
	Redirect string `gorm:"not null" form:"redirect" binding:"required"`
}

// get all links, in alphabetical order
func getAllLinkController(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		search := c.Query("search")

		if !searchPattern.Match([]byte(search)) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"errorMessage": "Incorrect search, must use pattern '[a-zA-Z0-9]*'",
			})
			return
		}

		links, err := a.getAllLink(c.Query("search"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"errorMessage": fmt.Sprintf("Unable to retrieve links, error: %s", err),
			})
			log.Errorf("unable to retrieve all links, error: %s", err)
			return
		}

		c.JSON(http.StatusOK, links)
	}
}

// create a link, required fields are name and redirect
func createLinkController(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get link body
		link, err := getLinkFromContext(c)
		if err != nil {
			log.Errorf("unable to parse request, error: %s", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"errorMessage": fmt.Sprintf("Unable to parse request, error: %s", err),
			})
			return
		}

		if !linkNamePattern.Match([]byte(link.Name)) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"errorMessage": "Incorrect link name, must use pattern '[a-zA-Z0-9]+'",
			})
			return
		}

		// Create link
		err = a.createLink(link)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"errorMessage": fmt.Sprintf("Unable to create link, error: %s", err),
			})
			log.WithFields(log.Fields{
				"name":     link.Name,
				"redirect": link.Redirect,
			}).Errorf("unable to create link, error: %s", err)
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

		if !linkNamePattern.Match([]byte(name)) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"errorMessage": "Incorrect link name, must use pattern '[a-zA-Z0-9]+'",
			})
			return
		}

		// Get link body
		link, err := getLinkFromContext(c)
		if err != nil {
			log.Errorf("unable to parse request, error: %s", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"errorMessage": fmt.Sprintf("Unable to parse request, error: %s", err),
			})
			return
		}

		// Update link
		err = a.updateLink(name, link)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"errorMessage": fmt.Sprintf("Unable to update link, error: %s", err),
			})
			log.WithFields(log.Fields{
				"name":     link.Name,
				"redirect": link.Redirect,
			}).Errorf("unable to update link, error: %s", err)
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

		if !linkNamePattern.Match([]byte(name)) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"errorMessage": "Incorrect link name, must use pattern '[a-zA-Z0-9]+'",
			})
			return
		}

		// Delete link on Name
		err := a.DB.Model(&Link{}).
			Where("name = ?", name).
			Delete(&Link{}).
			Error
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"errorMessage": fmt.Sprintf("Unable to delete link, error: %s", err),
			})
			log.Errorf("unable to delete link with name=%s, error: %s", name, err)
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}
