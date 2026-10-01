#!/usr/bin/env bash
set -e
# Speed Tunnel - Professional TUI Menu
# Github: https://github.com/SpeedwiT/SpeedTunnel
# Channel: @Speedw_IT | Support: @SpeedwIT

CONFIG="/etc/speedtunnel/config.json"
BIN="/usr/local/bin/speedtunnel"
SERVICE="speedtunnel"
VERSION="1.0.0"
REPO="SpeedwiT/SpeedTunnel"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
WHITE='\033[1;37m'
DIM='\033[2m'
NC='\033[0m'
BOLD='\033[1m'

# Helpers
need_root() {
  if [[ $EUID -ne 0 ]]; then
    echo -e "${RED}Please run with sudo${NC}"
    exit 1
  fi
}

has_jq() { command -v jq >/dev/null 2>&1; }

ensure_config() {
  mkdir -p /etc/speedtunnel
  if [[ ! -f "$CONFIG" ]]; then
    echo '{"version":"1.0.0","tunnels":[]}' > "$CONFIG"
    chmod 600 "$CONFIG"
  fi
}

banner() {
  clear
  echo -e "${CYAN}╔══════════════════════════════════════════════════════════╗${NC}"
  echo -e "${CYAN}║${NC}                                                          ${CYAN}║${NC}"
  echo -e "${CYAN}║${NC}         ███████╗██████╗ ███████╗███████╗██████╗          ${CYAN}║${NC}"
  echo -e "${CYAN}║${NC}         ██╔════╝██╔══██╗██╔════╝██╔════╝██╔══██╗         ${CYAN}║${NC}"
  echo -e "${CYAN}║${NC}         ███████╗██████╔╝█████╗  █████╗  ██║  ██║         ${CYAN}║${NC}"
  echo -e "${CYAN}║${NC}         ╚════██║██╔═══╝ ██╔══╝  ██╔══╝  ██║  ██║         ${CYAN}║${NC}"
  echo -e "${CYAN}║${NC}         ███████║██║     ███████╗███████╗██████╔╝         ${CYAN}║${NC}"
  echo -e "${CYAN}║${NC}         ╚══════╝╚═╝     ╚══════╝╚══════╝╚═════╝          ${CYAN}║${NC}"
  echo -e "${CYAN}║${NC}                                                          ${CYAN}║${NC}"
  echo -e "${CYAN}║${NC}    ████████╗██╗   ██╗███╗   ██╗███╗   ██╗███████╗██╗     ${CYAN}║${NC}"
  echo -e "${CYAN}║${NC}    ╚══██╔══╝██║   ██║████╗  ██║████╗  ██║██╔════╝██║     ${CYAN}║${NC}"
  echo -e "${CYAN}║${NC}       ██║   ██║   ██║██╔██╗ ██║██╔██╗ ██║█████╗  ██║     ${CYAN}║${NC}"
  echo -e "${CYAN}║${NC}       ██║   ██║   ██║██║╚██╗██║██║╚██╗██║██╔══╝  ██║     ${CYAN}║${NC}"
  echo -e "${CYAN}║${NC}     ██║   ╚██████╔╝██║ ╚████║██║ ╚████║███████╗███████╗  ${CYAN}║${NC}"
  echo -e "${CYAN}║${NC}     ╚═╝    ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝╚══════╝  ${CYAN}║${NC}"
  echo -e "${CYAN}║${NC}                                                          ${CYAN}║${NC}"
  echo -e "${CYAN}╚══════════════════════════════════════════════════════════╝${NC}"
  echo ""
  echo -e "  ${GREEN}⚡ Speed Tunnel v${VERSION}${NC}  ${DIM}— Fast • Secure • Anti-DPI${NC}"
  echo -e "  ${DIM}Github:${NC} ${WHITE}https://github.com/${REPO}${NC}"
  echo -e "  ${DIM}Channel:${NC} ${WHITE}@Speedw_IT${NC}  ${DIM}Support:${NC} ${WHITE}@SpeedwIT${NC}"
  echo ""
}

pause() { echo -ne "${DIM}Press Enter to continue...${NC}"; read -r _; }

service_status() {
  if systemctl is-active --quiet $SERVICE 2>/dev/null; then
    echo -e "${GREEN}● Active${NC}"
  else
    echo -e "${RED}● Inactive${NC}"
  fi
}

show_status() {
  banner
  echo -e "${BOLD}${WHITE}── Service Status ──${NC}"
  echo -e " Service: $(service_status)"
  if has_jq; then
    cnt=$(jq '.tunnels | length' "$CONFIG" 2>/dev/null || echo 0)
    enabled=$(jq '[.tunnels[] | select(.enabled==true)] | length' "$CONFIG" 2>/dev/null || echo 0)
    echo -e " Tunnels: ${CYAN}$cnt${NC} (Active: ${GREEN}$enabled${NC})"
    echo ""
    if [[ "$cnt" -gt 0 ]]; then
      echo -e "${BOLD}── Tunnel List ──${NC}"
      jq -r '.tunnels[] | " \(.id) | \(.name) | \(.role) | \(.transport) | SNI:\(.sni) | 0.0.0.0:\(.listen_port) → \(.remote_addr):\(.remote_port) | ctrl:\(.control_port) | enabled:\(.enabled)"' "$CONFIG" 2>/dev/null | while IFS= read -r line; do
        enabled=$(echo "$line" | grep -o 'enabled:true' || true)
        if [[ -n "$enabled" ]]; then
          echo -e "  ${GREEN}▸${NC} $line"
        else
          echo -e "  ${DIM}▸ $line${NC}"
        fi
      done
    fi
    echo ""
    echo -e "${DIM}── Port Health Check (TCP Ping) ──${NC}"
    jq -r '.tunnels[] | "\(.id) \(.name) \(.listen_port)"' "$CONFIG" 2>/dev/null | while read -r id name port; do
      if timeout 1 bash -c "cat < /dev/null > /dev/tcp/127.0.0.1/$port" 2>/dev/null; then
        echo -e "  ${GREEN}✔${NC} $name (127.0.0.1:$port) ${GREEN}open${NC}"
      else
        echo -e "  ${RED}✘${NC} $name (127.0.0.1:$port) ${RED}closed${NC}"
      fi
    done
  else
    cat "$CONFIG"
  fi
  echo ""
  systemctl status $SERVICE --no-pager -l 2>&1 | head -n 30 || true
  pause
}

create_tunnel() {
  banner
  echo -e "${BOLD}${WHITE}── Create New Tunnel ──${NC}\n"
  echo -ne "${CYAN}Tunnel name: ${NC}"
  read -r tname
  [[ -z "$tname" ]] && tname="tunnel-$(date +%s | tail -c 5)"

  echo -e "${YELLOW}Select server role:${NC}"
  echo "  1) iran   (Client - Iran server, connects outbound)"
  echo "  2) kharej (Server - Foreign server, listens for Iran)"
  echo -ne "${CYAN}Choice [1-2]: ${NC}"
  read -r role_sel
  if [[ "$role_sel" == "2" ]]; then role="kharej"; else role="iran"; fi

  echo -e "\n${YELLOW}Select transport:${NC}"
  echo "  1) SpeedTLS      (Recommended - TLS Spoof + Fragment - Fast & Anti-DPI)"
  echo "  2) SpeedHTTP     (HTTP/2 Fake - Stable, looks like browser traffic)"
  echo "  3) SpeedReverse  (For when Iran has no international internet - Iran connects out)"
  echo "  4) SpeedICMP     (Emergency - ICMP/DNS fallback)"
  echo -ne "${CYAN}Choice [1-4] default 1: ${NC}"
  read -r tr_sel
  case "$tr_sel" in
    2) transport="speedhttp" ;;
    3) transport="speedreverse" ;;
    4) transport="speedicmp" ;;
    *) transport="speedtls" ;;
  esac

  echo -ne "${CYAN}Listen port on foreign server [e.g. 443]: ${NC}"
  read -r listen_port
  echo -ne "${CYAN}Local service port on Iran [e.g. 443]: ${NC}"
  read -r remote_port
  echo -ne "${CYAN}Foreign server address (only for iran role) [e.g. 1.2.3.4]: ${NC}"
  read -r remote_addr
  echo -ne "${CYAN}Tunnel control port [default 7000]: ${NC}"
  read -r control_port
  echo -ne "${CYAN}Security secret (empty = auto-generate): ${NC}"
  read -r secret
  echo -ne "${CYAN}SNI spoof domain [default www.digikala.com]: ${NC}"
  read -r sni

  [[ -z "$listen_port" ]] && listen_port=443
  [[ -z "$remote_port" ]] && remote_port=443
  [[ -z "$control_port" ]] && control_port=7000
  [[ -z "$sni" ]] && sni="www.digikala.com"
  if [[ -z "$secret" ]]; then
    secret=$(openssl rand -hex 16 2>/dev/null || cat /proc/sys/kernel/random/uuid | tr -d '-')
  fi
  if [[ "$role" == "kharej" ]]; then remote_addr="0.0.0.0"; fi
  [[ -z "$remote_addr" ]] && remote_addr="YOUR_KHAREJ_IP"

  tid=$(cat /proc/sys/kernel/random/uuid 2>/dev/null | cut -c1-8 || echo "t$(date +%s | tail -c 6)")
  tid="st-${tid}"

  ensure_config
  if has_jq; then
    tmp=$(mktemp)
    jq --arg id "$tid" --arg name "$tname" --arg role "$role" --arg transport "$transport" --argjson listen_port "$listen_port" --argjson remote_port "$remote_port" --arg remote_addr "$remote_addr" --argjson control_port "$control_port" --arg secret "$secret" --arg sni "$sni" \
      '.tunnels += [{"id":$id,"name":$name,"role":$role,"transport":$transport,"listen_port":$listen_port,"remote_port":$remote_port,"remote_addr":$remote_addr,"control_port":$control_port,"secret":$secret,"sni":$sni,"enabled":true,"created_at": (now|todate)}]' "$CONFIG" > "$tmp" && mv "$tmp" "$CONFIG"
    chmod 600 "$CONFIG"
  else
    echo "jq not found, please install jq"
    return
  fi

  echo ""
  echo -e "${GREEN}✔ Tunnel created!${NC}"
  echo -e "  ID: ${WHITE}$tid${NC}"
  echo -e "  Secret: ${WHITE}$secret${NC}"
  echo -e "  ${DIM}Use the same Secret on both servers${NC}"
  echo -e "  ${YELLOW}Apply: option 8 (Restart Service)${NC}"
  pause
}

list_tunnels() {
  banner
  echo -e "${BOLD}${WHITE}── Manage Tunnels ──${NC}\n"
  if ! has_jq; then echo "jq required"; pause; return; fi
  cnt=$(jq '.tunnels | length' "$CONFIG")
  if [[ "$cnt" -eq 0 ]]; then
    echo -e "${YELLOW}No tunnels found.${NC}"
    pause; return
  fi
  jq -r '.tunnels | to_entries[] | "\(.key+1)) \(.value.id) | \(.value.name) | \(.value.role) | \(.value.transport) | :\(.value.listen_port) → :\(.value.remote_port) | enabled:\(.value.enabled)"' "$CONFIG" | while IFS= read -r line; do
    echo -e "  ${CYAN}$line${NC}"
  done
  echo ""
  echo "  d) Delete tunnel   e) Edit SNI/Port   t) Toggle enabled   b) Back"
  echo -ne "${CYAN}Choice: ${NC}"
  read -r sel
  case "$sel" in
    d|D)
      echo -ne "Tunnel ID to delete: "
      read -r del_id
      tmp=$(mktemp)
      jq --arg id "$del_id" '.tunnels |= map(select(.id != $id))' "$CONFIG" > "$tmp" && mv "$tmp" "$CONFIG"
      echo -e "${GREEN}Deleted${NC}"; sleep 1
      ;;
    e|E)
      echo -ne "Tunnel ID: "
      read -r eid
      echo -ne "New SNI (empty = no change): "
      read -r nsni
      echo -ne "New ListenPort (empty = no change): "
      read -r nlp
      tmp=$(mktemp)
      if [[ -n "$nsni" && -n "$nlp" ]]; then
        jq --arg id "$eid" --arg sni "$nsni" --argjson lp "$nlp" '(.tunnels[] | select(.id==$id) | .sni) |= $sni | (.tunnels[] | select(.id==$id) | .listen_port) |= $lp' "$CONFIG" > "$tmp" && mv "$tmp" "$CONFIG"
      elif [[ -n "$nsni" ]]; then
        jq --arg id "$eid" --arg sni "$nsni" '(.tunnels[] | select(.id==$id) | .sni) |= $sni' "$CONFIG" > "$tmp" && mv "$tmp" "$CONFIG"
      elif [[ -n "$nlp" ]]; then
        jq --arg id "$eid" --argjson lp "$nlp" '(.tunnels[] | select(.id==$id) | .listen_port) |= $lp' "$CONFIG" > "$tmp" && mv "$tmp" "$CONFIG"
      fi
      echo -e "${GREEN}Updated${NC}"; sleep 1
      ;;
    t|T)
      echo -ne "Tunnel ID: "
      read -r tid2
      tmp=$(mktemp)
      jq --arg id "$tid2" '(.tunnels[] | select(.id==$id) | .enabled) |= (if . then false else true end)' "$CONFIG" > "$tmp" && mv "$tmp" "$CONFIG"
      echo -e "${GREEN}Toggled${NC}"; sleep 1
      ;;
    *) ;;
  esac
}

bot_setup() {
  banner
  echo -e "${BOLD}${WHITE}── Telegram Bot Settings ──${NC}\n"
  cur_token=$(jq -r '.bot_token // ""' "$CONFIG")
  cur_admin=$(jq -r '.bot_admin_id // 0' "$CONFIG")
  echo -e " Current token: ${DIM}${cur_token:0:12}...${NC}"
  echo -e " Current Admin ID: ${WHITE}$cur_admin${NC}\n"
  echo -ne "${CYAN}Bot token (empty = no change): ${NC}"
  read -r ntok
  echo -ne "${CYAN}Admin numeric ID (empty = no change): ${NC}"
  read -r nadm
  tmp=$(mktemp)
  if [[ -n "$ntok" ]]; then
    jq --arg t "$ntok" '.bot_token = $t' "$CONFIG" > "$tmp" && mv "$tmp" "$CONFIG"
  fi
  if [[ -n "$nadm" ]]; then
    jq --argjson a "$nadm" '.bot_admin_id = $a' "$CONFIG" > "$tmp" && mv "$tmp" "$CONFIG"
  fi
  echo -e "${GREEN}Saved. Restart service to apply.${NC}"
  pause
}

update_from_github() {
  banner
  echo -e "${YELLOW}Checking for updates from GitHub...${NC}"
  echo -e "${DIM}Repo: https://github.com/${REPO}${NC}\n"
  tmpdir=$(mktemp -d)
  if curl -fsSL "https://github.com/${REPO}/archive/refs/heads/main.tar.gz" -o "$tmpdir/main.tar.gz"; then
    echo -e "${GREEN}Downloaded. Installing...${NC}"
    tar -xzf "$tmpdir/main.tar.gz" -C "$tmpdir"
    extracted=$(find "$tmpdir" -maxdepth 1 -type d -name "SpeedTunnel*" | head -1)
    if [[ -f "$extracted/install.sh" ]]; then
      bash "$extracted/install.sh" --update
    else
      echo -e "${YELLOW}install.sh not found, manual update:${NC}"
      echo "bash <(curl -s https://raw.githubusercontent.com/${REPO}/main/install.sh)"
    fi
  else
    echo -e "${RED}Download failed. Check your internet.${NC}"
    echo -e "${DIM}Manual: bash <(curl -s https://raw.githubusercontent.com/${REPO}/main/install.sh)${NC}"
  fi
  rm -rf "$tmpdir"
  pause
}

main_menu() {
  need_root
  ensure_config
  while true; do
    banner
    echo -e "${BOLD}${WHITE}  Main Menu${NC}  $(service_status)  ${DIM}Config: $CONFIG${NC}\n"
    echo -e "  ${GREEN}1)${NC} ${WHITE}Create New Tunnel${NC}"
    echo -e "  ${GREEN}2)${NC} ${WHITE}Manage Tunnels${NC}"
    echo -e "  ${GREEN}3)${NC} ${WHITE}View Status${NC}"
    echo -e "  ${GREEN}4)${NC} ${WHITE}Telegram Bot Settings${NC}"
    echo -e "  ${GREEN}5)${NC} ${WHITE}View Service Logs${NC}"
    echo -e "  ${GREEN}6)${NC} ${WHITE}Update from Github${NC}"
    echo -e "  ${GREEN}7)${NC} ${WHITE}Show Raw Config${NC}"
    echo -e "  ${GREEN}8)${NC} ${YELLOW}Restart Service${NC}"
    echo -e "  ${GREEN}9)${NC} ${RED}Uninstall Speed Tunnel${NC}"
    echo -e "  ${GREEN}0)${NC} Exit"
    echo ""
    echo -ne "${CYAN}Your choice [0-9]: ${NC}"
    read -r choice
    case "$choice" in
      1) create_tunnel ;;
      2) list_tunnels ;;
      3) show_status ;;
      4) bot_setup ;;
      5) banner; journalctl -u $SERVICE --no-pager -n 100 2>&1 | head -n 100; echo ""; systemctl status $SERVICE --no-pager 2>&1 | head -n 20; pause ;;
      6) update_from_github ;;
      7) banner; cat "$CONFIG" | head -n 100; echo ""; pause ;;
      8) systemctl restart $SERVICE && echo -e "${GREEN}✔ Service restarted${NC}" || echo -e "${RED}Restart failed${NC}"; sleep 1 ;;
      9) echo -ne "${RED}Are you sure? (y/N): ${NC}"; read -r c; if [[ "$c" == "y" || "$c" == "Y" ]]; then systemctl stop $SERVICE 2>/dev/null; systemctl disable $SERVICE 2>/dev/null; rm -f /etc/systemd/system/$SERVICE.service; systemctl daemon-reload; echo -e "${GREEN}Uninstalled (binary & config kept)${NC}"; fi; pause ;;
      0) echo -e "${GREEN}Goodbye! @Speedw_IT${NC}"; exit 0 ;;
      *) echo -e "${RED}Invalid option${NC}"; sleep 1 ;;
    esac
  done
}

main_menu
