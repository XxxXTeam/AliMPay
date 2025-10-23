#!/bin/bash
# 测试PHP后缀路由和双斜杠路径规范化

echo "========================================="
echo "测试 PHP 后缀路由支持"
echo "========================================="
echo ""

# 测试submit.php
echo "1. 测试 /submit.php"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "http://localhost:8080/submit.php?pid=1001001276912812&type=alipay&out_trade_no=TEST123&money=0.01&name=test&notify_url=http://example.com&return_url=http://example.com&sign=test&sign_type=MD5"
echo ""

# 测试双斜杠路径
echo "2. 测试 //submit (双斜杠，应自动规范化为 /submit)"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "http://localhost:8080//submit?pid=1001001276912812&type=alipay&out_trade_no=TEST123&money=0.01&name=test&notify_url=http://example.com&return_url=http://example.com&sign=test&sign_type=MD5"
echo ""

# 测试//submit.php
echo "3. 测试 //submit.php (双斜杠+.php)"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "http://localhost:8080//submit.php?pid=1001001276912812&type=alipay&out_trade_no=TEST123&money=0.01&name=test&notify_url=http://example.com&return_url=http://example.com&sign=test&sign_type=MD5"
echo ""

# 测试api.php
echo "4. 测试 /api.php"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "http://localhost:8080/api.php?act=order&pid=1001001276912812&trade_no=TEST123"
echo ""

# 测试mapi.php
echo "5. 测试 /mapi.php"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "http://localhost:8080/mapi.php?act=order&pid=1001001276912812&trade_no=TEST123"
echo ""

# 测试notify.php
echo "6. 测试 /notify.php"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "http://localhost:8080/notify.php"
echo ""

# 测试callback.php
echo "7. 测试 /callback.php"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "http://localhost:8080/callback.php"
echo ""

# 测试三斜杠
echo "8. 测试 ///api (三斜杠，应规范化为 /api)"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "http://localhost:8080///api?act=order&pid=test&trade_no=test"
echo ""

# 测试末尾斜杠
echo "9. 测试 /submit/ (末尾斜杠，应规范化为 /submit)"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "http://localhost:8080/submit/?pid=test"
echo ""

echo "========================================="
echo "测试完成！"
echo "========================================="
echo ""
echo "预期结果："
echo "- 所有带.php后缀的路由应返回200或其他正常状态码（非404）"
echo "- 双斜杠路径应自动规范化，返回与单斜杠相同的结果"
echo ""

