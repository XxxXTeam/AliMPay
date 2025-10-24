# 多二维码轮询功能文档

## 功能概述

多二维码轮询功能允许系统同时使用多个支付宝经营码进行收款，通过智能的轮询策略分配二维码给订单，有效解决以下问题：

1. **金额冲突**: 多个订单同时创建时可能产生相同的支付金额，导致匹配错误
2. **支付限流**: 单个二维码可能存在支付限制
3. **负载均衡**: 将支付流量分散到多个二维码，提高系统稳定性
4. **容灾能力**: 某个二维码失效时，系统可以继续使用其他二维码

## 功能特性

### 1. 多种轮询策略

#### Round Robin（轮询模式）- 默认推荐
- **工作原理**: 按顺序依次分配二维码
- **优点**: 
  - 负载分布最均匀
  - 可预测性强
  - 实现简单高效
- **适用场景**: 订单量稳定的场景
- **配置**: `polling_mode: "round_robin"`

#### Random（随机模式）
- **工作原理**: 随机选择一个可用的二维码
- **优点**: 
  - 分布相对均匀
  - 防止规律性攻击
- **适用场景**: 订单量不确定或需要随机性的场景
- **配置**: `polling_mode: "random"`

#### Least Used（最少使用模式）
- **工作原理**: 优先选择使用次数最少的二维码
- **优点**: 
  - 自动平衡负载
  - 适应动态变化
- **适用场景**: 长期运行且二维码性能不均衡的场景
- **配置**: `polling_mode: "least_used"`

### 2. 优先级支持

每个二维码可以设置优先级（priority），数字越小优先级越高。系统会先按优先级排序，然后应用轮询策略。

### 3. 动态启用/禁用

通过 `enabled` 字段可以动态启用或禁用某个二维码，无需重启服务。

### 4. 向后兼容

完全兼容原有的单二维码配置，现有配置无需修改即可继续使用。

## 配置说明

### 基础配置

在 `configs/config.yaml` 中添加多二维码配置：

```yaml
payment:
  business_qr_mode:
    enabled: true
    
    # 多二维码配置
    qr_code_paths:
      - id: "qr1"                               # 唯一标识
        path: "./qrcode/business_qr_1.png"      # 二维码图片路径
        code_id: "fkx123456"                    # 支付宝收款码ID（可选）
        enabled: true                           # 是否启用
        priority: 1                             # 优先级（越小越高）
        
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
    
    # 轮询策略
    polling_mode: "round_robin"
    
    # 其他配置...
    amount_offset: 0.01
    match_tolerance: 300
    payment_timeout: 300
```

### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 二维码唯一标识，用于日志和追踪 |
| path | string | 是 | 二维码图片文件路径 |
| code_id | string | 否 | 支付宝收款码ID，用于手机端拉起支付宝 |
| enabled | boolean | 是 | 是否启用此二维码 |
| priority | int | 是 | 优先级，数字越小优先级越高 |

### 轮询模式配置

| 模式 | 值 | 说明 |
|------|------|------|
| 轮询 | round_robin | 依次使用每个二维码（推荐） |
| 随机 | random | 随机选择二维码 |
| 最少使用 | least_used | 优先使用使用次数最少的二维码 |

## 使用指南

### 1. 准备二维码

从支付宝商家中心获取多个经营码：

```bash
# 将二维码保存到 qrcode 目录
cp /path/to/qr1.png ./qrcode/business_qr_1.png
cp /path/to/qr2.png ./qrcode/business_qr_2.png
cp /path/to/qr3.png ./qrcode/business_qr_3.png
```

### 2. 更新配置

编辑 `configs/config.yaml`，添加多二维码配置（参考上面的配置示例）。

### 3. 重启服务

```bash
# 停止服务
pkill alimpay

# 启动服务
./alimpay -config=./configs/config.yaml

# 或使用 systemd
systemctl restart alimpay
```

### 4. 验证配置

查看日志确认初始化成功：

```bash
tail -f logs/alimpay.log | grep "QR code selector"
```

应该看到类似输出：
```
{"level":"info","msg":"QR code selector initialized","qr_code_count":3,"polling_mode":"round_robin"}
```

## 工作流程

### 订单创建流程

1. 用户创建支付订单
2. 系统根据配置的轮询策略选择一个二维码
3. 将选中的二维码ID记录到订单中
4. 返回包含二维码信息的支付页面

### 支付页面显示

1. 用户访问支付页面
2. 系统根据订单中保存的二维码ID获取对应的二维码图片
3. 显示正确的二维码给用户扫描支付

### 支付监控

1. 监控服务定期查询支付宝账单
2. 根据金额匹配待支付订单
3. 更新订单状态为已支付

## 性能优化

### 推荐配置

根据不同的并发量，推荐以下配置：

| 日订单量 | 二维码数量 | 轮询模式 | 说明 |
|----------|------------|----------|------|
| < 100 | 1 | - | 使用单二维码即可 |
| 100-500 | 2-3 | round_robin | 基本够用 |
| 500-2000 | 3-5 | round_robin | 推荐配置 |
| > 2000 | 5-10 | least_used | 需要更多二维码 |

### 优化建议

1. **二维码数量**: 不要配置过多，3-5个通常足够
2. **优先级设置**: 可以将质量更好的二维码设置为更高优先级
3. **监控调整**: 根据实际使用情况调整二维码配置
4. **定期检查**: 定期检查各二维码的使用统计，必要时调整

## 监控与统计

### 查看统计信息

系统提供了统计接口来查看各二维码的使用情况（需要实现管理后台接口）：

```bash
# 查看二维码使用统计
curl http://localhost:8080/api/admin/qrcode/stats
```

返回示例：
```json
{
  "enabled": true,
  "qr_code_count": 3,
  "polling_mode": "round_robin",
  "stats": [
    {
      "id": "qr1",
      "usage_count": 150,
      "last_used_time": "2024-01-15T12:30:00Z",
      "priority": 1
    },
    {
      "id": "qr2",
      "usage_count": 148,
      "last_used_time": "2024-01-15T12:29:50Z",
      "priority": 2
    },
    {
      "id": "qr3",
      "usage_count": 152,
      "last_used_time": "2024-01-15T12:30:10Z",
      "priority": 3
    }
  ]
}
```

## 故障排除

### 问题1: 多二维码不生效

**症状**: 配置了多个二维码，但系统仍使用单个二维码

**排查步骤**:
1. 检查配置文件中 `qr_code_paths` 是否正确配置
2. 确认至少有2个二维码设置了 `enabled: true`
3. 查看日志确认 QRCodeSelector 是否初始化
4. 检查所有二维码文件是否存在

**解决方案**:
```bash
# 检查配置
cat configs/config.yaml | grep -A 20 "qr_code_paths"

# 检查文件
ls -lh qrcode/business_qr_*.png

# 查看日志
tail -f logs/alimpay.log | grep -i "qrcode\|selector"
```

### 问题2: 某个二维码使用频率过高

**症状**: 某个二维码使用次数远超其他二维码

**原因**: 可能是优先级设置不当或轮询模式选择不合适

**解决方案**:
1. 检查优先级配置，确保各二维码优先级相近
2. 尝试切换到 `least_used` 模式
3. 临时禁用使用过多的二维码，让其他二维码分担流量

### 问题3: 支付页面显示错误的二维码

**症状**: 订单分配了二维码A，但支付页面显示了二维码B

**排查步骤**:
1. 查看订单的 `qr_code_id` 字段
2. 检查二维码ID是否在配置中存在
3. 查看日志中的二维码分配记录

**解决方案**:
```bash
# 查询订单信息
sqlite3 data/alimpay.db "SELECT id, qr_code_id FROM codepay_orders WHERE id='订单号';"

# 检查配置
cat configs/config.yaml | grep -A 5 "id: \"qr"
```

## 迁移指南

### 从单二维码迁移到多二维码

#### 步骤1: 备份配置
```bash
cp configs/config.yaml configs/config.yaml.backup
```

#### 步骤2: 准备二维码文件
```bash
# 保留原二维码
cp qrcode/business_qr.png qrcode/business_qr_1.png

# 添加新二维码
cp /path/to/new_qr2.png qrcode/business_qr_2.png
cp /path/to/new_qr3.png qrcode/business_qr_3.png
```

#### 步骤3: 更新配置
```yaml
payment:
  business_qr_mode:
    enabled: true
    # 保留原配置（向后兼容）
    qr_code_path: "./qrcode/business_qr.png"
    qr_code_id: "fkx123456"
    
    # 添加多二维码配置
    qr_code_paths:
      - id: "qr1"
        path: "./qrcode/business_qr_1.png"
        code_id: "fkx123456"
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
    
    polling_mode: "round_robin"
```

#### 步骤4: 重启服务并验证
```bash
# 重启服务
systemctl restart alimpay

# 验证日志
tail -f logs/alimpay.log | grep "QR code"

# 测试创建订单
curl -X POST "http://localhost:8080/submit?..." 
```

## 安全建议

1. **二维码文件保护**: 确保二维码文件权限正确，不被未授权访问
2. **配置文件安全**: 妥善保管配置文件，特别是 code_id 信息
3. **日志监控**: 定期检查日志，发现异常及时处理
4. **定期更换**: 建议定期更换二维码，提高安全性

## 常见问题

**Q1: 可以动态添加二维码吗？**

A: 可以。修改配置文件添加新的二维码，然后重启服务即可。

**Q2: 如何临时禁用某个二维码？**

A: 在配置文件中将对应二维码的 `enabled` 设置为 `false`，然后重启服务。

**Q3: 轮询模式可以动态切换吗？**

A: 需要修改配置文件中的 `polling_mode`，然后重启服务。

**Q4: 二维码失效了怎么办？**

A: 将失效的二维码 `enabled` 设置为 `false`，系统会自动使用其他可用二维码。

**Q5: 多二维码会影响性能吗？**

A: 不会。QRCodeSelector 使用了高效的数据结构和算法，性能开销可以忽略不计。

## 更新日志

### v1.1.0 (2024-01-15)
- ✨ 新增多二维码轮询功能
- ✨ 支持三种轮询策略：round_robin、random、least_used
- ✨ 支持优先级配置
- ✨ 支持动态启用/禁用
- ✨ 完全向后兼容单二维码模式
- 📝 新增详细文档和配置示例

## 技术细节

### 数据库变更

添加了 `qr_code_id` 字段到订单表：

```sql
ALTER TABLE codepay_orders ADD COLUMN qr_code_id VARCHAR(32) DEFAULT '';
CREATE INDEX idx_qr_code_id ON codepay_orders(qr_code_id);
```

### API 变更

无需变更现有API，完全向后兼容。

### 核心类

- `QRCodeSelector`: 二维码选择器，负责选择和分配二维码
- `config.QRCode`: 二维码配置结构
- `config.BusinessQRMode`: 经营码模式配置（扩展）

## 贡献

欢迎提交 Issue 和 Pull Request 改进此功能！

## 许可

MIT License
