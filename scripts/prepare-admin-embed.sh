#!/bin/sh
# 把 Vite 构建产物整理成 Go embed 能收录的目录。
#
# 为什么需要这一步：
#   1. go:embed 会跳过以 "_" 或 "." 开头的文件，而 Vite 的 chunk 里存在
#      _plugin-vue_export-helper-xxxx.js 这类文件名，直接 embed 之后
#      前端会 404（文件根本没有被打进二进制）。
#   2. 因此这里给 "_" 开头的文件加 "x" 前缀，并同步替换 JS/CSS/HTML 里的引用。
#
# 用法：scripts/prepare-admin-embed.sh [dist 目录] [输出目录]
# 默认：web/dist -> internal/admin/static
#
# Makefile 的 web-build 和 deployments/Dockerfile 都调用本脚本，
# 保证「本地构建」与「镜像构建」产出一致。
set -eu

SRC=${1:-web/dist}
DST=${2:-internal/admin/static}

if [ ! -d "$SRC" ]; then
    echo "error: dist directory not found: $SRC" >&2
    exit 1
fi

echo "prepare-admin-embed: $SRC -> $DST"
rm -rf "$DST"
mkdir -p "$DST"
cp -R "$SRC"/. "$DST"/

# 1) 重命名 "_" 开头的文件（加 x 前缀）
find "$DST" -name '_*' -type f | while read -r f; do
    dir=$(dirname "$f")
    base=$(basename "$f")
    mv "$f" "$dir/x$base"
done

# 2) 替换引用（JS / CSS / HTML 里都可能有）
find "$DST" \( -name '*.js' -o -name '*.css' -o -name '*.html' \) -type f | while read -r f; do
    sed -i 's|_plugin-vue_export-helper|x_plugin-vue_export-helper|g' "$f"
done

echo "prepare-admin-embed: done"
