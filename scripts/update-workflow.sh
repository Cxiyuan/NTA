#!/bin/bash
# GitHub Actions Workflow 增量补丁
# 在现有 .github/workflows/build-offline-package.yml 中添加以下步骤

# 1. 在 "Pull base images" 步骤后添加：
echo "添加流处理镜像拉取..."
cat >> workflow_patch.txt << 'EOF'

      - name: Pull streaming images
        run: |
          echo "📥 拉取Kafka/Flink/Zookeeper镜像..."
          docker pull bitnami/zookeeper:3.9
          docker pull bitnami/kafka:3.6
          docker pull flink:1.18-scala_2.12-java11
          
          echo "✅ 流处理镜像拉取完成"

EOF

# 2. 在 "Build zeek image" 步骤后添加：
cat >> workflow_patch.txt << 'EOF'

      - name: Build kafka-consumer image
        run: |
          echo "🔨 构建 Kafka Consumer 镜像..."
          
          cd $GITHUB_WORKSPACE
          
          docker buildx build \
            --platform linux/amd64 \
            -t nta-kafka-consumer:v1.0.0 \
            -f docker/kafka-consumer/Dockerfile \
            --build-arg VERSION=${VERSION} \
            --build-arg BUILD_TIME=${BUILD_TIME} \
            --build-arg GIT_COMMIT=${GIT_COMMIT} \
            --output type=docker,dest=/tmp/nta-kafka-consumer.tar \
            .
          
          if [ ! -f /tmp/nta-kafka-consumer.tar ]; then
              echo "❌ Kafka Consumer 镜像构建失败"
              exit 1
          fi
          
          echo "✅ Kafka Consumer 镜像构建完成"

EOF

# 3. 在保存镜像步骤中添加：
cat >> workflow_patch.txt << 'EOF'

      - name: Save streaming images
        run: |
          echo "💾 保存流处理组件镜像..."
          
          mkdir -p /tmp/nta-deploy/images
          
          docker save bitnami/zookeeper:3.9 -o /tmp/zookeeper.tar
          docker save bitnami/kafka:3.6 -o /tmp/kafka.tar
          docker save flink:1.18-scala_2.12-java11 -o /tmp/flink.tar
          
          mv /tmp/zookeeper.tar /tmp/nta-deploy/images/
          mv /tmp/kafka.tar /tmp/nta-deploy/images/
          mv /tmp/flink.tar /tmp/nta-deploy/images/
          mv /tmp/nta-kafka-consumer.tar /tmp/nta-deploy/images/
          
          echo "✅ 流处理镜像保存完成"
          
          # 显示镜像大小
          ls -lh /tmp/nta-deploy/images/*.tar | tail -4

EOF

# 4. 在复制 flink-jobs 目录：
cat >> workflow_patch.txt << 'EOF'

      - name: Copy Flink jobs
        run: |
          echo "📋 复制 Flink 作业文件..."
          
          mkdir -p /tmp/nta-deploy/flink-jobs
          cp -r flink-jobs/* /tmp/nta-deploy/flink-jobs/
          chmod +x /tmp/nta-deploy/flink-jobs/deploy-jobs.sh
          
          echo "✅ Flink 作业文件复制完成"

EOF

# 5. 更新 summary 部分：
cat >> workflow_patch.txt << 'EOF'

          ### 📋 包含内容
          - ✅ Docker 24.0.7 离线安装包
          - ✅ Docker Compose 2.23.0
          - ✅ NTA 后端镜像 (nta-server)
          - ✅ NTA 前端镜像 (nta-web)
          - ✅ Zeek 探针镜像 (nta-zeek)
          - ✅ Kafka Consumer 镜像 (nta-kafka-consumer)
          - ✅ PostgreSQL 15 镜像
          - ✅ Redis 7 镜像
          - ✅ Zookeeper 3.9 镜像
          - ✅ Kafka 3.6 镜像
          - ✅ Flink 1.18 镜像
          - ✅ Prometheus 镜像
          - ✅ Grafana 镜像
          - ✅ Flink 流处理作业
          - ✅ 一键安装脚本
          - ✅ 配置文件模板
          - ✅ 部署文档

EOF

echo "✅ Workflow 补丁文件已生成: workflow_patch.txt"
echo ""
echo "📝 手动应用步骤："
echo "1. 打开 .github/workflows/build-offline-package.yml"
echo "2. 在相应位置插入 workflow_patch.txt 中的内容"
echo "3. 提交并推送到 GitHub"
