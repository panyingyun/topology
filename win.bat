@echo off
set GOOS=windows
set GOARCH=amd64
set GOHOSTOS=windows
set GOHOSTARCH=amd64
set CGO_ENABLED=0

go clean -cache
go clean -modcache
wails.exe build
