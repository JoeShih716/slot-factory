# Slot Factory 🎰

[![Go Version](https://img.shields.io/badge/go-1.25-blue)](https://go.dev/)
[![Docker](https://img.shields.io/badge/docker-ready-blue)](https://www.docker.com/)
[![Kubernetes](https://img.shields.io/badge/kubernetes-ready-blue)](https://kubernetes.io/)

Slot Factory 是一個使用現代化 Golang 架構構建的老虎機遊戲後端服務。專案採用 Clean Architecture (DDD) 設計，並支援完整的雲端原生 (Cloud Native) 開發與部署流程。

## 🌟 專案特色

*   **現代化架構**: 採用 Domain-Driven Design (DDD) 與 Clean Architecture，將核心邏輯、應用層與轉接層解耦。
*   **錢包服務**: 獨立的 Wallet Service 設計，支援 Mock 與 Real 金流轉接。
*   **開發者體驗**: 整合 `Air` 實現本地 Docker 環境下的 Hot Reload 開發。
*   **配置管理**: 符合 12-Factor App 的分層配置策略 (Local/Dev/Prod)。
*   **雲端原生**: 內建 Dockerfile 多階段建置與 Kubernetes (Deployment/Service) 部署清單。

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
使用 Docker Compose 一鍵啟動 (包含 Hot Reload)：

```bash
docker-compose up --build
```

服務啟動後：
*   **WebSocket Server**: `ws://localhost:8080/ws`
*   **測試工具**: 直接瀏覽器打開 `wstest.html` 即可連線遊玩。

## ☸️ Kubernetes 部署

本專案支援標準 K8s 部署。詳細操作請參考 [DevOps 筆記](DEVOPS_NOTES.md)。

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

## 📝 文件
*   [DevOps 完整筆記](DEVOPS_NOTES.md): 包含 Docker/K8s 詳細實作原理。
*   [Project Plan](PROJECT_PLAN.md): 專案規劃與進度。
