#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
H5 手机号验证码登录 · 端到端验证脚本

验证：
  1. 发送验证码（mock 模式返回 dev_code）
  2. 60 秒内重复发送、非法手机号被拒
  3. 错误验证码被拒
  4. 正确验证码登录成功，首次自动建号（need_profile=true）
  5. 验证码用后即焚
  6. 冷却期过后重发，同一手机号登录命中同一账号
  7. 错误次数达到上限后即使验证码正确也被拒
  8. 已存在的账号（种子用户手机号）按手机号命中，不重复建号
  9. 登录后 /user/me 可正常访问，openid 占位为 h5_<手机号>

用法（后端需先跑在 localhost:8080，且已应用 005 迁移）：
    python server/test_phone_login.py
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


def check(name, cond, detail=""):
    (PASS if cond else FAIL).append(name)
    print(("  [PASS] " if cond else "  [FAIL] ") + name + (f"  -> {detail}" if detail else ""))


def sql(query):
    """直连云库执行 SQL，用于造数与校验落库结果"""
    out = subprocess.run(
        ["mysql", "--host=101.35.245.54", "--user=marry_link_dev",
         "--password=HXchGX66LCMpmcAF", "marry_link_dev", "-N", "-B", "-e", query],
        capture_output=True, text=True,
    )
    return out.stdout.strip()


def send_code(phone):
    return call("POST", "/auth/sms/send", body={"phone": phone})


def sms_login(phone, code):
    return call("POST", "/auth/sms/login", body={"phone": phone, "code": code})


def age_last_code(phone, minutes=2):
    """把最近一条验证码的创建时间往前拨，绕过 60 秒冷却（仅测试造数）"""
    sql(f"UPDATE sms_code SET created_at = DATE_SUB(NOW(), INTERVAL {minutes} MINUTE) "
        f"WHERE phone = '{phone}' ORDER BY id DESC LIMIT 1;")


def main():
    ts = str(int(time.time()))
    phone = "139" + ts[2:10]
    # 种子用户，手机号固定，用于验证「老账号按手机号命中」
    seed_phone = "13810000000"
    seed_uid = sql(f"SELECT id FROM user WHERE phone = '{seed_phone}' LIMIT 1;")

    # ---------- 准备 ----------
    print("\n[0] 清理上轮残留")
    sql(f"DELETE FROM sms_code WHERE phone IN ('{phone}', '{seed_phone}');")
    sql(f"DELETE FROM user WHERE phone = '{phone}' AND openid LIKE 'h5_%';")
    check("测试手机号未被占用", sql(f"SELECT COUNT(*) FROM user WHERE phone = '{phone}';") == "0")

    # ---------- 1. 发送验证码 ----------
    print("\n[1] 发送验证码")
    r = send_code(phone)
    check("发送成功", r.get("code") == 0, r.get("message"))
    data = r.get("data") or {}
    dev_code = data.get("dev_code", "")
    check("开发环境返回 dev_code", len(dev_code) == 6 and dev_code.isdigit(), f"code={dev_code}")
    check("有效期 300 秒", data.get("expires_in") == 300, str(data.get("expires_in")))
    check("冷却 60 秒", data.get("resend_in") == 60, str(data.get("resend_in")))
    check("落库只存哈希不存明文",
          sql(f"SELECT code_hash FROM sms_code WHERE phone='{phone}' ORDER BY id DESC LIMIT 1;") != dev_code)

    r = send_code(phone)
    check("60 秒内重复发送被拒", r.get("code") != 0, r.get("message"))

    r = send_code("12345")
    check("非法手机号被拒", r.get("code") != 0, r.get("message"))

    # ---------- 2. 错误验证码 ----------
    print("\n[2] 错误验证码")
    wrong = "000000" if dev_code != "000000" else "111111"
    r = sms_login(phone, wrong)
    check("错误验证码被拒", r.get("code") != 0, r.get("message"))
    check("错误次数 +1",
          sql(f"SELECT attempts FROM sms_code WHERE phone='{phone}' ORDER BY id DESC LIMIT 1;") == "1")

    r = sms_login(phone, "123")
    check("格式不足 6 位被拒", r.get("code") != 0, r.get("message"))

    # ---------- 3. 正确验证码登录（首次建号） ----------
    print("\n[3] 正确验证码登录（首次自动建号）")
    r = sms_login(phone, dev_code)
    check("登录成功", r.get("code") == 0, r.get("message"))
    d = r.get("data") or {}
    token = d.get("token", "")
    uid = (d.get("user") or {}).get("id")
    check("返回 token", bool(token))
    check("首次登录需要完善资料", d.get("need_profile") is True, str(d.get("need_profile")))
    check("账号状态为待完善", d.get("status") == 0, str(d.get("status")))
    check("openid 占位为 h5_<手机号>",
          sql(f"SELECT openid FROM user WHERE id = {uid};") == f"h5_{phone}")
    check("库里 phone 已落库",
          sql(f"SELECT phone FROM user WHERE id = {uid};") == phone)

    r = call("GET", "/user/me", token)
    check("登录后 /user/me 可访问", r.get("code") == 0, r.get("message"))

    # ---------- 4. 验证码用后即焚 ----------
    print("\n[4] 验证码用后即焚")
    r = sms_login(phone, dev_code)
    check("同一个验证码不能复用", r.get("code") != 0, r.get("message"))

    # ---------- 5. 冷却后重发，二次登录命中同一账号 ----------
    print("\n[5] 二次登录命中同一账号")
    age_last_code(phone)
    r = send_code(phone)
    check("冷却期过后可重发", r.get("code") == 0, r.get("message"))
    code2 = (r.get("data") or {}).get("dev_code", "")
    r = sms_login(phone, code2)
    check("二次登录成功", r.get("code") == 0, r.get("message"))
    check("命中同一账号", (r.get("data") or {}).get("user", {}).get("id") == uid,
          f"uid={uid}")

    # ---------- 6. 错误次数上限 ----------
    print("\n[6] 错误次数上限")
    age_last_code(phone)
    r = send_code(phone)
    code3 = (r.get("data") or {}).get("dev_code", "")
    for _ in range(5):
        sms_login(phone, "000000" if code3 != "000000" else "111111")
    r = sms_login(phone, code3)
    check("错误 5 次后正确验证码也被拒", r.get("code") != 0, r.get("message"))

    # ---------- 7. 已存在账号按手机号命中 ----------
    print("\n[7] 已有账号按手机号命中")
    r = send_code(seed_phone)
    check("种子账号手机号可发送", r.get("code") == 0, r.get("message"))
    seed_code = (r.get("data") or {}).get("dev_code", "")
    r = sms_login(seed_phone, seed_code)
    check("种子账号登录成功", r.get("code") == 0, r.get("message"))
    check("命中已有账号而非新建",
          (r.get("data") or {}).get("user", {}).get("id") == int(seed_uid or 0),
          f"expect={seed_uid} got={(r.get('data') or {}).get('user', {}).get('id')}")
    check("已有账号资料仍为已审核状态",
          (r.get("data") or {}).get("status") == 2,
          str((r.get("data") or {}).get("status")))

    # ---------- 8. 无验证码 / 过期验证码 ----------
    print("\n[8] 无验证码 / 过期验证码")
    other_phone = "137" + ts[2:10]
    r = sms_login(other_phone, "123456")
    check("未发送过验证码时登录被拒", r.get("code") != 0, r.get("message"))
    sql(f"DELETE FROM sms_code WHERE phone = '{other_phone}';")

    # ---------- 收尾 ----------
    sql(f"DELETE FROM sms_code WHERE phone IN ('{phone}', '{seed_phone}', '{other_phone}');")
    if uid:
        for table in ("user_profile", "user_requirement", "user_photo",
                      "three_view_answer", "interaction", "invite_relation"):
            col = "user_id" if table != "interaction" else "from_user_id"
            sql(f"DELETE FROM {table} WHERE {col} = {uid};")
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
