.PHONY: lint test build verify

# 預設執行 verify
all: verify

# 1. 靜態分析 (需先安裝 golangci-lint)
lint:
	@echo "Running lint check..."
	golangci-lint run ./...

# 2. 單元測試
test:
	@echo "Running unit tests..."
	go test -v ./...


# 3. 完整驗證 (Lint + Test)
verify: lint test
	@echo "All checks passed!"
