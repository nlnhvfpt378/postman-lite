# Postman Lite

一个可下载即用的轻量 API 调试工具，基于 **Go + Fyne** 原生桌面 UI。

## 本次交付（v0.3.x）

- 原生桌面窗口：method / url / headers / body 输入 + send
- 响应展示：status / time / headers / body / size
- **Chrome 风格多标签**：
  - 新建 / 切换 / 关闭
  - 标签右侧关闭按钮
  - 鼠标中键关闭（桌面驱动支持时生效）
  - 右键菜单：关闭当前 / 关闭右侧 / 关闭其他
  - 至少保留一个标签
  - 持久化恢复标签和当前选中项
- **OpenAPI 3 JSON 导入**：
  - 导入入口按钮
  - Ctrl+V / Cmd+V 粘贴兜底触发
  - 导入后自动生成请求标签页
- **应用图标**：内置简洁圆角图标，并在 Fyne 应用与打包流程中生效
- GitHub Actions 自动构建：Linux amd64 / Windows amd64 / macOS amd64 / macOS arm64
- Git tag 自动 Release：上传真实可下载产物

## 技术方案

- UI：Fyne 原生桌面 UI
- 请求执行：复用现有 `internal/httpclient` 核心逻辑
- 持久化文件：
  - Linux: `~/.config/postman-lite/state.json`
  - macOS: `~/Library/Application Support/postman-lite/state.json`
  - Windows: `%AppData%/postman-lite/state.json`

## 目录

- `cmd/postman-lite`：程序入口
- `internal/app`：应用装配
- `internal/httpclient`：HTTP 请求执行
- `internal/model`：请求/响应模型
- `internal/state`：本地状态文件读写（含多标签恢复）
- `internal/ui`：Fyne 桌面 UI、图标、OpenAPI 导入
- `assets`：应用图标等静态资源
- `.github/workflows/release.yml`：tag 构建与自动发布
- `build/package-cross.sh`：Linux / Windows / macOS 打包脚本
- `FyneApp.toml`：Fyne 打包元数据

## 本地开发

### Linux 构建

```bash
cd /home/node/clawd/projects/postman-lite
PATH=/home/node/clawd/.local/go/bin:$PATH make build GO=/home/node/clawd/.local/go/bin/go
```

> 说明：Fyne 在 Linux 下需要图形相关开发依赖（如 `libgl1-mesa-dev`、`xorg-dev` 等）。

### 跨平台打包

```bash
cd /home/node/clawd/projects/postman-lite
PATH=/home/node/clawd/.local/go/bin:$PATH make package-cross GO=/home/node/clawd/.local/go/bin/go FYNE_BIN=/home/node/clawd/.local/go/bin/fyne VERSION=0.3.0
```

期望产物：

- `deliverables/postman-lite_0.3.0_linux_amd64.tar.gz`
- `deliverables/postman-lite_0.3.0_windows_amd64.zip`
- `deliverables/postman-lite_0.3.0_darwin_amd64.tar.gz`
- `deliverables/postman-lite_0.3.0_darwin_arm64.tar.gz`

## Release

推送 tag 即可自动发布：

```bash
git tag v0.3.0
git push origin v0.3.0
```

若 `v0.3.0` 已存在，则递增版本号后重试。

## UI 说明

### 多标签

- 点击 `新建` 创建请求标签
- 点击标签切换请求
- 点击标签右侧关闭按钮关闭
- 支持中键关闭（若当前平台的 Fyne 桌面驱动能提供中键事件）
- 支持右键菜单：关闭当前 / 关闭右侧 / 关闭其他
- 永远至少保留一个标签

### OpenAPI 导入

支持 **OpenAPI 3 JSON**：

- 点击 `导入 OpenAPI JSON`
- 或直接把 OpenAPI JSON 粘贴到窗口（Ctrl+V / Cmd+V）

导入后会按 `paths + method` 生成多个请求标签，并尽量带上：

- method
- url（优先拼接首个 server）
- request body example
- Content-Type

## 已知限制 / 降级说明

- Fyne 对标签原生中键/右键交互支持依赖桌面驱动；当前实现已提供：
  - 中键关闭：在支持 `MouseButtonTertiary` 的桌面环境下生效
  - 右键菜单：通过 Fyne popup menu 实现
  - 若平台事件受限，仍可通过关闭按钮和导入按钮完成核心操作
- OpenAPI 导入当前聚焦 **OpenAPI 3 JSON**，暂未覆盖 YAML、鉴权流和复杂 `$ref` 深度展开

## 持久化说明

以下内容会自动保存到本地状态文件：

- 所有请求标签
- 当前选中标签
- 每个标签的 Method / URL / Headers / Body

输入变更时即时保存，程序重启后自动恢复。
