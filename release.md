校验和文件：各平台使用独立配置文件发布，校验和文件名含平台信息。例如：`topology_1.3.0_Ubuntu22.04_checksums.txt`、`topology_1.3.0_Ubuntu24.04_checksums.txt`、`topology_1.3.0_Windows_checksums.txt`；macOS 使用 `Runtime.Goos/Goarch` 时为 `topology_1.3.0_darwin_amd64_checksums.txt`。

### 1、检查配置(不实际发布)

```
goreleaser check
```

### 2、测试配置(不实际发布)

```
goreleaser release --snapshot
```
### 3、设置环境变量
```
export GITHUB_TOKEN="your-github-token"
```

### 4、打Tag并发布
```
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
goreleaser release

goreleaser check -f .goreleaser.macos.yml
goreleaser check -f .goreleaser.ubuntu22.04.yml
goreleaser check -f .goreleaser.ubuntu24.04.yml
goreleaser check -f .goreleaser.windows.yml

# 快照测试（不发布）
goreleaser release -f .goreleaser.macos.yml --clean --snapshot
goreleaser release -f .goreleaser.ubuntu22.04.yml --clean --snapshot
goreleaser release -f .goreleaser.ubuntu24.04.yml --clean --snapshot
goreleaser release -f .goreleaser.windows.yml --clean --snapshot

# 正式发布（校验和文件名含平台：topology_版本_Ubuntu22.04_checksums.txt 等）
goreleaser release -f .goreleaser.macos.yml --clean
goreleaser release -f .goreleaser.ubuntu22.04.yml --clean
goreleaser release -f .goreleaser.ubuntu24.04.yml --clean
goreleaser release -f .goreleaser.windows.yml --clean
```