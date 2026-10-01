#!/usr/bin/env bash
set -e
# Speed Tunnel - Installer
# Usage: bash <(curl -fsSL https://raw.githubusercontent.com/SpeedwiT/SpeedTunnel/main/install.sh)
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
    echo -e "${RED}Please run with sudo: sudo bash install.sh${NC}"
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
  echo -e "${CYAN}[1/6] Installing dependencies...${NC}"
  apt-get update -qq 2>&1 | tail -1 || true
  apt-get install -y curl wget jq openssl tar golang-go 2>&1 | tail -5 || true
}

build_or_download() {
  echo -e "${CYAN}[2/6] Fetching/building binary...${NC}"
  arch=$(detect_arch)
  url="https://github.com/${REPO}/releases/latest/download/speedtunnel-linux-${arch}"
  if curl -fsSL "$url" -o "$BIN" 2>/dev/null && [[ -s "$BIN" ]]; then
    echo -e "${GREEN}✔ Binary downloaded from Release${NC}"
    chmod +x "$BIN"
    return 0
  fi
  echo -e "${YELLOW}Release not found, building from source...${NC}"
  if ! command -v go >/dev/null 2>&1; then
    echo -e "${YELLOW}Installing Go...${NC}"
    apt-get install -y golang-go 2>&1 | tail -3 || true
  fi
  tmpdir=$(mktemp -d)
  echo -e "Downloading source from GitHub..."
  if curl -fsSL "https://github.com/${REPO}/archive/refs/heads/main.tar.gz" -o "$tmpdir/src.tar.gz"; then
    tar -xzf "$tmpdir/src.tar.gz" -C "$tmpdir"
    src=$(find "$tmpdir" -maxdepth 1 -type d -name "SpeedTunnel*" | head -1)
    if [[ -n "$src" && -f "$src/go.mod" ]]; then
      echo -e "Building from $src ..."
      (cd "$src" && go build -o "$BIN" ./cmd/speedtunnel && chmod +x "$BIN")
      echo -e "${GREEN}✔ Binary built successfully${NC}"
    else
      echo -e "${RED}Error: source structure invalid${NC}"
      echo -e "${DIM}Contents of $tmpdir:${NC}"
      ls -la "$tmpdir" 2>/dev/null || true
      rm -rf "$tmpdir"
      exit 1
    fi
  else
    echo -e "${RED}Error downloading source${NC}"
    rm -rf "$tmpdir"
    exit 1
  fi
  rm -rf "$tmpdir"
}

setup_config() {
  echo -e "${CYAN}[3/6] Setting up config...${NC}"
  mkdir -p "$CONFIG_DIR"
  if [[ ! -f "$CONFIG" ]]; then
    echo '{"version":"1.0.0","tunnels":[]}' > "$CONFIG"
    chmod 600 "$CONFIG"
    echo -e "${GREEN}✔ Config created: $CONFIG${NC}"
  else
    echo -e "${GREEN}✔ Existing config preserved${NC}"
  fi
}

setup_service() {
  echo -e "${CYAN}[4/6] Installing systemd service...${NC}"
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
  echo -e "${GREEN}✔ Service installed${NC}"
}

setup_menu() {
  echo -e "${CYAN}[5/6] Installing menu...${NC}"
  for p in "./scripts/menu.sh" "$(dirname "$0")/scripts/menu.sh" "/root/project/scripts/menu.sh"; do
    if [[ -f "$p" ]]; then
      cp "$p" /usr/local/bin/speedtunnel-menu
      chmod +x /usr/local/bin/speedtunnel-menu
      ln -sf /usr/local/bin/speedtunnel-menu /usr/local/bin/st 2>/dev/null || true
      echo -e "${GREEN}✔ Menu installed: speedtunnel-menu / st${NC}"
      return
    fi
  done
  curl -fsSL "https://raw.githubusercontent.com/${REPO}/main/scripts/menu.sh" -o /usr/local/bin/speedtunnel-menu 2>/dev/null && chmod +x /usr/local/bin/speedtunnel-menu && ln -sf /usr/local/bin/speedtunnel-menu /usr/local/bin/st 2>/dev/null || echo -e "${YELLOW}Menu not found, add manually later${NC}"
}

finalize() {
  echo -e "${CYAN}[6/6] Starting...${NC}"
  if systemctl is-active --quiet $SERVICE 2>/dev/null; then
    systemctl restart $SERVICE || true
  else
    systemctl start $SERVICE 2>&1 | tail -3 || true
  fi
  sleep 1
  echo ""
  echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
  echo -e "${GREEN}║${NC}  ${WHITE}✔ Speed Tunnel installed!${NC}                        ${GREEN}║${NC}"
  echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
  echo -e "  Binary:  ${WHITE}$BIN${NC}"
  echo -e "  Config:  ${WHITE}$CONFIG${NC}"
  echo -e "  Service: ${WHITE}systemctl status $SERVICE${NC}"
  echo ""
  echo -e "  ${CYAN}Manage tunnels with:${NC}"
  echo -e "  ${WHITE}sudo speedtunnel-menu${NC}  ${WHITE}(or st)${NC}"
  echo ""
  echo -e "  ${WHITE}Github:  https://github.com/${REPO}${NC}"
  echo -e "  ${WHITE}Channel: @Speedw_IT  Support: @SpeedwIT${NC}"
  echo ""
  if ! $is_update; then
    echo -ne "${CYAN}Enter menu now? (y/N): ${NC}"
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
