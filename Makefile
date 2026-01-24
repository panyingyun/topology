# 获取版本信息
GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date +%FT%T%z)
VERSION    := $(shell git describe --tags --always --dirty)

# 定义变量注入路径 (根据你的实际 module 修改)
LDFLAGS := -X 'main.GitCommit=$(GIT_COMMIT)' \
           -X 'main.BuildTime=$(BUILD_TIME)' \
           -X 'main.Version=$(VERSION)'

.PHONY: env  clean lint build

all: env  clean lint build

env:
	@echo "=========install gofmt ==========="
	GOPROXY=https://goproxy.cn/,direct go install mvdan.cc/gofumpt@latest
	@echo "=========install gentool ==========="
	GOPROXY=https://goproxy.cn/,direct go install -v gorm.io/gen/tools/gentool@latest
	@echo "=========install goreleaser ==========="
	GOPROXY=https://goproxy.cn/,direct go install github.com/goreleaser/goreleaser/v2@latest
	@echo "=========install wails ==========="
	# //for ubuntu24.04
	# sudo apt install libwebkit2gtk-4.1-dev  libgtk-3-dev -y
	# //for ubuntu22.04
	# sudo apt install libwebkit2gtk-4.0-dev  libgtk-3-dev -y
	GOPROXY=https://goproxy.cn/,direct go install github.com/wailsapp/wails/v2/cmd/wails@latest
	@echo "=========install tsc ==========="
	npm install -g typescript
	echo "=========install vite ==========="
	npm install -g vite
gensql:
	gentool -dsn "root:Cjj123@tcp(192.168.1.120:6306)/realmdb?charset=utf8mb4&parseTime=True&loc=Local" -outPath "/home/yypan/panyingyun/Realm/realm-gui/dao/query"

dev_ubuntu2204:
	go mod tidy
	gofumpt -l -w .
	wails dev -tags webkit2_40

dev_ubuntu2404:
	go mod tidy
	gofumpt -l -w .
	wails dev -tags webkit2_41
			
clean:
	go clean -i .


build_ubuntu2204:
	go mod tidy
	gofumpt -l -w .
	wails build -tags webkit2_40

build_ubuntu2404:
	go mod tidy
	gofumpt -l -w .
	wails build -tags webkit2_41

build-windows:
	wails build 