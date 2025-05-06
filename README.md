# TraeCN 知识库系统

## 项目概述
基于Gin框架构建的知识共享平台，集成AI问答、内容爬取和用户行为分析功能。

## 核心功能

### 用户系统
- 多条件用户搜索（支持分页）
- JWT鉴权体系
- 用户行为追踪（搜索/阅读记录）

### 知识管理
- 多源内容爬取（腾讯云开发者社区）
- 文章标签分类
- 智能搜索推荐

### AI集成
- DeepSeek大模型对话接口
- 会话历史记录
- Token用量统计

## 技术栈
- **后端框架**: Gin
- **数据库**: MySQL + GORM
- **AI集成**: DeepSeek API
- **爬虫引擎**: Colly + goquery
- **基础设施**: JWT鉴权、RateLimiter

## API文档

### 用户模块
- `POST /api/login` 用户登录
- `GET /api/users` 用户搜索（需鉴权）

### 知识库模块
- `GET /api/articles` 文章搜索
- `POST /api/search-history` 记录搜索行为
- `POST /api/reading-history` 记录阅读行为

### AI模块
- `POST /api/chat` AI对话接口
- `GET /api/chat-sessions` 获取会话历史

## 部署指南
```bash
# 安装依赖
go mod tidy

# 配置环境变量
cp .env.example .env

# 启动服务
go run main.go
```

## 贡献规范
1. 新功能开发需创建feature分支
2. 提交前需通过单元测试
3. 接口变更需更新API文档