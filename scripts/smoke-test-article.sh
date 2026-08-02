#!/bin/bash
# 示例业务模块 article 端到端冒烟测试：验证业务模块经框架接入后 CRUD + 鉴权可用。
set -u
BASE=http://127.0.0.1:8080/api/v1
pass=0; fail=0
check() { if echo "$2" | grep -q "$3"; then pass=$((pass+1)); echo "PASS: $1";
  else fail=$((fail+1)); echo "FAIL: $1 => $2"; fi; }

# admin 登录（验证码从 Redis 取答案）
capid=$(curl -s $BASE/auth/captcha | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['captchaId'])")
code=$(redis-cli --no-raw get "captcha:$capid" | tr -d '"')
TOKEN=$(curl -s -X POST $BASE/auth/login -H 'Content-Type: application/json' \
  -d "{\"username\":\"admin\",\"password\":\"admin123\",\"captchaId\":\"$capid\",\"captchaCode\":\"$code\"}" \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['accessToken'])")
AUTH="Authorization: Bearer $TOKEN"

# 表已由模块 AutoMigrate 创建
check "biz_article 表已建" "$(su postgres -c "psql -d app -tAc \"SELECT to_regclass('biz_article')\"")" 'biz_article'

# 未认证被拒
check "未认证访问被拒" "$(curl -s $BASE/articles)" '"code":1001'

# admin (超管) 创建文章
R=$(curl -s -X POST $BASE/articles -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"title":"框架发布公告","author":"张三","content":"go-framework v0.1.0 发布","status":2}')
check "创建文章" "$R" '"title":"框架发布公告"'
AID=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")

# 列表
check "文章列表分页" "$(curl -s "$BASE/articles?page=1&pageSize=10" -H "$AUTH")" '"total":1'
check "标题模糊查询" "$(curl -s "$BASE/articles?title=公告" -H "$AUTH")" '框架发布公告'

# 详情（应自增浏览量）
curl -s $BASE/articles/$AID -H "$AUTH" >/dev/null
check "详情浏览量自增" "$(curl -s $BASE/articles/$AID -H "$AUTH")" '"views":2'

# 更新
curl -s -X PUT $BASE/articles/$AID -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"title":"框架发布公告(修订)","status":1}' >/dev/null
check "更新文章" "$(curl -s $BASE/articles/$AID -H "$AUTH")" '修订'

# 业务错误码：不存在的文章
check "业务错误码 10001" "$(curl -s $BASE/articles/99999 -H "$AUTH")" '"code":10001'

# 写操作已被系统审计日志记录
sleep 1
check "文章写操作已审计" "$(curl -s "$BASE/system/oper-logs?path=articles&pageSize=10" -H "$AUTH")" '/api/v1/articles'

# 普通用户无权限（common 角色未分配 article 权限）→ 403
capid=$(curl -s $BASE/auth/captcha | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['captchaId'])")
code=$(redis-cli --no-raw get "captcha:$capid" | tr -d '"')
curl -s -X POST $BASE/system/users -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"username":"editor1","password":"test123456","roleIds":[2]}' >/dev/null
capid=$(curl -s $BASE/auth/captcha | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['captchaId'])")
code=$(redis-cli --no-raw get "captcha:$capid" | tr -d '"')
UTOKEN=$(curl -s -X POST $BASE/auth/login -H 'Content-Type: application/json' \
  -d "{\"username\":\"editor1\",\"password\":\"test123456\",\"captchaId\":\"$capid\",\"captchaCode\":\"$code\"}" \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['accessToken'])")
check "普通用户无权限被拒" "$(curl -s $BASE/articles -H "Authorization: Bearer $UTOKEN")" '"code":1003'

# 授予 common 角色 GET /articles 权限后可访问（Casbin 实时生效）
curl -s -X PUT $BASE/system/roles/2/apis -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"apis":[{"path":"/api/v1/articles","method":"GET"}]}' >/dev/null
check "授权后可访问" "$(curl -s $BASE/articles -H "Authorization: Bearer $UTOKEN")" '"total":1'

# 删除
check "删除文章" "$(curl -s -X DELETE $BASE/articles/$AID -H "$AUTH")" '"code":0'

# OpenAPI 应包含业务模块路径
check "Swagger 含业务路径" "$(curl -s http://127.0.0.1:8080/swagger/doc.json)" '/api/v1/articles/{id}'

echo "----------------------------------------"
echo "通过 $pass 项, 失败 $fail 项"
exit $fail
