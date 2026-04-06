# Deploy sub

## 快速部署

```bash
bash <(curl -Ls https://raw.githubusercontent.com/pupmme/s-ui/main/install.sh)
```

或手动：

```bash
# 1. 下载 release
wget https://github.com/pupmme/s-ui/releases/latest/download/pupmsub-linux-amd64.tar.gz

# 2. 解压
tar -xzf pupmsub-linux-amd64.tar.gz

# 3. 安装
cd sub
chmod +x install.sh
./install.sh
```

## systemd 日志

```bash
journalctl -u pupmsub -f
```

## 二进制路径

- `/usr/local/bin/sui` — 主程序
- `/etc/sub/singbox.json` — 数据文件
- `/etc/sub/config.json` — 配置文件

## 交叉编译

```bash
./deploy/build.sh amd64   # → dist/pupmsub-linux-amd64.tar.gz
./deploy/build.sh arm64   # → dist/sub-linux-arm64.tar.gz
```
