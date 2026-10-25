-- ============================================================
-- 005 短信验证码（H5 手机号登录）
--
-- 为什么落库而不是放内存：
--   ① 服务重启后验证码不失效；② 多实例部署时天然共享；
--   ③ 发送频次（手机号/IP）需要可审计的历史记录。
--
-- 只存 HMAC 哈希，不存明文。6 位数字空间小，哈希带 JWT secret 作 pepper，
-- 避免库文件泄露后被直接读出可用验证码。
--
-- 兼容 MySQL 5.6 ~ 8.x。
-- ============================================================

CREATE TABLE IF NOT EXISTS `sms_code` (
  `id`         bigint unsigned NOT NULL AUTO_INCREMENT,
  `phone`      varchar(20)  NOT NULL DEFAULT '' COMMENT '手机号',
  `code_hash`  varchar(64)  NOT NULL DEFAULT '' COMMENT '验证码 HMAC-SHA256',
  `scene`      varchar(16)  NOT NULL DEFAULT 'login' COMMENT '用途场景',
  `ip`         varchar(64)  NOT NULL DEFAULT '' COMMENT '请求 IP，用于限频',
  `attempts`   tinyint      NOT NULL DEFAULT 0  COMMENT '校验失败次数',
  `used_at`    datetime     DEFAULT NULL COMMENT '使用时间，NULL 表示未使用',
  `expires_at` datetime     NOT NULL COMMENT '过期时间',
  `created_at` datetime     DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_phone_scene_created` (`phone`, `scene`, `created_at`),
  KEY `idx_ip_created` (`ip`, `created_at`),
  KEY `idx_expires` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='短信验证码';
