package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"net/http"
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

		links, err := a.getAllLink(c.Query("Search"))
		if err != nil {
			renderPage(c, "links.tmpl", &gin.H{
				"errorMessage": "Unable to retrieve links",
			})
			log.Errorf("unable to retrieve all links, error: %s", err.Error())
			return
		}

		renderPage(c, "links.tmpl", &gin.H{
			"links": links,
		})
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
			links, _ := a.getAllLink("")
			renderPage(c, "links.tmpl", &gin.H{
				"errorMessage": "Unable to retrieve link",
				"links": links,
			})
			log.Errorf("unable to retrieve link=%s, error: %s", name, err.Error())
			return
		}

		renderPage(c, "link-update.tmpl", &gin.H{
			"name":     link.Name,
			"redirect": link.Redirect,
		})
	}
}

// create a link, required fields are name and redirect
func createLinkController(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get link body
		link, err := getLinkFromContext(c)
		if err != nil {
			log.Errorf("unable to parse request, error: %s", err.Error())
			renderPage(c, "link-create.tmpl", &gin.H{
				"errorMessage": fmt.Sprintf("unable to parse request, error %s", err.Error()),
				"name":         link.Name,
				"redirect":     link.Redirect,
			})
			return
		}

		// Create link
		err = a.createLink(link)
		if err != nil {
			renderPage(c, "link-create.tmpl", &gin.H{
				"errorMessage": fmt.Sprintf("Unable to create link: %s", err.Error()),
				"name":         link.Name,
				"redirect":     link.Redirect,
			})
			log.WithFields(log.Fields{
				"name":     link.Name,
				"redirect": link.Redirect,
			}).Errorf("unable to create link, error: %s", err.Error())
			return
		}

		c.Redirect(http.StatusFound, "/link/all")
	}
}

// update a link, required fields are name and redirect
func updateLinkController(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get link body
		link, err := getLinkFromContext(c)
		if err != nil {
			log.Errorf("unable to parse request, error: %s", err.Error())
			renderPage(c, "link-update.tmpl", &gin.H{
				"errorMessage": fmt.Sprintf("unable to parse request, error %s", err.Error()),
				"name":         link.Name,
				"redirect":     link.Redirect,
			})
			return
		}

		// Update link
		err = a.updateLink(link)
		if err != nil {
			renderPage(c, "link-update.tmpl", &gin.H{
				"errorMessage": fmt.Sprintf("Unable to update link: %s", err.Error()),
				"name":         link.Name,
				"redirect":     link.Redirect,
			})
			log.WithFields(log.Fields{
				"name":     link.Name,
				"redirect": link.Redirect,
			}).Errorf("unable to update link, error: %s", err.Error())
			return
		}

		c.Redirect(http.StatusFound, "/link/all")
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
			links, _ := a.getAllLink("")
			renderPage(c, "links.tmpl", &gin.H{
				"errorMessage": fmt.Sprintf("Unable to delete link, error: %s", err.Error()),
				"links":        links,
			})
			log.Errorf("unable to delete link with name=%s, error: %s", name, err.Error())
			return
		}

		c.Redirect(http.StatusFound, "/link/all")
	}
}
