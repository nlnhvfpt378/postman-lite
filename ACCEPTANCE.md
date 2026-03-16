# 验收步骤与结果

## 环境

- 构建环境：Linux x86_64
- Go：1.22.12（用户态安装）
- 版本：0.1.0

## 验收步骤

### 1. 构建 Linux / Windows 交付物

```bash
cd /home/node/clawd/postman-lite
PATH=/home/node/clawd/.local/go/bin:$PATH make package-cross VERSION=0.1.0
```

预期产物：

- `deliverables/postman-lite_0.1.0_linux_amd64.tar.gz`
- `deliverables/postman-lite_0.1.0_windows_amd64.zip`
- `deliverables/postman-lite`
- `deliverables/postman-lite.exe`

### 2. 构建 Debian 安装包

```bash
cd /home/node/clawd/postman-lite
PATH=/home/node/clawd/.local/go/bin:$PATH make package-deb VERSION=0.1.0
```

预期产物：

- `deliverables/postman-lite_0.1.0_amd64.deb`

### 3. 启动应用

```bash
/home/node/clawd/postman-lite/deliverables/postman-lite
```

结果：应成功监听本地随机端口，并打印类似日志：

```text
Postman Lite listening on http://127.0.0.1:37033/
```

### 4. 健康检查验证

```bash
curl http://127.0.0.1:37033/healthz
```

结果：返回 `ok`

### 5. 请求发送验证

```bash
curl http://127.0.0.1:37033/api/send \
  -H 'Content-Type: application/json' \
  -d '{
    "method":"POST",
    "url":"https://httpbin.org/anything",
    "headers":[{"key":"Content-Type","value":"application/json"}],
    "body":"{\"name\":\"clawd\"}"
  }'
```

结果：成功返回 `200 OK`，包含响应头、响应体、大小等信息。

### 6. 历史记忆验证

打开页面后修改 `method/url/headers/body`，关闭页面并重新打开。

结果：四项字段应自动恢复；保存键为 `postman-lite.form.v1`。

## 结论

- Linux 打包：通过
- Windows 打包：通过
- `.deb` 打包：通过
- 基础请求发送：通过
- 历史记忆：通过
