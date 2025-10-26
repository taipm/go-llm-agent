package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/taipm/go-llm-agent/pkg/provider/gemini"
	"github.com/taipm/go-llm-agent/pkg/provider/ollama"
	"github.com/taipm/go-llm-agent/pkg/provider/openai"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// ProviderType represents the type of LLM provider
type ProviderType string

const (
	// ProviderOllama is the Ollama provider type
	ProviderOllama ProviderType = "ollama"
	// ProviderOpenAI is the OpenAI provider type
	ProviderOpenAI ProviderType = "openai"
	// ProviderGemini is the Google Gemini provider type
	ProviderGemini ProviderType = "gemini"
)

// Config holds the configuration for creating a provider
type Config struct {
	// Type specifies which provider to use (ollama, openai, gemini)
	Type ProviderType

	// APIKey is required for OpenAI and Gemini providers
	APIKey string

	// BaseURL is required for Ollama (e.g., "http://localhost:11434")
	// Optional for OpenAI (for Azure OpenAI or custom endpoints)
	BaseURL string

	// Model specifies the model to use (e.g., "llama3.2", "gpt-4o", "gemini-2.5-flash")
	Model string

	// ProjectID is required for Gemini Vertex AI
	ProjectID string

	// Location is required for Gemini Vertex AI (e.g., "us-central1")
	Location string
}

// New creates a new provider based on the configuration
func New(config Config) (types.LLMProvider, error) {
	// Validate config
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	switch config.Type {
	case ProviderOllama:
		return ollama.New(config.BaseURL, config.Model), nil

	case ProviderOpenAI:
		if config.BaseURL != "" {
			// Azure OpenAI or custom endpoint
			return openai.NewWithBaseURL(config.APIKey, config.Model, config.BaseURL), nil
		}
		return openai.New(config.APIKey, config.Model), nil

	case ProviderGemini:
		ctx := context.Background()
		if config.ProjectID != "" && config.Location != "" {
			// Vertex AI
			return gemini.NewWithVertexAI(ctx, config.ProjectID, config.Location, config.Model)
		}
		// Gemini API
		return gemini.New(ctx, config.APIKey, config.Model)

	default:
		return nil, fmt.Errorf("unsupported provider type: %s", config.Type)
	}
}

// FromEnv creates a provider based on environment variables
// Environment variables:
//   - LLM_PROVIDER: ollama, openai, or gemini
//   - LLM_MODEL: model name
//   - OLLAMA_BASE_URL: Ollama base URL (default: http://localhost:11434)
//   - OPENAI_API_KEY: OpenAI API key
//   - OPENAI_BASE_URL: Optional OpenAI base URL (for Azure)
//   - GEMINI_API_KEY: Gemini API key
//   - GEMINI_PROJECT_ID: Gemini Vertex AI project ID
//   - GEMINI_LOCATION: Gemini Vertex AI location
func FromEnv() (types.LLMProvider, error) {
	providerType := os.Getenv("LLM_PROVIDER")
	if providerType == "" {
		return nil, fmt.Errorf("LLM_PROVIDER environment variable is required")
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		return nil, fmt.Errorf("LLM_MODEL environment variable is required")
	}

	config := Config{
		Type:  ProviderType(providerType),
		Model: model,
	}

	switch config.Type {
	case ProviderOllama:
		config.BaseURL = os.Getenv("OLLAMA_BASE_URL")
		if config.BaseURL == "" {
			config.BaseURL = "http://localhost:11434"
		}

	case ProviderOpenAI:
		config.APIKey = os.Getenv("OPENAI_API_KEY")
		config.BaseURL = os.Getenv("OPENAI_BASE_URL") // Optional

	case ProviderGemini:
		config.APIKey = os.Getenv("GEMINI_API_KEY")
		config.ProjectID = os.Getenv("GEMINI_PROJECT_ID")   // Optional (for Vertex AI)
		config.Location = os.Getenv("GEMINI_LOCATION")       // Optional (for Vertex AI)

	default:
		return nil, fmt.Errorf("unsupported LLM_PROVIDER: %s (must be ollama, openai, or gemini)", providerType)
	}

	return New(config)
}

// validateConfig validates the provider configuration
func validateConfig(config Config) error {
	if config.Model == "" {
		return fmt.Errorf("model is required")
	}

	switch config.Type {
	case ProviderOllama:
		if config.BaseURL == "" {
			return fmt.Errorf("base URL is required for Ollama provider")
		}

	case ProviderOpenAI:
		if config.APIKey == "" {
			return fmt.Errorf("API key is required for OpenAI provider")
		}

	case ProviderGemini:
		// For Vertex AI, need ProjectID and Location
		if config.ProjectID != "" || config.Location != "" {
			if config.ProjectID == "" {
				return fmt.Errorf("project ID is required for Gemini Vertex AI")
			}
			if config.Location == "" {
				return fmt.Errorf("location is required for Gemini Vertex AI")
			}
		} else {
			// For Gemini API, need API key
			if config.APIKey == "" {
				return fmt.Errorf("API key is required for Gemini provider")
			}
		}

	default:
		return fmt.Errorf("invalid provider type: %s", config.Type)
	}

	return nil
}
