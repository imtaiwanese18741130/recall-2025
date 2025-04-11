#!/bin/bash

set -e  # 一旦有錯誤就中止腳本

# 檢查 .env 是否已存在
if [ -f .env ]; then
  echo "✅ .env 檔案已存在，略過建立。"
elif [ -f .example.env ]; then
  cp .example.env .env
  echo "✅ .env 檔案已從 .example.env 建立完成。"
else
  echo "⚠️ 無法建立 .env，因為找不到 .example.env"
  exit 1
fi

# 檢查 docker 是否安裝
if ! command -v docker &> /dev/null; then
  echo "❌ 請先安裝 Docker，請至 https://docs.docker.com/desktop/setup/install/mac-install/ 並根據說明安裝 docker-desktop。"
  exit 1
fi

# 判斷使用哪個 docker compose 指令
if command -v docker-compose &> /dev/null; then
  COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null; then
  COMPOSE_CMD="docker compose"
else
  echo "❌ 找不到 docker-compose，請至 https://docs.docker.com/desktop/setup/install/mac-install/ 並根據說明安裝 docker-desktop。"
  exit 1
fi

# 執行 compose 指令
echo "✅ Docker 與 Compose 工具已就緒，開始啟動服務..."
$COMPOSE_CMD -f docker-compose.yml up -d