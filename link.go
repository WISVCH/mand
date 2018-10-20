package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strconv"
)

// get all links, in alphabetical order
func getAllLink(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var links []*Link

		err := a.DB.Model(&Link{}).
			Order("name").
			Find(&links).
			Error
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Errorf("unable to retrieve all links, error: %s", err.Error())
			return
		}

		c.JSON(http.StatusOK, links)
	}
}

// create action for a link, required fields are name and redirect
func createLink(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get link body
		link, err := getLinkFromContext(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse and correct url if needed
		err = checkURL(link)
		if err != nil {
			log.Errorf("unable to check url: %s, error: %s", link.Redirect, err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url."})
			return
		}

		// Create link in DB
		err = a.DB.Model(&Link{}).
			Create(&link).
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
}

// update action for a link
func updateLink(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get link body
		link, err := getLinkFromContext(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse and correct url if needed
		err = checkURL(link)
		if err != nil {
			log.Errorf("unable to check url: %s, error: %s", link.Redirect, err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url."})
			return
		}

		var updated int
		// Update link in DB
		err = a.DB.Model(&Link{}).
			Where("name = ?", link.Name).
			Update(Link{Redirect: link.Redirect}).
			Count(&updated).
			Error
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.WithFields(log.Fields{
				"name":     link.Name,
				"redirect": link.Redirect,
			}).Errorf("unable to update link, error: %s", err.Error())
			return
		}

		// If zero records updated, return Not Found status
		if updated < 1 {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "link updated.",
		})
	}
}

// delete a link with path parameter ID
func deleteLink(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get path param ID
		id, err := strconv.Atoi(c.Param("ID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Delete link on ID
		err = a.DB.Model(&Link{}).
			Where("id = ?", id).
			Unscoped().
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
}

func getLinkFromContext(c *gin.Context) (Link, error) {
	var link Link
	err := c.ShouldBindJSON(&link)
	return link, err
}

func checkURL(link Link) error {
	u, err := url.Parse(link.Redirect)
	if err != nil {
		return err
	}

	if u.Scheme == "" {
		u.Scheme = "https:"
		link.Redirect = u.String()
	}

	return nil
}
