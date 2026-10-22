-- ============================================================
-- 003 权益体系重构：永久权益 + 原生家庭 + 行业标签 + 三观问答
--
-- 兼容 MySQL 5.6 ~ 8.x（勿用 8.0 才有的语法，云库是 5.6）
-- 幂等性说明：MySQL 5.6 不支持 ADD COLUMN IF NOT EXISTS，
-- 重复执行会报 1060 Duplicate column name，属预期，忽略即可。
-- ============================================================

-- ------------------------------------------------------------
-- 1. 权益等级：把「到期时间」换成「永久等级」
--
-- 为什么不再用 vip_expire_at：
--   旧机制是「邀请成功 → 双方各得 7 天」，用户看到「剩 7 天」却没有心理锚点
--   （付费产品用天数是因为能换算成钱，赠送权益没有这层换算），
--   且倒计时天然带压力感，与本项目「低压决策」的婚恋场景相悖。
--   改为永久解锁后：用户不需要理解任何规则，也不会因到期引发咨询工单。
--
-- 防刷不再靠时间，而靠「邀请审核制」—— 被邀请人必须通过人工审核才结算，
-- 成本已经足够高（详见 docs/08-权益体系重构方案.md 第四章）。
--
-- vip_expire_at 保留不删：历史数据与既有代码路径兼容，读取以 vip_level 为准。
-- ------------------------------------------------------------
ALTER TABLE `user`
  ADD COLUMN `vip_level` tinyint NOT NULL DEFAULT 0 COMMENT '权益等级 0无 1已解锁完整资料（永久）';

-- 历史数据迁移：凡是有未过期权益的，视为永久权益
UPDATE `user`
SET `vip_level` = 1
WHERE `vip_expire_at` IS NOT NULL AND `vip_expire_at` > NOW();

-- ------------------------------------------------------------
-- 2. 资料表新增：行业标签（免费） + 原生家庭（权益内容，默认私有）
--
-- 行业标签是「粗粒度」的（教育系统，而非「XX市第三中学」），
-- 理由是防止组合信息反查到具体的人 —— 体制内用户对可被识别的信息极敏感。
-- ------------------------------------------------------------
ALTER TABLE `user_profile`
  ADD COLUMN `industry`                tinyint      NOT NULL DEFAULT 0 COMMENT '行业标签(粗粒度) 0未填',
  ADD COLUMN `family_parent_occ`       tinyint      NOT NULL DEFAULT 0 COMMENT '父母职业类型 0未填',
  ADD COLUMN `family_atmosphere`       tinyint      NOT NULL DEFAULT 0 COMMENT '家庭氛围 0未填',
  ADD COLUMN `family_only_child`       tinyint      NOT NULL DEFAULT 0 COMMENT '是否独生子女 0未填 1是 2否',
  ADD COLUMN `family_city`             varchar(32)  NOT NULL DEFAULT '' COMMENT '家庭所在地(粗粒度，仅到市级)',
  ADD COLUMN `family_note`             varchar(200) NOT NULL DEFAULT '' COMMENT '家庭补充说明',
  ADD COLUMN `family_visible`          tinyint      NOT NULL DEFAULT 0 COMMENT '是否对权益用户展示原生家庭 0否(默认) 1是';

-- 索引：行业筛选是高频操作，且与状态/性别组合查询
ALTER TABLE `user_profile`
  ADD KEY `idx_industry` (`industry`);

-- ------------------------------------------------------------
-- 3. 三观问答答案表
--
-- 设计要点（见方案 4.2②）：
--   · 答题免费 —— 设成权益会没人填，而题库的价值恰恰在于让所有人都填
--   · 「看到对方的答案」才需权益
--   · 题目本身放代码常量（model.ThreeViewQuestions），不建题目表：
--     题目一年不会改一次，建表等于给自己加一套 CRUD
--
-- uk_user_question 保证一题一答；重复提交按覆盖处理
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `three_view_answer` (
  `id`         bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id`    bigint unsigned NOT NULL,
  `q_id`       smallint        NOT NULL COMMENT '题目编号，对应 model.ThreeViewQuestions 的 ID',
  `option_idx` tinyint         NOT NULL COMMENT '所选选项下标(从0开始)',
  `created_at` datetime        DEFAULT NULL,
  `updated_at` datetime        DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_question` (`user_id`, `q_id`),
  KEY `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='三观问答答案';
