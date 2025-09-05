# 🚨 MySQL数据一致性紧急问题 - 行动计划

## 问题严重程度：**CRITICAL**

**现象**：同一表相同数据，不同查询条件返回不同结果
**影响**：严重的数据一致性问题，可能导致业务逻辑错误

---

## 🔥 立即行动（15分钟内完成）

### 1. 紧急排查当前状态
```sql
-- 检查是否有长时间运行的事务
SELECT trx_id, trx_state, trx_started, 
       TIMESTAMPDIFF(SECOND, trx_started, NOW()) as duration_seconds,
       trx_mysql_thread_id, trx_query
FROM INFORMATION_SCHEMA.INNODB_TRX 
WHERE TIMESTAMPDIFF(SECOND, trx_started, NOW()) > 60
ORDER BY trx_started;
```

### 2. 检查锁等待情况
```sql
-- 如果有结果，说明存在锁等待问题
SELECT COUNT(*) as lock_wait_count
FROM INFORMATION_SCHEMA.INNODB_LOCK_WAITS;
```

### 3. 立即临时解决方案

#### 方案A：应用层修改（推荐）
**将原查询改为分步查询：**

```sql
-- 步骤1：先获取所有相关的whCode
SELECT DISTINCT whCode 
FROM sku_wh_attr 
WHERE skuCode = 'YAFEXPPP0419972001';

-- 步骤2：循环每个whCode进行精确查询
SELECT `skuCode`, `banBuy`, `safety`, `taskQty`, `waitQty`, `prodQty`, `tranQty`, `signQty`, `putOnQty`, `exchange`, `refunds`, `damage`, `moreQty`, `cancel`, `avaiQty`, `rSeller`, `mSeller`, `whCode`, `occuQty`, `readyQty`
FROM `sku_wh_attr`
WHERE `skuCode` = 'YAFEXPPP0419972001' AND `whCode` = '{具体的whCode}'
FOR UPDATE;
```

#### 方案B：移除FOR UPDATE（如果业务允许）
```sql
-- 暂时移除FOR UPDATE，避免锁等待
SELECT `skuCode`, `banBuy`, `safety`, `taskQty`, `waitQty`, `prodQty`, `tranQty`, `signQty`, `putOnQty`, `exchange`, `refunds`, `damage`, `moreQty`, `cancel`, `avaiQty`, `rSeller`, `mSeller`, `whCode`, `occuQty`, `readyQty`
FROM `sku_wh_attr`
WHERE `skuCode` = 'YAFEXPPP0419972001';
```

---

## ⚡ 30分钟内深度排查

### 1. 执行完整诊断
运行 `emergency_mysql_fix.sql` 中的所有诊断命令

### 2. 关键检查点

#### A. 事务状态检查
```sql
-- 查看所有活跃事务
SELECT trx_id, trx_state, trx_started, trx_mysql_thread_id, trx_query
FROM INFORMATION_SCHEMA.INNODB_TRX;
```

#### B. 锁状态检查
```sql
-- 查看锁等待
SELECT * FROM INFORMATION_SCHEMA.INNODB_LOCK_WAITS;
```

#### C. 分区表特殊检查
```sql
-- 检查分区裁剪是否正常
EXPLAIN PARTITIONS
SELECT * FROM sku_wh_attr WHERE skuCode = 'YAFEXPPP0419972001';
```

---

## 🛠️ 根本原因分析

### 最可能的原因（按概率排序）

1. **FOR UPDATE锁等待超时 (80%)**
   - 其他事务持有行锁
   - 默认50秒超时后返回空结果
   - 解决：使用NOWAIT或分步查询

2. **分区表查询优化器bug (15%)**
   - MySQL 5.7.25的已知问题
   - 解决：升级MySQL或改写查询

3. **索引损坏或统计信息错误 (5%)**
   - 解决：重建索引或更新统计信息

### 验证方法

#### 测试1：锁等待验证
```sql
-- 在新连接中执行，如果立即返回空结果，说明是锁问题
SELECT COUNT(*) FROM sku_wh_attr 
WHERE skuCode = 'YAFEXPPP0419972001' 
FOR UPDATE NOWAIT;
```

#### 测试2：分区表验证
```sql
-- 检查执行计划是否一致
EXPLAIN FORMAT=JSON
SELECT * FROM sku_wh_attr WHERE skuCode = 'YAFEXPPP0419972001';
```

---

## 📋 业务影响评估

### 当前风险
- ✅ 数据存在但查询不到 → **订单处理错误**
- ✅ 库存计算不准确 → **超卖或缺货**
- ✅ 业务逻辑判断错误 → **财务损失**

### 紧急措施
1. **监控业务报错** - 检查是否有相关业务异常
2. **数据核对** - 人工核对关键订单数据
3. **降级方案** - 必要时暂停相关功能

---

## 🔧 立即实施方案

### 代码层面（优先级1）
```python
# 示例：Python应用代码修改
def get_sku_wh_data(sku_code):
    # 原来的方法（有问题）
    # query = "SELECT ... FROM sku_wh_attr WHERE skuCode = %s FOR UPDATE"
    
    # 临时解决方案1：分步查询
    wh_codes = execute_query("SELECT DISTINCT whCode FROM sku_wh_attr WHERE skuCode = %s", [sku_code])
    
    results = []
    for wh_code in wh_codes:
        data = execute_query("""
            SELECT `skuCode`, `banBuy`, `safety`, `taskQty`, `waitQty`, `prodQty`, `tranQty`, 
                   `signQty`, `putOnQty`, `exchange`, `refunds`, `damage`, `moreQty`, 
                   `cancel`, `avaiQty`, `rSeller`, `mSeller`, `whCode`, `occuQty`, `readyQty`
            FROM `sku_wh_attr`
            WHERE `skuCode` = %s AND `whCode` = %s
            FOR UPDATE
        """, [sku_code, wh_code[0]])
        results.extend(data)
    
    return results
```

### 数据库层面（优先级2）
```sql
-- 如果发现是锁等待问题，临时调整参数
SET GLOBAL innodb_lock_wait_timeout = 5;  -- 降低等待时间，快速失败
```

---

## 📊 监控和预警

### 立即设置监控
```sql
-- 创建监控视图
CREATE VIEW emergency_lock_monitor AS
SELECT 
    COUNT(*) as active_trx_count,
    COUNT(CASE WHEN TIMESTAMPDIFF(SECOND, trx_started, NOW()) > 60 THEN 1 END) as long_trx_count,
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.INNODB_LOCK_WAITS) as lock_wait_count
FROM INFORMATION_SCHEMA.INNODB_TRX;

-- 每分钟检查一次
-- SELECT * FROM emergency_lock_monitor;
```

---

## ✅ 下一步行动清单

### 立即执行（现在）
- [ ] 运行诊断脚本确认问题原因
- [ ] 实施临时解决方案（分步查询或移除FOR UPDATE）
- [ ] 检查业务是否有异常报错

### 30分钟内
- [ ] 完成深度排查
- [ ] 确认根本原因
- [ ] 实施永久解决方案

### 1小时内
- [ ] 验证修复效果
- [ ] 业务功能回归测试
- [ ] 文档记录和总结

---

## 🚨 紧急联系和升级

如果问题无法在30分钟内解决：
1. **通知业务方** - 暂停相关功能
2. **DBA支持** - 寻求专业数据库管理员帮助
3. **考虑回滚** - 如果是新版本引入的问题

**记住：数据一致性问题绝不能拖延，必须立即处理！**