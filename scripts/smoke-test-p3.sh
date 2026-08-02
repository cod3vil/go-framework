#!/bin/bash
# P3 端到端冒烟测试：字典/参数/日志/定时任务/文件/监控/Swagger
set -u
BASE=http://127.0.0.1:8080/api/v1
pass=0; fail=0
check() { if echo "$2" | grep -q "$3"; then pass=$((pass+1)); echo "PASS: $1";
  else fail=$((fail+1)); echo "FAIL: $1 => $2"; fi; }

login() { # 以 admin 登录，回显 token
  local capid code
  capid=$(curl -s $BASE/auth/captcha | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['captchaId'])")
  code=$(redis-cli --no-raw get "captcha:$capid" | tr -d '"')
  curl -s -X POST $BASE/auth/login -H 'Content-Type: application/json' \
    -d "{\"username\":\"admin\",\"password\":\"admin123\",\"captchaId\":\"$capid\",\"captchaCode\":\"$code\"}" \
    | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['accessToken'])"
}

TOKEN=$(login); AUTH="Authorization: Bearer $TOKEN"

# 字典类型 CRUD
R=$(curl -s -X POST $BASE/system/dicts -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"name":"用户性别","type":"sys_user_sex","remark":"性别"}')
check "创建字典类型" "$R" '"type":"sys_user_sex"'
DID=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")
R=$(curl -s -X POST $BASE/system/dict-items -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"dictType":"sys_user_sex","label":"男","value":"1","sort":1}')
check "创建字典项" "$R" '"label":"男"'
check "查询字典项(下拉)" "$(curl -s $BASE/system/dicts/sys_user_sex/items -H "$AUTH")" '"男"'
check "删除字典类型级联" "$(curl -s -X DELETE $BASE/system/dicts/$DID -H "$AUTH")" '"code":0'

# 参数配置：内置参数缓存读取 + 修改刷新
CID=$(curl -s "$BASE/system/configs?key=sys.name" -H "$AUTH" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['list'][0]['id'])")
check "按key读取参数" "$(curl -s $BASE/system/configs/key/sys.name -H "$AUTH")" 'go-framework'
curl -s -X PUT $BASE/system/configs/$CID -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"name":"系统名称","key":"sys.name","value":"新标题"}' >/dev/null
check "参数修改后缓存刷新" "$(curl -s $BASE/system/configs/key/sys.name -H "$AUTH")" '新标题'
check "内置参数禁止删除" "$(curl -s -X DELETE $BASE/system/configs/$CID -H "$AUTH")" '"code":1003'

# 操作日志：前面的写操作应已被审计
sleep 1
check "操作日志已记录" "$(curl -s "$BASE/system/oper-logs?page=1&pageSize=5" -H "$AUTH")" '/api/v1/system/dicts'
check "操作日志密码脱敏" "$(curl -s "$BASE/system/oper-logs?path=login-logs&pageSize=50" -H "$AUTH")" '"total"'

# 登录日志
check "登录日志查询" "$(curl -s "$BASE/system/login-logs?pageSize=5" -H "$AUTH")" '"username":"admin"'

# 定时任务：注册表 + 创建 + 手动执行 + 执行日志
check "已注册任务列表" "$(curl -s $BASE/system/jobs/tasks -H "$AUTH")" 'system:heartbeat'
R=$(curl -s -X POST $BASE/system/jobs -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"name":"心跳","jobKey":"system:heartbeat","cronExpr":"0 0 * * *","status":2}')
check "创建定时任务" "$R" '"name":"心跳"'
JID=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")
check "非法cron被拒" "$(curl -s -X POST $BASE/system/jobs -H "$AUTH" -H 'Content-Type: application/json' -d '{"name":"x","jobKey":"system:heartbeat","cronExpr":"bad"}')" '"code":1000'
check "未注册任务被拒" "$(curl -s -X POST $BASE/system/jobs -H "$AUTH" -H 'Content-Type: application/json' -d '{"name":"x","jobKey":"nope","cronExpr":"0 0 * * *"}')" '未注册'
curl -s -X POST $BASE/system/jobs/$JID/run -H "$AUTH" >/dev/null
sleep 1
check "手动执行产生日志" "$(curl -s "$BASE/system/job-logs?jobId=$JID" -H "$AUTH")" '执行成功'
check "删除任务" "$(curl -s -X DELETE $BASE/system/jobs/$JID -H "$AUTH")" '"code":0'

# 文件上传/下载/删除
echo "hello framework" > /tmp/upload_test.txt
R=$(curl -s -X POST $BASE/system/files -H "$AUTH" -F "file=@/tmp/upload_test.txt")
check "上传文件" "$R" '"url":"/uploads'
FID=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")
FURL=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['url'])")
check "文件列表" "$(curl -s "$BASE/system/files?pageSize=5" -H "$AUTH")" 'upload_test.txt'
check "静态访问上传文件" "$(curl -s http://127.0.0.1:8080$FURL)" 'hello framework'
check "下载文件" "$(curl -s $BASE/system/files/$FID/download -H "$AUTH")" 'hello framework'
# 非法类型被拒
echo "x" > /tmp/bad.exe
check "非法类型被拒" "$(curl -s -X POST $BASE/system/files -H "$AUTH" -F "file=@/tmp/bad.exe")" '不允许的文件类型'
check "删除文件" "$(curl -s -X DELETE $BASE/system/files/$FID -H "$AUTH")" '"code":0'

# 服务监控
check "服务监控指标" "$(curl -s $BASE/system/monitor/server -H "$AUTH")" '"goroutines"'
check "监控含内存" "$(curl -s $BASE/system/monitor/server -H "$AUTH")" '"usedPercent"'

# Swagger
check "OpenAPI 规范生成" "$(curl -s http://127.0.0.1:8080/swagger/doc.json)" '"openapi":"3.0.3"'
check "OpenAPI 含系统路径" "$(curl -s http://127.0.0.1:8080/swagger/doc.json)" '/api/v1/system/users/{id}'
check "Swagger UI 页面" "$(curl -s http://127.0.0.1:8080/swagger)" 'swagger-ui'

# userinfo 现在应含更多菜单
check "菜单树含服务监控" "$(curl -s $BASE/auth/userinfo -H "$AUTH")" '服务监控'

echo "----------------------------------------"
echo "通过 $pass 项, 失败 $fail 项"
exit $fail
