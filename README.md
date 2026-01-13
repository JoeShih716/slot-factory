# Slot Factory 🎰

[![Go Version](https://img.shields.io/badge/go-1.25-blue)](https://go.dev/)
[![CI Status](https://github.com/JoeShih716/slot-factory/actions/workflows/ci.yaml/badge.svg)](https://github.com/JoeShih716/slot-factory/actions)
[![Docker](https://img.shields.io/badge/docker-ready-blue)](https://www.docker.com/)
[![Kubernetes](https://img.shields.io/badge/kubernetes-ready-blue)](https://kubernetes.io/)

Slot Factory 是一個使用現代化 Golang 架構構建的老虎機遊戲後端服務。專案採用 Clean Architecture (DDD) 設計，並支援完整的雲端原生 (Cloud Native) 開發與部署流程。

## 🌟 專案特色

*   **現代化微服務架構**: 採用 Domain-Driven Design (DDD) 與 Clean Architecture，並將服務拆分為 `wsserver` (連線) 與 `api` (管理/讀取) 獨立服務。
*   **全域狀態管理 (Redis)**: 整合 Redis 實現跨實體的人數統計 (Counter) 與指令廣播 (Pub/Sub)，支援分散式水平擴展。
*   **介面隔離原則 (ISP)**: 透過窄介面定義 (`GameProvider`, `AdminProvider`, `HistoryProvider`)，精確控制服務間的依賴。
*   **無縫錢包 (Seamless Wallet)**: 支援「代理模式 (Proxy Mode)」—— 由外部平台管理資金，本地非同步記錄交易流水。
*   **開發者體驗**: 整合 `Air` 支援多容器同時開發的 Hot Reload，並提供 Multi-binary Dockerfile。
*   **配置管理**: 統一的 `configs` 目錄，支援一套軟體多重角色的分層配置策略。

## 📂 目錄結構

```text
.
├── backend/                # Go 後端核心代碼
│   ├── cmd/                # 程式進入點
│   ├── configs/            # 設定檔 (Local/Dev/Prod)
│   ├── internal/           # 內部包 (Domain/Application/Adapter)
│   └── pkg/                # 公用包 (Config/WSS)
├── deploy/                 # 部署相關檔案
│   └── k8s/                # Kubernetes Manifests
├── docker-compose.yaml     # 本地開發編排
├── wstest.html            # WebSocket 測試工具
└── README.md               # 專案說明
```

## 🚀 快速開始 (Local Development)

### 前置需求
*   Docker & Docker Compose
*   (Optional) Make

### 啟動服務
使用 Docker Compose 一鍵啟動 (包含 MySQL, phpMyAdmin 與 Hot Reload)：

```bash
docker-compose up -d --build
```

服務啟動後：
*   **WebSocket Server**: `ws://localhost:8080/ws` (處理遊戲連線)
*   **REST API Gateway**: [http://localhost:8081](http://localhost:8081) (查詢列表、歷史、管理員指令)
*   **phpMyAdmin**: [http://localhost:8088](http://localhost:8088) (帳: root / 密: root)
*   **Redis**: `localhost:6379` (全域狀態儲存)
*   **測試工具**: 直接瀏覽器打開 `wstest.html` 即可連線遊玩（請確保 WS 地址正確）。

### 本地驗證 (Local Verification)
為了確保程式碼品質，我們提供了 `Makefile` 讓開發者在 Commit 前快速檢查：

```bash
cd backend
make verify
```

此指令會自動執行：
1.  **Lint**: `golangci-lint` (檢查程式碼風格)
2.  **Test**: `go test` (單元測試)

### 核心演示
在本地 `local` 環境下，專案展示了以下進階特性：
1.  **分散式人數統計**: 透過 Redis，`api` 服務能即時查詢所有伺服器實體上的玩家總量。
2.  **全域廣播指令**: 呼叫 `api` 的 `/kick_all` 端點，會透過 Redis Pub/Sub 同步踢除所有 `wsserver` 內的線上玩家。
3.  **職責分離**: 核心業務邏輯僅寫在 `internal/application`，但透過不同介面暴露給連線層與管理層，實現高內聚低耦合。
4.  **台灣時區支援**: 資料庫流水與查詢系統完整對接 `Asia/Taipei`，符合在地營運需求。

## ☸️ Kubernetes 部署

本專案支援標準 K8s 部署。

```bash
# 部署至當前 K8s Context
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
```

## 🔄 CI/CD 自動化流程

本專案採用 **GitHub Actions** 進行持續整合，確保每次 Commit 的品質。

*   **Workflow**: `.github/workflows/ci.yaml`
*   **Pipeline Stages**:
    1.  **Lint**: 使用 `golangci-lint` 進行靜態分析。
    2.  **Test**: 執行所有單元測試。
    3.  **Build**: 確保 `wsserver` 與 `api` 雙服務皆可成功編譯。
*   **Strategy**: 使用 `go install` 現場編譯最新版 Linter，以支援最新的 Go 1.25 特性。

## 🛠 技術棧

*   **Language**: Golang 1.25+
*   **Tech**: Redis (State), MySQL (Audit), Gin (HTTP), Gorilla WebSocket
*   **Strategy**: Clean Architecture / DDD / ISP
*   **DevOps**: Docker (Multi-binary), Docker Compose, Air (Hot Reload)


