-- MySQL 查询问题诊断脚本
-- 针对 sku_wh_attr 表的查询异常分析

-- 1. 检查当前事务和锁状态
SELECT 
    trx_id,
    trx_state,
    trx_started,
    trx_requested_lock_id,
    trx_wait_started,
    trx_weight,
    trx_mysql_thread_id,
    trx_query
FROM INFORMATION_SCHEMA.INNODB_TRX 
WHERE trx_query LIKE '%sku_wh_attr%' OR trx_query LIKE '%YAFEXPPP0419972001%';

-- 2. 检查锁等待情况
SELECT 
    r.trx_id waiting_trx_id,
    r.trx_mysql_thread_id waiting_thread,
    r.trx_query waiting_query,
    b.trx_id blocking_trx_id,
    b.trx_mysql_thread_id blocking_thread,
    b.trx_query blocking_query
FROM INFORMATION_SCHEMA.INNODB_LOCK_WAITS w
INNER JOIN INFORMATION_SCHEMA.INNODB_TRX b ON b.trx_id = w.blocking_trx_id
INNER JOIN INFORMATION_SCHEMA.INNODB_TRX r ON r.trx_id = w.requesting_trx_id;

-- 3. 检查表的分区信息和数据分布
SELECT 
    PARTITION_NAME,
    TABLE_ROWS,
    AVG_ROW_LENGTH,
    DATA_LENGTH,
    INDEX_LENGTH
FROM INFORMATION_SCHEMA.PARTITIONS 
WHERE TABLE_SCHEMA = 'your_database_name' -- 请替换为实际数据库名
  AND TABLE_NAME = 'sku_wh_attr'
  AND PARTITION_NAME IS NOT NULL
ORDER BY PARTITION_NAME;

-- 4. 检查特定skuCode的分区位置
SELECT 
    'YAFEXPPP0419972001' as skuCode,
    MOD(CRC32('YAFEXPPP0419972001'), 250) as partition_number;

-- 5. 验证数据是否真实存在（不使用FOR UPDATE）
SELECT COUNT(*) as total_count
FROM sku_wh_attr 
WHERE skuCode = 'YAFEXPPP0419972001';

-- 6. 检查具体的记录详情
SELECT *
FROM sku_wh_attr 
WHERE skuCode = 'YAFEXPPP0419972001';

-- 7. 分析查询执行计划
EXPLAIN PARTITIONS 
SELECT `skuCode`, `banBuy`, `safety`, `taskQty`, `waitQty`, `prodQty`, `tranQty`, `signQty`, `putOnQty`, `exchange`, `refunds`, `damage`, `moreQty`, `cancel`, `avaiQty`, `rSeller`, `mSeller`, `whCode`, `occuQty`, `readyQty`
FROM `sku_wh_attr`
WHERE `skuCode` = 'YAFEXPPP0419972001';

EXPLAIN PARTITIONS 
SELECT `skuCode`, `whCode`, `tranQty`, `putOnQty` 
FROM `sku_wh_attr` 
WHERE `skuCode` = 'YAFEXPPP0419972001' AND `whCode` = 'D02';

-- 8. 检查MySQL版本和配置
SELECT @@version, @@transaction_isolation, @@innodb_lock_wait_timeout;

-- 9. 检查是否有长时间运行的事务
SELECT 
    id,
    user,
    host,
    db,
    command,
    time,
    state,
    info
FROM INFORMATION_SCHEMA.PROCESSLIST 
WHERE time > 60 OR state LIKE '%lock%'
ORDER BY time DESC;

-- 10. 测试在不同事务隔离级别下的查询
-- 注意：这会影响当前会话，测试后请恢复原设置
SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;
SELECT COUNT(*) as count_rc
FROM sku_wh_attr 
WHERE skuCode = 'YAFEXPPP0419972001';

SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;
SELECT COUNT(*) as count_rr
FROM sku_wh_attr 
WHERE skuCode = 'YAFEXPPP0419972001';