# Personal Assistant Example - Quick Start Guide

## 🎯 Mục đích

Đây là ví dụ **thực tế và hữu ích nhất** của thư viện go-llm-agent - một trợ lý cá nhân thông minh có thể:

- ✅ Trả lời câu hỏi về ngày giờ
- ✅ Tìm kiếm thông tin trên web
- ✅ Tính toán phức tạp (lãi kép, phần trăm, thống kê)
- ✅ Quản lý file (đọc, liệt kê, tìm kiếm)
- ✅ Giám sát hệ thống (CPU, RAM, disk)
- ✅ Công cụ mạng (DNS lookup, ping, SSL check)
- ✅ Tính toán thời gian (đếm ngày, so sánh ngày tháng)

**Và quan trọng nhất**: Nó **TỰ HỌC** từ kinh nghiệm để ngày càng thông minh hơn!

## 📦 Cài đặt nhanh

### Bước 1: Cài Ollama
```bash
# macOS/Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Download model (chọn 1 trong 3)
ollama pull qwen2.5:3b   # ⭐ Khuyến nghị
ollama pull llama3.2:3b  # Thay thế
ollama pull gemma2:2b    # Nhẹ hơn
```

### Bước 2 (Tùy chọn): Cài Qdrant để có learning
```bash
docker run -d -p 6334:6334 -p 6333:6333 --name qdrant qdrant/qdrant
```

### Bước 3: Chạy ví dụ
```bash
cd examples/personal_assistant
go run main.go
```

## 🎬 Demo Scenarios

Chương trình có 7 kịch bản thực tế:

### 1. Thông tin hàng ngày
```
You: What's today's date and what day of the week is it?
🤖: Today is January 27, 2025, Monday
```

### 2. Tìm kiếm web
```
You: Search for the latest news about artificial intelligence
🤖: [Tìm kiếm và tóm tắt tin tức AI mới nhất]
```

### 3. Tính toán lãi kép
```
You: Calculate compound interest for $10,000 at 5% for 3 years
🤖: Final amount: $11,576.25, Total interest: $1,576.25
```

### 4. Quản lý file
```
You: Show me what files are in the current directory
🤖: Found 3 files: main.go, README.md, .gitignore
```

### 5. Giám sát hệ thống
```
You: What's the current CPU and memory usage?
🤖: CPU: 25.3%, Memory: 8.2GB / 16GB (51% used)
```

### 6. Network tools
```
You: Do a DNS lookup for google.com
🤖: google.com resolves to: 142.250.185.78 (IPv4), 2404:6800:4003:c00::64 (IPv6)
```

### 7. Tính ngày tháng
```
You: How many days until Christmas 2025?
🤖: There are 332 days until Christmas 2025 (December 25, 2025)
```

## 💬 Interactive Mode

Sau demo, bạn có thể chat tự do:

```
You: Calculate 15% of 250
🤖: 15% of 250 is 37.5

You: status
📊 PERSONAL ASSISTANT STATUS
==================================================
🤖 Provider: ollama/qwen2.5:3b
💬 Conversations: 12 messages in memory
🔧 Available Tools: 24

🧠 LEARNING SYSTEM:
   ✅ Enabled: true
   📚 Total Experiences: 42
   ✨ Success Rate: 87.5%
   🎯 Tool Selector: Active
   🔍 Error Analyzer: Active

You: help
📖 EXAMPLE QUESTIONS
==================================================
  • What's the current time and date?
  • Search the web for Python tutorials
  • Calculate 15% of 250
  • What files are in the current directory?
  • Check system CPU usage
  [...]
```

## 🧠 Tính năng Self-Learning

Khi có Qdrant, agent sẽ:

1. **Ghi nhớ mọi tương tác**:
   - Query nào thành công/thất bại
   - Tool nào được dùng
   - Thời gian phản hồi

2. **Học từ kinh nghiệm**:
   - Tool nào hiệu quả cho từng loại câu hỏi
   - Pattern lỗi thường gặp
   - Cách sửa lỗi tốt nhất

3. **Cải thiện liên tục**:
   - Success rate tăng dần theo thời gian
   - Chọn tool chính xác hơn
   - Phản hồi nhanh hơn

## 🔧 Customization

### Đổi model
```go
// main.go line 28
llm := ollama.New("http://localhost:11434", "llama3.2:3b")
```

### Tắt learning (nếu không có Qdrant)
```go
// main.go line 34
agent.WithLearning(false),
```

### Tăng log level để debug
```go
// main.go line 35
agent.WithLogLevel(logger.LogLevelDebug),
```

## 🚀 Use Cases thực tế

Bạn có thể dùng ví dụ này làm base để build:

- **DevOps Assistant**: Giám sát server, deploy code, check logs
- **Data Analyst**: Đọc CSV/JSON, tính toán thống kê, vẽ biểu đồ
- **Personal Secretary**: Quản lý email, lịch hẹn, nhắc nhở
- **Research Assistant**: Tìm kiếm papers, tóm tắt tài liệu
- **System Admin**: Check services, restart apps, analyze metrics

## 📚 Tài liệu liên quan

- [Main README](../../README.md) - Tổng quan thư viện
- [Learning Agent Example](../learning_agent/) - Chi tiết về learning system
- [Built-in Tools](../agent_with_builtin_tools/) - Danh sách 24 tools

## 💡 Tips

1. **Lần đầu chạy**: Chọn mode 1 (Demo) để xem các use case
2. **Thử nghiệm**: Chọn mode 2 (Interactive) để chat tự do
3. **Check status**: Gõ `status` để xem metrics
4. **Cần gợi ý**: Gõ `help` để xem câu hỏi mẫu
5. **Thoát**: Gõ `exit` hoặc `quit`

## ❓ Troubleshooting

**❌ "VectorMemory not available"**
→ Agent vẫn chạy nhưng không có learning. Cài Qdrant để bật learning.

**❌ "Model not found"**
→ Chạy `ollama pull qwen2.5:3b`

**❌ Phản hồi chậm**
→ Thử model nhẹ hơn: `gemma2:2b`

**❌ "No tools available"**
→ Check log, có thể cần restart Ollama

## ✨ Tại sao example này hay?

1. **Thực tế**: Không phải demo đơn giản, thực sự hữu ích
2. **Học được**: Minh họa tất cả tính năng của thư viện
3. **Dễ dùng**: Menu rõ ràng, hướng dẫn chi tiết
4. **Tự cải thiện**: Learning system làm nó thông minh hơn mỗi ngày
5. **Customizable**: Dễ mở rộng cho use case riêng

---

**Chúc bạn build agent thành công! 🚀**
