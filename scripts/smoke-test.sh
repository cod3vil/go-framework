#!/bin/bash
# P2 端到端冒烟测试
set -u
BASE=http://127.0.0.1:8080/api/v1
pass=0; fail=0
check() { # check <名称> <实际> <期望子串>
  if echo "$2" | grep -q "$3"; then pass=$((pass+1)); echo "PASS: $1";
  else fail=$((fail+1)); echo "FAIL: $1 => $2"; fi
}

# 1. 获取验证码，从 Redis 读取答案（测试专用后门）
CAP=$(curl -s $BASE/auth/captcha)
CAPID=$(echo "$CAP" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['captchaId'])")
CODE=$(redis-cli --no-raw get "captcha:$CAPID" | tr -d '"')
check "获取验证码" "$CAP" '"captchaId"'

# 2. 错误密码登录 → 报错并计入失败
R=$(curl -s -X POST $BASE/auth/login -H 'Content-Type: application/json' \
  -d "{\"username\":\"admin\",\"password\":\"wrong\",\"captchaId\":\"$CAPID\",\"captchaCode\":\"$CODE\"}")
check "错误密码被拒绝" "$R" '用户名或密码错误'

# 3. 正确登录（新验证码）
CAPID=$(curl -s $BASE/auth/captcha | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['captchaId'])")
CODE=$(redis-cli --no-raw get "captcha:$CAPID" | tr -d '"')
R=$(curl -s -X POST $BASE/auth/login -H 'Content-Type: application/json' \
  -d "{\"username\":\"admin\",\"password\":\"admin123\",\"captchaId\":\"$CAPID\",\"captchaCode\":\"$CODE\"}")
check "管理员登录成功" "$R" '"accessToken"'
TOKEN=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['accessToken'])")
REFRESH=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['refreshToken'])")
AUTH="Authorization: Bearer $TOKEN"

# 4. userinfo：角色/权限/菜单树
R=$(curl -s $BASE/auth/userinfo -H "$AUTH")
check "userinfo 返回角色" "$R" '"roles":\["admin"\]'
check "userinfo 返回菜单树" "$R" '系统管理'
check "userinfo 返回权限标识" "$R" '\*:\*:\*'

# 5. 未带 token 访问受保护接口 → 401
R=$(curl -s $BASE/system/users)
check "无令牌返回未认证" "$R" '"code":1001'

# 6. 管理员访问用户列表
R=$(curl -s "$BASE/system/users?page=1&pageSize=10" -H "$AUTH")
check "用户列表分页" "$R" '"total":1'

# 7. 创建测试用户 tester（角色: common id=2）
R=$(curl -s -X POST $BASE/system/users -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"username":"tester","password":"test123456","nickname":"测试员","roleIds":[2]}')
check "创建用户" "$R" '"username":"tester"'
TESTER_ID=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")

# 8. tester 登录（验证码）
CAPID=$(curl -s $BASE/auth/captcha | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['captchaId'])")
CODE=$(redis-cli --no-raw get "captcha:$CAPID" | tr -d '"')
R=$(curl -s -X POST $BASE/auth/login -H 'Content-Type: application/json' \
  -d "{\"username\":\"tester\",\"password\":\"test123456\",\"captchaId\":\"$CAPID\",\"captchaCode\":\"$CODE\"}")
TTOKEN=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['accessToken'])")
check "tester 登录成功" "$R" '"accessToken"'

# 9. tester 无策略访问用户列表 → 403
R=$(curl -s $BASE/system/users -H "Authorization: Bearer $TTOKEN")
check "无权限返回 403" "$R" '"code":1003'

# 10. 管理员给 common 角色授予 GET /system/users 权限（casbin 实时生效）
R=$(curl -s -X PUT $BASE/system/roles/2/apis -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"apis":[{"path":"/api/v1/system/users","method":"GET"}]}')
check "授予角色 API 权限" "$R" '"code":0'
R=$(curl -s $BASE/system/users -H "Authorization: Bearer $TTOKEN")
check "授权后可访问" "$R" '"total":2'

# 11. 但 POST 仍被拒
R=$(curl -s -X POST $BASE/system/users -H "Authorization: Bearer $TTOKEN" -H 'Content-Type: application/json' -d '{"username":"x","password":"12345678"}')
check "未授权方法仍 403" "$R" '"code":1003'

# 12. 刷新令牌（轮换）
R=$(curl -s -X POST $BASE/auth/refresh -H 'Content-Type: application/json' -d "{\"refreshToken\":\"$REFRESH\"}")
check "刷新令牌" "$R" '"accessToken"'
R2=$(curl -s -X POST $BASE/auth/refresh -H 'Content-Type: application/json' -d "{\"refreshToken\":\"$REFRESH\"}")
check "旧刷新令牌已失效" "$R2" '"code":1001'

# 13. 登出 → 原 access token 进黑名单
curl -s -X POST $BASE/auth/logout -H "$AUTH" > /dev/null
R=$(curl -s $BASE/auth/userinfo -H "$AUTH")
check "登出后令牌失效" "$R" '"code":1001'

# 14. 菜单树、部门树、API 清单、角色菜单绑定
# 重新登录管理员
CAPID=$(curl -s $BASE/auth/captcha | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['captchaId'])")
CODE=$(redis-cli --no-raw get "captcha:$CAPID" | tr -d '"')
TOKEN=$(curl -s -X POST $BASE/auth/login -H 'Content-Type: application/json' \
  -d "{\"username\":\"admin\",\"password\":\"admin123\",\"captchaId\":\"$CAPID\",\"captchaCode\":\"$CODE\"}" \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['accessToken'])")
AUTH="Authorization: Bearer $TOKEN"
check "菜单树" "$(curl -s $BASE/system/menus/tree -H "$AUTH")" '用户管理'
check "部门树" "$(curl -s $BASE/system/depts/tree -H "$AUTH")" '总公司'
check "API清单" "$(curl -s $BASE/system/apis -H "$AUTH")" '/api/v1/system/roles/:id/apis'
R=$(curl -s -X PUT $BASE/system/roles/2/menus -H "$AUTH" -H 'Content-Type: application/json' -d '{"menuIds":[1,2]}')
check "绑定角色菜单" "$R" '"code":0'
check "读取角色菜单" "$(curl -s $BASE/system/roles/2/menus -H "$AUTH")" '\[1,2\]'

# 15. 登录日志已记录
check "登录日志" "$(su postgres -c "psql -d app -tAc 'SELECT count(*) FROM sys_login_log;'")" '[1-9]'

echo "----------------------------------------"
echo "通过 $pass 项, 失败 $fail 项"
exit $fail
