#!/bin/bash

# 测试支付宝深链接功能
# Test Alipay Deep Link Feature

echo "=========================================="
echo "  支付宝深链接功能测试脚本"
echo "  Alipay Deep Link Feature Test"
echo "=========================================="
echo ""

# 配置
BASE_URL="http://localhost:8080"
QR_CODE_ID="fkx12345678901234"  # 示例二维码ID，请替换为真实ID
AMOUNT="1.23"
REMARK="测试订单"

echo "配置信息："
echo "  服务地址: $BASE_URL"
echo "  二维码ID: $QR_CODE_ID"
echo "  支付金额: ¥$AMOUNT"
echo "  备注信息: $REMARK"
echo ""

# 测试1: 使用配置的默认二维码ID
echo "测试1: 使用默认二维码ID生成深链接"
echo "----------------------------------------"
echo "请求: GET /alipay/link?amount=$AMOUNT&remark=$REMARK"
echo ""

curl -s "$BASE_URL/alipay/link?amount=$AMOUNT&remark=$REMARK" | jq '.' || echo "服务未启动或jq未安装"
echo ""

# 测试2: 使用自定义二维码ID
echo ""
echo "测试2: 使用自定义二维码ID生成深链接"
echo "----------------------------------------"
echo "请求: GET /alipay/link?qr_code_id=$QR_CODE_ID&amount=$AMOUNT&remark=$REMARK"
echo ""

curl -s "$BASE_URL/alipay/link?qr_code_id=$QR_CODE_ID&amount=$AMOUNT&remark=$REMARK" | jq '.' || echo "服务未启动或jq未安装"
echo ""

# 测试3: 不带金额和备注
echo ""
echo "测试3: 仅生成二维码链接（不带金额）"
echo "----------------------------------------"
echo "请求: GET /alipay/link?qr_code_id=$QR_CODE_ID"
echo ""

curl -s "$BASE_URL/alipay/link?qr_code_id=$QR_CODE_ID" | jq '.' || echo "服务未启动或jq未安装"
echo ""

# 测试4: 错误场景 - 缺少二维码ID
echo ""
echo "测试4: 错误场景 - 缺少二维码ID"
echo "----------------------------------------"
echo "请求: GET /alipay/link?amount=$AMOUNT"
echo ""

curl -s "$BASE_URL/alipay/link?amount=$AMOUNT" | jq '.' || echo "服务未启动或jq未安装"
echo ""

# 测试5: 错误场景 - 无效金额
echo ""
echo "测试5: 错误场景 - 无效金额格式"
echo "----------------------------------------"
echo "请求: GET /alipay/link?qr_code_id=$QR_CODE_ID&amount=invalid"
echo ""

curl -s "$BASE_URL/alipay/link?qr_code_id=$QR_CODE_ID&amount=invalid" | jq '.' || echo "服务未启动或jq未安装"
echo ""

# 测试6: 直接重定向到支付宝（仅显示重定向URL）
echo ""
echo "测试6: 直接重定向到支付宝"
echo "----------------------------------------"
echo "请求: GET /alipay/pay?qr_code_id=$QR_CODE_ID&amount=$AMOUNT&remark=$REMARK"
echo ""
echo "重定向URL:"
curl -s -I "$BASE_URL/alipay/pay?qr_code_id=$QR_CODE_ID&amount=$AMOUNT&remark=$REMARK" | grep -i "Location:" || echo "服务未启动"
echo ""

echo ""
echo "=========================================="
echo "  测试完成"
echo "=========================================="
echo ""
echo "提示："
echo "1. 请确保服务已启动: make run"
echo "2. 深链接仅在移动设备上有效"
echo "3. 需要在配置文件中设置正确的 qr_code_id"
echo ""
