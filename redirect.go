package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func redirect(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := strings.Split(c.Request.RequestURI[1:], "/")[0]

		if path == "" || path == "/" {
			c.Redirect(http.StatusFound, a.Config.EmptyRedirect)
		}

		var link Link
		err := a.DB.Model(&Link{}).
			Where("name = ?", path).
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

		c.Redirect(http.StatusFound, link.Redirect)
	}
}
