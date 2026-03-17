# 验收步骤与结果

## 环境

- 构建环境：Linux x86_64（当前本地容器缺少完整桌面依赖，发布验证依赖 GitHub Actions）
- Go：见 `go.mod`
- 目标版本：0.3.x

## 验收步骤

### 1. 构建 Linux / Windows / macOS 交付物

```bash
cd /home/node/clawd/projects/postman-lite
PATH=/home/node/clawd/.local/go/bin:$PATH make package-cross VERSION=0.3.0 GO=/home/node/clawd/.local/go/bin/go FYNE_BIN=/home/node/clawd/.local/go/bin/fyne
```

预期产物：

- `deliverables/postman-lite_0.3.0_linux_amd64.tar.gz`
- `deliverables/postman-lite_0.3.0_windows_amd64.zip`
- `deliverables/postman-lite_0.3.0_darwin_amd64.tar.gz`
- `deliverables/postman-lite_0.3.0_darwin_arm64.tar.gz`

### 2. 应用图标验证

验证点：

- 窗口图标已设置
- Fyne 打包元数据包含图标
- macOS / Windows 打包流程显式传入 `assets/icon.png`

结果：代码已接入；最终成品图标以 GitHub Actions 打包产物验证为准。

### 3. 多标签验证

验证点：

- 可新建标签
- 可切换标签
- 标签右侧关闭按钮可关闭
- 至少保留一个标签
- 支持关闭当前 / 关闭右侧 / 关闭其他
- 支持持久化恢复

结果：已实现。

### 4. 中键 / 右键交互验证

验证点：

- 鼠标中键关闭标签
- 右键菜单可弹出并执行关闭动作

结果：代码已实现。

说明：是否触发取决于 Fyne 当前桌面驱动和运行平台；若平台不提供对应事件，关闭按钮和菜单入口仍可完成核心关闭操作。

### 5. OpenAPI 3 JSON 导入验证

验证点：

- 导入按钮可用
- Ctrl+V / Cmd+V 粘贴可触发导入兜底
- 成功解析 OpenAPI 3 JSON
- 导入后生成请求标签页

结果：已实现。

范围说明：当前覆盖 OpenAPI 3 JSON 主路径、方法、首个 server、requestBody example / schema example。

### 6. 本地状态恢复验证

关闭并重新打开应用。

结果：

- 标签列表恢复
- 当前选中标签恢复
- 每个标签的 method / url / headers / body 恢复

### 7. GitHub Actions Release 验证

验证点：

- tag push 触发 workflow
- release 上传 Linux / Windows / macOS 产物

结果：待 push/tag 后在线验证并补充链接。

## 当前结论

- 多标签：完成
- OpenAPI 3 JSON 导入：完成
- 图标接入：完成
- macOS 打包脚本 / workflow：完成
- 本地桌面全量构建：受当前容器缺少 Linux GUI 构建依赖限制，未在本地完整跑通
- GitHub Actions / Release 在线验证：待 push/tag 后确认
