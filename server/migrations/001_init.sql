-- 婚介小程序 初始化建表脚本
-- MySQL 8.4 / utf8mb4
-- 执行：mysql -uroot -proot123456 hunjie < 001_init.sql

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 1. 用户账号（含管理员，role 区分）
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id`             bigint unsigned NOT NULL AUTO_INCREMENT,
  `openid`         varchar(64)  NOT NULL DEFAULT ''  COMMENT '微信 openid',
  `phone`          varchar(20)  NOT NULL DEFAULT ''  COMMENT '手机号',
  `nickname`       varchar(32)  NOT NULL DEFAULT ''  COMMENT '昵称',
  `avatar`         varchar(255) NOT NULL DEFAULT ''  COMMENT '头像 URL',
  `gender`         tinyint      NOT NULL DEFAULT 0   COMMENT '1男 2女，注册后不可改',
  `role`           tinyint      NOT NULL DEFAULT 1   COMMENT '1普通 2审核员 3管理员 4超管',
  `status`         tinyint      NOT NULL DEFAULT 0   COMMENT '0待完善 1待审核 2正常 3已拒绝 4已封禁',
  `ban_reason`     varchar(255) NOT NULL DEFAULT ''  COMMENT '封禁原因',
  `last_active_at` datetime     DEFAULT NULL          COMMENT '最后活跃时间，推荐流排序依据',
  `created_at`     datetime     DEFAULT NULL,
  `updated_at`     datetime     DEFAULT NULL,
  `deleted_at`     datetime     DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_openid` (`openid`),
  KEY `idx_status_gender` (`status`, `gender`),
  KEY `idx_last_active` (`last_active_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户账号';

-- ----------------------------
-- 2. 用户详细资料
-- ----------------------------
DROP TABLE IF EXISTS `user_profile`;
CREATE TABLE `user_profile` (
  `id`             bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id`        bigint unsigned NOT NULL,
  `birthday`       date         DEFAULT NULL,
  `age`            tinyint      NOT NULL DEFAULT 0   COMMENT '冗余字段，保存资料时计算',
  `height`         smallint     NOT NULL DEFAULT 0   COMMENT '身高 cm',
  `weight`         smallint     NOT NULL DEFAULT 0   COMMENT '体重 kg，选填',
  `marital_status` tinyint      NOT NULL DEFAULT 0   COMMENT '1未婚 2离异 3丧偶',
  `has_children`   tinyint      NOT NULL DEFAULT 0   COMMENT '0无 1有同住 2有不同住',
  `education`      tinyint      NOT NULL DEFAULT 0   COMMENT '1高中及以下 2大专 3本科 4硕士 5博士',
  `school`         varchar(64)  NOT NULL DEFAULT ''  COMMENT '毕业院校，选填',
  `occupation`     varchar(64)  NOT NULL DEFAULT ''  COMMENT '职业',
  `income_level`   tinyint      NOT NULL DEFAULT 0   COMMENT '1(<5k) 2(5-10k) 3(10-20k) 4(20-30k) 5(30-50k) 6(50k+)',
  `house_status`   tinyint      NOT NULL DEFAULT 0   COMMENT '1已购房 2与父母同住 3租房 4计划购房',
  `car_status`     tinyint      NOT NULL DEFAULT 0   COMMENT '1已购车 2暂未购车',
  `province`       varchar(32)  NOT NULL DEFAULT ''  COMMENT '现居省',
  `city`           varchar(32)  NOT NULL DEFAULT ''  COMMENT '现居市',
  `city_code`      varchar(16)  NOT NULL DEFAULT ''  COMMENT '现居城市编码，筛选用',
  `home_province`  varchar(32)  NOT NULL DEFAULT ''  COMMENT '户籍省',
  `home_city`      varchar(32)  NOT NULL DEFAULT ''  COMMENT '户籍市',
  `self_intro`     varchar(1000) NOT NULL DEFAULT '' COMMENT '自我介绍',
  `complete_rate`  tinyint      NOT NULL DEFAULT 0   COMMENT '资料完整度 0-100',
  `audit_status`   tinyint      NOT NULL DEFAULT 0   COMMENT '0待审核 1通过 2拒绝',
  `audit_remark`   varchar(255) NOT NULL DEFAULT ''  COMMENT '拒绝原因',
  `audited_at`     datetime     DEFAULT NULL,
  `created_at`     datetime     DEFAULT NULL,
  `updated_at`     datetime     DEFAULT NULL,
  `deleted_at`     datetime     DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_id` (`user_id`),
  KEY `idx_audit_status` (`audit_status`),
  KEY `idx_filter` (`audit_status`, `age`, `education`, `income_level`),
  KEY `idx_city_code` (`city_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户详细资料';

-- ----------------------------
-- 3. 择偶要求
-- ----------------------------
DROP TABLE IF EXISTS `user_requirement`;
CREATE TABLE `user_requirement` (
  `id`              bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id`         bigint unsigned NOT NULL,
  `age_min`         tinyint      NOT NULL DEFAULT 0,
  `age_max`         tinyint      NOT NULL DEFAULT 0,
  `height_min`      smallint     NOT NULL DEFAULT 0,
  `height_max`      smallint     NOT NULL DEFAULT 0,
  `education_min`   tinyint      NOT NULL DEFAULT 0  COMMENT '学历下限',
  `income_min`      tinyint      NOT NULL DEFAULT 0  COMMENT '收入下限',
  `marital_status`  varchar(32)  NOT NULL DEFAULT '' COMMENT '可接受婚史，逗号分隔 "1,2"',
  `accept_children` tinyint      NOT NULL DEFAULT 1  COMMENT '0不接受 1接受',
  `city_codes`      varchar(255) NOT NULL DEFAULT '' COMMENT '期望城市，逗号分隔',
  `other`           varchar(500) NOT NULL DEFAULT '' COMMENT '其他要求',
  `created_at`      datetime     DEFAULT NULL,
  `updated_at`      datetime     DEFAULT NULL,
  `deleted_at`      datetime     DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='择偶要求';

-- ----------------------------
-- 4. 相册
-- ----------------------------
DROP TABLE IF EXISTS `user_photo`;
CREATE TABLE `user_photo` (
  `id`           bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id`      bigint unsigned NOT NULL,
  `url`          varchar(255) NOT NULL DEFAULT '',
  `sort`         smallint     NOT NULL DEFAULT 0,
  `audit_status` tinyint      NOT NULL DEFAULT 0  COMMENT '0待审 1通过 2拒绝',
  `created_at`   datetime     DEFAULT NULL,
  `updated_at`   datetime     DEFAULT NULL,
  `deleted_at`   datetime     DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_audit_status` (`audit_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='相册';

-- ----------------------------
-- 5. 互动（浏览/喜欢/不喜欢）
-- ----------------------------
DROP TABLE IF EXISTS `interaction`;
CREATE TABLE `interaction` (
  `id`           bigint unsigned NOT NULL AUTO_INCREMENT,
  `from_user_id` bigint unsigned NOT NULL,
  `to_user_id`   bigint unsigned NOT NULL,
  `type`         tinyint  NOT NULL DEFAULT 0 COMMENT '1浏览 2喜欢 3不喜欢',
  `is_read`      tinyint  NOT NULL DEFAULT 0 COMMENT '接收方是否已读',
  `created_at`   datetime DEFAULT NULL,
  `updated_at`   datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_from_to_type` (`from_user_id`, `to_user_id`, `type`),
  KEY `idx_to_type` (`to_user_id`, `type`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='互动记录';

-- ----------------------------
-- 6. 会话（= 配对关系）
-- ----------------------------
DROP TABLE IF EXISTS `conversation`;
CREATE TABLE `conversation` (
  `id`              bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_a_id`       bigint unsigned NOT NULL COMMENT '约定：较小的 user_id',
  `user_b_id`       bigint unsigned NOT NULL,
  `last_message`    varchar(255) NOT NULL DEFAULT '' COMMENT '最后一条消息摘要',
  `last_message_at` datetime     DEFAULT NULL,
  `a_unread`        int          NOT NULL DEFAULT 0,
  `b_unread`        int          NOT NULL DEFAULT 0,
  `status`          tinyint      NOT NULL DEFAULT 1 COMMENT '1正常 2已解除',
  `matched_at`      datetime     DEFAULT NULL,
  `created_at`      datetime     DEFAULT NULL,
  `updated_at`      datetime     DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pair` (`user_a_id`, `user_b_id`),
  KEY `idx_a_time` (`user_a_id`, `last_message_at`),
  KEY `idx_b_time` (`user_b_id`, `last_message_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会话/配对';

-- ----------------------------
-- 7. 消息
-- ----------------------------
DROP TABLE IF EXISTS `message`;
CREATE TABLE `message` (
  `id`              bigint unsigned NOT NULL AUTO_INCREMENT,
  `conversation_id` bigint unsigned NOT NULL,
  `from_user_id`    bigint unsigned NOT NULL,
  `type`            tinyint       NOT NULL DEFAULT 1 COMMENT '1文字 2图片 3系统',
  `content`         varchar(1000) NOT NULL DEFAULT '',
  `is_read`         tinyint       NOT NULL DEFAULT 0,
  `created_at`      datetime      DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_conv_id` (`conversation_id`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息';

-- ----------------------------
-- 8. 黑名单
-- ----------------------------
DROP TABLE IF EXISTS `blacklist`;
CREATE TABLE `blacklist` (
  `id`             bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id`        bigint unsigned NOT NULL COMMENT '拉黑发起人',
  `target_user_id` bigint unsigned NOT NULL,
  `created_at`     datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_target` (`user_id`, `target_user_id`),
  KEY `idx_target` (`target_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='黑名单';

-- ----------------------------
-- 9. 举报
-- ----------------------------
DROP TABLE IF EXISTS `report`;
CREATE TABLE `report` (
  `id`             bigint unsigned NOT NULL AUTO_INCREMENT,
  `from_user_id`   bigint unsigned NOT NULL,
  `target_user_id` bigint unsigned NOT NULL,
  `reason`         tinyint       NOT NULL DEFAULT 0 COMMENT '1虚假资料 2骚扰 3广告 4违法 5其他',
  `description`    varchar(500)  NOT NULL DEFAULT '',
  `images`         varchar(1000) NOT NULL DEFAULT '' COMMENT '截图 URL，逗号分隔',
  `status`         tinyint       NOT NULL DEFAULT 0 COMMENT '0待处理 1已处理 2已驳回',
  `handle_remark`  varchar(255)  NOT NULL DEFAULT '',
  `created_at`     datetime      DEFAULT NULL,
  `updated_at`     datetime      DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_target` (`target_user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='举报';

SET FOREIGN_KEY_CHECKS = 1;
