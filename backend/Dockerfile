# === Stage Base: Common ===
FROM golang:1.25 AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# === Stage Dev: Hot Reload ===
FROM base AS dev
RUN go install github.com/air-verse/air@latest
CMD ["air", "-c", ".air.toml"]

# === Stage Builder: Compile ===
FROM base AS builder

# 預設編譯目標 (wsserver 或 api)
ARG TARGET=wsserver

# 編譯應用程式
RUN CGO_ENABLED=0 go build -o server ./cmd/${TARGET}

# === Stage Runner: Alpine ===
FROM alpine:latest
WORKDIR /app

# 從 builder 複製執行檔與設定
COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs

# 啟動時根據 TARGET 決定行為 (預設執行 server)
CMD ["./server"]
