BINARY=taskapi
OUT_DIR=bin

# Detect OS
ifeq ($(OS),Windows_NT)
	SHELL := powershell.exe
	.SHELLFLAGS := -NoProfile -Command
	BUILD_TIME := $(shell (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ"))
	MKDIR = if (-not (Test-Path $(OUT_DIR))) { New-Item -ItemType Directory -Path $(OUT_DIR) | Out-Null }
	RM = if (Test-Path $(OUT_DIR)) { Remove-Item -Recurse -Force $(OUT_DIR) }
	PATH_SEP = \\
	BUILD_CMD = $$env:CGO_ENABLED='0'; $$env:GOOS='linux'; $$env:GOARCH='amd64'; go build -ldflags="-s -w -X main.buildTime=$(BUILD_TIME)" -o $(OUT_DIR)$(PATH_SEP)$(BINARY) .
else
	BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
	MKDIR = mkdir -p $(OUT_DIR)
	RM = rm -rf $(OUT_DIR)
	PATH_SEP = /
	BUILD_CMD = CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.buildTime=$(BUILD_TIME)" -o $(OUT_DIR)$(PATH_SEP)$(BINARY) .
endif

.PHONY: build run test docker-build docker-up docker-down docker-run fmt clean

build:
	$(MKDIR)
	$(BUILD_CMD)

run:
	go run main.go

test:
	go test ./... -v

docker-build:
	docker build -t repoentrevistaanjr022026:latest .

docker-up:
	docker-compose up --build -d

docker-run:
	docker compose -f docker-compose.yml up -d api

docker-down:
	docker-compose down

fmt:
	gofmt -w .

clean:
	$(RM)
