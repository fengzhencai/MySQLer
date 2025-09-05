-- 紧急MySQL数据一致性问题排查脚本
-- 针对 sku_wh_attr 表的查询结果不一致问题

-- ==================== 第一阶段：立即诊断 ====================

-- 1. 检查当前活跃的事务（重要！）
SELECT 
    trx_id,
    trx_state,
    trx_started,
    TIMESTAMPDIFF(SECOND, trx_started, NOW()) as duration_seconds,
    trx_mysql_thread_id,
    trx_query,
    trx_operation_state,
    trx_tables_in_use,
    trx_tables_locked,
    trx_lock_structs,
    trx_rows_locked
FROM INFORMATION_SCHEMA.INNODB_TRX 
ORDER BY trx_started;

-- 2. 检查锁等待情况（关键诊断）
SELECT 
    r.trx_id as waiting_trx_id,
    r.trx_mysql_thread_id as waiting_thread,
    r.trx_query as waiting_query,
    b.trx_id as blocking_trx_id,
    b.trx_mysql_thread_id as blocking_thread,
    b.trx_query as blocking_query,
    TIMESTAMPDIFF(SECOND, r.trx_wait_started, NOW()) as wait_seconds
FROM INFORMATION_SCHEMA.INNODB_LOCK_WAITS w
INNER JOIN INFORMATION_SCHEMA.INNODB_TRX b ON b.trx_id = w.blocking_trx_id
INNER JOIN INFORMATION_SCHEMA.INNODB_TRX r ON r.trx_id = w.requesting_trx_id;

-- 3. 立即测试问题查询（在新会话中执行）
-- 测试1：原问题查询（不要在生产环境的事务中执行）
-- SELECT `skuCode`, `banBuy`, `safety`, `taskQty`, `waitQty`, `prodQty`, `tranQty`, `signQty`, `putOnQty`, `exchange`, `refunds`, `damage`, `moreQty`, `cancel`, `avaiQty`, `rSeller`, `mSeller`, `whCode`, `occuQty`, `readyQty`
-- FROM `sku_wh_attr`
-- WHERE `skuCode` = 'YAFEXPPP0419972001'
-- FOR UPDATE;

-- 测试2：不带FOR UPDATE的查询
SELECT COUNT(*) as without_for_update_count
FROM `sku_wh_attr`
WHERE `skuCode` = 'YAFEXPPP0419972001';

-- 测试3：查看所有相关记录
SELECT `skuCode`, `whCode`, `avaiQty`, `occuQty`
FROM `sku_wh_attr`
WHERE `skuCode` = 'YAFEXPPP0419972001';

-- ==================== 第二阶段：执行计划分析 ====================

-- 4. 比较执行计划
EXPLAIN FORMAT=JSON
SELECT `skuCode`, `banBuy`, `safety`, `taskQty`, `waitQty`, `prodQty`, `tranQty`, `signQty`, `putOnQty`, `exchange`, `refunds`, `damage`, `moreQty`, `cancel`, `avaiQty`, `rSeller`, `mSeller`, `whCode`, `occuQty`, `readyQty`
FROM `sku_wh_attr`
WHERE `skuCode` = 'YAFEXPPP0419972001';

EXPLAIN FORMAT=JSON
SELECT `skuCode`, `whCode`, `tranQty`, `putOnQty` 
FROM `sku_wh_attr` 
WHERE `skuCode` = 'YAFEXPPP0419972001' AND `whCode` = 'D02';

-- 5. 检查分区裁剪情况
EXPLAIN PARTITIONS
SELECT `skuCode`, `banBuy`, `safety`, `taskQty`, `waitQty`, `prodQty`, `tranQty`, `signQty`, `putOnQty`, `exchange`, `refunds`, `damage`, `moreQty`, `cancel`, `avaiQty`, `rSeller`, `mSeller`, `whCode`, `occuQty`, `readyQty`
FROM `sku_wh_attr`
WHERE `skuCode` = 'YAFEXPPP0419972001';

-- ==================== 第三阶段：紧急修复测试 ====================

-- 6. 测试不同的查询方式
-- 方式A：使用FORCE INDEX
SELECT `skuCode`, `banBuy`, `safety`, `taskQty`, `waitQty`, `prodQty`, `tranQty`, `signQty`, `putOnQty`, `exchange`, `refunds`, `damage`, `moreQty`, `cancel`, `avaiQty`, `rSeller`, `mSeller`, `whCode`, `occuQty`, `readyQty`
FROM `sku_wh_attr` FORCE INDEX (PRIMARY)
WHERE `skuCode` = 'YAFEXPPP0419972001';

-- 方式B：使用NOWAIT（避免锁等待）
-- 注意：在新会话中测试，不要在业务事务中执行
-- SELECT `skuCode`, `banBuy`, `safety`, `taskQty`, `waitQty`, `prodQty`, `tranQty`, `signQty`, `putOnQty`, `exchange`, `refunds`, `damage`, `moreQty`, `cancel`, `avaiQty`, `rSeller`, `mSeller`, `whCode`, `occuQty`, `readyQty`
-- FROM `sku_wh_attr`
-- WHERE `skuCode` = 'YAFEXPPP0419972001'
-- FOR UPDATE NOWAIT;

-- ==================== 第四阶段：系统状态检查 ====================

-- 7. 检查MySQL系统状态
SHOW STATUS LIKE 'Innodb_rows_locked';
SHOW STATUS LIKE 'Innodb_lock_wait%';
SHOW STATUS LIKE 'Innodb_deadlocks';

-- 8. 检查表状态
SHOW TABLE STATUS LIKE 'sku_wh_attr';

-- 9. 检查是否有表损坏
CHECK TABLE sku_wh_attr;

-- ==================== 第五阶段：配置检查 ====================

-- 10. 关键配置参数
SELECT 
    @@innodb_lock_wait_timeout as lock_wait_timeout,
    @@transaction_isolation as isolation_level,
    @@innodb_deadlock_detect as deadlock_detect,
    @@innodb_print_all_deadlocks as print_deadlocks;

-- ==================== 紧急解决方案测试 ====================

-- 11. 临时解决方案：改写查询
-- 解决方案1：分步查询（推荐用于紧急修复）
-- 第一步：获取所有whCode
SELECT DISTINCT whCode 
FROM sku_wh_attr 
WHERE skuCode = 'YAFEXPPP0419972001';

-- 第二步：针对每个whCode单独查询并锁定
-- SELECT `skuCode`, `banBuy`, `safety`, `taskQty`, `waitQty`, `prodQty`, `tranQty`, `signQty`, `putOnQty`, `exchange`, `refunds`, `damage`, `moreQty`, `cancel`, `avaiQty`, `rSeller`, `mSeller`, `whCode`, `occuQty`, `readyQty`
-- FROM `sku_wh_attr`
-- WHERE `skuCode` = 'YAFEXPPP0419972001' AND `whCode` = 'D02'  -- 替换为实际的whCode
-- FOR UPDATE;