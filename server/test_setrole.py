import pymysql, json, urllib.request, sys

DB = dict(host='101.35.245.54', port=3306, user='marry_link_dev',
          password='HXchGX66LCMpmcAF', database='marry_link_dev', autocommit=True)
API = 'http://127.0.0.1:8080/api/v1'
op = urllib.request.build_opener(urllib.request.ProxyHandler({}))


def q(sql, args=None):
    c = pymysql.connect(**DB); cur = c.cursor()
    cur.execute(sql, args or ()); rows = cur.fetchall(); c.close(); return rows


def api(method, path, body=None, token=None):
    req = urllib.request.Request(API + path, method=method)
    req.add_header('Content-Type', 'application/json')
    if token:
        req.add_header('Authorization', 'Bearer ' + token)
    data = json.dumps(body).encode() if body is not None else None
    with op.open(req, data, timeout=15) as r:
        return json.loads(r.read().decode())


def login(code):
    return api('POST', '/auth/wx-login', {'code': code})['data']


results = []


def chk(name, ok, extra=''):
    results.append(ok)
    print(f'  {"✅" if ok else "❌"} {name}{" — " + str(extra) if extra else ""}')


# 用一个脏数据账号 uid=5 (test_useryyy, role=2, status=1) 验证扶正
row = q("SELECT id,openid,role,status FROM user WHERE id=5")
print('目标账号:', row)
uid = 5

print('\n=== 场景 A：运营号登录时状态扶正（原 bug：Auditing 不扶正） ===')
d = login('test_useryyy')
print(f'  登录返回 status={d["status"]} role={d["role"]}')
s = q("SELECT status FROM user WHERE id=%s", (uid,))[0][0]
chk('登录后 status 扶正为 2', s == 2, f'status={s}')

print('\n=== 场景 B：降为普通用户时按资料实况重算 ===')
at = login('admin')['token']
r = api('POST', f'/admin/user/{uid}/role', {'role': 1}, at)
print('  降级:', r.get('code'), r.get('message'))
u = q("SELECT role,status FROM user WHERE id=%s", (uid,))[0]
pr = q("SELECT audit_status FROM user_profile WHERE user_id=%s", (uid,))
print(f'  role={u[0]} status={u[1]} profile={pr}')
if pr:
    ast = pr[0][0]
    expect = {0: 1, 1: 2, 2: 3}[ast]
    chk('status 与资料实况一致', u[1] == expect, f'profile.audit_status={ast} → 期望 status={expect}，实得 {u[1]}')
else:
    chk('无资料 → status=0（回向导）', u[1] == 0, f'status={u[1]}')

print('\n=== 场景 C：再提为审核员，状态扶正 ===')
r = api('POST', f'/admin/user/{uid}/role', {'role': 2}, at)
print('  提权:', r.get('code'), r.get('message'))
u = q("SELECT role,status FROM user WHERE id=%s", (uid,))[0]
chk('提权后 status=2 正常', u[0] == 2 and u[1] == 2, f'role={u[0]} status={u[1]}')

print('\n=== 场景 D：运营号不进待审队列（回归） ===')
d = api('GET', '/admin/audit/list?page=1&page_size=50', None, at)
items = d['data']['list'] if isinstance(d['data'], dict) else d['data']
ids = [it.get('user_id') for it in items]
staff = {r[0] for r in q("SELECT id FROM user WHERE role>1")}
bad = [i for i in ids if i in staff]
chk('待审队列中无运营号', len(bad) == 0, f'队列 {len(ids)} 条（{ids}），运营号命中 {bad}')

print('\n================ 汇总 ================')
print(f'通过 {sum(results)}/{len(results)}')
sys.exit(0 if all(results) else 1)
