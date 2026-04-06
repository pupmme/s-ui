# Contributing to pupmsub

欢迎贡献。

## 开发环境

```bash
git clone https://github.com/pupmme/s-ui.git
cd s-ui

# 前端
cd frontend && npm i && npm run build && cd ..

# 后端
CGO_ENABLED=0 go build -ldflags '-w -s' -o pupmsub .

# 运行
SUI_DB_FOLDER=db SUI_DEBUG=true ./pupmsub run
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SUI_DB_FOLDER` | 数据库目录 | `db` |
| `SUI_DEBUG` | 调试模式 | `true` |
| `SUI_BIN_FOLDER` | 二进制目录 | `bin` |
