# Postman Lite

一个轻量版 Postman 风格 API 调试工具，使用 Go 实现，交付为 Linux / Windows 可运行程序，并保留 `.deb` 安装包。

## 功能

- HTTP 方法：GET / POST / PUT / DELETE / PATCH
- URL 输入
- Headers 编辑（每行 `Key: Value`）
- Body 输入（raw / JSON）
- 发送请求
- 响应展示：状态码、耗时、响应头、响应体、大小
- JSON 自动美化（请求体手动美化、响应体自动识别 JSON）
- 历史记忆：关闭重开后自动恢复 method / url / headers / body

## 实现说明

本版本采用：

- Go 后端单二进制
- 内嵌静态 Web UI（本地监听 `127.0.0.1` 随机端口）
- 启动后自动尝试调用系统默认浏览器
- 前端通过 `localStorage` 持久化表单，键名：`postman-lite.form.v1`

这样做的原因：当前构建环境缺少 Wails/Fyne 所需的系统图形开发依赖，且无可用 root 提权安装链路；为了保证按时交付，优先选择零运行时开发依赖、可复现、可打包的方案。

## 目录

- `cmd/postman-lite`：程序入口
- `internal/app`：应用装配
- `internal/httpclient`：HTTP 请求执行
- `internal/model`：请求/响应模型
- `internal/ui`：本地 HTTP 服务与静态资源分发
- `build/package-cross.sh`：Linux / Windows 打包脚本
- `build/package-deb.sh`：`.deb` 打包脚本
- `deliverables/`：最终交付物

## 构建

前置：本仓库使用用户态 Go 1.22.12 构建。

### 普通构建

```bash
cd /home/node/clawd/postman-lite
PATH=/home/node/clawd/.local/go/bin:$PATH make build
```

### 跨平台打包（Linux + Windows）

```bash
cd /home/node/clawd/postman-lite
PATH=/home/node/clawd/.local/go/bin:$PATH make package-cross VERSION=0.1.0
```

产物：

- `deliverables/postman-lite_0.1.0_linux_amd64.tar.gz`
- `deliverables/postman-lite_0.1.0_windows_amd64.zip`
- `deliverables/postman-lite`
- `deliverables/postman-lite.exe`

### 构建 `.deb`

```bash
cd /home/node/clawd/postman-lite
PATH=/home/node/clawd/.local/go/bin:$PATH make package-deb VERSION=0.1.0
```

### 一次性构建全部交付物

```bash
cd /home/node/clawd/postman-lite
PATH=/home/node/clawd/.local/go/bin:$PATH make package-all VERSION=0.1.0
```

## 运行

### Linux

```bash
/home/node/clawd/postman-lite/deliverables/postman-lite
```

或：

```bash
/home/node/clawd/postman-lite/deliverables/bin/postman-lite
```

### Windows

解压后运行：

```text
postman-lite.exe
```

启动后会：

1. 监听 `127.0.0.1` 随机端口
2. 尝试自动打开默认浏览器
3. 若自动打开失败，终端日志会打印本地访问地址，手动复制到浏览器即可

## 历史记忆说明

以下字段会自动保存并在重新打开后恢复：

- Method
- URL
- Headers
- Body

触发时机：

- 页面加载时自动恢复
- 输入变更时自动保存
- 点击“发送请求”前再次保存

## 验收要点

- 可成功启动
- 可访问健康检查 `/healthz`
- 可向 `https://httpbin.org/anything` 发送 POST JSON 请求
- 返回状态、响应头、响应体与大小信息正常
- Linux / Windows 压缩包已生成
- `.deb` 已成功构建

详见：`ACCEPTANCE.md`

## 已知问题

1. 当前不是原生 WebView 桌面壳，而是“本地服务 + 默认浏览器”模式。
2. AppImage 未产出，原因是当前环境缺少 AppImage 工具链与图形依赖，且不在本次最小改动范围内。
3. 若目标机没有浏览器打开命令，程序不会自动弹浏览器，但仍会在终端打印可访问地址。
4. 当前未实现请求历史列表、环境变量、认证助手、集合管理等高级能力。
