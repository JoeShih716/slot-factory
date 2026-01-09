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

# 編譯應用程式
# CGO_ENABLED=0: 關閉 CGO，生成純靜態連結的執行檔，這樣在 alpine 裡跑才不會缺 library
# -o server: 輸出檔名為 server
RUN CGO_ENABLED=0 go build -o server ./cmd/wsserver

# === Stage 2: Runner ===
# 使用極簡的 alpine 映像檔作為執行環境
FROM alpine:latest

# 設定工作目錄
WORKDIR /app

# 從 builder 階段把編譯好的 server 執行檔複製過來
# --from=builder: 指定來源是剛剛命名為 builder 的那個階段
COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs

# 黃金法則 EXPOSE = app listen port
EXPOSE 8080

# 容器啟動時執行的指令
CMD ["./server"]
