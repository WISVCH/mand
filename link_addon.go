package main

import (
	"fmt"
	"net/url"

	"strings"

	"github.com/gin-gonic/gin"
)

func (a App) createLink(l *Link) error {
	return a.DB.Model(&Link{}).
		Create(l).
		Error
}

func (a App) updateLink(name string, l *Link) error {
	err := a.DB.Model(&Link{}).
		Where("name = ?", name).
		Update(&l).
		Error
	if err != nil {
		return err
	}
	return nil
}

func getLinkFromContext(c *gin.Context) (*Link, error) {
	link := &Link{}
	err := c.ShouldBindJSON(link)
	if err != nil {
		return nil, err
	}

	// Lowercase name of the link
	link.Name = strings.ToLower(link.Name)

	// Check correctness of the destination url
	u, err := url.Parse(link.Redirect)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		u.Scheme = "https"
		link.Redirect = u.String()
	}

	return link, nil
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
