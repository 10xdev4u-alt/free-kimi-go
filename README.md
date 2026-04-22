# 🦾 Kimi Go Beast

A high-performance, ultra-lightweight Golang rewrite of the `kimi-free-api`. This beast transforms Kimi AI's web protocol into a production-grade, OpenAI-compatible API with native concurrency and microscopic memory footprint.

## 🚀 Why Go Beast?

-   **Memory King**: Idles at **~10-15MB RAM** (compared to ~100MB+ for Node.js).
-   **Native Streaming**: Powered by `Fiber` and `fasthttp` for insane SSE concurrency.
-   **Zero-Leak Tool Engine**: Surgical XML/JSON trapping that promotes Kimi's reasoning into real local tool calls (perfect for **Claude Code** and **Cursor**).
-   **Indestructible**: No need for external watchdogs; Go's runtime handles the heavy lifting.
-   **English-First**: Clean, sanitized codebase with full English logs and comments.

## ✨ Core Features

*   ✅ **100% OpenAI Compatible**: Drop-in replacement for `/v1/chat/completions`.
*   ✅ **Token Rotation**: Load-balance across multiple refresh tokens (comma-separated).
*   ✅ **High-Fidelity SSE**: Real-time streaming with initial role engagement.
*   ✅ **File/Image Support**: Native support for Base64 and URL-based uploads to Kimi OSS.
*   ✅ **Stealth Mode**: Automated background `fakeRequest` calls to mimic human browsing.
*   ✅ **SSRF Protection**: Hardened file fetcher to protect your internal network.

## 🛠️ Installation

```bash
# Clone the repo
git clone https://github.com/10xdev4u-alt/kimi-go-beast.git
cd kimi-go-beast

# Build the beast
go build -o kimi-api cmd/api/main.go

# Launch the binary
export PORT=8788
./kimi-api
```

## 🛰️ API Configuration

| Setting | Value |
| :--- | :--- |
| **Base URL** | `http://localhost:8788/v1` |
| **API Key** | `eyJhbGci...` (Your Kimi Refresh Token) |
| **Model ID** | `kimi-k2.6-instant` or `kimi-silent_search` |

## 🏗️ Architecture

Built using **Domain-Driven Design (Clean Architecture)**:
- `internal/chat`: The core engine for session management and stream parsing.
- `internal/token`: Concurrent-safe token manager with thundering herd protection.
- `internal/core`: Stealth headers, cookie synthesis, and secure utilities.

---
**Lets GO!** 🚀🛰️🦾
