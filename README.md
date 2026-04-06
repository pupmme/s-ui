# pupmmesub

基于 [alireza0/s-ui](https://github.com/alireza0/s-ui) 改造的 sing-box 节点管理器（xboard sub-node agent），移除了本地用户管理和订阅功能。

## 核心定位

节点侧 agent，通过 xboard v2 协议（handshake → pull config/users → report traffic）接入 xboard 管理面板。每个节点独立运行，通过 xboard 下发的差异化 sing-box inbound 配置工作。

## 架构

```
xboard (管理面板)
    │  ← xboard v2 协议 (handshake / get_config / get_users / push_traffic)
    ↓
pupmmesub (节点侧 agent)
    │  ← /etc/sub/config.json (xboard 连接配置)
    ↓
pupmmesub/sub (sing-box 核心)
    │  ← /etc/sub/singbox.json (节点 inbound + 用户配置)
    ↓
sing-box (sing-box 内核)
```

## 功能

- ✅ xboard v2 协议对接（handshake → sync → report）
- ✅ 节点差异化 inbound 配置下发
- ✅ 流量上报（per-user delta）
- ✅ 系统状态上报（CPU / 内存 / 磁盘）
- ✅ 移除订阅、注册登录、ruleset 下载、telegram bot
- ✅ 前端只读展示 inbounds / outbounds / clients

## 构建

```bash
# 前端
cd frontend && npm i && npm run build
cp -R frontend/dist web/

# 后端 (glibc)
CGO_ENABLED=0 go build -ldflags '-w -s' -o pupmmesub .

# 后端 (musl, for Alpine/Docker)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -s -extldflags "-static"' -tags "with_quic,with_grpc,with_utls,with_acme,with_gvisor" -o pupmmesub .
```

## 运行

```bash
# 节点模式（接入 xboard）
./pupmmesub run
./pupmmesub node      # 查看节点状态
./pupmmesub sync      # 手动触发一次同步
./pupmmesub restart   # 重启 sing-box

# 面板模式（独立 Web UI）
./pupmmesub web
```

## 配置

- `/etc/sub/config.json` — xboard 连接信息（apiHost / apiKey / nodeId / nodeType）
- `/etc/sub/singbox.json` — sing-box 配置（inbounds / outbounds / clients）

## 协议

[xboard v2 Server API]: https://github.com/w panels/xboard/blob/master/doc/v2-server-api.md
