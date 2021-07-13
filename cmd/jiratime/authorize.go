package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/smlx/jiratime/internal/client"
	"github.com/smlx/jiratime/internal/config"
	"golang.org/x/oauth2"
)

// AuthorizeCmd represents the `authorize` or `auth` command.
type AuthorizeCmd struct{}

func startRedirectServer(ctx context.Context, state string, c chan<- string) {
	mux := http.NewServeMux()
	// Create a new redirect route
	mux.HandleFunc("/oauth/redirect", func(w http.ResponseWriter, r *http.Request) {
		// First, we need to get the value of the `code` query param
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not parse query: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		c <- r.FormValue("code")
		w.Write([]byte("Authorization successful. You may now close this page."))
	})
	s := &http.Server{
		Addr:        ":8080",
		Handler:     mux,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}
	log.Println(s.ListenAndServe())
}

// randomHex returns a string consisting of n hex-encoded random bytes. Because
// the string is hex encoded its length will be 2*n.
func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("couldn't read random bytes: %v", err)
	}
	return hex.EncodeToString(bytes), nil
}

// Run the Authorize command.
func (cmd *AuthorizeCmd) Run() error {
	ctx, cancel := getContext(30 * time.Second)
	defer cancel()
	// read the config file to get the oauth2 clientID and secret
	auth, err := config.ReadAuth("./auth.yml")
	if err != nil {
		return fmt.Errorf("couldn't load config: %v", err)
	}
	// sanity check configuration
	if auth == nil {
		return fmt.Errorf("couldn't find oauth2 configuration")
	}
	if auth.ClientID == "" || auth.Secret == "" {
		return fmt.Errorf("missing ClientID or Secret in oauth2 configuration")
	}
	// generate a random state
	state, err := randomHex(16)
	if err != nil {
		return fmt.Errorf("couldn't generate state: %v", err)
	}
	// get the OAuth2 config object
	conf := client.GetOAuth2Config(auth)
	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL(state,
		oauth2.SetAuthURLParam("audience", "api.atlassian.com"),
		oauth2.SetAuthURLParam("prompt", "consent"))
	fmt.Printf("Visit this URL to authorize jiratime: %v", url)
	// start the server to handle the redirect after authorization
	c := make(chan string, 1)
	go startRedirectServer(ctx, state, c)
	var code string
	select {
	case code = <-c:
		// received code
	case <-ctx.Done():
		return fmt.Errorf("timed out waiting for code")
	}
	// use a custom HTTP client with a reasonable timeout
	tok, err := conf.Exchange(
		context.WithValue(ctx, oauth2.HTTPClient,
			&http.Client{Timeout: 4 * time.Second}), code)
	if err != nil {
		log.Fatal(err)
	}
	auth.Token = tok
	if err = config.WriteAuth(auth, "./auth.yml"); err != nil {
		return fmt.Errorf("couldn't write config: %v", err)
	}
	return nil
}
