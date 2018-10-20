package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func login(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "admin login here.",
	})
}
