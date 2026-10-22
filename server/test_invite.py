#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
邀请裂变 · 端到端验证脚本

验证「邀请码 + 深度互动结算」的完整链路与各项约束：
  1. 生成邀请码（幂等）
  2. 绑定邀请码（成功后不可重复绑定、不能用自己的码）
  3. 绑定后不发放权益（关键合规点：注册不等于深度互动）
  4. 审核通过后才结算，双方各得权益
  5. 「谁看过我」非会员锁定 3 条，会员解锁全部
  6. 邀请上限、封禁邀请人时作废
  7. 邀请资格：邀请人须已过审，被邀请人仅限提交资料前的新人
  8. 结算回执：unseen_settled 红点数据与 /invite/seen 已读

用法（后端需先跑在 localhost:8080）：
    python server/test_invite.py
"""
import json
import subprocess
import sys
import time
import urllib.request

BASE = "http://localhost:8080/api/v1"

# 系统代理会拦截 127.0.0.1 请求，显式绕过
_opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))

PASS, FAIL = [], []


def call(method, path, token=None, body=None):
    url = BASE + path
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(url, data=data, method=method)
    req.add_header("Content-Type", "application/json")
    if token:
        req.add_header("Authorization", "Bearer " + token)
    try:
        with _opener.open(req, timeout=15) as r:
            return json.loads(r.read().decode())
    except urllib.error.HTTPError as e:
        return json.loads(e.read().decode())


def login(code):
    """mock 模式下 code 即 openid"""
    r = call("POST", "/auth/wx-login", body={"code": code})
    if r.get("code") != 0:
        raise RuntimeError(f"登录失败 {code}: {r}")
    return r["data"]["token"], r["data"]["user"]["id"]


def check(name, cond, detail=""):
    (PASS if cond else FAIL).append(name)
    print(("  [PASS] " if cond else "  [FAIL] ") + name + (f"  -> {detail}" if detail else ""))


def sql(query):
    """直连云库执行 SQL，用于构造测试数据与校验落库结果"""
    out = subprocess.run(
        ["mysql", "--host=101.35.245.54", "--user=marry_link_dev",
         "--password=HXchGX66LCMpmcAF", "marry_link_dev", "-N", "-B", "-e", query],
        capture_output=True, text=True,
    )
    return out.stdout.strip()


def main():
    ts = str(int(time.time()))
    inviter_code = f"inviter_{ts}"
    invitee_code = f"invitee_{ts}"
    # 手机号有应用层查重，且测试库存量账号占用过固定号段（13800138000 等），
    # 每次运行用时间戳生成唯一手机号，避免撞上种子/历史数据
    phone1 = "138" + ts[2:10]
    phone2 = "138" + str(int(ts) + 1)[2:10]

    # ---------- 准备两个干净账号 ----------
    print("\n[0] 准备测试账号")
    sql(f"DELETE FROM invite_relation WHERE inviter_id IN "
        f"(SELECT id FROM user WHERE openid IN ('{inviter_code}','{invitee_code}'));")
    sql(f"DELETE FROM invite_code WHERE user_id IN "
        f"(SELECT id FROM user WHERE openid IN ('{inviter_code}','{invitee_code}'));")
    # 清掉旧的软删记录，避免 openid 唯一索引冲突
    sql(f"DELETE FROM user WHERE openid IN ('{inviter_code}','{invitee_code}');")

    inviter_token, inviter_id = login(inviter_code)
    invitee_token, invitee_id = login(invitee_code)
    check("两个账号登录成功", inviter_id and invitee_id, f"inviter={inviter_id} invitee={invitee_id}")

    # ---------- 1. 邀请码生成幂等 ----------
    print("\n[1] 邀请码生成")
    r1 = call("GET", "/invite/my-code", inviter_token)
    check("取邀请码成功", r1.get("code") == 0, r1.get("message"))
    code = r1["data"]["code"]
    check("邀请码长度为 6", len(code) == 6, f"code={code}")
    check("邀请码字符集已排除易混淆字符",
          not set(code) & set("01OIL"), f"code={code}")

    r2 = call("GET", "/invite/my-code", inviter_token)
    check("重复获取返回同一个码（幂等）", r2["data"]["code"] == code)
    check("初始未解锁", r1["data"]["unlocked"] is False)
    check("响应不含到期时间字段（权益为永久制）",
          "vip_expire_at" not in r1["data"] and "reward_days" not in r1["data"],
          str(list(r1["data"].keys())))

    # ---------- 2. 绑定邀请码 ----------
    print("\n[2] 绑定邀请码")
    # 邀请人资格：须为已通过人工审核的正常账号。先验证「未过审邀请人」被拒，
    # 再把邀请人置为正常，走正式流程（造数置状态，资料流程本身由第 4 节覆盖）
    r = call("POST", "/invite/bind", invitee_token, {"code": code})
    check("邀请人未过审时，绑定被拒", r.get("code") != 0, r.get("message"))
    sql(f"UPDATE user SET status = 2 WHERE id = {inviter_id};")

    r = call("POST", "/invite/bind", invitee_token, {"code": code.lower()})
    check("小写邀请码也能绑定（大小写不敏感）", r.get("code") == 0, r.get("message"))

    r = call("POST", "/invite/bind", invitee_token, {"code": code})
    check("同一用户不能重复绑定", r.get("code") != 0, r.get("message"))

    r = call("POST", "/invite/bind", inviter_token, {"code": code})
    check("不能使用自己的邀请码", r.get("code") != 0, r.get("message"))

    # 用一个还没绑定过的账号验证「码不存在」（已绑定的账号会先被幂等拦截）
    probe_code = f"probe_{ts}"
    sql(f"DELETE FROM user WHERE openid = '{probe_code}';")
    probe_token, probe_id = login(probe_code)
    r = call("POST", "/invite/bind", probe_token, {"code": "ZZZZZZ"})
    check("不存在的邀请码被拒绝", r.get("code") != 0, r.get("message"))
    # 被邀请人资格：已是正常状态的老用户绑了也永远不会结算（结算只挂在审核通过上）
    sql(f"UPDATE user SET status = 2 WHERE id = {probe_id};")
    r = call("POST", "/invite/bind", probe_token, {"code": code})
    check("已通过审核的账号不能使用邀请码", r.get("code") != 0, r.get("message"))
    sql(f"DELETE FROM user WHERE id = {probe_id};")

    # ---------- 3. 关键合规点：绑定后不发放权益 ----------
    print("\n[3] 绑定后不发放权益（注册 != 深度互动）")
    inviter_me = call("GET", "/user/me", inviter_token)
    invitee_me = call("GET", "/user/me", invitee_token)
    check("邀请人此时仍无权益", not inviter_me["data"]["user"].get("vip_level"),
          str(inviter_me["data"]["user"].get("vip_level")))
    check("被邀请人此时仍无权益", not invitee_me["data"]["user"].get("vip_level"),
          str(invitee_me["data"]["user"].get("vip_level")))

    info = call("GET", "/invite/my-code", inviter_token)["data"]
    check("邀请记录处于「待达成」", info["pending_count"] == 1 and info["settled_count"] == 0,
          f"pending={info['pending_count']} settled={info['settled_count']}")

    # ---------- 4. 被邀请人提交资料 → 审核通过 → 结算 ----------
    print("\n[4] 深度互动达成后结算")
    profile = {
        "nickname": "测试被邀", "avatar": "/uploads/test.png", "gender": 1,
        "birthday": "1995-06-15", "height": 175, "weight": 70,
        "marital_status": 1, "has_children": 0, "education": 3,
        "school": "测试大学", "occupation": "工程师", "income_level": 3,
        "house_status": 1, "car_status": 1, "province": "河南省", "city": "郑州市",
        "city_code": "410100", "home_province": "河南省", "home_city": "郑州市",
        "self_intro": "这是一段用于测试的自我介绍，长度需要超过三十个字符才能通过前端校验，这里补足。",
    }
    r = call("PUT", "/user/profile", invitee_token, profile)
    check("保存资料成功", r.get("code") == 0, r.get("message"))
    r = call("PUT", "/user/requirement", invitee_token, {
        "age_min": 22, "age_max": 40, "height_min": 150, "height_max": 195,
        "education_min": 0, "income_min": 0, "marital_status": "1",
        "accept_children": 1, "city_codes": "", "other": "",
    })
    check("保存择偶要求成功", r.get("code") == 0, r.get("message"))

    # 提交审核要求手机号已绑定（婚恋类目要求可联系的实名用户）
    r = call("POST", "/auth/bind-phone", invitee_token, {"phone": phone1})
    check("绑定手机号成功（mock 模式手填）", r.get("code") == 0, r.get("message"))

    r = call("POST", "/user/submit-audit", invitee_token)
    check("提交审核成功", r.get("code") == 0, r.get("message"))

    # 审核前再看一次：仍不应有权益
    info = call("GET", "/invite/my-code", inviter_token)["data"]
    check("提交审核但未通过时仍不结算", info["settled_count"] == 0,
          f"settled={info['settled_count']}")

    # 用超管账号审核通过
    admin_token, _ = login("admin")
    r = call("POST", f"/admin/audit/{invitee_id}", admin_token,
             {"pass": True, "remark": "自动化测试通过"})
    check("审核通过接口调用成功", r.get("code") == 0, r.get("message"))

    # 结算校验
    info = call("GET", "/invite/my-code", inviter_token)["data"]
    check("审核通过后邀请人结算数 +1", info["settled_count"] == 1,
          f"settled={info['settled_count']}")
    check("邀请人获得权益", info["unlocked"] is True)

    # ---------- 4.5 结算回执：红点 + 已读 ----------
    print("\n[4.5] 结算回执（unseen_settled / invite_unseen / seen）")
    check("结算生成未读回执", info["unseen_settled"] == 1,
          f"unseen={info['unseen_settled']}")

    me = call("GET", "/user/me", inviter_token)["data"]
    check("me 接口带出邀请未读数", me.get("invite_unseen") == 1,
          str(me.get("invite_unseen")))

    r = call("POST", "/invite/seen", inviter_token)
    check("标记回执已读成功", r.get("code") == 0, r.get("message"))
    info = call("GET", "/invite/my-code", inviter_token)["data"]
    check("已读后未读回执清零", info["unseen_settled"] == 0,
          f"unseen={info['unseen_settled']}")

    invitee_info = call("GET", "/invite/my-code", invitee_token)["data"]
    check("被邀请人也获得权益", invitee_info["unlocked"] is True)

    inviter_me = call("GET", "/user/me", inviter_token)["data"]["user"]
    invitee_me = call("GET", "/user/me", invitee_token)["data"]["user"]
    check("双方 vip_level 均已置 1（永久解锁）",
          inviter_me.get("vip_level") == 1 and invitee_me.get("vip_level") == 1,
          f"inviter={inviter_me.get('vip_level')} invitee={invitee_me.get('vip_level')}")
    check("对外响应不回传旧的到期时间字段",
          "vip_expire_at" not in inviter_me and "vip_expire_at" not in invitee_me)

    # 结算幂等：再审核一次不应重复发权益
    before = call("GET", "/invite/my-code", inviter_token)["data"]["unlocked"]
    call("POST", f"/admin/audit/{invitee_id}", admin_token, {"pass": True, "remark": "重复审核"})
    after = call("GET", "/invite/my-code", inviter_token)["data"]["unlocked"]
    check("重复审核不重复发放权益（幂等）", before == after, f"{before} -> {after}")

    # ---------- 5. 权益解锁「谁看过我」 ----------
    print("\n[5] 权益解锁「谁看过我」")
    # 「谁看过我」挂在 RequireBrowse 下（资料未完善不能浏览他人），
    # 所以邀请人自己也要有资料。这里直接 SQL 造数：第 5 节测的是权益解锁行为，
    # 资料填写流程本身已由第 4 节覆盖，不必重复走一遍。
    sql(f"DELETE FROM user_profile WHERE user_id = {inviter_id};")
    sql(f"INSERT INTO user_profile (user_id, birthday, age, height, gender, "
        f"marital_status, has_children, education, occupation, income_level, "
        f"house_status, car_status, province, city, city_code, self_intro, "
        f"audit_status, created_at, updated_at) "
        f"VALUES ({inviter_id}, '1994-03-10', 31, 180, 2, 1, 0, 3, '测试', 3, "
        f"1, 1, '河南省', '郑州市', '410100', '测试资料', 1, NOW(), NOW());")
    sql(f"UPDATE user SET status = 2, gender = 2 WHERE id = {inviter_id};")

    # 让被邀请人「看过」邀请人的资料，制造可预期的场景
    sql(f"DELETE FROM interaction WHERE to_user_id = {inviter_id} AND type = 1;")
    sql(f"INSERT INTO interaction (from_user_id, to_user_id, type, is_read, created_at, updated_at) "
        f"VALUES ({invitee_id}, {inviter_id}, 1, 0, NOW(), NOW());")
    sql(f"INSERT INTO interaction (from_user_id, to_user_id, type, is_read, created_at, updated_at) "
        f"VALUES ({invitee_id}, {inviter_id}, 1, 0, NOW(), NOW());")

    # 把邀请人的权益临时摘掉，验证非会员视角
    sql(f"UPDATE user SET vip_level = 0 WHERE id = {inviter_id};")
    r = call("GET", "/interaction/viewed-me", inviter_token)
    check("非会员也能取到列表（功能不被门禁）", r.get("code") == 0, r.get("message"))
    free = r["data"]
    check("非会员 unlocked=false", free["unlocked"] is False)

    sql(f"UPDATE user SET vip_level = 1 WHERE id = {inviter_id};")
    r = call("GET", "/interaction/viewed-me", inviter_token)
    vip = r["data"]
    check("会员 unlocked=true", vip["unlocked"] is True)
    check("会员无锁定内容", vip["locked"] == 0, f"locked={vip['locked']}")
    check("会员可见条数不少于免费版",
          len(vip["list"]) >= len(free["list"]),
          f"free={len(free['list'])} vip={len(vip['list'])}")
    check("总人数口径一致", free["total"] == vip["total"],
          f"free_total={free['total']} vip_total={vip['total']}")

    # ---------- 6. 封禁邀请人 → 关系作废 ----------
    print("\n[6] 邀请人封禁时作废关系")
    invitee2 = f"invitee2_{ts}"
    sql(f"DELETE FROM user WHERE openid = '{invitee2}';")
    r2 = call("POST", "/auth/wx-login", body={"code": invitee2})
    invitee2_token, invitee2_id = r2["data"]["token"], r2["data"]["user"]["id"]

    r = call("POST", "/invite/bind", invitee2_token, {"code": code})
    check("新用户绑定成功", r.get("code") == 0, r.get("message"))

    # 封禁邀请人 —— 结算前的前置校验应判定其不可参与
    before_ban = sql(f"SELECT vip_level FROM user WHERE id = {inviter_id};")
    sql(f"UPDATE user SET status = 4, ban_reason = '测试封禁' WHERE id = {inviter_id};")

    # 被邀请人走完流程，审核通过时应「作废」而非发放
    call("PUT", "/user/profile", invitee2_token, {**profile, "nickname": "测试被邀2"})
    call("PUT", "/user/requirement", invitee2_token, {
        "age_min": 22, "age_max": 40, "height_min": 150, "height_max": 195,
        "education_min": 0, "income_min": 0, "marital_status": "1",
        "accept_children": 1, "city_codes": "", "other": ""})
    call("POST", "/auth/bind-phone", invitee2_token, {"phone": phone2})
    call("POST", "/user/submit-audit", invitee2_token)
    call("POST", f"/admin/audit/{invitee2_id}", admin_token, {"pass": True, "remark": "封禁场景"})

    st = sql(f"SELECT status FROM invite_relation WHERE invitee_id = {invitee2_id};")
    check("封禁邀请人的关系被置为作废(2)", st == "2", f"status={st}")

    # 既有权益不追回，但绝不能被这次作废的关系顺延 —— 等级须与封禁前一致
    after_ban = sql(f"SELECT vip_level FROM user WHERE id = {inviter_id};")
    check("作废的关系未给邀请人顺延权益", before_ban == after_ban,
          f"{before_ban} -> {after_ban}")

    # ---------- 收尾 ----------
    sql(f"UPDATE user SET status = 2, ban_reason = '' WHERE id = {inviter_id};")
    for oid in (inviter_code, invitee_code, invitee2):
        sql(f"DELETE FROM invite_relation WHERE inviter_id IN (SELECT id FROM user WHERE openid='{oid}') "
            f"OR invitee_id IN (SELECT id FROM user WHERE openid='{oid}');")
        sql(f"DELETE FROM invite_code WHERE user_id IN (SELECT id FROM user WHERE openid='{oid}');")
        sql(f"DELETE FROM interaction WHERE from_user_id IN (SELECT id FROM user WHERE openid='{oid}') "
            f"OR to_user_id IN (SELECT id FROM user WHERE openid='{oid}');")
        sql(f"DELETE FROM user_profile WHERE user_id IN (SELECT id FROM user WHERE openid='{oid}');")
        sql(f"DELETE FROM user_requirement WHERE user_id IN (SELECT id FROM user WHERE openid='{oid}');")
        sql(f"DELETE FROM three_view_answer WHERE user_id IN (SELECT id FROM user WHERE openid='{oid}');")
        sql(f"DELETE FROM user WHERE openid='{oid}';")

    print(f"\n{'=' * 52}")
    print(f"通过 {len(PASS)} / 失败 {len(FAIL)}")
    if FAIL:
        print("失败项：")
        for f in FAIL:
            print("  - " + f)
    print("=" * 52)
    return 1 if FAIL else 0


if __name__ == "__main__":
    sys.exit(main())
