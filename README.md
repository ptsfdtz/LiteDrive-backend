# 个人网盘后端（Go + Gin）

本项目是一个仅后端的在线网盘服务，使用 Go 语言与 Gin 框架构建，提供文件上传/下载、目录管理、分享与权限等核心能力，支持本地磁盘或 S3 兼容对象存储。

## 技术栈

- 语言：Go 1.21+
- Web 框架：Gin
- 数据库：PostgreSQL / MySQL / SQLite（三选一）
- 对象存储：本地文件系统或 S3 兼容（MinIO、AWS S3 等）
- 认证：JWT
- 其他：GORM（ORM，可选）、Zap/Logrus（日志，可选）、golangci-lint（静态检查，可选）

## 目录结构（建议）

以下是建议的 Gin 项目结构，便于后续扩展与维护：

```
online-disk-server/
├─ README.md
├─ go.mod / go.sum                # Go 依赖管理
├─ .env                           # 环境变量（本地开发，勿提交）
├─ .env.example                   # 环境变量示例（可提交）
├─ cmd/
│  └─ server/
│     └─ main.go                  # 程序入口，加载配置、初始化依赖、启动 HTTP 服务
├─ internal/                      # 业务实现（不对外暴露）
│  ├─ config/                     # 配置加载（env/yaml），默认值与校验
│  ├─ server/                     # Gin 引擎、全局中间件、路由注册
│  ├─ router/                     # 路由与分组（/auth, /files, /folders, /shares）
│  ├─ handler/                    # 控制器（参数校验、请求处理）
│  ├─ service/                    # 领域服务（存储、用户、分享、缩略图等）
│  ├─ repository/                 # 数据访问（GORM/SQL）
│  ├─ model/                      # 领域模型/实体（文件/用户/分享）
│  ├─ middleware/                 # 认证、限流、日志、恢复、CORS 等
│  ├─ storage/                    # 存储接口与实现（local、s3/minio）
│  ├─ auth/                       # JWT、密码哈希、权限检查
│  └─ pkg/                        # 工具库（日志、错误、响应格式化、字符串/时间）
├─ api/
│  └─ openapi.yaml                # OpenAPI 规范（可选）
├─ scripts/                       # 迁移、种子数据、备份脚本
├─ docker/                        # Dockerfile、docker-compose.yaml（可选）
├─ test/                          # 集成测试/基准测试（可选）
└─ LICENSE                        # 许可证（可选）
```

> 提示：当前仓库尚未包含上述所有目录，这是推荐结构。后续实现时可按需创建。

## 架构与模块

- cmd/server/main.go：程序入口，组合 config + server + router。
- internal/server：Gin Engine 初始化，注册全局中间件（恢复、CORS、请求日志、限流）。
- internal/router：按领域划分路由组，如 /auth, /files, /folders, /shares。
- internal/handler：解析请求、调用 service、统一返回格式。
- internal/service：业务逻辑（文件上传/下载、分享、回收站、去重、缩略图）。
- internal/storage：定义 Storage 接口并提供本地与 S3 实现。
- internal/repository：持久化（用户、文件元信息、分享记录）。
- internal/auth：JWT 发放与校验、密码哈希（bcrypt/argon2）。
- internal/middleware：鉴权、限流、日志、Tracing/Prometheus（可选）。

---

## 🛠️ 技术栈验证

- Go + Gin: Web 服务器正常运行
- MySQL + GORM: 数据库连接和操作正常
- JWT 认证: 安全认证机制工作正常
- MinIO S3: 对象存储服务集成成功
- OpenAPI 文档: API 文档生成和展示正常

## 部署建议

- 容器化：提供 `docker/Dockerfile` 与 `docker-compose.yaml`（可选）。
- 生产环境：
  - Gin 运行在 `release` 模式，前置 Nginx/Traefik 反代与 HTTPS。
  - 对象存储建议使用 MinIO/S3，避免本地磁盘成为单点。
  - 开启按需的缓存/CDN 加速公开文件。
