#!/bin/bash
set -e

# =============================================
# pupmme/s-ui 安装脚本
# =============================================

red='\033[0;31m'
green='\033[0;32m'
yellow='\033[0;33m'
blue='\033[0;34m'
plain='\033[0m'

NAME="s-ui"
BINARY_NAME="sui"
BIN_DIR="/usr/local/${NAME}"
CFG_DIR="/etc/sub"
BIN_PATH="${BIN_DIR}/${BINARY_NAME}"
CMD_PATH="/usr/bin/${BINARY_NAME}"
SERVICE_NAME="s-ui"
SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}.service"

GIT_RAW="https://raw.githubusercontent.com/pupmme/s-ui"
REPO_API="https://api.github.com/repos/pupmme/s-ui/releases"

error()   { echo -e "${red}[错误]${plain} $*"; }
info()    { echo -e "${blue}[信息]${plain} $*"; }
success() { echo -e "${green}[成功]${plain} $*"; }
warn()    { echo -e "${yellow}[警告]${plain} $*"; }

# check root
if [[ $EUID -ne 0 ]]; then
    error "必须使用 root 用户运行此脚本"
    exit 1
fi

arch_detect() {
    case "$(uname -m)" in
    x86_64|x64|amd64)   echo 'amd64' ;;
    aarch64|arm64)      echo 'arm64' ;;
    armv7l)             echo 'armv7' ;;
    *)                  echo 'amd64' ;;
    esac
}

systemd_reload() {
    if systemctl is-active --quiet ${SERVICE_NAME}; then
        systemctl stop ${SERVICE_NAME} 2>/dev/null || true
    fi
    if systemctl is-enabled --quiet ${SERVICE_NAME} 2>/dev/null; then
        systemctl disable ${SERVICE_NAME} 2>/dev/null || true
    fi
    systemctl daemon-reload
}

uninstall_old() {
    # 清理旧版 pupmsub / sub 安装
    if [[ -f /etc/systemd/system/sub.service ]] || [[ -f /etc/systemd/system/sing-box.service ]]; then
        warn "检测到旧版安装，正在清理..."
        systemd_reload 2>/dev/null || true
        rm -f /etc/systemd/system/sub.service
        rm -f /etc/systemd/system/sing-box.service
        rm -rf /usr/local/sub/
        rm -f /usr/bin/sub
        info "旧版清理完成"
    fi
    # 清理本工具旧安装
    if [[ -d "${BIN_DIR}" ]] || [[ -f "${CMD_PATH}" ]]; then
        warn "检测到本工具旧版安装，正在卸载..."
        systemd_reload
        rm -rf "${BIN_DIR}"
        rm -f "${CMD_PATH}"
        info "旧版卸载完成"
    fi
}

install_base() {
    info "安装基础依赖..."
    if command -v apt-get &>/dev/null; then
        apt-get update -qq && apt-get install -y -qq wget curl tar tzdata > /dev/null 2>&1
    elif command -v yum &>/dev/null; then
        yum install -y -q wget curl tar > /dev/null 2>&1
    elif command -v dnf &>/dev/null; then
        dnf install -y -q wget curl tar > /dev/null 2>&1
    fi
    success "基础依赖安装完成"
}

write_systemd() {
    cat > "${SERVICE_PATH}" << EOF
[Unit]
Description=pupmme s-ui panel
After=network.target
Wants=network.target

[Service]
Type=simple
ExecStart=${BIN_PATH}
Restart=on-failure
RestartSec=5s
StateDirectory=sub
StateDirectoryMode=0755
PermissionsStartOnly=true
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOF
    success "systemd service 已写入"
}

write_default_config() {
    mkdir -p "${CFG_DIR}"
    if [[ ! -f "${CFG_DIR}/config.json" ]]; then
        cat > "${CFG_DIR}/config.json" << EOF
{
  "log": {
    "level": "info",
    "output": ""
  },
  "web": {
    "port": 2053,
    "cert": "",
    "key": "",
    "username": "admin",
    "password": "$(head -c 12 /dev/urandom | base64)"
  },
  "node": false,
  "xboard": {
    "apiHost": "",
    "apiKey": "",
    "nodeId": 0,
    "nodeType": "",
    "timeout": 30,
    "listenIP": "0.0.0.0",
    "sendIP": "0.0.0.0",
    "tcpFastOpen": false,
    "sniffEnabled": false,
    "inboundConfig": {
      "protocol_options": {
        "heartbeat_interval": "3s",
        "idle_timeout": "300s"
      }
    },
    "certConfig": {
      "certMode": "dns",
      "certFile": "",
      "keyFile": "",
      "provider": "cloudflare",
      "dnsEnv": {}
    }
  }
}
EOF
        success "配置文件已写入 (${CFG_DIR}/config.json)"
    fi
}

fetch_binary() {
    local arch=$(arch_detect)
    local tmp_dir=$(mktemp -d)
    local tmp_tar="${tmp_dir}/s-ui.tar.gz"

    info "下载 s-ui ${arch} ..."

    if [[ $# -eq 0 ]]; then
        local latest=$(curl -sL "${REPO_API}/latest" | grep -oP '"tag_name":\s*"\K[^"]+')
        if [[ -z "${latest}" ]]; then
            error "无法获取最新版本，请检查网络连接"
            rm -rf "${tmp_dir}"
            exit 1
        fi
    else
        local latest=$1
    fi

    info "版本: ${latest}"

    local url="${REPO_API}/download/${latest}/s-ui-linux-${arch}.tar.gz"
    wget -q -O "${tmp_tar}" "${url}"
    if [[ $? -ne 0 ]]; then
        error "下载失败: ${url}"
        rm -rf "${tmp_dir}"
        exit 1
    fi

    tar -xzf "${tmp_tar}" -C "${tmp_dir}"
    local extracted_dir=$(ls "${tmp_dir}" | grep -v 'tar.gz' | head -1)
    if [[ -z "${extracted_dir}" ]]; then
        error "解压失败"
        rm -rf "${tmp_dir}"
        exit 1
    fi

    mkdir -p "${BIN_DIR}"
    cp "${tmp_dir}/${extracted_dir}/sui" "${BIN_PATH}"
    chmod +x "${BIN_PATH}"
    cp "${tmp_dir}/${extracted_dir}/s-ui.sh" "${BIN_DIR}/"
    chmod +x "${BIN_DIR}/s-ui.sh"
    ln -sf "${BIN_DIR}/s-ui.sh" "${CMD_PATH}"

    # 清理
    rm -rf "${tmp_dir}"

    success "二进制安装完成 (${BIN_PATH})"
}

enable_start() {
    systemctl daemon-reload
    systemctl enable ${SERVICE_NAME} --now
    sleep 1

    if systemctl is-active --quiet ${SERVICE_NAME}; then
        success "s-ui 服务已启动"
    else
        error "服务启动失败，请查看 journalctl -u ${SERVICE_NAME} -n 20"
        exit 1
    fi
}

show_info() {
    echo ""
    echo "============================================"
    success "s-ui 安装完成！"
    echo "============================================"
    echo ""
    echo "管理命令: ${CMD_PATH}"
    echo ""
    ${CMD_PATH} admin -show 2>/dev/null || true
    echo ""
    echo "systemd:  systemctl ${SERVICE_NAME} status|start|stop|restart"
    echo "日志:     journalctl -u ${SERVICE_NAME} -f"
    echo ""
}

# =============================================
# 主流程
# =============================================
main() {
    echo -e "${green}============================================"
    echo -e "  pupmme/s-ui 安装脚本"
    echo -e "============================================${plain}"
    echo ""

    install_base
    uninstall_old
    write_default_config
    fetch_binary "$@"
    write_systemd
    enable_start
    show_info
}

main "$@"
