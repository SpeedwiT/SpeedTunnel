#!/usr/bin/env bash
set -e
# Speed Tunnel - Installer
# Usage: bash <(curl -s https://raw.githubusercontent.com/SpeedwiT/SpeedTunnel/main/install.sh)
#        bash install.sh --update
# Github: https://github.com/SpeedwiT/SpeedTunnel

REPO="SpeedwiT/SpeedTunnel"
BIN="/usr/local/bin/speedtunnel"
CONFIG_DIR="/etc/speedtunnel"
CONFIG="$CONFIG_DIR/config.json"
SERVICE="speedtunnel"
VERSION="1.0.0"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m'

is_update=false
[[ "$1" == "--update" ]] && is_update=true

need_root() {
  if [[ $EUID -ne 0 ]]; then
    echo -e "${RED}لطفا با sudo اجرا کنید: sudo bash install.sh${NC}"
    exit 1
  fi
}

detect_arch() {
  arch=$(uname -m)
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) echo "amd64" ;;
  esac
}

install_deps() {
  echo -e "${CYAN}[1/6] نصب پیش‌نیازها...${NC}"
  apt-get update -qq 2>&1 | tail -1 || true
  apt-get install -y curl wget jq openssl tar 2>&1 | tail -5 || true
}

build_or_download() {
  echo -e "${CYAN}[2/6] دریافت/ساخت باینری...${NC}"
  arch=$(detect_arch)
  # Try download release first
  url="https://github.com/${REPO}/releases/latest/download/speedtunnel-linux-${arch}"
  if curl -fsSL "$url" -o "$BIN" 2>/dev/null && [[ -s "$BIN" ]]; then
    echo -e "${GREEN}✔ باینری از Release دانلود شد${NC}"
    chmod +x "$BIN"
    return 0
  fi
  echo -e "${YELLOW}Release یافت نشد، در حال ساخت از سورس...${NC}"
  # Build from source
  if ! command -v go >/dev/null 2>&1; then
    echo -e "${YELLOW}نصب Go...${NC}"
    apt-get install -y golang-go 2>&1 | tail -3 || true
  fi
  tmpdir=$(mktemp -d)
  echo -e "دانلود سورس از گیتهاب..."
  if curl -fsSL "https://github.com/${REPO}/archive/refs/heads/main.tar.gz" -o "$tmpdir/src.tar.gz"; then
    tar -xzf "$tmpdir/src.tar.gz" -C "$tmpdir"
    src=$(find "$tmpdir" -maxdepth 1 -type d -name "SpeedTunnel*")
    if [[ -f "$src/go.mod" ]]; then
      (cd "$src" && go build -o "$BIN" ./cmd/speedtunnel && chmod +x "$BIN")
      echo -e "${GREEN}✔ باینری ساخته شد${NC}"
    else
      # Fallback: try local build if running from repo
      if [[ -f "./go.mod" && -f "./cmd/speedtunnel/main.go" ]]; then
        go build -o "$BIN" ./cmd/speedtunnel && chmod +x "$BIN"
        echo -e "${GREEN}✔ باینری از سورس لوکال ساخته شد${NC}"
      else
        echo -e "${RED}خطا: سورس یافت نشد${NC}"
        exit 1
      fi
    fi
  else
    # Local fallback
    if [[ -f "./go.mod" ]]; then
      go build -o "$BIN" ./cmd/speedtunnel && chmod +x "$BIN"
      echo -e "${GREEN}✔ باینری از سورس لوکال ساخته شد${NC}"
    else
      echo -e "${RED}خطا در دانلود سورس${NC}"
      exit 1
    fi
  fi
  rm -rf "$tmpdir"
}

setup_config() {
  echo -e "${CYAN}[3/6] تنظیم کانفیگ...${NC}"
  mkdir -p "$CONFIG_DIR"
  if [[ ! -f "$CONFIG" ]]; then
    echo '{"version":"1.0.0","tunnels":[]}' > "$CONFIG"
    chmod 600 "$CONFIG"
    echo -e "${GREEN}✔ کانفیگ اولیه ساخته شد: $CONFIG${NC}"
  else
    echo -e "${GREEN}✔ کانفیگ موجود حفظ شد${NC}"
  fi
}

setup_service() {
  echo -e "${CYAN}[4/6] نصب سرویس systemd...${NC}"
  cat > /etc/systemd/system/${SERVICE}.service <<EOF
[Unit]
Description=Speed Tunnel - Fast Secure Tunnel
After=network.target nss-lookup.target

[Service]
Type=simple
ExecStart=${BIN} run
Restart=always
RestartSec=3
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF
  systemctl daemon-reload
  systemctl enable $SERVICE >/dev/null 2>&1 || true
  echo -e "${GREEN}✔ سرویس نصب شد${NC}"
}

setup_menu() {
  echo -e "${CYAN}[5/6] نصب منو...${NC}"
  # Find menu.sh
  for p in "./scripts/menu.sh" "$(dirname "$0")/scripts/menu.sh" "/root/project/scripts/menu.sh"; do
    if [[ -f "$p" ]]; then
      cp "$p" /usr/local/bin/speedtunnel-menu
      chmod +x /usr/local/bin/speedtunnel-menu
      # alias command
      ln -sf /usr/local/bin/speedtunnel-menu /usr/local/bin/st 2>/dev/null || true
      echo -e "${GREEN}✔ منو نصب شد: speedtunnel-menu / st${NC}"
      return
    fi
  done
  # fallback: download from github
  curl -fsSL "https://raw.githubusercontent.com/${REPO}/main/scripts/menu.sh" -o /usr/local/bin/speedtunnel-menu 2>/dev/null && chmod +x /usr/local/bin/speedtunnel-menu && ln -sf /usr/local/bin/speedtunnel-menu /usr/local/bin/st 2>/dev/null || echo -e "${YELLOW}منو یافت نشد، بعدا دستی اضافه کنید${NC}"
}

finalize() {
  echo -e "${CYAN}[6/6] راه‌اندازی...${NC}"
  if systemctl is-active --quiet $SERVICE 2>/dev/null; then
    systemctl restart $SERVICE || true
  else
    systemctl start $SERVICE 2>&1 | tail -3 || true
  fi
  sleep 1
  echo ""
  echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
  echo -e "${GREEN}║${NC}  ${WHITE}✔ Speed Tunnel نصب شد!${NC}                        ${GREEN}║${NC}"
  echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
  echo -e "  باینری: ${WHITE}$BIN${NC}"
  echo -e "  کانفیگ: ${WHITE}$CONFIG${NC}"
  echo -e "  سرویس:  ${WHITE}systemctl status $SERVICE${NC}"
  echo ""
  echo -e "  ${CYAN}برای مدیریت تانل‌ها دستور زیر را بزنید:${NC}"
  echo -e "  ${WHITE}sudo speedtunnel-menu${NC}  ${WHITE}(یا st)${NC}"
  echo ""
  echo -e "  ${WHITE}Github:  https://github.com/${REPO}${NC}"
  echo -e "  ${WHITE}Channel: @Speedw_IT  Support: @SpeedwIT${NC}"
  echo ""
  if ! $is_update; then
    echo -ne "${CYAN}آیا می‌خواهید الان وارد منو شوید؟ (y/N): ${NC}"
    read -r ans
    if [[ "$ans" == "y" || "$ans" == "Y" ]]; then
      exec /usr/local/bin/speedtunnel-menu 2>/dev/null || exec bash /usr/local/bin/speedtunnel-menu
    fi
  fi
}

main() {
  need_root
  echo -e "${WHITE}🚀 Speed Tunnel Installer v${VERSION}${NC}"
  echo -e "${CYAN}Repo: https://github.com/${REPO}${NC}\n"
  install_deps
  build_or_download
  setup_config
  setup_service
  setup_menu
  finalize
}

main
