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


uid = q("SELECT id FROM user WHERE openid='test_user_1'")[0][0]
at = api('POST', '/auth/wx-login', {'code': 'admin'})['data']['token']
results = []


def chk(name, ok, extra=''):
    results.append(ok)
    print(f'  {"✅" if ok else "❌"} {name}{" — " + str(extra) if extra else ""}')


def snap():
    u = q("SELECT status FROM user WHERE id=%s", (uid,))[0][0]
    pr = q("SELECT audit_status FROM user_profile WHERE user_id=%s", (uid,))[0][0]
    ph = [x[0] for x in q("SELECT audit_status FROM user_photo WHERE user_id=%s", (uid,))]
    return u, pr, ph


print('=== 场景 A：误拒 → 管理员直接改判通过（不过用户手） ===')
api('POST', f'/admin/audit/{uid}', {'pass': False, 'remark': '测试误拒'}, at)
u, pr, ph = snap()
chk('被拒后 user=3 profile=2 photos 全 2', u == 3 and pr == 2 and all(x == 2 for x in ph), f'u={u} p={pr} ph={ph}')

api('POST', f'/admin/audit/{uid}', {'pass': True, 'remark': ''}, at)
u, pr, ph = snap()
chk('改判后 user=2 profile=1 photos 全 1', u == 2 and pr == 1 and all(x == 1 for x in ph), f'u={u} p={pr} ph={ph}')

print('\n=== 场景 B：封禁 → 解封（原 bug：audit_status=Rejected 会卡死） ===')
api('POST', f'/admin/audit/{uid}', {'pass': False, 'remark': '再拒一次'}, at)
u, pr, ph = snap()
print(f'  当前 u={u} p={pr} ph={ph}')
api('POST', f'/admin/user/{uid}/ban', {'reason': '测试封禁'}, at)
u, pr, ph = snap()
chk('封禁后 user=4', u == 4, f'u={u}')

api('POST', f'/admin/user/{uid}/unban', {}, at)
u, pr, ph = snap()
# 解封后 user 应为 1（审核中），profile 回到 0（待审）—— 否则永远等不到审核
chk('解封后 user=1（审核中）', u == 1, f'u={u}')
chk('解封后 profile=0（待审，进了队列）', pr == 0, f'p={pr}')
chk('解封后 photos 全部=0（跟随排队）', all(x == 0 for x in ph), f'ph={ph}')

print('\n=== 场景 C：相册只展示已通过的 ===')
api('POST', f'/admin/audit/{uid}', {'pass': True, 'remark': ''}, at)
ut = api('POST', '/auth/wx-login', {'code': 'test_user_1'})['data']['token']
photos = api('GET', '/user/me', token=ut)['data']['photos']
u, pr, ph = snap()
chk('通过后 user=2 profile=1', u == 2 and pr == 1, f'u={u} p={pr}')
chk('photo 表全部=1', all(x == 1 for x in ph), f'ph={ph}')

print('\n================ 汇总 ================')
print(f'通过 {sum(results)}/{len(results)}')
sys.exit(0 if all(results) else 1)
