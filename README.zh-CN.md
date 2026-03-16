<img src="web/frontend/public/favicon.svg" alt="TeamSphere" width="20%" height="20%" align="center" style="display: block; margin: 0 auto;"> 

# TeamSphere


企业级团队协作聊天系统，提供房间与私聊的实时沟通、通知与文件传输能力。

[English](README.md)

[![License](https://img.shields.io/badge/License-MIT-blue)](License)
[![Go](https://img.shields.io/badge/Go-1.25.0-00ADD8)](https://go.dev/dl/)
[![Node](https://img.shields.io/badge/Node-22.21.1-339933)](https://nodejs.org/)



## 目录

- [简介](#简介)
- [主要特性](#主要特性)
- [截图](#截图)
- [技术栈](#技术栈)
- [快速开始](#快速开始)
- [配置](#配置)
- [常用命令](#常用命令)
- [目录结构](#目录结构)
- [License](#license)


## 简介

TeamSphere 以 Go 为后端、React 为前端，提供 WebSocket 实时消息、房间管理、私聊、好友与邀请等核心协作能力。


## 主要特性

- 实时聊天与房间管理
- 私聊与好友系统
- 系统级通知
- 邀请链接与权限控制
- 文件上传与下载



## 截图

- ![截图 1](docs/assets/screenshot-1.png)

- ![截图 2](docs/assets/screenshot-2.png)

- ![截图 3](docs/assets/screenshot-3.png)
## 技术栈

- 后端：Go，Gin，PostgreSQL，Redis，JWT，WebSocket
- 前端：React，Vite，TypeScript，Zustand


## 快速开始

### Docker Compose

1. 启动服务。
2. 打开 Web UI 完成 Setup 引导，会在 `/app/data/config.yaml` 生成配置文件。

```bash
mkdir -p data

# 创建 .env 文件，内容示例：
# POSTGRES_USER=team_sphere
# POSTGRES_PASSWORD=your_db_password
# POSTGRES_DB=team_sphere

docker compose up --build
```

服务端默认端口为 8080。

### 本地开发

1. 启动后端。
2. 打开 Web UI 完成 Setup 引导（会生成 `config.yaml`）。
3. 启动前端。

```bash
# 后端
go run ./cmd/server/main.go 
或者
make run

# 前端
cd web\frontend
npm install
npm run dev
或者
make web
```

前端默认地址为 http://localhost:5173 。
## 配置

Setup 引导会生成 `config.yaml`，通常无需手动创建或编辑。
默认路径为 `config.yaml`，也可通过 `TEAMSPHERE_CONFIG_PATH` 覆盖。

配置项支持环境变量覆盖，格式为 `TEAMSPHERE_` 前缀，例如 `TEAMSPHERE_DATABASE_HOST`。

安全密钥请按注释生成：

```bash
openssl rand -hex 32
openssl rand -hex 64
```

默认配置文件为 `config.yaml`，可通过环境变量 `TEAMSPHERE_CONFIG_PATH` 指定路径。

配置项支持环境变量覆盖，格式为 `TEAMSPHERE_` 前缀，例如 `TEAMSPHERE_DATABASE_HOST`。

安全密钥请按注释生成：

```bash
openssl rand -hex 32
openssl rand -hex 64
```


## 常用命令

```bash
# 构建
make build

# 迁移
make migrate

# 清理
make clean
```


## 目录结构

- `cmd/`：入口与启动逻辑
- `internal/`：核心业务逻辑与服务实现
- `web/`：前端项目
- `docs/`：文档
- `uploads/`：上传文件目录


## License

MIT License。
