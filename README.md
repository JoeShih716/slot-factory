# Slot Factory 🎰

[![Go Version](https://img.shields.io/badge/go-1.25-blue)](https://go.dev/)
[![Docker](https://img.shields.io/badge/docker-ready-blue)](https://www.docker.com/)
[![Kubernetes](https://img.shields.io/badge/kubernetes-ready-blue)](https://kubernetes.io/)

Slot Factory 是一個使用現代化 Golang 架構構建的老虎機遊戲後端服務。專案採用 Clean Architecture (DDD) 設計，並支援完整的雲端原生 (Cloud Native) 開發與部署流程。

## 🌟 專案特色

*   **現代化架構**: 採用 Domain-Driven Design (DDD) 與 Clean Architecture，將核心邏輯、應用層與轉接層解耦。
*   **無縫錢包 (Seamless Wallet)**: 支援「代理模式 (Proxy Mode)」—— 由外部平台管理資金，本地非同步記錄交易流水，符合老虎機產業主流架構。
*   **開發者體驗**: 整合 `Air` 實現本地 Docker 環境下的 Hot Reload 開發。
*   **配置管理**: 符合 12-Factor App 的分層配置策略，支援 `Auth` 與 `Wallet` 的獨立 Mock/Real 切換。
*   **雲端原生**: 內建 Dockerfile 多階段建置與 Kubernetes (Deployment/Service) 部署清單。
*   **在地化支援**: 完整支援台灣時區 (Asia/Taipei) 的資料庫紀錄與系統顯示。

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
*   **WebSocket Server**: `ws://localhost:8080/ws`
*   **phpMyAdmin**: [http://localhost:8088](http://localhost:8088) (帳: root / 密: root)
*   **測試工具**: 直接瀏覽器打開 `wstest.html` 即可連線遊玩。

### 核心演示
在本地 `local` 環境下，預設啟動 **Proxy Wallet + MySQL Logging**:
1.  **錢包餘額**: 由 `internal/adapter/wallet/proxy` 模擬呼叫外部平台 (固定回傳 100,000)。
2.  **交易流水**: 每一次 Spin 的結果都會非同步寫入本地 MySQL 的 `wallet_transactions` 表，並使用 **台灣時區 (Asia/Taipei)**。
3.  **靈活開發**: 您可以修改 `backend/configs/wsServer/config.local.yaml` 來獨立切換 Auth 或 Wallet 為 Mock 模式。

## ☸️ Kubernetes 部署

本專案支援標準 K8s 部署。

```bash
# 部署至當前 K8s Context
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
```

## 🛠 技術棧

*   **Language**: Golang 1.25+
*   **Framework**: Gin (HTTP), Gorilla WebSocket
*   **Config**: Viper
*   **DevOps**: Docker, Kubernetes, Air


