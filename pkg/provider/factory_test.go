package provider

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "ollama provider",
			config: Config{
				Type:    ProviderOllama,
				BaseURL: "http://localhost:11434",
				Model:   "llama3.2",
			},
			wantErr: false,
		},
		{
			name: "openai provider",
			config: Config{
				Type:   ProviderOpenAI,
				APIKey: "sk-test123",
				Model:  "gpt-4o",
			},
			wantErr: false,
		},
		{
			name: "openai provider with base URL (Azure)",
			config: Config{
				Type:    ProviderOpenAI,
				APIKey:  "sk-test123",
				BaseURL: "https://mycompany.openai.azure.com",
				Model:   "gpt-4o",
			},
			wantErr: false,
		},
		{
			name: "gemini provider",
			config: Config{
				Type:   ProviderGemini,
				APIKey: "test-api-key",
				Model:  "gemini-2.5-flash",
			},
			wantErr: false,
		},
		// Skip vertex AI test - requires GCP credentials
		// {
		// 	name: "gemini vertex AI provider",
		// 	config: Config{
		// 		Type:      ProviderGemini,
		// 		ProjectID: "my-project",
		// 		Location:  "us-central1",
		// 		Model:     "gemini-2.5-flash",
		// 	},
		// 	wantErr: false,
		// },
		{
			name: "missing model",
			config: Config{
				Type:    ProviderOllama,
				BaseURL: "http://localhost:11434",
			},
			wantErr: true,
		},
		{
			name: "ollama missing base URL",
			config: Config{
				Type:  ProviderOllama,
				Model: "llama3.2",
			},
			wantErr: true,
		},
		{
			name: "openai missing API key",
			config: Config{
				Type:  ProviderOpenAI,
				Model: "gpt-4o",
			},
			wantErr: true,
		},
		{
			name: "gemini missing API key",
			config: Config{
				Type:  ProviderGemini,
				Model: "gemini-2.5-flash",
			},
			wantErr: true,
		},
		{
			name: "gemini vertex AI missing location",
			config: Config{
				Type:      ProviderGemini,
				ProjectID: "my-project",
				Model:     "gemini-2.5-flash",
			},
			wantErr: true,
		},
		{
			name: "unsupported provider type",
			config: Config{
				Type:  "claude",
				Model: "claude-3",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("New() returned nil provider without error")
			}
		})
	}
}

func TestFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "ollama from env",
			envVars: map[string]string{
				"LLM_PROVIDER":    "ollama",
				"LLM_MODEL":       "llama3.2",
				"OLLAMA_BASE_URL": "http://localhost:11434",
			},
			wantErr: false,
		},
		{
			name: "ollama with default base URL",
			envVars: map[string]string{
				"LLM_PROVIDER": "ollama",
				"LLM_MODEL":    "llama3.2",
			},
			wantErr: false,
		},
		{
			name: "openai from env",
			envVars: map[string]string{
				"LLM_PROVIDER":   "openai",
				"LLM_MODEL":      "gpt-4o",
				"OPENAI_API_KEY": "sk-test123",
			},
			wantErr: false,
		},
		{
			name: "openai with base URL (Azure)",
			envVars: map[string]string{
				"LLM_PROVIDER":    "openai",
				"LLM_MODEL":       "gpt-4o",
				"OPENAI_API_KEY":  "sk-test123",
				"OPENAI_BASE_URL": "https://mycompany.openai.azure.com",
			},
			wantErr: false,
		},
		{
			name: "gemini from env",
			envVars: map[string]string{
				"LLM_PROVIDER":   "gemini",
				"LLM_MODEL":      "gemini-2.5-flash",
				"GEMINI_API_KEY": "test-api-key",
			},
			wantErr: false,
		},
		// Skip vertex AI test - requires GCP credentials
		// {
		// 	name: "gemini vertex AI from env",
		// 	envVars: map[string]string{
		// 		"LLM_PROVIDER":      "gemini",
		// 		"LLM_MODEL":         "gemini-2.5-flash",
		// 		"GEMINI_PROJECT_ID": "my-project",
		// 		"GEMINI_LOCATION":   "us-central1",
		// 	},
		// 	wantErr: false,
		// },
		{
			name: "missing LLM_PROVIDER",
			envVars: map[string]string{
				"LLM_MODEL": "llama3.2",
			},
			wantErr: true,
		},
		{
			name: "missing LLM_MODEL",
			envVars: map[string]string{
				"LLM_PROVIDER": "ollama",
			},
			wantErr: true,
		},
		{
			name: "unsupported provider",
			envVars: map[string]string{
				"LLM_PROVIDER": "claude",
				"LLM_MODEL":    "claude-3",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			clearEnvVars()

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer clearEnvVars()

			provider, err := FromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("FromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("FromEnv() returned nil provider without error")
			}
		})
	}
}

func clearEnvVars() {
	os.Unsetenv("LLM_PROVIDER")
	os.Unsetenv("LLM_MODEL")
	os.Unsetenv("OLLAMA_BASE_URL")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_BASE_URL")
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("GEMINI_PROJECT_ID")
	os.Unsetenv("GEMINI_LOCATION")
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid ollama config",
			config: Config{
				Type:    ProviderOllama,
				BaseURL: "http://localhost:11434",
				Model:   "llama3.2",
			},
			wantErr: false,
		},
		{
			name: "valid openai config",
			config: Config{
				Type:   ProviderOpenAI,
				APIKey: "sk-test",
				Model:  "gpt-4o",
			},
			wantErr: false,
		},
		{
			name: "valid gemini config",
			config: Config{
				Type:   ProviderGemini,
				APIKey: "test-key",
				Model:  "gemini-2.5-flash",
			},
			wantErr: false,
		},
		{
			name: "invalid - empty model",
			config: Config{
				Type:    ProviderOllama,
				BaseURL: "http://localhost:11434",
			},
			wantErr: true,
		},
		{
			name: "invalid - empty provider type",
			config: Config{
				Model: "llama3.2",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
