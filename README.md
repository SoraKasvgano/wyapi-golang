# WyAPI-Golang

Go 版本的网易云音乐解析服务，前后端分离但可通过 `embed` 打包为单一二进制运行。内置 Vue 3 + Element Plus 前端，提供标准 API 与离线 Swagger UI 文档。

## 原项目

- [backend](https://github.com/Suxiaoqinx/Netease_url)
- [frontend](https://github.com/Suxiaoqinx/netease-vue)
- [demo](https://wyapi.toubiec.cn/)

## 特性

- Go 1.22 实现，默认关闭 cgo
- 前后端分离，生产环境通过 `embed` 打包前端静态资源
- 首次启动自动生成 `config.json` 与 `api_token`
- 可选 API Token 认证（默认关闭，方便本地使用）
- 支持 cookie.txt，提升高音质解析成功率
- 临时文件支持内存模式（可配置落盘）
- Swagger UI 离线文档：`/swagger/index.html`
- 兼容旧接口与新版前端接口

## 目录结构

```
wyapi-golang/
├── cmd/server/           # 服务入口
├── internal/             # 核心业务逻辑
├── frontend/             # 前端源码（Vue 3）
├── docs/openapi.json     # Swagger/OpenAPI 文档
├── assets.go             # go:embed 入口
├── config.json           # 运行时配置（自动生成）
├── cookie.txt            # 网易云 cookie（可选）
└── dist/                 # 构建产物（build 脚本生成）
```

## 快速开始（本地）

### 依赖

- Go 1.22+
- Node.js 20.19+ / 22.12+（仅构建前端需要）

### 构建并运行

Windows:

```
build.bat
dist\wyapi-golang_windows_amd64.exe
```

Linux/macOS:

```
chmod +x build.sh
./build.sh
./dist/wyapi-golang_linux_amd64
```

默认端口：`8000`  
前端入口：`http://127.0.0.1:8000/`

> 如果你直接打开 `frontend/dist/index.html` 或通过 `npm run dev` 访问，可能会导致 API 不同源。

### 跳过前端构建

```
SKIP_FRONTEND=1 ./build.sh
```

Windows:

```
set SKIP_FRONTEND=1
build.bat
```

## 配置说明（config.json）

首次启动自动生成，示例：

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8000
  },
  "security": {
    "api_token": "xxxxxxxx",
    "require_token": false
  },
  "cookie": {
    "file": "cookie.txt"
  },
  "download": {
    "dir": "downloads",
    "in_memory": true
  }
}
```

启用 Token 后，可在请求头添加 `X-API-Token` 或 `Authorization: Bearer <token>`。

## Cookie 说明（cookie.txt）

`cookie.txt` 建议写一行：

```
MUSIC_U=...; MUSIC_A=...; __csrf=...; ...
```

支持：
- `Cookie:` 前缀
- BOM 自动处理
- 单行或换行分隔

高音质（lossless/hires 等）通常需要有效 cookie 才能解析。

## API 与文档

- Swagger UI：`/swagger/index.html`
- OpenAPI：`/openapi.json`

常用接口：

```
POST /api/music/url
POST /api/music/detail
POST /api/music/lyric
POST /api/music/playlist
POST /api/music/album
POST /netease/search
```

示例：

```
curl -X POST http://127.0.0.1:8000/api/music/url \
  -H "Content-Type: application/json" \
  -d "{\"id\":\"476899057\",\"level\":\"standard\"}"
```

## Docker 部署

构建二进制：

```
./build.sh
```

构建并更新容器：

```
chmod +x docker-update.sh
./docker-update.sh
```

`docker-compose.yml` 默认映射 `./data:/data`，因此：

- `/data/config.json`
- `/data/cookie.txt`
- `/data/downloads`

都会持久化到宿主机 `./data` 目录。

## 常见问题

1) 提示“该歌曲已下架或无法获取”  
请先尝试切换到 `standard` 音质，并检查 `cookie.txt` 是否包含 `MUSIC_A` 和 `MUSIC_U`。

2) 前端与后端未结合  
务必通过 `http://127.0.0.1:8000/` 访问，不要打开本地 `dist/index.html`。

## 免责声明

仅供学习交流使用，请勿用于商业用途。
