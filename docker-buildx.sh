#!/usr/bin/env bash
# =============================================================================
# docker-buildx.sh
# 多架构 Docker 镜像构建 & 推送到 Docker Hub
#
# 用法:
#   ./docker-buildx.sh                              # 使用 git short hash 作为 tag
#   ./docker-buildx.sh v1.0.0                       # 指定 tag
#   DOCKER_REPO=myuser/myapp ./docker-buildx.sh     # 自定义镜像仓库名
#   PUSH=true ./docker-buildx.sh                    # 构建并推送
#   PUSH=true ./docker-buildx.sh v1.0.0             # 指定 tag 并推送
#   PLATFORMS="linux/amd64" ./docker-buildx.sh      # 仅构建指定平台
#   LOAD=true PLATFORMS="linux/amd64" ./docker-buildx.sh  # 构建并加载到本地（单平台）
# =============================================================================
set -euo pipefail

# ---------------------- 可配置变量 ----------------------
# Docker Hub 仓库地址（格式：用户名/仓库名）
DOCKER_REPO="${DOCKER_REPO:-your-dockerhub-username/wish-fullfilement-fiction}"

# 构建平台（多架构，逗号分隔）
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"

# Buildx builder 实例名称
BUILDER_NAME="${BUILDER_NAME:-multiarch-builder}"

# 是否推送到 Docker Hub（设为 true 则推送）
PUSH="${PUSH:-false}"

# 是否加载到本地 docker images（仅支持单平台，设为 true 则加载）
LOAD="${LOAD:-false}"

# ---------------------- 计算镜像 Tag ----------------------
if [ -n "${1:-}" ]; then
    TAG="$1"
else
    TAG=$(git rev-parse --short HEAD 2>/dev/null || echo "latest")
    # 如果工作区有未提交的修改，加上 .dirty 后缀
    if [ -n "$(git status --porcelain 2>/dev/null)" ]; then
        TAG="${TAG}.dirty"
    fi
fi

IMAGE="${DOCKER_REPO}:${TAG}"
IMAGE_LATEST="${DOCKER_REPO}:latest"

echo "============================================"
echo " Docker 多架构构建"
echo "============================================"
echo " 镜像:     ${IMAGE}"
echo " 平台:     ${PLATFORMS}"
echo " 推送:     ${PUSH}"
echo " 本地加载: ${LOAD}"
echo "============================================"

# ---------------------- 检查 Docker Buildx ----------------------
if ! docker buildx version &>/dev/null; then
    echo "❌ 错误: docker buildx 未安装，请升级 Docker Desktop 或安装 buildx 插件"
    exit 1
fi

# ---------------------- LOAD 与 PUSH 互斥校验 ----------------------
if [ "${LOAD}" = "true" ] && [ "${PUSH}" = "true" ]; then
    echo "❌ 错误: LOAD=true 与 PUSH=true 不能同时使用"
    exit 1
fi

# LOAD 模式仅支持单平台
if [ "${LOAD}" = "true" ] && [[ "${PLATFORMS}" == *","* ]]; then
    echo "❌ 错误: LOAD=true 时不支持多平台（PLATFORMS 只能指定一个）"
    exit 1
fi

# ---------------------- 创建/使用 Buildx Builder ----------------------
if ! docker buildx inspect "${BUILDER_NAME}" &>/dev/null; then
    echo "📦 创建 buildx builder: ${BUILDER_NAME}"
    docker buildx create \
        --name "${BUILDER_NAME}" \
        --driver docker-container \
        --bootstrap \
        --use
else
    echo "📦 使用已有 builder: ${BUILDER_NAME}"
    docker buildx use "${BUILDER_NAME}"
fi

# ---------------------- 登录 Docker Hub ----------------------
if [ "${PUSH}" = "true" ]; then
    # 尝试检测是否已登录（docker info 在新版可能不含 Username）
    if ! docker info 2>/dev/null | grep -q "Username"; then
        echo "🔑 请先登录 Docker Hub:"
        docker login
    fi
fi

# ---------------------- 构建（并可选推送/加载） ----------------------
BUILD_ARGS=(
    --platform "${PLATFORMS}"
    --tag      "${IMAGE}"
    --tag      "${IMAGE_LATEST}"
    --file     Dockerfile
)

if [ "${PUSH}" = "true" ]; then
    BUILD_ARGS+=(--push)
    echo "🚀 构建并推送多架构镜像..."
elif [ "${LOAD}" = "true" ]; then
    BUILD_ARGS+=(--load)
    echo "🔨 构建并加载到本地 docker images（单平台）..."
else
    echo "🔨 纯构建（不推送、不加载），可用于验证 Dockerfile..."
    echo "   ⚠️  多平台本地构建需要 PUSH=true 推送，或使用 LOAD=true（单平台）加载到本地"
fi

docker buildx build "${BUILD_ARGS[@]}" .

echo ""
echo "============================================"
if [ "${PUSH}" = "true" ]; then
    echo "✅ 镜像已推送到 Docker Hub:"
    echo "   ${IMAGE}"
    echo "   ${IMAGE_LATEST}"
    echo ""
    echo "   拉取命令: docker pull ${IMAGE}"
elif [ "${LOAD}" = "true" ]; then
    echo "✅ 镜像已加载到本地:"
    echo "   ${IMAGE}"
    echo "   运行命令: docker run --rm ${IMAGE}"
else
    echo "✅ 构建完成（未推送）"
    echo "   镜像 Tag: ${IMAGE}"
    echo ""
    echo "   推送到 Docker Hub:"
    echo "   PUSH=true ./docker-buildx.sh ${TAG}"
    echo ""
    echo "   加载到本地（单平台）:"
    echo "   LOAD=true PLATFORMS=linux/amd64 ./docker-buildx.sh ${TAG}"
fi
echo "============================================"

