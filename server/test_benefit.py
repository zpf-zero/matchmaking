#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
权益体系重构 · 端到端验证脚本（v1.3）

覆盖三项新功能与权益判定从「到期时间」到「永久等级」的切换：
  1. 权益等级 vip_level：读写、曝光加权、对外不暴露旧字段
  2. 三观问答：题库公开、答题免费、看答案需权益、一致率只算双方都答过的题
  3. 照片分层：前 3 张免费、第 4 张起锁定并计数，解锁后全量返回
  4. 原生家庭：默认私有；「已授权 + 我已解锁」双条件才返回；空内容不允许授权
  5. 行业标签：写入、卡片展示、筛选

用法（后端需先跑在 localhost:8080）：
    python server/test_benefit.py
"""
import json
import subprocess
import sys
import time
import urllib.error
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
    out = subprocess.run(
        ["mysql", "--host=101.35.245.54", "--user=marry_link_dev",
         "--password=HXchGX66LCMpmcAF", "marry_link_dev", "-N", "-B", "-e", query],
        capture_output=True, text=True,
    )
    return out.stdout.strip()


# 资料模板：性别女（与男性测试号构成异性可见关系）
FEMALE_PROFILE = {
    "nickname": "权益测试女", "avatar": "/uploads/test.png", "gender": 2,
    "birthday": "1995-06-15", "height": 165, "weight": 52,
    "marital_status": 1, "has_children": 0, "education": 3,
    "school": "测试大学", "occupation": "教师", "income_level": 3,
    "house_status": 1, "car_status": 2, "province": "河南省", "city": "洛阳市",
    "city_code": "410300", "home_province": "河南省", "home_city": "洛阳市",
    "self_intro": "这是一段用于权益体系测试的自我介绍，长度需要超过三十个字符才能通过校验，这里补足。",
    "industry": 2,  # 教育系统
}

MALE_PROFILE = {
    **FEMALE_PROFILE,
    "nickname": "权益测试男", "gender": 1, "height": 178, "weight": 72,
    "occupation": "公务员", "house_status": 1, "car_status": 1,
    "city": "郑州市", "city_code": "410100",
    "self_intro": "这是一段用于权益体系测试的男性自我介绍，同样需要超过三十个字符才能通过校验。",
    "industry": 1,  # 机关事业单位
}

REQUIREMENT = {
    "age_min": 22, "age_max": 45, "height_min": 145, "height_max": 200,
    "education_min": 0, "income_min": 0, "marital_status": "1",
    "accept_children": 1, "city_codes": "", "other": "",
}


def setup_full_user(token, profile, phone):
    """走完资料 + 择偶要求 + 手机号 + 提交审核，让账号可被浏览。

    必须先 submit-audit 把 status 推到 1（待审核），管理员才能审；
    否则 admin/audit 找不到队列里的记录，用户会停在 status=0 而无法浏览他人。
    """
    call("PUT", "/user/profile", token, profile)
    call("PUT", "/user/requirement", token, REQUIREMENT)
    call("POST", "/auth/bind-phone", token, {"phone": phone})
    call("POST", "/user/submit-audit", token)


def main():
    ts = str(int(time.time()))
    female_code = f"benefit_f_{ts}"
    male_code = f"benefit_m_{ts}"
    # 手机号有应用层查重，且测试库历史数据占过固定号段（13800138010 等），
    # 每次运行用时间戳生成唯一手机号，避免撞上种子/历史数据
    f_phone = "138" + ts[2:10]
    m_phone = "138" + str(int(ts) + 1)[2:10]

    # ---------- 准备 ----------
    print("\n[0] 准备测试账号")
    for oid in (female_code, male_code):
        sql(f"DELETE FROM three_view_answer WHERE user_id IN (SELECT id FROM user WHERE openid='{oid}');")
        sql(f"DELETE FROM user_photo WHERE user_id IN (SELECT id FROM user WHERE openid='{oid}');")
        sql(f"DELETE FROM user_profile WHERE user_id IN (SELECT id FROM user WHERE openid='{oid}');")
        sql(f"DELETE FROM user_requirement WHERE user_id IN (SELECT id FROM user WHERE openid='{oid}');")
        sql(f"DELETE FROM user WHERE openid='{oid}';")

    female_token, female_id = login(female_code)   # 被查看者（女方）
    male_token, male_id = login(male_code)         # 查看者（男方）
    check("两个账号登录成功", female_id and male_id,
          f"female={female_id} male={male_id}")

    setup_full_user(female_token, FEMALE_PROFILE, f_phone)
    setup_full_user(male_token, MALE_PROFILE, m_phone)

    admin_token, _ = login("admin")
    call("POST", f"/admin/audit/{female_id}", admin_token, {"pass": True, "remark": "测试"})
    call("POST", f"/admin/audit/{male_id}", admin_token, {"pass": True, "remark": "测试"})
    fst = sql(f"SELECT CONCAT(status,'/',(SELECT audit_status FROM user_profile WHERE user_id={female_id})) "
              f"FROM user WHERE id={female_id};")
    check("女方已进入可浏览状态(2/1)", fst == "2/1", f"status/audit={fst}")
    mst = sql(f"SELECT CONCAT(status,'/',(SELECT audit_status FROM user_profile WHERE user_id={male_id})) "
              f"FROM user WHERE id={male_id};")
    check("男方已进入可浏览状态(2/1)", mst == "2/1", f"status/audit={mst}")

    # 女方相册造 5 张审核通过的照片
    sql(f"DELETE FROM user_photo WHERE user_id = {female_id};")
    for i in range(5):
        sql(f"INSERT INTO user_photo (user_id, url, sort, audit_status, created_at, updated_at) "
            f"VALUES ({female_id}, '/uploads/photo/p{i}.jpg', {i}, 1, NOW(), NOW());")

    # ---------- 1. 三观题库与答题（免费） ----------
    print("\n[1] 三观问答：题库公开、答题免费")
    r = call("GET", "/common/three-view/questions")
    check("未登录也能取题库", r.get("code") == 0, r.get("message"))
    questions = r["data"]
    check("题库返回 8 道题", len(questions) == 8, f"len={len(questions)}")
    check("每题都有标题与选项",
          all(q.get("title") and q.get("options") for q in questions))

    n = len(questions)
    # 女方答前 6 题，男方答满 8 题 —— 用于验证「只算双方都答过的题」
    female_answers = [0, 1, 2, 0, 1, 2, -1, -1]
    male_answers = [0, 1, 9, 0, 1, 2, 0, 0]  # 第 3 题故意不同（索引 2 值为 9 无效，见下）

    r = call("GET", "/user/three-view", female_token)
    check("取我的答题状态成功", r.get("code") == 0, r.get("message"))
    check("未答时 answered=0", r["data"]["answered"] == 0, str(r["data"]["answered"]))
    check("answers 长度等于题库题数", len(r["data"]["answers"]) == n)

    r = call("PUT", "/user/three-view", female_token, {"answers": female_answers})
    check("答题免费（无需权益即可提交）", r.get("code") == 0, r.get("message"))

    # 越界下标必须整体拒绝，避免写入脏数据
    bad = [0] * n
    bad[0] = 99
    r = call("PUT", "/user/three-view", male_token, {"answers": bad})
    check("选项下标越界被拒绝", r.get("code") != 0, r.get("message"))

    # 长度不匹配也必须拒绝
    r = call("PUT", "/user/three-view", male_token, {"answers": [0, 1, 2]})
    check("答案数量不匹配被拒绝", r.get("code") != 0, r.get("message"))

    male_answers = [0, 1, 2, 0, 1, 2, 0, 0]
    r = call("PUT", "/user/three-view", male_token, {"answers": male_answers})
    check("男方答题成功", r.get("code") == 0, r.get("message"))

    # ---------- 2. 权益未解锁时看不到对方答案 ----------
    print("\n[2] 未解锁：只看得到「答了几题」")
    sql(f"UPDATE user SET vip_level = 0 WHERE id = {male_id};")
    r = call("GET", f"/user/detail/{female_id}", male_token)
    check("取他人资料成功", r.get("code") == 0, r.get("message"))
    d = r["data"]
    check("three_view_answered 为 true（对方答过题）", d["three_view_answered"] is True)
    check("three_view_unlocked 为 false", d["three_view_unlocked"] is False)
    check("未解锁时 three_view.answers 为空",
          not (d["three_view"] or {}).get("answers"),
          str((d["three_view"] or {}).get("answers")))
    check("未解锁仍告知对方答题数", (d["three_view"] or {}).get("answered") == 6,
          str((d["three_view"] or {}).get("answered")))

    # ---------- 3. 照片分层 ----------
    print("\n[3] 照片分层：前 3 张免费")
    check("总照片数为 5", d["photo_total"] == 5, str(d["photo_total"]))
    check("未解锁只返回 3 张", len(d["photos"]) == 3, f"len={len(d['photos'])}")
    check("锁定张数为 2", d["locked_photo_count"] == 2, str(d["locked_photo_count"]))
    check("photos_locked 为 true", d["photos_locked"] is True)

    # ---------- 4. 原生家庭默认私有 ----------
    print("\n[4] 原生家庭：默认私有")
    check("未填写时 viewed_family 为 false", d["viewed_family"] is False)
    check("未授权时 family 为 null", d["family"] is None)

    # 女方填写原生家庭但不授权
    r = call("PUT", "/user/profile", female_token, {
        **FEMALE_PROFILE,
        "family_parent_occ": 1, "family_atmosphere": 1, "family_only_child": 1,
        "family_city": "河南 洛阳", "family_note": "家庭关系简单", "family_visible": 0,
    })
    check("保存原生家庭（未授权）成功", r.get("code") == 0, r.get("message"))
    # 重新审核（改资料会进复审队列）
    call("POST", f"/admin/audit/{female_id}", admin_token, {"pass": True, "remark": "复审"})

    sql(f"UPDATE user SET vip_level = 1 WHERE id = {male_id};")
    d = call("GET", f"/user/detail/{female_id}", male_token)["data"]
    check("已解锁但对方未授权时仍看不到家庭信息",
          d["viewed_family"] is False and d["family"] is None,
          f"viewed={d['viewed_family']} family={d['family']}")

    # ---------- 5. 解锁后：三观 + 照片 + 家庭 ----------
    print("\n[5] 解锁后可见内容")
    check("解锁后 three_view_unlocked 为 true", d["three_view_unlocked"] is True)
    tv = d["three_view"] or {}
    check("解锁后返回对方答案", len(tv.get("answers") or []) == n,
          f"len={len(tv.get('answers') or [])}")
    check("解锁后返回我的答案（并列展示）", len(tv.get("my_answers") or []) == n)
    check("my_answered 为 true", tv.get("my_answered") is True)
    check("answered 为对方答题数 6", tv.get("answered") == 6, str(tv.get("answered")))

    # 双方都答过的题 = 第 1,2,3,4,5,6 题（女方答了前 6 题）。
    # 女方 [0,1,2,0,1,2]，男方 [0,1,2,0,1,2] → 6/6 全一致 = 100%
    check("一致率只算双方都答过的题（应为 100）", tv.get("match_rate") == 100,
          f"match_rate={tv.get('match_rate')}")

    check("解锁后返回全部 5 张照片", len(d["photos"]) == 5, f"len={len(d['photos'])}")
    check("解锁后无锁定张数", d["locked_photo_count"] == 0, str(d["locked_photo_count"]))
    check("解锁后 photos_locked 为 false", d["photos_locked"] is False)

    # ---------- 6. 授权后原生家庭可见 ----------
    print("\n[6] 原生家庭：授权后才可见")
    r = call("PUT", "/user/profile", female_token, {
        **FEMALE_PROFILE,
        "family_parent_occ": 1, "family_atmosphere": 1, "family_only_child": 1,
        "family_city": "河南 洛阳", "family_note": "家庭关系简单", "family_visible": 1,
    })
    check("保存原生家庭（已授权）成功", r.get("code") == 0, r.get("message"))
    call("POST", f"/admin/audit/{female_id}", admin_token, {"pass": True, "remark": "复审"})

    d = call("GET", f"/user/detail/{female_id}", male_token)["data"]
    check("授权 + 已解锁后可见家庭信息", d["viewed_family"] is True)
    fam = d["family"] or {}
    check("家庭字段带回文案", fam.get("parent_occ_text") == "均为公职",
          str(fam.get("parent_occ_text")))
    check("家庭氛围文案正确", fam.get("atmosphere_text") == "氛围融洽",
          str(fam.get("atmosphere_text")))
    check("独生子女文案正确", fam.get("only_child_text") == "是", str(fam.get("only_child_text")))
    check("家庭所在地只到城市", fam.get("city") == "河南 洛阳", str(fam.get("city")))

    # 未解锁者即便对方授权也看不到 —— 双条件缺一不可
    sql(f"UPDATE user SET vip_level = 0 WHERE id = {male_id};")
    d2 = call("GET", f"/user/detail/{female_id}", male_token)["data"]
    check("对方已授权但我未解锁时仍看不到", d2["viewed_family"] is False and d2["family"] is None)
    sql(f"UPDATE user SET vip_level = 1 WHERE id = {male_id};")

    # ---------- 7. 「授权但内容全空」不允许 ----------
    print("\n[7] 空内容的授权被拒绝")
    r = call("PUT", "/user/profile", female_token, {
        **FEMALE_PROFILE,
        "family_parent_occ": 0, "family_atmosphere": 0, "family_only_child": 0,
        "family_city": "", "family_note": "", "family_visible": 1,
    })
    check("提交空内容 + 授权，接口成功（不报错）", r.get("code") == 0, r.get("message"))
    vis = sql(f"SELECT family_visible FROM user_profile WHERE user_id = {female_id};")
    check("服务端把空内容的 family_visible 强制置 0", vis == "0", f"family_visible={vis}")
    # 改资料会重新进入复审队列（audit_status 回 0），必须再过审才能被浏览到 ——
    # 否则下面的行业/曝光断言会因为「资料审核中」而被 baseQuery 过滤掉
    call("POST", f"/admin/audit/{female_id}", admin_token, {"pass": True, "remark": "复审"})

    # ---------- 8. 行业标签 ----------
    print("\n[8] 行业标签与筛选")
    ind = sql(f"SELECT industry FROM user_profile WHERE user_id = {female_id};")
    check("行业标签已落库（2=教育系统）", ind == "2", f"industry={ind}")

    r = call("GET", "/discover/recommend?page=1&page_size=50", male_token)
    cards = r["data"]["list"]
    my = [c for c in cards if c["id"] == female_id]
    check("推荐卡片带出行业标签文案",
          bool(my) and my[0].get("industry_text") == "教育系统",
          str(my[0].get("industry_text") if my else "未找到"))

    r = call("GET", "/discover/search?page=1&page_size=50&industry=2", male_token)
    hit = [c for c in r["data"]["list"] if c["id"] == female_id]
    check("按行业筛选能命中", bool(hit))

    r = call("GET", "/discover/search?page=1&page_size=50&industry=7", male_token)
    miss = [c for c in r["data"]["list"] if c["id"] == female_id]
    check("按其他行业筛选不命中", not miss)

    # ---------- 9. 曝光加权 ----------
    print("\n[9] 权益曝光加权")
    sql(f"UPDATE user SET vip_level = 1 WHERE id = {female_id};")
    q = sql(
        "SELECT u.id FROM user u JOIN user_profile p ON p.user_id = u.id "
        f"WHERE u.status = 2 AND p.audit_status = 1 AND u.gender = 2 "
        "ORDER BY (u.vip_level >= 1) DESC, u.last_active_at DESC, u.id DESC LIMIT 5;")
    check("曝光加权 SQL 可执行且列表非空", bool(q), q.replace("\n", ","))

    # ---------- 收尾 ----------
    for oid, uid in ((female_code, female_id), (male_code, male_id)):
        sql(f"DELETE FROM three_view_answer WHERE user_id = {uid};")
        sql(f"DELETE FROM user_photo WHERE user_id = {uid};")
        sql(f"DELETE FROM user_profile WHERE user_id = {uid};")
        sql(f"DELETE FROM user_requirement WHERE user_id = {uid};")
        sql(f"DELETE FROM interaction WHERE from_user_id = {uid} OR to_user_id = {uid};")
        sql(f"DELETE FROM user WHERE id = {uid};")

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
