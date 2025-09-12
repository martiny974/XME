#!/usr/bin/env bash

set -euo pipefail

# ------------------------------
# xiaohongshu-mcp 启动脚本
# ------------------------------

MIN_GO_VERSION="1.23.5"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色输出
if [ -t 1 ]; then
  GREEN='\033[0;32m'
  YELLOW='\033[0;33m'
  RED='\033[0;31m'
  BLUE='\033[0;34m'
  BOLD='\033[1m'
  NC='\033[0m'
else
  GREEN=""; YELLOW=""; RED=""; BLUE=""; BOLD=""; NC=""
fi

log() { printf "%b\n" "$1"; }
info() { log "${BLUE}$1${NC}"; }
ok() { log "${GREEN}$1${NC}"; }
warn() { log "${YELLOW}$1${NC}"; }
err() { log "${RED}$1${NC}"; }

usage() {
  cat <<EOF
${BOLD}用法:${NC} $0 [选项]

选项:
  --with-ui            非无头模式启动，有浏览器界面 (等价于 -headless=false)
  --skip-login         即使缺少 Cookies 也跳过登录流程
  --only-login         仅运行登录流程，完成后退出
  -h, --help           显示帮助

示例:
  $0                   # 安装依赖，若需要则先登录，然后以无头模式启动服务
  $0 --with-ui         # 同上，但以有界面模式启动
  $0 --only-login      # 仅执行一次登录以保存 Cookies 后退出
EOF
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    err "Required command '$1' is not installed or not in PATH."
    exit 1
  fi
}

version_ge() {
  # returns 0 (true) if $1 >= $2 (semantic version comparison)
  # Normalize to three segments
  local IFS=.
  local -a v1=($1) v2=($2)
  for i in 0 1 2; do
    v1[i]="${v1[i]:-0}"
    v2[i]="${v2[i]:-0}"
    if ((10#${v1[i]} > 10#${v2[i]})); then return 0; fi
    if ((10#${v1[i]} < 10#${v2[i]})); then return 1; fi
  done
  return 0
}

detect_temp_dir() {
  # Prefer POSIX TMPDIR, then Windows TEMP/TMP (works in Git Bash), fallback to /tmp
  printf "%s" "${TMPDIR:-${TEMP:-${TMP:-/tmp}}}"
}

# Parse args
WITH_UI=false
SKIP_LOGIN=false
ONLY_LOGIN=false
while (( "$#" )); do
  case "$1" in
    --with-ui)
      WITH_UI=true; shift ;;
    --skip-login)
      SKIP_LOGIN=true; shift ;;
    --only-login)
      ONLY_LOGIN=true; shift ;;
    -h|--help)
      usage; exit 0 ;;
    *)
      err "Unknown option: $1"; usage; exit 1 ;;
  esac
done

info "检查运行环境..."
if ! command -v go >/dev/null 2>&1; then
  err "未检测到 Go，请先安装 Go ($MIN_GO_VERSION 或更高) 后重试。"
  OS_NAME="$(uname -s 2>/dev/null || echo unknown)"
  case "$OS_NAME" in
    MINGW*|MSYS*|CYGWIN*)
      cat <<'WIN'
Windows 安装指引（以管理员身份在 PowerShell 中执行其一）:
  1) winget（推荐）:
     winget install --id GoLang.Go -e
  2) Chocolatey:
     choco install golang -y
  3) 官网安装包:
     https://go.dev/dl/

安装完成后，请关闭并重新打开终端，让 PATH 生效。
WIN
      ;;
    Darwin)
      cat <<'MAC'
macOS 安装指引:
  brew install go
  或下载官方安装包: https://go.dev/dl/
MAC
      ;;
    Linux)
      cat <<'LIN'
Linux 安装指引（任选其一）:
  Debian/Ubuntu: sudo apt update && sudo apt install -y golang
  Fedora/CentOS: sudo dnf install -y golang
  Arch:          sudo pacman -S --noconfirm go
  或使用官方二进制包: https://go.dev/dl/
LIN
      ;;
    *)
      info "请参考官网安装 Go: https://go.dev/dl/"
      ;;
  esac
  exit 1
fi

GO_VER_RAW="$(go version 2>/dev/null | awk '{print $3}')" || true
if [[ -z "$GO_VER_RAW" ]]; then
  err "无法检测 Go 版本，请确认 Go 已正确安装。"
  exit 1
fi
GO_VER="${GO_VER_RAW#go}"

if version_ge "$GO_VER" "$MIN_GO_VERSION"; then
  ok "Go 版本: $GO_VER (>= $MIN_GO_VERSION)"
else
  err "检测到 Go 版本较低：$GO_VER，需要 >= $MIN_GO_VERSION。请升级后重试：https://go.dev/dl/"
  exit 1
fi

info "下载依赖 (go mod download)..."
GO111MODULE=on go mod download
ok "依赖就绪。"

TEMP_DIR="$(detect_temp_dir)"
COOKIES_FILE="$TEMP_DIR/cookies.json"
info "Cookies 文件路径: $COOKIES_FILE"

needs_login() {
  if [[ ! -s "$COOKIES_FILE" ]]; then
    return 0
  fi
  # Optionally, validate JSON structure (best-effort)
  return 1
}

if [[ "$ONLY_LOGIN" == true ]]; then
  info "开始运行登录流程..."
  go run cmd/login/main.go || {
    err "登录流程失败。"
    exit 1
  }
  ok "登录完成。Cookies 应已保存到: $COOKIES_FILE"
  exit 0
fi

if [[ "$SKIP_LOGIN" == false ]] && needs_login; then
  warn "未找到或为空的 Cookies，先进行登录..."
  go run cmd/login/main.go || {
    err "登录流程失败。"
    exit 1
  }
  ok "登录完成。"
else
  ok "检测到 Cookies，跳过登录。"
fi

HEADLESS_ARG=""
if [[ "$WITH_UI" == true ]]; then
  HEADLESS_ARG="-headless=false"
fi

info "启动 xiaohongshu-mcp 服务..."
set +e
go run . ${HEADLESS_ARG} 
EXIT_CODE=$?
set -e

if [[ $EXIT_CODE -ne 0 ]]; then
  err "服务退出，错误码：$EXIT_CODE"
  exit $EXIT_CODE
fi

ok "服务已停止。"


