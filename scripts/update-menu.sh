#!/bin/bash

# 菜单数据管理脚本

# 用法: ./scripts/update-menu.sh [json_file]
# 示例: ./scripts/update-menu.sh init/c_menu.json
#       ./scripts/update-menu.sh /path/to/custom_menu.json

set -e

# 默认文件路径
DEFAULT_JSON_FILE="init/c_menu.json"

# 获取 JSON 文件路径
JSON_FILE="${1:-$DEFAULT_JSON_FILE}"

echo "🔄 菜单数据管理工具"
echo ""
echo "💡 使用说明:"
echo "   1. 修改菜单 JSON 文件来更新菜单数据"
echo "   2. 运行 './scripts/update-menu.sh [json_file]' 重新生成代码"
echo "   3. 运行 './ecmdb init' 来初始化系统"
echo "   4. 菜单数据会在系统初始化时自动加载到数据库"
echo ""
# 检查 JSON 文件是否存在
if [ ! -f "$JSON_FILE" ]; then
    echo "❌ 错误: $JSON_FILE 文件不存在"
    echo "💡 用法: $0 [json_file]"
    echo "   示例: $0 init/c_menu.json"
    echo "         $0 /path/to/custom_menu.json"
    exit 1
fi

echo "📋 当前菜单数据文件: $JSON_FILE"
echo "📊 菜单数据统计:"
echo "   - 总菜单项: $(jq length "$JSON_FILE")"
echo "   - 目录数量: $(jq '[.[] | select(.type == "1")] | length' "$JSON_FILE")"
echo "   - 菜单数量: $(jq '[.[] | select(.type == "2")] | length' "$JSON_FILE")"
echo "   - 按钮数量: $(jq '[.[] | select(.type == "3")] | length' "$JSON_FILE")"

echo ""
echo "🔧 生成菜单代码..."

# 生成菜单代码
cd cmd/tools/menu-generator
# 将相对路径转换为绝对路径
if [[ "$JSON_FILE" != /* ]]; then
    JSON_FILE="../../../$JSON_FILE"
fi
go run ast_generator.go "$JSON_FILE"
# 移动生成的文件到正确位置
mv menu_data.go ../../initial/menu/
cd ../../..

echo "✅ 菜单代码生成完成!"

# 只有在使用了自定义文件时才显示示例
# 注意：JSON_FILE 在代码生成时可能被修改为绝对路径，所以需要比较原始输入
if [ "$1" != "" ] && [ "$1" != "$DEFAULT_JSON_FILE" ]; then
    echo ""
    echo "📝 示例:"
    echo "   ./scripts/update-menu.sh                    # 使用默认文件 init/c_menu.json"
    echo "   ./scripts/update-menu.sh init/c_menu.json   # 指定文件路径"
    echo "   ./scripts/update-menu.sh /path/to/menu.json # 使用绝对路径"
fi
