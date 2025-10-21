# 经营码二维码目录

## 说明

此目录用于存放支付宝经营码二维码图片。

## 使用步骤

### 1. 获取经营码

登录支付宝商家中心，获取您的经营码二维码图片。

### 2. 保存二维码

将二维码图片保存为 `business_qr.png` 并放置在此目录。

```bash
# 示例
cp /path/to/your/alipay_qrcode.png business_qr.png
```

### 3. 验证

启动服务后，访问：

```
http://localhost:8080/qrcode?type=business&token=今日token
```

应该能看到您上传的二维码图片。

## 注意事项

1. **文件格式**: 支持 PNG、JPG、JPEG 格式
2. **文件大小**: 建议不超过 2MB
3. **图片质量**: 确保二维码清晰可扫描
4. **权限设置**: 确保程序有读取权限

## 安全

- 二维码访问需要 token 验证
- Token 每天自动更新
- 只有知道 token 的人才能访问

## 配置

在 `configs/config.yaml` 中配置：

```yaml
payment:
  business_qr_mode:
    enabled: true
    qr_code_path: "./qrcode/business_qr.png"
```

## 故障排除

### 问题：无法访问二维码

**解决方案**:
1. 检查文件是否存在
2. 检查文件权限
3. 检查 token 是否正确
4. 查看服务日志

### 问题：二维码显示不清晰

**解决方案**:
1. 使用更高分辨率的图片
2. 确保原始二维码质量
3. 避免过度压缩

## 示例

```bash
# 上传二维码
cp ~/Downloads/alipay_business_qr.png business_qr.png

# 验证上传
ls -lh business_qr.png

# 测试访问（获取今日token）
TOKEN=$(echo -n "qrcode_access_$(date +%Y-%m-%d)" | md5sum | cut -d' ' -f1)
curl "http://localhost:8080/qrcode?type=business&token=$TOKEN" > test.png
```

