package main

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func redirect(a *App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse url to split it into components
		parsedURL, err := url.ParseRequestURI(c.Request.RequestURI)
		if err != nil {
			log.Errorf("unable to parse request uri, error: %s", err)
		}

		// Only use the path for redirecting
		path := strings.Split(parsedURL.Path[1:], "/")[0]

		if path == "" || path == "/" {
			c.Redirect(http.StatusFound, a.Config.EmptyRedirect)
		}

		linkPath := strings.ToLower(path)

		var link Link
		err = a.DB.Model(&Link{}).
			Where("name = ?", linkPath).
			Find(&link).
			Error
		if err != nil {
			log.Errorf("unable to retrieve link: '%s', error: %s", c.Request.RequestURI, err)
			if gorm.IsRecordNotFoundError(err) {
				// 404 redirect
				c.Redirect(http.StatusFound, a.Config.NotFoundURL)
			} else {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}

		err = a.DB.Exec("UPDATE links SET visits = visits + 1 WHERE name = ?", linkPath).Error
		if err != nil {
			log.Errorf("unable to update visiting counter: '%s', error: %s", c.Request.RequestURI, err)
		}

		c.Redirect(http.StatusFound, link.Redirect)
	}
}
