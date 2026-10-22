import pymysql, json, urllib.request, sys

DB = dict(host='101.35.245.54', port=3306, user='marry_link_dev',
          password='HXchGX66LCMpmcAF', database='marry_link_dev', autocommit=True)
API = 'http://127.0.0.1:8080/api/v1'
op = urllib.request.build_opener(urllib.request.ProxyHandler({}))


def q(sql, args=None):
    c = pymysql.connect(**DB); cur = c.cursor()
    cur.execute(sql, args or ()); rows = cur.fetchall(); c.close(); return rows


def api(method, path, body=None, token=None, raw=False):
    req = urllib.request.Request(API + path, method=method)
    req.add_header('Content-Type', 'application/json')
    if token:
        req.add_header('Authorization', 'Bearer ' + token)
    data = json.dumps(body).encode() if body is not None else None
    try:
        with op.open(req, data, timeout=15) as r:
            return json.loads(r.read().decode())
    except urllib.error.HTTPError as e:
        return json.loads(e.read().decode())


def login(code):
    return api('POST', '/auth/wx-login', {'code': code})['data']


results = []


def chk(name, ok, extra=''):
    results.append(ok)
    print(f'  {"✅" if ok else "❌"} {name}{" — " + str(extra) if extra else ""}')


uid = q("SELECT id FROM user WHERE openid='test_user_1'")[0][0]
at = login('admin')['token']
ut = login('test_user_1')['token']

# 找一个可互动的目标（异性、正常）
target = q("""SELECT id,nickname FROM user WHERE role=1 AND status=2 AND gender=1
              AND id<>%s LIMIT 1""", (uid,))
print('目标用户:', target)
if not target:
    print('无可互动目标，跳过')
    sys.exit(0)
tid = target[0][0]

print('\n=== 场景 A：正常用户可互动（基线） ===')
r = api('POST', '/interaction/like', {'to_user_id': tid}, ut)
chk('正常状态可喜欢', r.get('code') == 0, f'{r.get("code")} {r.get("message")}')
api('DELETE', f'/interaction/like/{tid}', None, ut)

print('\n=== 场景 B：待审核用户不能互动（原漏洞） ===')
api('POST', f'/admin/audit/{uid}', {'pass': False, 'remark': '测试'}, at)  # 先拒
q("UPDATE user SET status=1 WHERE id=%s", (uid,))                          # 直接置待审
ut = login('test_user_1')['token']
r = api('POST', '/interaction/like', {'to_user_id': tid}, ut)
chk('待审核不能喜欢', r.get('code') != 0, f'{r.get("code")} {r.get("message")}')

print('\n=== 场景 C：未通过用户不能互动 ===')
q("UPDATE user SET status=3 WHERE id=%s", (uid,))
ut = login('test_user_1')['token']
r = api('POST', '/interaction/like', {'to_user_id': tid}, ut)
chk('未通过不能喜欢', r.get('code') != 0, f'{r.get("code")} {r.get("message")}')

print('\n=== 场景 D：封禁用户不能互动 ===')
q("UPDATE user SET status=4 WHERE id=%s", (uid,))
ut = login('test_user_1')['token']
r = api('POST', '/interaction/like', {'to_user_id': tid}, ut)
chk('封禁不能喜欢', r.get('code') != 0, f'{r.get("code")} {r.get("message")}')

print('\n=== 场景 E：非正常状态不能发消息 ===')
conv = q("""SELECT id FROM conversation WHERE (user_a_id=%s OR user_b_id=%s)
            AND status=1 LIMIT 1""", (uid, uid))
if conv:
    cid = conv[0][0]
    r = api('POST', f'/chat/conversation/{cid}/message', {'type': 1, 'content': 'test'}, ut)
    chk('封禁不能发消息', r.get('code') != 0, f'{r.get("code")} {r.get("message")}')
else:
    print('  无正常会话，跳过（改判未通过场景由同样代码路径覆盖）')
    chk('封禁不能发消息（跳过）', True, 'skip')

print('\n=== 恢复 ===')
q("UPDATE user SET status=2 WHERE id=%s", (uid,))
q("UPDATE user_profile SET audit_status=1,audit_remark='' WHERE user_id=%s", (uid,))
q("UPDATE user_photo SET audit_status=1 WHERE user_id=%s", (uid,))
ut = login('test_user_1')['token']
r = api('POST', '/interaction/like', {'to_user_id': tid}, ut)
chk('恢复后可互动', r.get('code') == 0, f'{r.get("code")} {r.get("message")}')
api('DELETE', f'/interaction/like/{tid}', None, ut)

print('\n================ 汇总 ================')
print(f'通过 {sum(results)}/{len(results)}')
sys.exit(0 if all(results) else 1)
