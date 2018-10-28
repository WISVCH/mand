package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
)

func (a App) createLink(l *Link) error {
	return a.DB.Model(&Link{}).
		Create(l).
		Error
}

func (a App) updateLink(l *Link) error {
	var updated int
	err := a.DB.Model(&Link{}).
		Where("name = ?", l.Name).
		Update("redirect", l.Redirect).
		Count(&updated).
		Error
	if err != nil {
		return err
	}
	// If zero records updated, return error
	if updated < 1 {
		return fmt.Errorf("unable to update link with name=%s, link does not exist.", l.Name)
	}
	return nil
}

func getLinkFromContext(c *gin.Context) (*Link, error) {
	link := &Link{}
	err := c.ShouldBind(link)
	if err != nil {
		return nil, err
	}

	// For some reason, the ShouldBind() sets all values to empty values (0001-01-01 00:00:00 +0000 UTC), deletedAt should be nil...
	link.DeletedAt = nil

	u, err := url.Parse(link.Redirect)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		u.Scheme = "https:"
		link.Redirect = u.String()
	}

	return link, nil
}

func renderPage(c *gin.Context, template string, data *gin.H) {
	c.HTML(http.StatusOK, template, data)
}

func (a App) getAllLink(search string) ([]*Link, error) {
	var links []*Link

	searchString := fmt.Sprintf("%%%s%%", search)

	q := a.DB.Model(&Link{})
	if search != "" {
		q = q.Where("name LIKE ? OR redirect LIKE ?", searchString, searchString)
	}
	err := q.Order("name").
		Find(&links).
		Error
	return links, err
}
