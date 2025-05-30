package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/joho/godotenv"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func NewAuth() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	store := sessions.NewCookieStore([]byte(securecookie.GenerateRandomKey(32)))
	store.MaxAge(86400 * 30)

	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = false
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	goth.UseProviders(google.New(googleClientID, googleClientSecret, "https://backend-ph82.onrender.com/auth/google/callback"))
}
