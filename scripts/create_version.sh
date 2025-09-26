#!/bin/bash

# 创建新版本目录和文件的脚本
# 使用方法: ./create_version.sh v1.10.0

if [ $# -eq 0 ]; then
    echo "使用方法: $0 <版本号>"
    echo "示例: $0 v1.10.0"
    exit 1
fi

VERSION=$1
VERSION_DIR="cmd/initial/incr/$VERSION"
TEMPLATE_DIR="cmd/initial/incr/template"

# 验证版本号格式
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "错误: 版本号格式不正确，应为 v{major}.{minor}.{patch}"
    echo "示例: v1.10.0"
    exit 1
fi

# 检查版本目录是否已存在
if [ -d "$VERSION_DIR" ]; then
    echo "错误: 版本目录 $VERSION_DIR 已存在"
    exit 1
fi

echo "创建版本 $VERSION 的目录和文件..."

# 创建版本目录
mkdir -p "$VERSION_DIR"

# 从模板创建文件
echo "创建实现文件..."
# 提取版本号用于包名（去掉 v 前缀）
VERSION_PACKAGE=$(echo $VERSION | sed 's/v//' | sed 's/\.//g')

# 计算前一个版本和后一个版本
MAJOR=$(echo $VERSION | sed 's/v\([0-9]*\)\.\([0-9]*\)\.\([0-9]*\)/\1/')
MINOR=$(echo $VERSION | sed 's/v\([0-9]*\)\.\([0-9]*\)\.\([0-9]*\)/\2/')
PATCH=$(echo $VERSION | sed 's/v\([0-9]*\)\.\([0-9]*\)\.\([0-9]*\)/\3/')

if [ $PATCH -eq 0 ]; then
    if [ $MINOR -eq 0 ]; then
        PREV_VERSION="v$((MAJOR-1)).9.9"
    else
        PREV_VERSION="v$MAJOR.$((MINOR-1)).9"
    fi
else
    PREV_VERSION="v$MAJOR.$MINOR.$((PATCH-1))"
fi

NEXT_VERSION="v$MAJOR.$MINOR.$((PATCH+1))"

sed "s/{version_package}/$VERSION_PACKAGE/g; s/{version}/$VERSION/g; s/backup\.BackupOptions/backup.Options/g; s/return \"vv/return \"v/g" "$TEMPLATE_DIR/incr-v{version}.go.template" > "$VERSION_DIR/incr-$VERSION.go"

echo "创建测试文件..."
sed "s/{version_package}/$VERSION_PACKAGE/g; s/{version}/$VERSION/g; s/{prev_version}/$PREV_VERSION/g; s/{next_version}/$NEXT_VERSION/g; s/vv/v/g" "$TEMPLATE_DIR/incr_v{version}_test.go.template" > "$VERSION_DIR/incr_v${VERSION_PACKAGE}_test.go"

echo "创建变更文档..."
sed "s/{version}/$VERSION/g" "$TEMPLATE_DIR/CHANGES.md.template" > "$VERSION_DIR/CHANGES.md"

echo "版本 $VERSION 创建完成！"
echo ""
echo "目录结构:"
echo "  $VERSION_DIR/"
echo "  ├── incr-$VERSION.go"
echo "  ├── incr_v${VERSION_PACKAGE}_test.go"
echo "  └── CHANGES.md"
echo ""
echo "下一步:"
echo "1. 编辑 $VERSION_DIR/incr-$VERSION.go 实现版本逻辑"
echo "2. 编辑 $VERSION_DIR/incr_v${VERSION_PACKAGE}_test.go 编写测试"
echo "3. 编辑 $VERSION_DIR/CHANGES.md 填写变更说明"
echo "4. 在 cmd/initial/incr/types.go 中注册新版本"
echo ""
echo "注册示例:"
echo "  import \"github.com/Duke1616/ecmdb/cmd/initial/incr/$VERSION\""
echo "  registerIncr(v${VERSION_PACKAGE}.NewIncrV${VERSION_PACKAGE}(app))"
