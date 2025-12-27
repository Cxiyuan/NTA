# GitHub Actions Workflow 构建内容详解

## 工作流触发条件
- 推送到 `main` 或 `master` 分支
- 手动触发 (workflow_dispatch)

## 构建环境
- **运行环境**: `ubuntu-24.04`
- **Go 版本**: 1.23
- **Node.js 版本**: 20

---

## 📦 Workflow 会做的事情

### 1️⃣ **编译 Go 后端程序** (在 GitHub Actions 上编译)

#### nta-server
```bash
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-X main.Version=v2.0.0 -X main.BuildTime=2025-12-27..." \
  -o nta-server \
  ./cmd/nta-server
```
- **输出**: `bin/nta-server` (约 30MB)
- **架构**: Linux x86_64
- **CGO**: 启用 (依赖系统库)
- **版本信息**: 编译时注入

#### nta-kafka-consumer
```bash
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
  -o nta-kafka-consumer \
  ./cmd/kafka-consumer
```
- **输出**: `bin/nta-kafka-consumer` (约 20MB)
- **架构**: Linux x86_64
- **CGO**: 启用

---

### 2️⃣ **构建 Vue.js 前端** (在 GitHub Actions 上构建)

```bash
cd web
npm install
npm run build
```
- **输出**: `web/dist/*` (约 10MB)
- **内容**: 静态文件 (HTML, CSS, JS)
- **框架**: Vue.js + Vite
- **生产优化**: 压缩、Tree-shaking

---

### 3️⃣ **下载基础设施源码/二进制包** (wget 下载)

#### PostgreSQL 15.5 (源码)
```bash
wget https://ftp.postgresql.org/pub/source/v15.5/postgresql-15.5.tar.gz
```
- **类型**: 源码包
- **大小**: ~30MB (压缩)
- **说明**: 将在目标系统编译

#### Redis 7.2.3 (源码)
```bash
wget https://download.redis.io/releases/redis-7.2.3.tar.gz
```
- **类型**: 源码包
- **大小**: ~3MB (压缩)
- **说明**: 将在目标系统编译

#### Kafka 3.6.1 (二进制)
```bash
wget https://archive.apache.org/dist/kafka/3.6.1/kafka_2.13-3.6.1.tgz
```
- **类型**: Java 二进制包 (含 Zookeeper)
- **大小**: ~100MB (压缩)
- **说明**: 无需编译，直接运行

#### Zeek 6.0.3 (源码)
```bash
wget https://download.zeek.org/zeek-6.0.3.tar.gz
```
- **类型**: 源码包
- **大小**: ~50MB (压缩)
- **说明**: 将在目标系统编译 (耗时最长)

---

### 4️⃣ **打包配置文件和脚本**

#### 应用配置
- `config/nta.yaml` - 主配置文件
- `config/threat_feed.json` - 威胁情报配置 (可选)
- `config/license.key` - 许可证文件 (可选)
- `config/public.pem` - 公钥文件 (可选)

#### Zeek 脚本
- `zeek-scripts/*` - 自定义 Zeek 检测脚本
  - `main.zeek`
  - `lateral-scan.zeek`
  - `lateral-auth.zeek`
  - 等...

#### 部署脚本
- `install.sh` - 安装脚本
- `uninstall.sh` - 卸载脚本
- `scripts/init-databases.sh` - 数据库初始化脚本

---

### 5️⃣ **生成元数据文件**

#### VERSION.txt
```
NTA Network Traffic Analysis System
Version: v2.0.0
Build Time: 2025-12-27T12:34:56Z
Git Commit: a1b2c3d
Build Environment: Ubuntu 24.04 LTS

Components:
- Go: 1.23
- Node.js: 20
- PostgreSQL: 15.5 (source)
- Redis: 7.2.3 (source)
- Kafka: 3.6.1 (binary)
- Zeek: 6.0.3 (source)

Target Platform: Ubuntu 24.04 LTS x86_64
```

#### README.txt
- 系统要求
- 部署步骤
- 默认账户
- 访问地址
- 注意事项
- 故障排查

---

### 6️⃣ **创建离线部署包**

```bash
tar -czf nta-offline-deploy-a1b2c3d-20251227.tar.gz nta-build/
sha256sum nta-offline-deploy-a1b2c3d-20251227.tar.gz > nta-offline-deploy-a1b2c3d-20251227.tar.gz.sha256
```

#### 部署包结构
```
nta-offline-deploy-a1b2c3d-20251227/
├── bin/
│   ├── nta-server              (已编译 - 30MB)
│   └── nta-kafka-consumer      (已编译 - 20MB)
├── web/                        (已构建 - 10MB)
│   ├── index.html
│   ├── assets/
│   └── ...
├── packages/                   (源码包 - 183MB)
│   ├── postgresql-15.5.tar.gz  (30MB)
│   ├── redis-7.2.3.tar.gz      (3MB)
│   ├── kafka_2.13-3.6.1.tgz    (100MB)
│   └── zeek-6.0.3.tar.gz       (50MB)
├── config/
│   └── nta.yaml
├── zeek-scripts/
│   └── *.zeek
├── scripts/
│   └── init-databases.sh
├── install.sh
├── uninstall.sh
├── VERSION.txt
└── README.txt
```

**总大小**: 约 250-300MB (压缩后)

---

### 7️⃣ **上传构建产物**

```yaml
- name: nta-offline-deploy-ubuntu-24.04-a1b2c3d
  files:
    - nta-offline-deploy-a1b2c3d-20251227.tar.gz
    - nta-offline-deploy-a1b2c3d-20251227.tar.gz.sha256
  retention: 30 days
```

---

## 📊 编译 vs 下载 对比表

| 组件 | 操作 | 在哪里执行 | 输出大小 |
|------|------|-----------|----------|
| **nta-server** | ✅ 编译 | GitHub Actions (Ubuntu 24.04) | 30MB |
| **nta-kafka-consumer** | ✅ 编译 | GitHub Actions (Ubuntu 24.04) | 20MB |
| **Vue.js 前端** | ✅ 构建 | GitHub Actions | 10MB |
| **PostgreSQL** | ⬇️ 下载源码 | - | 30MB |
| **Redis** | ⬇️ 下载源码 | - | 3MB |
| **Kafka** | ⬇️ 下载二进制 | - | 100MB |
| **Zeek** | ⬇️ 下载源码 | - | 50MB |

---

## 🔄 完整流程图

```
GitHub Actions (Ubuntu 24.04)
│
├─[1] 编译 Go 后端
│     ├─ nta-server (CGO_ENABLED=1)
│     └─ nta-kafka-consumer
│
├─[2] 构建 Vue.js 前端
│     └─ npm install + npm run build
│
├─[3] 下载基础设施包
│     ├─ wget postgresql-15.5.tar.gz
│     ├─ wget redis-7.2.3.tar.gz
│     ├─ wget kafka_2.13-3.6.1.tgz
│     └─ wget zeek-6.0.3.tar.gz
│
├─[4] 打包配置文件
│     ├─ config/
│     ├─ zeek-scripts/
│     └─ scripts/
│
├─[5] 生成元数据
│     ├─ VERSION.txt
│     └─ README.txt
│
├─[6] 创建 tar.gz 压缩包
│     └─ nta-offline-deploy-*.tar.gz (250-300MB)
│
└─[7] 上传到 GitHub Artifacts
      └─ 保留 30 天
```

---

## ✅ 总结

### Workflow 编译的内容 (在 GitHub Actions 上)
1. ✅ **nta-server** - Go 后端主程序
2. ✅ **nta-kafka-consumer** - Go Kafka 消费者
3. ✅ **Web 前端** - Vue.js 静态文件

### Workflow 下载的内容 (wget)
1. ⬇️ **PostgreSQL 15.5** - 源码包 (目标系统编译)
2. ⬇️ **Redis 7.2.3** - 源码包 (目标系统编译)
3. ⬇️ **Kafka 3.6.1** - Java 二进制包 (直接运行)
4. ⬇️ **Zeek 6.0.3** - 源码包 (目标系统编译)

### 为什么这样设计？
- **编译 Go 程序**: 确保与 Ubuntu 24.04 兼容
- **下载源码包**: PostgreSQL/Redis/Zeek 需要在目标系统编译以适配具体环境
- **下载二进制包**: Kafka 是 Java 应用，平台无关
- **构建前端**: 生产环境优化，减小体积

### 最终产物
一个 **250-300MB** 的离线部署包，包含：
- 预编译的 NTA 应用程序
- 基础设施源码/二进制包
- 配置文件和部署脚本
- 完整的安装说明

用户只需下载这一个文件，在 Ubuntu 24.04 上运行 `install.sh` 即可完成部署。✅
