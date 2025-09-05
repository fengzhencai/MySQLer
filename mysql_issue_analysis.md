# MySQL查询异常问题分析报告

## 问题概述

在MySQL 5.7.25环境下，针对分区表`sku_wh_attr`执行相似查询时出现结果不一致的问题：
- 查询1（带FOR UPDATE）：返回空结果
- 查询2（带完整主键条件）：返回正常结果

## 环境信息
- **MySQL版本**: 5.7.25
- **隔离级别**: REPEATABLE-READ
- **表类型**: InnoDB分区表（250个分区，按skuCode进行KEY分区）
- **主键**: (skuCode, whCode)

## 问题分析

### 1. FOR UPDATE锁机制问题

#### 可能原因：
- **锁等待超时**: 默认`innodb_lock_wait_timeout=50秒`，如果其他事务持有锁超时会返回空结果
- **死锁检测**: InnoDB检测到死锁风险，主动回滚查询
- **锁粒度问题**: 在分区表中，FOR UPDATE可能产生更大范围的锁

#### 验证方法：
```sql
-- 检查锁等待超时设置
SHOW VARIABLES LIKE 'innodb_lock_wait_timeout';

-- 检查当前锁状态
SELECT * FROM INFORMATION_SCHEMA.INNODB_LOCKS;
SELECT * FROM INFORMATION_SCHEMA.INNODB_LOCK_WAITS;
```

### 2. 分区表查询优化差异

#### 查询1分析：
```sql
WHERE `skuCode` = 'YAFEXPPP0419972001' FOR UPDATE
```
- 只有分区键条件，应该能分区裁剪
- 但FOR UPDATE可能影响查询优化器的执行计划

#### 查询2分析：
```sql
WHERE `skuCode` = 'YAFEXPPP0419972001' AND `whCode` = 'D02' LIMIT 1
```
- 完整主键条件，查询效率最高
- 无锁操作，不会有锁等待问题

### 3. 并发事务影响

在REPEATABLE-READ隔离级别下：
- 可能存在幻读保护机制
- 其他事务的未提交更改可能影响FOR UPDATE查询
- 长时间运行的事务可能持有相关锁

## 解决方案

### 立即解决方案

#### 方案1：移除FOR UPDATE（推荐）
```sql
-- 如果不需要行锁，直接移除FOR UPDATE
SELECT `skuCode`, `banBuy`, `safety`, `taskQty`, `waitQty`, `prodQty`, `tranQty`, `signQty`, `putOnQty`, `exchange`, `refunds`, `damage`, `moreQty`, `cancel`, `avaiQty`, `rSeller`, `mSeller`, `whCode`, `occuQty`, `readyQty`
FROM `sku_wh_attr`
WHERE `skuCode` = 'YAFEXPPP0419972001';
```

#### 方案2：添加完整主键条件
```sql
-- 如果知道whCode，添加完整主键条件
SELECT `skuCode`, `banBuy`, `safety`, `taskQty`, `waitQty`, `prodQty`, `tranQty`, `signQty`, `putOnQty`, `exchange`, `refunds`, `damage`, `moreQty`, `cancel`, `avaiQty`, `rSeller`, `mSeller`, `whCode`, `occuQty`, `readyQty`
FROM `sku_wh_attr`
WHERE `skuCode` = 'YAFEXPPP0419972001' AND `whCode` = 'D02'
FOR UPDATE;
```

#### 方案3：降低锁等待超时
```sql
-- 在会话级别设置更短的锁等待时间
SET SESSION innodb_lock_wait_timeout = 5;
```

### 长期优化方案

#### 1. 优化查询策略
```sql
-- 使用NOWAIT避免长时间等待
SELECT * FROM sku_wh_attr 
WHERE skuCode = 'YAFEXPPP0419972001' 
FOR UPDATE NOWAIT;

-- 使用SKIP LOCKED跳过被锁定的行
SELECT * FROM sku_wh_attr 
WHERE skuCode = 'YAFEXPPP0419972001' 
FOR UPDATE SKIP LOCKED;
```

#### 2. 应用层面优化
- 缩短事务持续时间
- 避免在长时间事务中使用FOR UPDATE
- 实现重试机制处理锁等待

#### 3. 监控和预警
```sql
-- 创建监控视图，定期检查锁等待
CREATE VIEW lock_monitor AS
SELECT 
    r.trx_id as waiting_trx,
    r.trx_mysql_thread_id as waiting_thread,
    b.trx_id as blocking_trx,
    b.trx_mysql_thread_id as blocking_thread,
    TIMESTAMPDIFF(SECOND, r.trx_wait_started, NOW()) as wait_seconds
FROM INFORMATION_SCHEMA.INNODB_LOCK_WAITS w
INNER JOIN INFORMATION_SCHEMA.INNODB_TRX b ON b.trx_id = w.blocking_trx_id
INNER JOIN INFORMATION_SCHEMA.INNODB_TRX r ON r.trx_id = w.requesting_trx_id;
```

## 诊断步骤

1. **执行诊断脚本**: 运行提供的`mysql_diagnosis.sql`
2. **检查锁状态**: 确认是否存在锁等待或死锁
3. **分析执行计划**: 对比两个查询的EXPLAIN结果
4. **验证数据存在性**: 不使用FOR UPDATE确认数据确实存在
5. **监控并发事务**: 检查是否有长时间运行的事务

## 预防措施

1. **代码层面**:
   - 避免在读取操作中不必要地使用FOR UPDATE
   - 优先使用完整主键条件查询
   - 实现查询超时和重试机制

2. **数据库层面**:
   - 定期监控锁等待情况
   - 适当调整`innodb_lock_wait_timeout`参数
   - 考虑使用READ COMMITTED隔离级别（如果业务允许）

3. **应用设计**:
   - 尽可能缩短事务范围
   - 避免在事务中执行耗时操作
   - 合理设计并发控制策略

## 结论

该问题最可能的原因是**FOR UPDATE导致的锁等待或死锁**。建议首先尝试移除FOR UPDATE子句，如果确实需要行锁，则应添加完整的主键条件或使用NOWAIT/SKIP LOCKED选项来避免长时间等待。