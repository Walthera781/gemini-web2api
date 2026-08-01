# gemini-web2api

High-performance Go proxy converting Google Gemini web interface into OpenAI-compatible (`/v1/chat/completions`, `/v1/responses`) and Google Gemini native (`/v1beta/models/*`) APIs.

## Features

- **OpenAI Compatible**: Seamless drop-in replacement for `/v1/chat/completions` and `/v1/responses` (Codex CLI).
- **Google Native API**: Endpoints for `/v1beta/models/{model}:generateContent` and `:streamGenerateContent`.
- **Streaming & SSE**: High-throughput SSE streaming out of the box with zero memory cap limits.
- **Multimodal Support**: Scotty resumable upload for images with automatic token caching and optional compression.
- **TLS Fingerprinting**: Native optional TLS fingerprint impersonation (`chrome_146`, etc.) powered by `tls-client`.
- **Tool Calling**: OpenAI tool call parsing and 3-pattern Google function call extraction.
- **Micro-Container**: Multi-stage build into a static binary on a `scratch` container (~15MB image).

## Installation

### Prerequisites
- Go `1.26+`

### Build from source
```bash
git clone https://github.com/ikhsan3adi/gemini-web2api.git
cd gemini-web2api
go build -o gemini-web2api .
```

### Install with Go
```bash
go install github.com/ikhsan3adi/gemini-web2api@latest
```

## Running the Server

```bash
./gemini-web2api --port 8081 --config ./config.json
```

### Command-line Options
- `--port`: Server listening port (default `8081`).
- `--config`: Path to `config.json`.
- `--cookie-file`: Path to cookie file (JSON or raw Netscape string).
- `--proxy`: HTTP/HTTPS/SOCKS proxy URL.
- `--impersonate`: TLS impersonation profile (e.g. `chrome_146`).
- `--version`: Show version and exit.

## Docker Deployment

### Local Docker Build
```bash
docker build -t gemini-web2api .
docker run -p 8081:8081 -v $(pwd)/config.json:/app/config.json gemini-web2api
```

### Docker Compose
```bash
docker-compose -f docker-compose.local.yml up -d
```

## Endpoints

| Method | Path | Auth Required | Description |
|---|---|---|---|
| `GET` | `/` | No | Health check & model list |
| `GET` | `/v1/models` | Yes | List OpenAI models |
| `POST` | `/v1/chat/completions` | Yes | OpenAI Chat Completions |
| `POST` | `/v1/responses` | Yes | OpenAI Responses API (Codex CLI) |
| `GET` | `/v1beta/models` | Yes | List Google Gemini models |
| `POST` | `/v1beta/models/{model}:generateContent` | Yes | Google Gemini content generation |
| `POST` | `/v1beta/models/{model}:streamGenerateContent` | Yes | Google Gemini streaming generation |

## License

MIT License.
