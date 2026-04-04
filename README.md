# s-ui (sub edition)

基于 [alireza0/s-ui](https://github.com/alireza0/s-ui) 改造的面板，移除了本地用户管理和订阅功能。

## 改造内容

- ✅ 移除订阅面板和订阅链接生成
- ✅ 移除本地用户注册/登录
- ✅ 移除 ruleset 在线订阅下载
- ✅ 移除 telegram/bot 订阅服务
- ✅ 移除 sub 独立订阅服务 (`sub.Server`)
- ✅ 移除 s-ui 内嵌 sing-box（节点侧由 pupmsub/sub 接管）
- ✅ 移除 Token 认证（APIv2）
- ✅ 保留节点入站配置展示
- ✅ 保留用户列表只读展示（从 `/etc/sub/singbox.json` 读取）
- ✅ 保留 sing-box 配置读取
- ✅ 保留路由/出站配置展示

## 依赖

- Go 1.24+
- Node.js (前端构建，如需重新打包)
- sing-box v1.13.4+

## 构建

```bash
go build -o s-ui .
```

## 运行

```bash
./s-ui
```

## 数据来源

节点入站信息和用户列表从 sub (pupmsub) 的本地文件读取：
- `/etc/sub/singbox.json` - sing-box 原生配置
- `/etc/sub/config.json` - V2bX 节点配置
