# Postman Lite

一个可下载即用的轻量 API 调试工具，现已改为 **Go + Fyne 原生桌面 UI**，不再依赖默认浏览器承载界面。

## 本次交付

- 原生桌面窗口：method / url / headers / body 输入 + send
- 响应展示：status / time / headers / body / size
- 本地文件持久化：重启后恢复 method / url / headers / body
- GitHub Actions 自动构建：Linux amd64 + Windows amd64
- Git tag 自动 Release：上传真实可下载产物

## 技术方案

- UI：Fyne 原生桌面 UI
- 请求执行：复用现有 `internal/httpclient` 核心逻辑
- 持久化文件：
  - Linux: `~/.config/postman-lite/state.json`
  - Windows: `%AppData%/postman-lite/state.json`

## 目录

- `cmd/postman-lite`：程序入口
- `internal/app`：应用装配
- `internal/httpclient`：HTTP 请求执行
- `internal/model`：请求/响应模型
- `internal/state`：本地状态文件读写
- `internal/ui`：Fyne 桌面 UI
- `.github/workflows/release.yml`：tag 构建与自动发布
- `build/package-cross.sh`：Linux / Windows 打包脚本

## 本地开发

### Linux 构建

```bash
cd /home/node/clawd/projects/postman-lite
PATH=/home/node/clawd/.local/go/bin:$PATH make build GO=/home/node/clawd/.local/go/bin/go
```

> 说明：Fyne 在 Linux 下需要图形相关开发依赖（如 `libgl1-mesa-dev`、`xorg-dev` 等）。GitHub Actions 已自动安装这些依赖。

### 跨平台打包

```bash
cd /home/node/clawd/projects/postman-lite
PATH=/home/node/clawd/.local/go/bin:$PATH make package-cross GO=/home/node/clawd/.local/go/bin/go VERSION=0.2.0
```

产物：

- `deliverables/postman-lite_0.2.0_linux_amd64.tar.gz`
- `deliverables/postman-lite_0.2.0_windows_amd64.zip`

## Release

推送 tag 即可自动发布：

```bash
git tag v0.2.0
git push origin v0.2.0
```

也支持 `workflow_dispatch` 手动补发。

## UI 说明

最小功能保持明确：

- Method 选择
- URL 输入
- Headers 多行输入（每行 `Key: Value`）
- Body 多行输入
- Send
- 响应状态 / 耗时 / 响应头 / 响应体 / 大小
- JSON 美化
- 复制响应体
- 查看本地状态文件位置

## 持久化说明

以下字段会自动保存到本地状态文件：

- Method
- URL
- Headers
- Body

输入变更时即时保存，程序重启后自动恢复。
