package main

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var connectConfig oauth2.Config
var provider oidc.Provider
var verifier *oidc.IDTokenVerifier

func connect(URL string, clientID string, clientSecret string, redirectURL string) {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, URL)
	if err != nil {
		log.Fatalf("unable to create new authentication provider, error: %s", err.Error())
	}

	verifier = provider.Verifier(&oidc.Config{ClientID: "2e7726ae-1547-4a95-a61f-f84567578808"})

	// Configure an OpenID Connect aware OAuth2 client.
	connectConfig = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "ldap", "ldap_groups"},
	}

	url := connectConfig.AuthCodeURL("login")
	spew.Dump(url)
}

func loginController(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Redirect(http.StatusFound, connectConfig.AuthCodeURL("login"))
	}
}

func callbackController(a App) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := connectConfig.Exchange(context.TODO(), c.Query("code"))
		if err != nil {
			log.Errorf("unable to exchange token, error: %s", err.Error())
			return
		}

		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			log.Errorf("unable to get id_token from login")
			return
		}
		idToken, err := verifier.Verify(context.TODO(), rawIDToken)
		if err != nil {
			log.Errorf("unable to verify id_token, error: %s", err.Error())
			return
		}

		var claims struct {
			Groups []string `json:"ldap_groups"`
		}
		if err := idToken.Claims(&claims); err != nil {
			log.Errorf("unable to read ldap_groups from id_token, error: %s", err.Error())
		}
		spew.Dump(claims)
	}
}
