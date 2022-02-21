package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var connectConfig oauth2.Config
var verifier *oidc.IDTokenVerifier
var allowedGroup string
var states = map[string]bool{}

func connect(URL, clientID, clientSecret, redirectURL, group string) {
	ctx := context.Background()

	allowedGroup = group

	var err error
	provider, err := oidc.NewProvider(ctx, URL)
	if err != nil {
		log.Fatalf("unable to create new authentication provider, error: %s", err)
	}

	verifier = provider.Verifier(&oidc.Config{ClientID: clientID})

	// Configure an OpenID Connect aware OAuth2 client.
	connectConfig = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "ldap"},
	}

}

func connectMiddleware(a *App) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("authorization")
		auth := strings.Split(authHeader, " ")

		if len(auth) != 2 || auth[0] != "Bearer" {
			log.Errorf("Wrong authorization header, was %s", authHeader)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"errorMessage": "Incorrect authorization header",
			})
			return
		}

		if checkAuth(auth[1]) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"errorMessage": "Missing authentication",
		})
	}
}

func loginController(a *App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get new random string
		state := randSeq(30)
		// Store the state for the user
		states[state] = true

		c.Redirect(http.StatusFound, connectConfig.AuthCodeURL(state))
	}
}

func callbackController(a *App) gin.HandlerFunc {
	return func(c *gin.Context) {
		state := c.Query("state")
		_, ok := states[state]
		if !ok {
			log.Errorf("state was not found, possible CSRF attack, state: %s", state)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := connectConfig.Exchange(context.TODO(), c.Query("code"))
		if err != nil {
			log.Errorf("unable to exchange token \"%s\", error: %s", c.Query("code"), err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			log.Errorf("unable to get id_token from login")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if checkAuth(rawIDToken) {
			c.JSON(http.StatusOK, gin.H{
				"token": rawIDToken,
			})
		} else {
			_, err := c.Writer.WriteString("permission denied")
			if err != nil {
				log.Errorf("unable to write response to user")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			c.Status(http.StatusUnauthorized)
		}
	}
}

func checkAuth(rawIDToken string) bool {
	idToken, err := verifier.Verify(context.TODO(), rawIDToken)
	if err != nil {
		log.Errorf("unable to verify id_token, error: %s", err)
		return false
	}

	var claims struct {
		Groups []string `json:"ldap_groups"`
	}
	if err := idToken.Claims(&claims); err != nil {
		log.Errorf("unable to read ldap_groups from id_token, error: %s", err)
		return false
	}

	for _, group := range claims.Groups {
		if group == allowedGroup {
			return true
		}
	}
	return false
}
