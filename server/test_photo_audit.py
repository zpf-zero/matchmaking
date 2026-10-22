import pymysql, json, urllib.request, sys

DB = dict(host='101.35.245.54', port=3306, user='marry_link_dev',
          password='HXchGX66LCMpmcAF', database='marry_link_dev', autocommit=True)
API = 'http://127.0.0.1:8080/api/v1'

# 绕过系统代理
opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))


def conn():
    return pymysql.connect(**DB)


def q(sql, args=None):
    c = conn()
    cur = c.cursor()
    cur.execute(sql, args or ())
    rows = cur.fetchall()
    c.close()
    return rows


def api(method, path, body=None, token=None):
    """直连不上就报错，由调用方决定重试"""
    req = urllib.request.Request(API + path, method=method)
    req.add_header('Content-Type', 'application/json')
    if token:
        req.add_header('Authorization', 'Bearer ' + token)
    data = json.dumps(body).encode() if body is not None else None
    with opener.open(req, data, timeout=15) as r:
        return json.loads(r.read().decode())


def login(code):
    r = api('POST', '/auth/wx-login', {'code': code})
    return r['data']['token']


uid = q("SELECT id FROM user WHERE openid='test_user_1'")[0][0]
print(f'test_user_1 uid={uid}')

at = login('admin')
ut = login('test_user_1')

# 0. 确保有照片在库；没有就直插一条（相册上传走 COS，这里只验证状态机）
photos = q("SELECT id,audit_status FROM user_photo WHERE user_id=%s ORDER BY sort", (uid,))
if not photos:
    print('库中无照片，插入一条测试数据')
    c = conn()
    cur = c.cursor()
    cur.execute(
        "INSERT INTO user_photo (user_id,url,sort,audit_status,created_at,updated_at) "
        "VALUES (%s,%s,%s,%s,NOW(),NOW())",
        (uid, 'https://example.com/test.jpg', 0, 0))
    c.close()
    photos = q("SELECT id,audit_status FROM user_photo WHERE user_id=%s", (uid,))
print('照片:', photos)
r = api('POST', f'/admin/audit/{uid}', {'pass': False, 'remark': '头像非真人照片'}, at)
print('拒绝:', r)
u = q("SELECT status FROM user WHERE id=%s", (uid,))[0]
pr = q("SELECT audit_status FROM user_profile WHERE user_id=%s", (uid,))[0]
ph = q("SELECT audit_status FROM user_photo WHERE user_id=%s", (uid,))
print(f'  被拒后: user.status={u[0]} profile.audit_status={pr[0]} photos={[x[0] for x in ph]}')

# 2. 用户改资料 → 应重新排队
ut = login('test_user_1')
me = api('GET', '/user/me', token=ut)['data']
p = me['profile']
form = {
    'nickname': me['user']['nickname'],
    'avatar': me['user']['avatar'] or '',
    'gender': me['user']['gender'],
    'birthday': (p['birthday'] or '')[:10],
    'height': p['height'],
    'weight': p['weight'],
    'marital_status': p['marital_status'],
    'has_children': p['has_children'],
    'education': p['education'],
    'school': p.get('school') or '',
    'occupation': p.get('occupation') or '',
    'income_level': p['income_level'],
    'house_status': p['house_status'],
    'car_status': p['car_status'],
    'province': p.get('province') or '',
    'city': p['city'],
    'city_code': p.get('city_code') or '',
    'home_province': p.get('home_province') or '',
    'home_city': p.get('home_city') or '',
    'self_intro': p['self_intro'],
}
r = api('PUT', '/user/profile', form, ut)
print('保存资料:', r.get('code'), r.get('message'))
u = q("SELECT status FROM user WHERE id=%s", (uid,))[0]
pr = q("SELECT audit_status FROM user_profile WHERE user_id=%s", (uid,))[0]
ph = q("SELECT audit_status FROM user_photo WHERE user_id=%s", (uid,))
print(f'  改资料后: user.status={u[0]} profile.audit_status={pr[0]} photos={[x[0] for x in ph]}')

ok_profile = u[0] == 1 and pr[0] == 0
ok_photo = all(x[0] == 0 for x in ph)
print(f'  ✅ 资料重新排队' if ok_profile else '  ❌ 资料未重新排队')
print(f'  ✅ 相册重新排队' if ok_photo else '  ❌ 相册未重新排队（原 bug）')

# 3. 管理员通过 → 相册应一并通过
r = api('POST', f'/admin/audit/{uid}', {'pass': True, 'remark': ''}, at)
print('通过:', r.get('code'), r.get('message'))
u = q("SELECT status FROM user WHERE id=%s", (uid,))[0]
pr = q("SELECT audit_status FROM user_profile WHERE user_id=%s", (uid,))[0]
ph = q("SELECT audit_status FROM user_photo WHERE user_id=%s", (uid,))
print(f'  通过后: user.status={u[0]} profile.audit_status={pr[0]} photos={[x[0] for x in ph]}')
ok = u[0] == 2 and pr[0] == 1 and all(x[0] == 1 for x in ph)
print('  ✅ 全部通过，相册可正常展示' if ok else '  ❌ 状态不一致')
sys.exit(0 if (ok_profile and ok_photo and ok) else 1)
