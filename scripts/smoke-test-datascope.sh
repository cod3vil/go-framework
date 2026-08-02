#!/bin/bash
# 数据权限（数据范围）端到端测试：
# 构造 部门树 + 两个不同部门的编辑用户，验证「仅本人 / 本部门 / 本部门及以下 / 自定义」过滤生效。
set -u
BASE=http://127.0.0.1:8080/api/v1
pass=0; fail=0
check() { if echo "$2" | grep -q "$3"; then pass=$((pass+1)); echo "PASS: $1";
  else fail=$((fail+1)); echo "FAIL: $1 => $2"; fi; }
jqf() { python3 -c "import sys,json;d=json.load(sys.stdin);print(eval(\"d$1\"))"; }

login() { # login <user> <pass> -> token
  local capid code
  capid=$(curl -s $BASE/auth/captcha | jqf "['data']['captchaId']")
  code=$(redis-cli --no-raw get "captcha:$capid" | tr -d '"')
  curl -s -X POST $BASE/auth/login -H 'Content-Type: application/json' \
    -d "{\"username\":\"$1\",\"password\":\"$2\",\"captchaId\":\"$capid\",\"captchaCode\":\"$code\"}" | jqf "['data']['accessToken']"
}

ADMIN=$(login admin admin123); A="Authorization: Bearer $ADMIN"

# 1. 建部门树：研发部(顶级) -> 前端组 / 后端组
RD=$(curl -s -X POST $BASE/system/depts -H "$A" -H 'Content-Type: application/json' -d '{"name":"研发部","parentId":0}' | jqf "['data']['id']")
FE=$(curl -s -X POST $BASE/system/depts -H "$A" -H 'Content-Type: application/json' -d "{\"name\":\"前端组\",\"parentId\":$RD}" | jqf "['data']['id']")
BE=$(curl -s -X POST $BASE/system/depts -H "$A" -H 'Content-Type: application/json' -d "{\"name\":\"后端组\",\"parentId\":$RD}" | jqf "['data']['id']")
check "部门树创建" "$RD-$FE-$BE" '[0-9]'

# 2. 建角色：编辑(本部门数据) 授予 articles 全部权限
ROLE=$(curl -s -X POST $BASE/system/roles -H "$A" -H 'Content-Type: application/json' \
  -d '{"name":"部门编辑","key":"depteditor","dataScope":3}' | jqf "['data']['id']")
check "角色创建(本部门范围)" "$(curl -s "$BASE/system/roles?name=部门编辑" -H "$A")" '"dataScope":3'
curl -s -X PUT $BASE/system/roles/$ROLE/apis -H "$A" -H 'Content-Type: application/json' \
  -d '{"apis":[{"path":"/api/v1/articles","method":"GET"},{"path":"/api/v1/articles","method":"POST"},{"path":"/api/v1/system/users","method":"GET"}]}' >/dev/null

# 3. 两个用户：前端组 alice、后端组 bob，都属于“部门编辑”角色
curl -s -X POST $BASE/system/users -H "$A" -H 'Content-Type: application/json' \
  -d "{\"username\":\"alice\",\"password\":\"test123456\",\"deptId\":$FE,\"roleIds\":[$ROLE]}" >/dev/null
curl -s -X POST $BASE/system/users -H "$A" -H 'Content-Type: application/json' \
  -d "{\"username\":\"bob\",\"password\":\"test123456\",\"deptId\":$BE,\"roleIds\":[$ROLE]}" >/dev/null
ALICE=$(login alice test123456); AL="Authorization: Bearer $ALICE"
BOB=$(login bob test123456); BO="Authorization: Bearer $BOB"

# 4. alice 与 bob 各建一篇文章（DeptID 自动取各自部门）
curl -s -X POST $BASE/articles -H "$AL" -H 'Content-Type: application/json' -d '{"title":"前端组文章"}' >/dev/null
curl -s -X POST $BASE/articles -H "$BO" -H 'Content-Type: application/json' -d '{"title":"后端组文章"}' >/dev/null

# 5. 本部门范围：alice 只看到本部门(前端组)的文章，看不到后端组
R=$(curl -s "$BASE/articles?pageSize=50" -H "$AL")
check "本部门范围-看到本组" "$R" '前端组文章'
if echo "$R" | grep -q '后端组文章'; then fail=$((fail+1)); echo "FAIL: 本部门范围应看不到他组 => $R"; else pass=$((pass+1)); echo "PASS: 本部门范围-看不到他组"; fi

# 6. admin(全部范围) 两篇都能看到
R=$(curl -s "$BASE/articles?pageSize=50" -H "$A")
check "全部范围-两篇都可见(前端)" "$R" '前端组文章'
check "全部范围-两篇都可见(后端)" "$R" '后端组文章'

# 7. 改为“仅本人”：alice 只看到自己创建的
curl -s -X PUT $BASE/system/roles/$ROLE -H "$A" -H 'Content-Type: application/json' \
  -d '{"name":"部门编辑","key":"depteditor","dataScope":5}' >/dev/null
# alice 再建一篇，bob 也建一篇（同前端组无关，仅本人维度）
ALICE=$(login alice test123456); AL="Authorization: Bearer $ALICE"
R=$(curl -s "$BASE/articles?pageSize=50" -H "$AL")
check "仅本人范围-看到自己的" "$R" '前端组文章'
if echo "$R" | grep -q '后端组文章'; then fail=$((fail+1)); echo "FAIL: 仅本人不应看到他人 => $R"; else pass=$((pass+1)); echo "PASS: 仅本人范围-看不到他人"; fi

# 8. 改为“本部门及以下”：给 alice 换到研发部(顶级)，应能看到前端组+后端组
curl -s -X PUT $BASE/system/roles/$ROLE -H "$A" -H 'Content-Type: application/json' \
  -d '{"name":"部门编辑","key":"depteditor","dataScope":4}' >/dev/null
AID=$(curl -s "$BASE/system/users?username=alice" -H "$A" | jqf "['data']['list'][0]['id']")
curl -s -X PUT $BASE/system/users/$AID -H "$A" -H 'Content-Type: application/json' \
  -d "{\"username\":\"alice\",\"nickname\":\"alice\",\"deptId\":$RD,\"roleIds\":[$ROLE],\"status\":1}" >/dev/null
ALICE=$(login alice test123456); AL="Authorization: Bearer $ALICE"
R=$(curl -s "$BASE/articles?pageSize=50" -H "$AL")
check "本部门及以下-看到子部门(前端)" "$R" '前端组文章'
check "本部门及以下-看到子部门(后端)" "$R" '后端组文章'

# 9. 自定义部门：只授权后端组，alice(研发部) 应只看到后端组文章
curl -s -X PUT $BASE/system/roles/$ROLE -H "$A" -H 'Content-Type: application/json' \
  -d "{\"name\":\"部门编辑\",\"key\":\"depteditor\",\"dataScope\":2,\"deptIds\":[$BE]}" >/dev/null
check "自定义部门回显" "$(curl -s $BASE/system/roles/$ROLE/depts -H "$A")" "$BE"
ALICE=$(login alice test123456); AL="Authorization: Bearer $ALICE"
R=$(curl -s "$BASE/articles?pageSize=50" -H "$AL")
check "自定义部门-看到授权组(后端)" "$R" '后端组文章'
if echo "$R" | grep -q '前端组文章'; then fail=$((fail+1)); echo "FAIL: 自定义部门不应看到未授权组 => $R"; else pass=$((pass+1)); echo "PASS: 自定义部门-看不到未授权组"; fi

# 10. 用户列表也受数据范围约束：alice(自定义=后端组) 只能看到后端组的用户(bob)，看不到前端组的自己?
# alice 现在在研发部，自定义仅后端组 -> 看不到自己(研发部)与前端组用户；能看到 bob(后端组)
R=$(curl -s "$BASE/system/users?pageSize=100" -H "$AL")
check "用户列表-数据范围(可见后端组bob)" "$R" '"username":"bob"'
if echo "$R" | grep -q '"username":"admin"'; then fail=$((fail+1)); echo "FAIL: alice 不应看到 admin(顶级部门) => 用户列表越权"; else pass=$((pass+1)); echo "PASS: 用户列表-看不到未授权部门用户"; fi

echo "----------------------------------------"
echo "通过 $pass 项, 失败 $fail 项"
exit $fail
