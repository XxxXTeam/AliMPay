# 经营码二维码目录

## 说明

此目录用于存放支付宝经营码二维码图片。支持单个二维码和多个二维码模式。

## 使用步骤

### 单二维码模式（适合小规模使用）

#### 1. 获取经营码

登录支付宝商家中心，获取您的经营码二维码图片。

#### 2. 保存二维码

将二维码图片保存为 `business_qr.png` 并放置在此目录。

```bash
# 示例
cp /path/to/your/alipay_qrcode.png business_qr.png
```

#### 3. 配置

在 `configs/config.yaml` 中配置：

```yaml
payment:
  business_qr_mode:
    enabled: true
    qr_code_path: "./qrcode/business_qr.png"
    qr_code_id: ""  # 可选：支付宝收款码ID
```

### 多二维码模式（推荐用于高并发场景）

**优势**：
- 分散支付负载，减少金额冲突
- 提高支付成功率
- 支持多种轮询策略

#### 1. 准备多个经营码

从支付宝商家中心获取多个经营码二维码图片。

#### 2. 保存二维码

将多个二维码保存到此目录：

```bash
# 示例：保存3个二维码
cp /path/to/qr1.png business_qr_1.png
cp /path/to/qr2.png business_qr_2.png
cp /path/to/qr3.png business_qr_3.png
```

#### 3. 配置多二维码

在 `configs/config.yaml` 中配置：

```yaml
payment:
  business_qr_mode:
    enabled: true
    # 多二维码配置
    qr_code_paths:
      - id: "qr1"
        path: "./qrcode/business_qr_1.png"
        code_id: "fkx123456"  # 支付宝收款码ID
        enabled: true
        priority: 1
      - id: "qr2"
        path: "./qrcode/business_qr_2.png"
        code_id: "fkx789012"
        enabled: true
        priority: 2
      - id: "qr3"
        path: "./qrcode/business_qr_3.png"
        code_id: "fkx345678"
        enabled: true
        priority: 3
    # 轮询模式：round_robin（轮询）、random（随机）、least_used（最少使用）
    polling_mode: "round_robin"
```

#### 4. 轮询策略说明

**round_robin（轮询模式）**：
- 依次使用每个二维码
- 负载均衡最好
- 推荐用于订单量稳定的场景

**random（随机模式）**：
- 随机选择二维码
- 分布较均匀
- 适合订单量不确定的场景

**least_used（最少使用模式）**：
- 优先使用使用次数最少的二维码
- 自动平衡负载
- 适合长期运行的服务

### 验证配置

启动服务后，访问：

```bash
# 单二维码模式
TOKEN=$(echo -n "qrcode_access_$(date +%Y-%m-%d)" | md5sum | cut -d' ' -f1)
curl "http://localhost:8080/qrcode?type=business&token=$TOKEN" > test.png

# 多二维码模式（指定ID）
curl "http://localhost:8080/qrcode?type=business&token=$TOKEN&id=qr1" > test_qr1.png
curl "http://localhost:8080/qrcode?type=business&token=$TOKEN&id=qr2" > test_qr2.png
```

## 注意事项

1. **文件格式**: 支持 PNG、JPG、JPEG 格式
2. **文件大小**: 建议不超过 2MB
3. **图片质量**: 确保二维码清晰可扫描
4. **权限设置**: 确保程序有读取权限
5. **多二维码建议**: 建议配置 3-5 个二维码，过多会增加管理成本

## 安全

- 二维码访问需要 token 验证
- Token 每天自动更新
- 只有知道 token 的人才能访问
- 多二维码模式下，系统自动分配二维码，用户无法指定

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

### 问题：多二维码模式不生效

**解决方案**:
1. 确认配置了 `qr_code_paths` 且至少有2个启用的二维码
2. 检查所有二维码文件是否存在
3. 查看日志确认 QRCodeSelector 是否初始化成功
4. 检查 `enabled: true` 是否设置

## 示例

### 单二维码示例

```bash
# 上传二维码
cp ~/Downloads/alipay_business_qr.png business_qr.png

# 验证上传
ls -lh business_qr.png

# 测试访问
TOKEN=$(echo -n "qrcode_access_$(date +%Y-%m-%d)" | md5sum | cut -d' ' -f1)
curl "http://localhost:8080/qrcode?type=business&token=$TOKEN" > test.png
```

### 多二维码示例

```bash
# 上传多个二维码
cp ~/Downloads/qr1.png business_qr_1.png
cp ~/Downloads/qr2.png business_qr_2.png
cp ~/Downloads/qr3.png business_qr_3.png

# 验证上传
ls -lh business_qr_*.png

# 测试访问不同的二维码
TOKEN=$(echo -n "qrcode_access_$(date +%Y-%m-%d)" | md5sum | cut -d' ' -f1)
curl "http://localhost:8080/qrcode?type=business&token=$TOKEN&id=qr1" > test_qr1.png
curl "http://localhost:8080/qrcode?type=business&token=$TOKEN&id=qr2" > test_qr2.png
curl "http://localhost:8080/qrcode?type=business&token=$TOKEN&id=qr3" > test_qr3.png
```

## 性能优化建议

1. **二维码数量**: 根据并发订单量配置，一般 3-5 个即可
2. **轮询策略**: 默认使用 round_robin，简单高效
3. **优先级设置**: 可以将某些二维码设置为更高优先级
4. **动态调整**: 可以通过修改配置文件动态启用/禁用某个二维码

