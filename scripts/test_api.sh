#!/bin/bash

# AliMPay API 测试脚本
# Usage: ./scripts/test_api.sh

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
API_URL="${API_URL:-http://localhost:8080}"
PID="${PID:-1001003549245339}"
KEY="${KEY:-your_merchant_key}"

echo -e "${GREEN}=====================================${NC}"
echo -e "${GREEN}  AliMPay API 测试工具${NC}"
echo -e "${GREEN}=====================================${NC}"
echo ""
echo "API地址: $API_URL"
echo "商户ID: $PID"
echo ""

# 测试健康检查
test_health() {
    echo -e "${YELLOW}[TEST] 健康检查${NC}"
    response=$(curl -s "$API_URL/health?action=status")
    if echo "$response" | grep -q "success"; then
        echo -e "${GREEN}✓ 健康检查通过${NC}"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        echo -e "${RED}✗ 健康检查失败${NC}"
        echo "$response"
    fi
    echo ""
}

# 测试查询商户信息
test_query_merchant() {
    echo -e "${YELLOW}[TEST] 查询商户信息${NC}"
    response=$(curl -s "$API_URL/api?action=query&pid=$PID&key=$KEY")
    if echo "$response" | grep -q '"code":1'; then
        echo -e "${GREEN}✓ 查询商户信息成功${NC}"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        echo -e "${RED}✗ 查询商户信息失败${NC}"
        echo "$response"
    fi
    echo ""
}

# 生成MD5签名
generate_sign() {
    local params="$1"
    local key="$2"
    echo -n "${params}${key}" | md5sum | awk '{print $1}'
}

# 测试创建订单
test_create_order() {
    echo -e "${YELLOW}[TEST] 创建支付订单${NC}"
    
    # 订单参数
    OUT_TRADE_NO="TEST$(date +%s)"
    NAME="测试商品"
    MONEY="1.00"
    NOTIFY_URL="http://example.com/notify"
    RETURN_URL="http://example.com/return"
    TYPE="alipay"
    
    # 生成签名字符串（按字母顺序）
    SIGN_STR="money=${MONEY}&name=${NAME}&notify_url=${NOTIFY_URL}&out_trade_no=${OUT_TRADE_NO}&pid=${PID}&return_url=${RETURN_URL}&type=${TYPE}"
    SIGN=$(generate_sign "$SIGN_STR" "$KEY")
    
    # 发送请求
    response=$(curl -s -X POST "$API_URL/submit" \
        -d "pid=$PID" \
        -d "type=$TYPE" \
        -d "out_trade_no=$OUT_TRADE_NO" \
        -d "notify_url=$NOTIFY_URL" \
        -d "return_url=$RETURN_URL" \
        -d "name=$NAME" \
        -d "money=$MONEY" \
        -d "sign=$SIGN" \
        -d "sign_type=MD5")
    
    if echo "$response" | grep -q '"code":1'; then
        echo -e "${GREEN}✓ 创建订单成功${NC}"
        echo "订单号: $OUT_TRADE_NO"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        
        # 保存订单号供后续测试使用
        echo "$OUT_TRADE_NO" > /tmp/alimpay_test_order.txt
    else
        echo -e "${RED}✗ 创建订单失败${NC}"
        echo "$response"
    fi
    echo ""
}

# 测试查询订单
test_query_order() {
    echo -e "${YELLOW}[TEST] 查询订单状态${NC}"
    
    # 读取之前创建的订单号
    if [ -f /tmp/alimpay_test_order.txt ]; then
        OUT_TRADE_NO=$(cat /tmp/alimpay_test_order.txt)
    else
        echo -e "${YELLOW}⚠ 未找到测试订单，请先运行创建订单测试${NC}"
        echo ""
        return
    fi
    
    response=$(curl -s "$API_URL/api/order?pid=$PID&out_trade_no=$OUT_TRADE_NO")
    
    if echo "$response" | grep -q '"code":1'; then
        echo -e "${GREEN}✓ 查询订单成功${NC}"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        echo -e "${RED}✗ 查询订单失败${NC}"
        echo "$response"
    fi
    echo ""
}

# 测试标记已支付
test_mark_paid() {
    echo -e "${YELLOW}[TEST] 标记订单已支付${NC}"
    
    # 读取之前创建的订单号
    if [ -f /tmp/alimpay_test_order.txt ]; then
        OUT_TRADE_NO=$(cat /tmp/alimpay_test_order.txt)
    else
        echo -e "${YELLOW}⚠ 未找到测试订单，请先运行创建订单测试${NC}"
        echo ""
        return
    fi
    
    response=$(curl -s -X POST "$API_URL/admin?action=mark_paid&pid=$PID&key=$KEY&out_trade_no=$OUT_TRADE_NO")
    
    if echo "$response" | grep -q '"success":true'; then
        echo -e "${GREEN}✓ 标记支付成功${NC}"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        echo -e "${RED}✗ 标记支付失败${NC}"
        echo "$response"
    fi
    echo ""
}

# 主菜单
show_menu() {
    echo -e "${GREEN}请选择测试项：${NC}"
    echo "1) 健康检查"
    echo "2) 查询商户信息"
    echo "3) 创建支付订单"
    echo "4) 查询订单状态"
    echo "5) 标记订单已支付"
    echo "6) 运行所有测试"
    echo "0) 退出"
    echo ""
    read -p "请输入选项 [0-6]: " choice
    
    case $choice in
        1) test_health ;;
        2) test_query_merchant ;;
        3) test_create_order ;;
        4) test_query_order ;;
        5) test_mark_paid ;;
        6) 
            test_health
            test_query_merchant
            test_create_order
            test_query_order
            ;;
        0) 
            echo "退出测试"
            exit 0
            ;;
        *)
            echo -e "${RED}无效选项${NC}"
            ;;
    esac
    
    echo ""
    show_menu
}

# 检查依赖
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        echo -e "${RED}错误: 未找到 curl 命令${NC}"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        echo -e "${YELLOW}警告: 未找到 jq 命令，JSON输出可能不美观${NC}"
    fi
}

# 主程序
main() {
    check_dependencies
    
    if [ "$1" == "auto" ]; then
        # 自动运行所有测试
        test_health
        test_query_merchant
        test_create_order
        test_query_order
    else
        # 交互式菜单
        show_menu
    fi
}

main "$@"

