# recall2025

## 聲明
1. 本 repo 由 OurTaiwan (以下簡稱本團隊) 所維護。
1. 本 repo 之設計與程式碼，為本團隊所有，僅供查看，禁止未經授權的分發或商業用途。所有權利受法律保護，違者必究。
1. 如有任何建議或問題，歡迎來信：<imtaiwanese18741130@gmail.com>

#### macOS 本機端部署方法
- 請至 Repo Releases 下載最新版本程式，並解壓縮 [連結](https://github.com/imtaiwanese18741130/recall-2025/releases/tag/localhost-server-v0.0.2)
- 請安裝 Docker Desktop [連結](https://docs.docker.com/desktop/setup/install/mac-install/)
- 開啟 終端機(terminal)，可參考該[連結](https://support.apple.com/zh-tw/guide/terminal/apd5265185d-f365-44cb-8b09-71a064a42125/mac)
- 前往程式檔案位置 (從上一步提取的資料夾根目錄)
    - (作法1) 在終端機中，使用 cd 指令前往剛剛解壓縮的資料夾(或者其他路徑)，例如：
        ```
        cd ~/Downloads/recall-2025-localhost-server-v0.0.2
        ```
    - (作法2) 在終端機中輸入 `cd `（包含空格），然後將剛剛解壓縮的資料夾從 Finder 拖曳到終端機視窗，接著按下 Enter 鍵。
- 在當前路徑執行程式 `/bin/sh local-depolyment/start.sh`
- 開啟瀏覽器，並前往 `http://localhost:8964/`

#### Windows 本機端部署方法
- 請至 Repo Releases 下載最新版本程式，並解壓縮 [連結](https://github.com/imtaiwanese18741130/recall-2025/releases/tag/localhost-server-v0.0.2)
- 請安裝 Docker Desktop [連結](https://docs.docker.com/desktop/setup/install/windows-install/)
- 開啟 PowerShell，可以參考該[連結](https://learn.microsoft.com/zh-tw/powershell/scripting/learn/ps101/01-getting-started?view=powershell-7.4)
- 前往程式檔案位置 (從上一步提取的資料夾根目錄)
    - (作法1) 在 PowerShell 中，使用 cd 指令前往剛剛解壓縮的資料夾(或者其他路徑)，例如：
        ```
        cd "C:\Users\你的使用者名稱\Downloads\recall-2025-localhost-server-v0.0.2"
        ```
    - (作法2) 在 PowerShell 中輸入 `cd `（包含空格），然後將剛剛解壓縮的資料夾從檔案總管拖曳到 PowerShell 視窗，接著按下 Enter 鍵。
- 在當前路徑執行程式 `.\local-deployment\start.ps1`
- 開啟瀏覽器，並前往 `http://localhost:8964/`