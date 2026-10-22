-- 邀请裂变 · 增量建表脚本
-- MySQL 5.6 ~ 8.x 兼容（勿用 8.x 专有语法）
-- 执行：mysql --host=<HOST> --user=<USER> -p <DATABASE> < 002_invite.sql
--
-- 设计说明见 docs/06-微信登录与分享裂变调研.md 第六章。
-- 核心：绑定在注册时、结算在人工审核通过时 —— 这是合规的前提，
-- 被邀请人必须完成「注册 + 完善资料 + 通过审核」这段真实业务流程，
-- 才构成《微信小程序平台运营规范》认定的「深度互动」。

SET NAMES utf8mb4;

-- ----------------------------
-- 1. user 表增加权益到期时间
-- ----------------------------
-- 一期用「到期时间」而非独立会员表：避免引入订单/续费/退款一整套。
-- NULL 表示无权益；判定 vip_expire_at > NOW() 即生效。
-- 二期接微信支付时再演进为独立 user_membership 表。
ALTER TABLE `user`
  ADD COLUMN `vip_expire_at` datetime DEFAULT NULL COMMENT '权益到期时间，NULL 表示无权益';

-- ----------------------------
-- 2. 邀请码
-- ----------------------------
-- code 必须短：小程序码 scene 参数上限 32 字符，且只支持可见 ASCII 字符。
-- 一个用户恒定一个码，重复进入时只回显，不重复生成。
DROP TABLE IF EXISTS `invite_code`;
CREATE TABLE `invite_code` (
  `id`         bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id`    bigint unsigned NOT NULL COMMENT '归属用户',
  `code`       varchar(12)  NOT NULL COMMENT '短码，唯一',
  `created_at` datetime     DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  UNIQUE KEY `uk_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邀请码';

-- ----------------------------
-- 3. 邀请关系与结算
-- ----------------------------
-- invitee_id 上的唯一索引是防刷的第一道闸：一个人只能被邀请一次，
-- 从数据库层面杜绝「同一个人被多个邀请人重复计入」。
-- 关系在注册（或首次填码）时写入 status=0，审核通过时才置为 1 并发放权益。
DROP TABLE IF EXISTS `invite_relation`;
CREATE TABLE `invite_relation` (
  `id`         bigint unsigned NOT NULL AUTO_INCREMENT,
  `inviter_id` bigint unsigned NOT NULL COMMENT '邀请人',
  `invitee_id` bigint unsigned NOT NULL COMMENT '被邀请人',
  `code`       varchar(12)  NOT NULL DEFAULT '' COMMENT '使用的邀请码',
  `status`     tinyint      NOT NULL DEFAULT 0 COMMENT '0待达成 1已结算 2已作废',
  `settled_at` datetime     DEFAULT NULL COMMENT '结算时间',
  `created_at` datetime     DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_invitee` (`invitee_id`),
  KEY `idx_inviter_status` (`inviter_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邀请关系与结算';
