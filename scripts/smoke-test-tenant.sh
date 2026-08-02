#!/bin/bash
# 多租户（PG schema 隔离）端到端测试：
# 主租户开通两个新租户 → 各租户独立管理员登录 → 各自建数据 → 验证跨租户完全隔离。
# 需以 APP_TENANT_ENABLED=true 启动服务。
set -u
BASE=http://127.0.0.1:8080/api/v1
pass=0; fail=0
check() { if echo "$2" | grep -q "$3"; then pass=$((pass+1)); echo "PASS: $1";
  else fail=$((fail+1)); echo "FAIL: $1 => $2"; fi; }
jqf() { python3 -c "import sys,json;d=json.load(sys.stdin);print(eval(\"d$1\"))"; }

# 登录：$1 用户名 $2 密码 $3 租户编码(可空)
login() {
  local capid code hdr=()
  [ -n "$3" ] && hdr=(-H "X-Tenant: $3")
  capid=$(curl -s "${hdr[@]}" $BASE/auth/captcha | jqf "['data']['captchaId']")
  code=$(redis-cli --no-raw get "captcha:$capid" | tr -d '"')
  curl -s "${hdr[@]}" -X POST $BASE/auth/login -H 'Content-Type: application/json' \
    -d "{\"username\":\"$1\",\"password\":\"$2\",\"captchaId\":\"$capid\",\"captchaCode\":\"$code\"}" | jqf "['data']['accessToken']"
}

check "多租户已启用" "$(curl -s $BASE/tenant-enabled)" '"enabled":true'

# 主租户 admin 登录（无 X-Tenant → primary/public）
ADMIN=$(login admin admin123 ""); A="Authorization: Bearer $ADMIN"
check "主租户管理员登录" "$ADMIN" '.'

# 开通两个租户 acme、globex
check "开通租户 acme" "$(curl -s -X POST $BASE/system/tenants -H "$A" -H 'Content-Type: application/json' -d '{"code":"acme","name":"Acme 公司"}')" '"schema":"tenant_acme"'
check "开通租户 globex" "$(curl -s -X POST $BASE/system/tenants -H "$A" -H 'Content-Type: application/json' -d '{"code":"globex","name":"Globex 公司"}')" '"schema":"tenant_globex"'
check "租户列表含主租户" "$(curl -s "$BASE/system/tenants?pageSize=50" -H "$A")" '"primary":true'
check "非法编码被拒" "$(curl -s -X POST $BASE/system/tenants -H "$A" -H 'Content-Type: application/json' -d '{"code":"1bad","name":"x"}')" '"code":1000'

# 各租户用各自 schema 里的默认管理员登录（provision 种子了 admin/admin123）
ACME=$(login admin admin123 "acme"); AC="Authorization: Bearer $ACME"
GLOBEX=$(login admin admin123 "globex"); GX="Authorization: Bearer $GLOBEX"
check "acme 管理员登录" "$ACME" '.'
check "globex 管理员登录" "$GLOBEX" '.'

# 每个租户在各自 schema 建用户
curl -s -X POST $BASE/system/users -H "$AC" -H 'Content-Type: application/json' -d '{"username":"acme_alice","password":"test123456"}' >/dev/null
curl -s -X POST $BASE/system/users -H "$GX" -H 'Content-Type: application/json' -d '{"username":"globex_bob","password":"test123456"}' >/dev/null

# 隔离校验：acme 只看到自己的用户，看不到 globex 的
RA=$(curl -s "$BASE/system/users?pageSize=50" -H "$AC")
check "acme 看到自己的用户" "$RA" 'acme_alice'
if echo "$RA" | grep -q 'globex_bob'; then fail=$((fail+1)); echo "FAIL: acme 越权看到 globex 用户"; else pass=$((pass+1)); echo "PASS: acme 看不到 globex 用户"; fi
RG=$(curl -s "$BASE/system/users?pageSize=50" -H "$GX")
check "globex 看到自己的用户" "$RG" 'globex_bob'
if echo "$RG" | grep -q 'acme_alice'; then fail=$((fail+1)); echo "FAIL: globex 越权看到 acme 用户"; else pass=$((pass+1)); echo "PASS: globex 看不到 acme 用户"; fi

# 业务模块也隔离：各租户建文章
curl -s -X POST $BASE/articles -H "$AC" -H 'Content-Type: application/json' -d '{"title":"Acme 内部公告"}' >/dev/null
curl -s -X POST $BASE/articles -H "$GX" -H 'Content-Type: application/json' -d '{"title":"Globex 内部公告"}' >/dev/null
RA=$(curl -s "$BASE/articles?pageSize=50" -H "$AC")
check "acme 文章可见自己的" "$RA" 'Acme 内部公告'
if echo "$RA" | grep -q 'Globex'; then fail=$((fail+1)); echo "FAIL: acme 越权看到 globex 文章"; else pass=$((pass+1)); echo "PASS: acme 文章与 globex 隔离"; fi

# 主租户看不到任何租户的数据（primary/public 独立）
RP=$(curl -s "$BASE/system/users?pageSize=50" -H "$A")
if echo "$RP" | grep -qE 'acme_alice|globex_bob'; then fail=$((fail+1)); echo "FAIL: 主租户越权看到租户用户"; else pass=$((pass+1)); echo "PASS: 主租户与各租户隔离"; fi

# 直接查 DB 确认 schema 落位
check "DB: tenant_acme schema 存在" "$(su postgres -c "psql -d app -tAc \"SELECT schema_name FROM information_schema.schemata WHERE schema_name='tenant_acme'\"")" 'tenant_acme'
check "DB: acme 用户在其 schema" "$(su postgres -c "psql -d app -tAc \"SELECT username FROM tenant_acme.sys_user WHERE username='acme_alice'\"")" 'acme_alice'
check "DB: acme 用户不在 public" "$(su postgres -c "psql -d app -tAc \"SELECT count(*) FROM public.sys_user WHERE username='acme_alice'\"")" '^0$'

# 停用租户后无法登录
TID=$(curl -s "$BASE/system/tenants?pageSize=50" -H "$A" | python3 -c "import sys,json;print([t['id'] for t in json.load(sys.stdin)['data']['list'] if t['code']=='globex'][0])")
curl -s -X PUT $BASE/system/tenants/$TID/status -H "$A" -H 'Content-Type: application/json' -d '{"status":2}' >/dev/null
check "停用租户后登录被拒" "$(curl -s -H 'X-Tenant: globex' $BASE/auth/captcha)" '租户已停用'

echo "----------------------------------------"
echo "通过 $pass 项, 失败 $fail 项"
exit $fail
