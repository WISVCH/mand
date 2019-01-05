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
var provider oidc.Provider
var verifier *oidc.IDTokenVerifier
var allowedGroup string

func connect(URL, clientID, clientSecret, redirectURL, group string) {
	ctx := context.Background()

	allowedGroup = group

	provider, err := oidc.NewProvider(ctx, URL)
	if err != nil {
		log.Fatalf("unable to create new authentication provider, error: %s", err.Error())
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

func connectMiddleware(a App) gin.HandlerFunc {
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

func loginController(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// change to hash of the session or some other sort unique session identifiable data for the user to avoid csrf attacks
		c.Redirect(http.StatusFound, connectConfig.AuthCodeURL("login"))
	}
}

func callbackController(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := connectConfig.Exchange(context.TODO(), c.Query("code"))
		if err != nil {
			log.Errorf("unable to exchange token \"%s\", error: %s", c.Query("code"), err.Error())
			return
		}

		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			log.Errorf("unable to get id_token from login")
			return
		}

		if checkAuth(rawIDToken) {
			c.JSON(http.StatusOK, gin.H{
				"token": rawIDToken,
			})
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

func checkAuth(rawIDToken string) bool {
	idToken, err := verifier.Verify(context.TODO(), rawIDToken)
	if err != nil {
		log.Errorf("unable to verify id_token, error: %s", err.Error())
		return false
	}

	var claims struct {
		Groups []string `json:"ldap_groups"`
	}
	if err := idToken.Claims(&claims); err != nil {
		log.Errorf("unable to read ldap_groups from id_token, error: %s", err.Error())
		return false
	}

	for _, group := range claims.Groups {
		if group == allowedGroup {
			return true
		}
	}
	return false
}
