# GitHub Actions Workflows

## build-offline-package.yml

构建 NTA 离线部署包的 GitHub Actions 工作流。

### 触发条件

- 推送到 `main` 或 `master` 分支
- 手动触发 (workflow_dispatch)

### 构建内容

1. **Docker 镜像**
   - nta-server (后端服务)
   - nta-web (前端服务)
   - postgres:15-alpine
   - redis:7-alpine
   - prom/prometheus:latest
   - grafana/grafana:latest

2. **离线安装包**
   - Docker 24.0.7
   - Docker Compose 2.23.0

3. **部署脚本**
   - install.sh (一键安装脚本)
   - uninstall.sh (卸载脚本)

4. **配置文件**
   - nta.yaml.example (配置模板)
   - docker-compose.yml (编排文件)

### 输出产物

- 文件名: `nta-offline-deploy-{VERSION}-{DATE}.zip`
- 校验文件: `nta-offline-deploy-{VERSION}-{DATE}.zip.sha256`
- 保留时间: 30天

### 使用方法

下载 Artifacts 中的部署包，解压后运行：

```bash
sudo bash install.sh
```

### 支持平台

- CentOS 7+
- Ubuntu 20.04+
- Anolis OS (龙蜥) 8
- x86_64 架构
