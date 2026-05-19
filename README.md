# CPA-SUB2 Conv

> 一个轻量级 Go Web 工具，用于在 CPA (`codex`) 与 Sub2API 账号 JSON 格式之间相互转换。

[在线体验](https://conv.wangmin.xyz) · [示例文件](./examples) · [Docker Compose](./docker-compose.yml)

## 目录

- [功能亮点](#功能亮点)
- [快速开始](#快速开始)
- [Docker 部署](#docker-部署)
- [API](#api)
- [示例文件](#示例文件)
- [项目结构](#项目结构)
- [开发](#开发)
- [发布](#发布)

## 功能亮点

| 功能 | 说明 |
| --- | --- |
| 双向转换 | 支持 `CPA -> Sub2API` 与 `Sub2API -> CPA` |
| 自动识别 | `auto` 模式会自动判断输入格式并转换到另一种格式 |
| 文本转换 | 可直接在浏览器中粘贴 JSON 并转换 |
| 文件转换 | 支持上传单个 `.json` 文件并下载转换结果 |
| 批量转换 | 支持上传 `.zip`，批量转换压缩包内所有 `.json` 文件 |
| ZIP 输出 | 批量结果会打包为 `converted_<timestamp>.zip` |

### 下载命名方式

| 模式 | 说明 |
| --- | --- |
| 时间戳 | 使用当前时间生成文件名 |
| 格式 + 原文件名 | 例如 `sub2_accounts.json` |
| 格式 + JSON 名称 | 优先读取 JSON 中的 `name`、`email` 或 `account_id` |
| 自定义前缀 | 使用自定义前缀生成下载文件名 |

## 快速开始

### 本地运行

需要 Go `1.22+`。

```bash
go build -o converter_server .
./converter_server
```

Windows PowerShell:

```powershell
go build -o converter_server.exe .
.\converter_server.exe
```

启动后访问 [http://localhost:8080](http://localhost:8080)。

### 指定端口

服务默认监听 `8080`，可以通过 `PORT` 环境变量覆盖。

```bash
PORT=9090 ./converter_server
```

Windows PowerShell:

```powershell
$env:PORT = "9090"
.\converter_server.exe
```

## Docker 部署

### 使用 Docker Compose

```bash
docker compose up -d
```

启动后访问 [http://localhost:8080](http://localhost:8080)。

停止服务：

```bash
docker compose down
```

### 修改宿主机端口

```bash
HOST_PORT=9090 docker compose up -d
```

Windows PowerShell:

```powershell
$env:HOST_PORT = "9090"
docker compose up -d
```

默认 Compose 文件会拉取 GitHub Container Registry 上的镜像：

```text
ghcr.io/wm1634208243/cpa-sub2-conv:latest
```

如果包仍是私有状态，需要先登录 `ghcr.io`：

```bash
echo <YOUR_GITHUB_TOKEN> | docker login ghcr.io -u <YOUR_GITHUB_USERNAME> --password-stdin
```

### 本地构建镜像

```bash
docker build -t cpa-sub2-conv .
docker run --rm -p 8080:8080 cpa-sub2-conv
```

## API

| Method | Path | 说明 |
| --- | --- | --- |
| `POST` | `/api/detect` | 识别输入 JSON 格式 |
| `POST` | `/api/convert` | 转换文本 JSON |
| `POST` | `/api/convert-file` | 转换上传的 `.json` 或 `.zip` 文件 |

### `POST /api/detect`

Request:

```json
{
  "input": "<json string>"
}
```

Response:

```json
{
  "format": "cpa"
}
```

`format` 可能的值：

- `cpa`
- `sub2`
- `unknown`

### `POST /api/convert`

Request:

```json
{
  "input": "<json string>",
  "target": "auto"
}
```

`target` 可能的值：

- `cpa`
- `sub2`
- `auto`

Response:

```json
{
  "output": "<converted json string>",
  "detected": "cpa"
}
```

### `POST /api/convert-file`

Multipart form fields:

| Field | 说明 |
| --- | --- |
| `file` | `.json` 或 `.zip` 文件 |
| `target` | `cpa`、`sub2` 或 `auto` |

返回行为：

- 上传 `.json` 时，返回转换后的 JSON 文件。
- 上传 `.zip` 时，转换压缩包内的所有 `.json` 文件，并返回 `converted_<timestamp>.zip`。
- 上传文件大小限制为 `32MB`。

## 示例文件

示例输入文件位于 [examples](./examples)：

- [cpa.sample.json](./examples/cpa.sample.json)
- [sub2.sample.json](./examples/sub2.sample.json)

## 项目结构

```text
.
|-- main.go
|-- go.mod
|-- Dockerfile
|-- docker-compose.yml
|-- Makefile
|-- static/
|   |-- index.html
|   `-- favicon.svg
|-- internal/
|   |-- converter/
|   `-- handler/
|-- examples/
|-- .github/
|   `-- workflows/
`-- README.md
```

## 开发

运行测试：

```bash
go test ./...
```

构建项目：

```bash
go build ./...
```

如果本地安装了 `make`，也可以使用：

```bash
make build
make run
```

Windows PowerShell 下如需将 Go 构建缓存放在项目目录内：

```powershell
$env:GOCACHE = "$PWD\.gocache"
go build ./...
go test ./...
```

## 发布

仓库包含 GitHub Actions 工作流：

- [ci.yml](./.github/workflows/ci.yml)：持续集成检查。
- [docker-publish.yml](./.github/workflows/docker-publish.yml)：推送到 `main` 分支或推送 `v*` 标签时，自动构建并发布容器镜像。

工作流首次成功运行后，用户可以克隆仓库，也可以只下载 [docker-compose.yml](./docker-compose.yml) 并直接启动容器。

## Roadmap

- ZIP 批量转换支持部分成功，并生成失败报告。
- 增强输入校验与错误提示。
- 为批量转换提供更清晰的单文件进度或状态。
- 补充版本化二进制发布包。

## 致谢

感谢真诚、友善、团结、专业的 [LinuxDo](https://linux.do) 社区，让我学到很多开发和 AI 相关的知识和玩法。

[LinuxDo](https://linux.do) - 学 AI，上 L 站。

## Contributing

欢迎提交 Issue 或 Pull Request。更多说明见 [CONTRIBUTING.md](./CONTRIBUTING.md)。

## License

Released under the [MIT License](./LICENSE).
