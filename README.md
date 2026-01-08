# Ollama2OpenAI Proxy

将 Ollama API 转换为完全兼容 OpenAI API 格式的代理服务，支持 OpenAI SDK、LangChain、n8n 等工具无感替换。

## 功能特性

- **Chat Completions** - 文本对话、Vision 图像理解、流式响应
- **Embeddings** - 向量生成，支持 string 和 []string 输入
- **Streaming (SSE)** - 服务器发送事件流式响应
- **API Key 鉴权** - 多 Key 支持，带别名统计
- **Usage 统计** - 按 API Key 维度统计 token 使用量
- **Models API** - 模型列表与详情

## 快速开始

### 1. 安装

```bash
git clone https://github.com/yourname/ollama2openai.git
cd ollama2openai
go build -o ollama2openai .
```

### 2. 配置

编辑 `config/config.yaml`：

```yaml
# HTTP Server
host: "0.0.0.0"
port: 8080

# Ollama Server
ollama_url: "http://localhost:11434"

# API Keys (key -> alias mapping)
api_keys:
  sk-1234567890: "user1"
  sk-0987654321: "user2"
  sk-default-key: "default"

# Request timeout (seconds)
timeout: 300

# Log level: debug, info, warn, error
log_level: "info"
```

### 3. 启动

```bash
# 使用默认配置
./ollama2openai

# 或指定配置路径
CONFIG_PATH=/path/to/config.yaml ./ollama2openai
```

启动时验证 Ollama 连接并显示可用模型：

```
2026/01/08 17:30:05 Starting Ollama2OpenAI Proxy on 0.0.0.0:8080
2026/01/08 17:30:05 Ollama URL: http://localhost:11434
2026/01/08 17:30:05 Verifying Ollama connection...
2026/01/08 17:30:05 Ollama is connected. Available models:
2026/01/08 17:30:05   - llama3:latest
2026/01/08 17:30:05   - qwen2.5-vl:latest
2026/01/08 17:30:05   - nomic-embed-text:latest
```

## API 使用示例

### Chat Completions

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-1234567890" \
  -d '{
    "model": "llama3",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ]
  }'
```

### Streaming

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-1234567890" \
  -d '{
    "model": "llama3",
    "messages": [{"role": "user", "content": "Tell me a story"}],
    "stream": true
  }'
```

### Vision (多模态)

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-1234567890" \
  -d '{
    "model": "qwen2.5-vl",
    "messages": [
      {
        "role": "user",
        "content": [
          {"type": "text", "text": "Describe this image:"},
          {"type": "image_url", "image_url": {"url": "data:image/jpeg;base64,/9j/4AAQ..."}}
        ]
      }
    ]
  }'
```

### Embeddings

```bash
# 单文本
curl -X POST http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-1234567890" \
  -d '{
    "model": "nomic-embed-text",
    "input": "Hello world"
  }'

# 多文本
curl -X POST http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-1234567890" \
  -d '{
    "model": "nomic-embed-text",
    "input": ["Hello world", "Goodbye world"]
  }'
```

### List Models

```bash
curl http://localhost:8080/v1/models \
  -H "Authorization: Bearer sk-1234567890"
```

### Usage 统计

```bash
curl http://localhost:8080/usage
```

响应示例：
```json
{
  "user1": {
    "prompt_tokens": 150,
    "completion_tokens": 85,
    "embedding_tokens": 120,
    "total_requests": 5,
    "embedding_requests": 2
  }
}
```

## 推荐模型

| 能力 | 模型 |
|------|------|
| Chat | llama3, qwen2.5 |
| Vision | llava, qwen2.5-vl |
| Embedding | nomic-embed-text, qwen3-embedding |

## OpenAI SDK 使用

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-1234567890",
    base_url="http://localhost:8080/v1"
)

response = client.chat.completions.create(
    model="llama3",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

## 日志级别

支持 `debug`, `info`, `warn`, `error` 四个级别，通过配置文件设置：

```yaml
log_level: "info"
```

## License

MIT
