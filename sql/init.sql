-- MySQLer 数据库初始化脚本
-- 保证客户端/脚本以 utf8mb4 方式与服务端交互，避免中文乱码
SET NAMES utf8mb4;
SET character_set_client = utf8mb4;
-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS `mysqler` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `mysqler`;

-- 连接信息表
CREATE TABLE IF NOT EXISTS `connections` (
    `id` VARCHAR(36) NOT NULL PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL COMMENT '连接名称',
    `environment` ENUM('prod','test','dev') DEFAULT 'test' COMMENT '环境类型',
    `host` VARCHAR(255) NOT NULL COMMENT '主机地址',
    `port` INT DEFAULT 3306 COMMENT '端口号',
    `username` VARCHAR(100) NOT NULL COMMENT '用户名',
    `password` TEXT NOT NULL COMMENT '加密密码',
    `database_name` VARCHAR(100) NOT NULL COMMENT '数据库名',
    `description` TEXT COMMENT '描述信息',
    `connect_timeout` INT DEFAULT 5 COMMENT '连接超时(秒)',
    `charset` VARCHAR(20) DEFAULT 'utf8mb4' COMMENT '字符集',
    `use_ssl` BOOLEAN DEFAULT FALSE COMMENT '是否使用SSL',
    `created_by` VARCHAR(100) COMMENT '创建人',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_environment` (`environment`),
    INDEX `idx_name` (`name`)
) COMMENT='数据库连接配置表';

-- 执行记录表
CREATE TABLE IF NOT EXISTS `execution_records` (
    `id` VARCHAR(36) NOT NULL PRIMARY KEY,
    `connection_id` VARCHAR(36) NOT NULL COMMENT '连接ID',
    `table_name` VARCHAR(200) NOT NULL COMMENT '目标表名',
    `database_name` VARCHAR(100) NOT NULL COMMENT '数据库名',
    `ddl_type` ENUM('fragment','add_column','modify_column','drop_column','add_index','drop_index','other') COMMENT 'DDL类型',
    `original_ddl` TEXT COMMENT '原始DDL语句',
    `generated_command` TEXT NOT NULL COMMENT '生成的pt命令',
    `execution_params` JSON COMMENT '执行参数配置',
    `status` ENUM('pending','running','completed','failed','cancelled') DEFAULT 'pending',
    `start_time` TIMESTAMP NULL COMMENT '开始时间',
    `end_time` TIMESTAMP NULL COMMENT '结束时间',
    `duration_seconds` INT COMMENT '执行耗时(秒)',
    `processed_rows` BIGINT DEFAULT 0 COMMENT '已处理行数',
    `total_rows` BIGINT DEFAULT 0 COMMENT '总行数',
    `avg_speed` DECIMAL(10,2) COMMENT '平均处理速度(rows/sec)',
    `container_id` VARCHAR(64) COMMENT 'Docker容器ID',
    `execution_logs` LONGTEXT COMMENT '执行日志',
    `error_message` TEXT COMMENT '错误信息',
    `created_by` VARCHAR(100) COMMENT '执行人',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`connection_id`) REFERENCES `connections`(`id`) ON DELETE CASCADE,
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_table` (`database_name`, `table_name`),
    INDEX `idx_created_by` (`created_by`)
) COMMENT='DDL执行记录表';

-- 用户表
CREATE TABLE IF NOT EXISTS `users` (
    `id` VARCHAR(36) NOT NULL PRIMARY KEY,
    `username` VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名',
    `password_hash` VARCHAR(255) NOT NULL COMMENT '密码哈希',
    `display_name` VARCHAR(100) COMMENT '显示名称',
    `email` VARCHAR(255) COMMENT '邮箱',
    `role` ENUM('admin','operator','viewer') DEFAULT 'viewer' COMMENT '角色权限',
    `is_active` BOOLEAN DEFAULT TRUE COMMENT '是否激活',
    `last_login_at` TIMESTAMP NULL COMMENT '最后登录时间',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_username` (`username`),
    INDEX `idx_role` (`role`)
) COMMENT='用户表';

-- 操作审计表
CREATE TABLE IF NOT EXISTS `audit_logs` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `user_id` VARCHAR(36) COMMENT '用户ID',
    `username` VARCHAR(50) COMMENT '用户名',
    `action` VARCHAR(100) NOT NULL COMMENT '操作动作',
    `resource_type` VARCHAR(50) COMMENT '资源类型',
    `resource_id` VARCHAR(36) COMMENT '资源ID',
    `request_data` JSON COMMENT '请求数据',
    `response_data` JSON COMMENT '响应数据',
    `ip_address` VARCHAR(45) COMMENT '客户端IP',
    `user_agent` TEXT COMMENT '用户代理',
    `execution_time_ms` INT COMMENT '执行时间(毫秒)',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_action` (`action`),
    INDEX `idx_created_at` (`created_at`)
) COMMENT='操作审计日志表';

-- 插入默认管理员用户
-- 密码: admin123 (使用bcrypt加密)
INSERT IGNORE INTO `users` (`id`, `username`, `password_hash`, `display_name`, `email`, `role`, `is_active`) 
VALUES (
    'user-admin-001', 
    'admin', 
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 
    '系统管理员', 
    'admin@mysqler.local', 
    'admin', 
    TRUE
);

-- 插入测试连接数据
INSERT IGNORE INTO `connections` (`id`, `name`, `environment`, `host`, `port`, `username`, `password`, `database_name`, `description`, `created_by`) 
VALUES (
    'conn-test-001',
    '本地测试数据库',
    'dev',
    'localhost',
    3306,
    'root',
    'encrypted:root123456', -- 实际使用时需要加密
    'test_db',
    '用于开发测试的本地MySQL数据库',
    'admin'
);

-- 创建测试数据库和表（用于测试DDL操作）
CREATE DATABASE IF NOT EXISTS `test_db` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `test_db`;

-- 创建一个测试表
CREATE TABLE IF NOT EXISTS `test_users` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(50) NOT NULL,
    `email` VARCHAR(100),
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_username` (`username`)
) COMMENT='测试用户表';

-- 插入一些测试数据
INSERT IGNORE INTO `test_users` (`username`, `email`) VALUES
('test_user1', 'user1@test.com'),
('test_user2', 'user2@test.com'),
('test_user3', 'user3@test.com');

-- 切换回主数据库
USE `mysqler`;