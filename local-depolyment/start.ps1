# start.ps1

# 檢查 .env 檔案是否存在，若無則從 .example.env 建立
if (Test-Path ".env") {
    Write-Host "✅ .env 檔案已存在，略過建立。"
} elseif (Test-Path ".example.env") {
    Copy-Item ".example.env" ".env"
    Write-Host "✅ .env 檔案已從 .example.env 建立完成。"
} else {
    Write-Host "⚠️ 無法建立 .env，因為找不到 .example.env"
    exit 1
}

# 檢查 Docker 是否安裝
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "❌ 請先安裝 Docker，請至 https://docs.docker.com/desktop/setup/install/windows-install/ 並根據說明安裝 docker-desktop。"
    exit 1
}

# 判斷 docker compose 版本
$composeCmd = ""
if (Get-Command docker-compose -ErrorAction SilentlyContinue) {
    $composeCmd = "docker-compose"
} elseif ((docker compose version) -ne $null) {
    $composeCmd = "docker compose"
} else {
    Write-Host "❌ 沒有找到 docker-compose，請至 https://docs.docker.com/desktop/setup/install/windows-install/ 並根據說明安裝 docker-desktop。"
    exit 1
}

# 執行 compose 指令
Write-Host "✅ 開始啟動 Docker Compose..."
Invoke-Expression "$composeCmd -f docker-compose.yml up -d"