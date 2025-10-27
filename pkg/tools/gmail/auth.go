package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmailapi "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// GmailConfig contains configuration for Gmail API access
type GmailConfig struct {
	// CredentialsFile is the path to OAuth2 credentials JSON file
	// Download from Google Cloud Console: https://console.cloud.google.com/apis/credentials
	CredentialsFile string

	// TokenFile is the path where OAuth2 token will be cached
	TokenFile string

	// Scopes are the Gmail API scopes to request
	// Default: gmailapi.GmailModifyScope (read and send)
	Scopes []string
}

// DefaultGmailConfig provides sensible defaults for Gmail API
var DefaultGmailConfig = GmailConfig{
	CredentialsFile: "credentials.json",
	TokenFile:       "token.json",
	Scopes: []string{
		gmailapi.GmailModifyScope, // Read, send, delete, and manage email
	},
}

// AuthHelper manages Gmail API authentication
type AuthHelper struct {
	config GmailConfig
}

// NewAuthHelper creates a new Gmail authentication helper
func NewAuthHelper(config GmailConfig) *AuthHelper {
	return &AuthHelper{config: config}
}

// GetService creates an authenticated Gmail service
func (h *AuthHelper) GetService(ctx context.Context) (*gmailapi.Service, error) {
	// Read credentials file
	credBytes, err := os.ReadFile(h.config.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file %s: %w (download from https://console.cloud.google.com/apis/credentials)", h.config.CredentialsFile, err)
	}

	// Parse OAuth2 config
	oauthConfig, err := google.ConfigFromJSON(credBytes, h.config.Scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %w", err)
	}

	// Get OAuth2 token (from cache or user authorization)
	token, err := h.getToken(oauthConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to get OAuth2 token: %w", err)
	}

	// Create Gmail service
	client := oauthConfig.Client(ctx, token)
	service, err := gmailapi.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail service: %w", err)
	}

	return service, nil
}

// getToken retrieves a token from cache or prompts user for authorization
func (h *AuthHelper) getToken(config *oauth2.Config) (*oauth2.Token, error) {
	// Try to load cached token
	token, err := h.loadTokenFromFile()
	if err == nil {
		return token, nil
	}

	// Token not found or invalid, need user authorization
	// This will only work in interactive environments
	// For production, you should pre-authorize and cache the token
	return nil, fmt.Errorf("no cached token found at %s. Please run authorization flow first (see documentation)", h.config.TokenFile)
}

// loadTokenFromFile loads OAuth2 token from file
func (h *AuthHelper) loadTokenFromFile() (*oauth2.Token, error) {
	file, err := os.Open(h.config.TokenFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(file).Decode(token)
	return token, err
}

// SaveToken saves OAuth2 token to file
func (h *AuthHelper) SaveToken(token *oauth2.Token) error {
	file, err := os.Create(h.config.TokenFile)
	if err != nil {
		return fmt.Errorf("unable to create token file: %w", err)
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(token)
}

// AuthorizeInteractive performs interactive OAuth2 flow (for initial setup)
// This should be called separately to set up credentials
func (h *AuthHelper) AuthorizeInteractive(ctx context.Context) (*oauth2.Token, error) {
	credBytes, err := os.ReadFile(h.config.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w", err)
	}

	config, err := google.ConfigFromJSON(credBytes, h.config.Scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %w", err)
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n", authURL)
	fmt.Print("Enter authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token: %w", err)
	}

	// Save token for future use
	if err := h.SaveToken(token); err != nil {
		return nil, fmt.Errorf("unable to save token: %w", err)
	}

	return token, nil
}

// ValidateCredentials checks if credentials and token are available
func (h *AuthHelper) ValidateCredentials() error {
	// Check credentials file
	if _, err := os.Stat(h.config.CredentialsFile); err != nil {
		return fmt.Errorf("credentials file not found: %s (download from Google Cloud Console)", h.config.CredentialsFile)
	}

	// Check token file
	if _, err := os.Stat(h.config.TokenFile); err != nil {
		return fmt.Errorf("token file not found: %s (run authorization flow first)", h.config.TokenFile)
	}

	return nil
}
