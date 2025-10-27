# Copy & Test in 60 Seconds

Muốn test thư viện này ở dự án khác? Làm theo 3 bước:

## Bước 1: Copy Example

```bash
# Copy toàn bộ thư mục standalone_demo sang dự án của bạn
cp -r /path/to/go-llm-agent/examples/standalone_demo /path/to/your-project/test-agent

cd /path/to/your-project/test-agent
```

## Bước 2: Setup

```bash
# Initialize Go module
go mod init test-agent

# Copy environment config
cp .env.example .env

# Download dependencies  
go mod tidy
```

## Bước 3: Chạy

### Với Ollama (local, free):

```bash
# Đảm bảo Ollama đang chạy
ollama serve

# Pull model (nếu chưa có)
ollama pull qwen2.5:7b

# Chạy
go run main.go
```

### Với OpenAI:

Sửa file `.env`:
```
LLM_PROVIDER=openai
LLM_MODEL=gpt-4o-mini
OPENAI_API_KEY=sk-your-key-here
```

Rồi chạy:
```bash
go run main.go
```

### Với Gemini:

Sửa file `.env`:
```
LLM_PROVIDER=gemini
LLM_MODEL=gemini-2.0-flash-exp
GEMINI_API_KEY=your-key-here
```

Rồi chạy:
```bash
go run main.go
```

## Bạn Sẽ Có Gì?

✅ **Agent tương tác** với 20 tools sẵn có  
✅ **Logging đầy đủ** với màu sắc và biểu tượng  
✅ **Memory** - nhớ ngữ cảnh hội thoại  
✅ **Vietnamese support** - hỗ trợ tiếng Việt  
✅ **Sẵn sàng dùng** - không cần config gì thêm

## Test Ngay

Sau khi chạy, thử hỏi:

```
👤 You: Tính 25 * 4 + 100
👤 You: What time is it now?
👤 You: List files in current directory
👤 You: Tôi sinh ngày 15/03/1990, năm nay tôi bao nhiêu tuổi?
```

## Tùy Chỉnh

Muốn tắt logging? Sửa trong `main.go`:

```go
a := agent.New(llm, agent.WithMemory(mem), agent.DisableLogging())
```

Muốn thêm tools riêng? Xem hướng dẫn trong README.md

## Chạy Nhanh Từ Repo

Nếu bạn đang ở trong repo này:

```bash
cd examples/standalone_demo
cp .env.example .env
go run main.go
```

---

**Chỉ 3 bước**: Copy → Setup → Chạy. Vậy thôi! 🚀
