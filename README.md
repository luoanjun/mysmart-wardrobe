# MySmart Wardrobe - 智能衣柜管理系统

基于 Go + React + SQLite 的智能衣柜管理应用，支持本地AI图片识别，零Token成本。

## 功能特性

- **图片优先上传** - 上传即保存，异步识别，无需等待
- **本地AI识别** - 集成阿里万物识别模型，按需启停Docker容器，零Token消耗
- **云端AI备选** - 兼容 OpenAI Vision API 格式的云端识别
- **智能分类** - 三级分类体系（类别 → 子类别 → 颜色大类）
- **图片压缩** - 自动压缩至 800×800，JPEG 80% 质量
- **识别状态追踪** - 实时显示 pending/processing/completed/failed
- **失败重试** - 识别失败支持一键重新识别
- **Docker 一键部署** - 单二进制嵌入前端，开箱即用

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.21 + Gin |
| 前端 | React 18 + Vite + TailwindCSS |
| 数据库 | SQLite (纯Go驱动 glebarez/sqlite) |
| 缓存 | 内存缓存 + TTL |
| AI识别 | 阿里万物识别模型 (本地) / OpenAI Vision API (云端) |
| 部署 | Docker + docker-compose |

## 项目结构

```
├── main.go                 # 入口，路由，静态文件服务
├── embed.go                # go:embed 前端资源
├── Dockerfile              # 多阶段构建
├── docker-compose.yml      # 部署配置
├── cache/
│   └── cache.go            # 内存缓存
├── config/
│   └── config.go           # 配置管理
├── database/
│   └── database.go         # SQLite 初始化
├── handlers/
│   └── handlers.go         # HTTP 请求处理
├── models/
│   └── models.go           # 数据模型
├── services/
│   ├── ai.go               # AI识别服务（本地/云端双模式）
│   ├── docker_manager.go   # 按需启停AI容器
│   └── image.go            # 图片处理与异步识别
├── local-ai/
│   ├── app.py              # 本地AI识别服务 (Flask)
│   ├── Dockerfile          # AI服务镜像
│   └── requirements.txt    # Python依赖
└── frontend/
    ├── src/
    │   ├── pages/
    │   │   ├── Wardrobe.tsx     # 衣柜主页面
    │   │   ├── AddClothes.tsx   # 添加衣服
    │   │   └── Settings.tsx     # AI设置
    │   ├── components/
    │   │   └── Layout.tsx       # 侧边栏布局
    │   ├── services/
    │   │   └── api.ts           # API调用
    │   └── types/
    │       └── index.ts         # 类型定义
    └── ...配置文件
```

## 快速开始

### Docker 部署（推荐）

```bash
# 1. 构建主应用镜像
docker build -t wardrobe:latest .

# 2. 构建本地AI镜像
cd local-ai && docker build -t wardrobe-local-ai:latest . && cd ..

# 3. 启动服务
docker compose up -d
```

访问 `http://localhost:15156`

### 本地开发

```bash
# 后端
go run .

# 前端（开发模式）
cd frontend && npm install && npm run dev
```

## AI 识别模式

### 本地模式（默认，免费）

- 使用阿里万物识别模型（186MB，12000+中文类别）
- DockerManager 按需启停容器，5分钟无请求自动释放内存
- 适合低配NAS设备，识别速度约 25-40秒/张
- 完全免费，无Token消耗

### 云端模式（备选）

- 兼容 OpenAI Vision API 格式
- 支持豆包、GPT-4V 等视觉模型
- 需配置 API URL、模型名称和 API Key

## 分类体系

### 服装类别

| 大类 | 子类 |
|------|------|
| 上衣 | T恤、衬衫、卫衣、针织衫/毛衣、打底衫、马甲/背心 |
| 外套 | 夹克、牛仔外套、西装外套、风衣、棉服/棉袄、羽绒服、大衣 |
| 下装 | 牛仔裤、休闲裤、运动裤、西裤、短裤、半身裙 |
| 裙装 | 连衣裙、吊带裙、背带裙 |
| 鞋 | 板鞋/帆布鞋、运动鞋、休闲皮鞋、凉鞋、短靴、长靴 |

### 颜色分类

| 大类 | 颜色 |
|------|------|
| 无彩色 | 黑、白、灰 |
| 中性色 | 卡其、驼色、牛仔蓝、藏青 |
| 暖色 | 红、橙、黄、粉 |
| 冷色 | 蓝、绿、紫 |

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DB_PATH` | `./data/wardrobe.db` | 数据库路径 |
| `UPLOAD_PATH` | `./uploads` | 图片存储路径 |

## License

MIT
