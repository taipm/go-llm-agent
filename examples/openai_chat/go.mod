module example/openai_chat

go 1.25.3

replace github.com/taipm/go-llm-agent => ../..

require github.com/taipm/go-llm-agent v0.0.0-00010101000000-000000000000

require (
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/openai/openai-go/v3 v3.6.1 // indirect
	github.com/tidwall/gjson v1.14.4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
)
