# TaskGo接口文档

## 用户管理

### 1. 用户注册

- URL: /register
- Method: POST
- Auth: N

请求体示例:

```json
{
    "username": "xzw",
    "password": "123456",
    "email": "u202314382@hust.edu.cn",
    "role": 2
}
```

响应体示例:

```json
{
    "code": 200,
    "data": {
        "id": 2,
        "username": "xzw",
        "password": "e10adc3949ba59abbe56e057f20f883e",
        "email": "u202314382@hust.edu.cn",
        "role": 2,
        "created": 1772768230,
        "updated": 0
    },
    "msg": "register success"
}
```

### 2. 用户登录

- URL: /login
- Method: POST
- Auth: N

请求体示例:

```json
{
    "username": "xzw",
    "password": "123456"
}
```

响应体示例:

```json
{
    "code": 200,
    "data": {
        "user": {
            "id": 2,
            "username": "xzw",
            "password": "123456",
            "email": "u202314382@hust.edu.cn",
            "role": 2,
            "created": 1772768230,
            "updated": 0
        },
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MiwiVXNlck5hbWUiOiJ4enciLCJCdWZmZXJUaW1lIjo4NjQwMCwiaXNzIjoieHp3IiwiZXhwIjoxNzczMzczMjg4LCJuYmYiOjE3NzI3Njc0ODgsImlhdCI6MTc3Mjc2ODQ4OH0.Z6GoNFFdgzpCLhmLCvNWA20NWoQdVxBWgfU8Vl04Xyc"	// jwt token
    },
    "msg": "login success"
}
```

## 节点管理

### 1. 删除节点

- URL: /node/delete
- Method: POST
- Auth: Y

请求体示例:

```json
{
    "uuid": "e0c491c5-16fe-11f1-9657-e0c264b3a993"
}
```

响应体示例:

```json
// 删除成功 (删除未运行节点)
{
    "code": 200,
    "data": {},
    "msg": "delete success"
}

// 删除失败 (删除正在运行的节点)
{
    "code": 1000,
    "data": {},
    "msg": "[delete node] failed"
}
```

### 2. 搜索节点

- URL: /node/search
- Method: POST
- Auth: Y

请求体示例:

```json
{
    "page": 1, 
    "page_size": 4,
    "ip": "",
    "uuid": "",
    "up": 0,
    "status": 1
}
```

响应体示例:

```json
{
    "code": 200,
    "data": {
        "list": [
            {
                "id": 8,
                "pid": "6798",
                "ip": "10.10.201.11",
                "hostname": "Archlinux",
                "uuid": "a65a95a5-1920-11f1-bf4f-e0c264b3a993",
                "version": "v1.1.0",
                "status": 1,
                "up": 1772776344,
                "down": 0,
                "task_count": 0
            },
            {
                "id": 7,
                "pid": "6487",
                "ip": "10.10.201.11",
                "hostname": "Archlinux",
                "uuid": "9e71666e-1920-11f1-82ed-e0c264b3a993",
                "version": "v1.1.0",
                "status": 1,
                "up": 1772776331,
                "down": 0,
                "task_count": 0
            },
            {
                "id": 6,
                "pid": "6062",
                "ip": "10.10.201.11",
                "hostname": "Archlinux",
                "uuid": "8a893ff4-1920-11f1-b2b5-e0c264b3a993",
                "version": "v1.1.0",
                "status": 1,
                "up": 1772776297,
                "down": 0,
                "task_count": 0
            }
        ],
        "total": 3,
        "page": 1,
        "page_size": 4
    },
    "msg": "search success"
}
```

## 脚本管理

### 1. 添加或更新脚本

- URL: /script/add
- Method: POST
- Auth: Y

请求体示例:

```json
{
    "id": 0,
    "name": "install_gcc",
    "command": "sudo pacman -Ss gcc",
    "cmd": [],
    "created": 0,
    "updated": 0
}
```

响应体示例:

```json
{
    "code": 200,
    "data": {
        "id": 1,
        "name": "install_gcc",
        "command": "sudo pacman -Ss gcc",
        "created": 1772777831,
        "updated": 0,
        "cmd": []
    },
    "msg": "operate success"
}
```

### 2. 删除指定脚本

- URL: /script/delete
- Method: POST
- Auth: Y

请求体示例:

```json
{
    "ids": [1]
}
```

响应体示例:

```json
{
    "code": 200,
    "data": {},
    "msg": "delete success"
}
```

### 3. 查询指定脚本

- URL: /script/find?id=n
- Method: GET
- Auth: Y

请求体示例: 无

响应体示例:

```json
{
    "code": 200,
    "data": {
        "id": 2,
        "name": "install_gcc",
        "command": "sudo pacman -Ss gcc",
        "created": 1772778192,
        "updated": 0,
        "cmd": null
    },
    "msg": "find success"
}
```

### 4. 分页查询脚本

- URL: /script/search
- Method: POST
- Auth: Y

请求体示例:

```json
{
    "page": 1,
    "page_size": 3,
    "id": 0,
    "name": "install" // 脚本名称模糊匹配
}
```

响应体示例:

```json
{
    "code": 200,
    "data": {
        "list": [
            {
                "id": 2,
                "name": "install_gcc",
                "command": "sudo pacman -Ss gcc",
                "created": 1772778192,
                "updated": 0,
                "cmd": null
            },
            {
                "id": 3,
                "name": "install_gdb",
                "command": "sudo pacman -Ss gdb",
                "created": 1772778202,
                "updated": 0,
                "cmd": null
            },
            {
                "id": 4,
                "name": "install_go",
                "command": "sudo pacman -Ss go",
                "created": 1772778215,
                "updated": 0,
                "cmd": null
            }
        ],
        "total": 4,
        "page": 1,
        "page_size": 3
    },
    "msg": "search success"
}
```

## 任务管理

### 1. 添加或更新任务

- URL: /task/add
- Method: POST
- Auth: Y

请求体格式:

```json
{
  "id": 0,
  "name": "task_echo_hello",
  "command": "echo hello",
  "script_id": [],
  "timeout": 60,
  "retry_times": 0,
  "retry_interval": 10,
  "task_type": 1,	// cmd task
  "http_method": 0,
  "notify_type": 0,
  "status": 0,
  "notify_to": [1],
  "spec": "55 14 * * *",
  "run_on": "",
  "note": "for api test",
  "created": 0,
  "updated": 0,

  "host_name": "",
  "ip": "",
  "cmd": [],

  "allocation": 2
}
```

响应体格式:

```json
{
    "code": 200,
    "data": {
        "id": 1,
        "name": "task_echo_hello",
        "command": "echo hello",
        "script_id": [],
        "timeout": 60,
        "retry_times": 0,
        "retry_interval": 10,
        "task_type": 1,
        "http_method": 0,
        "notify_type": 0,
        "status": 0,
        "notify_to": [
            1
        ],
        "spec": "55 14 * * *",
        "run_on": "8a893ff4-1920-11f1-b2b5-e0c264b3a993",
        "note": "for api test",
        "created": 1772779846,
        "updated": 0,
        "host_name": "",
        "ip": "",
        "cmd": [
            "echo",
            "hello"
        ],
        "allocation": 2
    },
    "msg": "operate success"
}
```

### 2. 删除任务

- URL: /task/delete
- Method: POST
- Auth: Y

请求体示例:

```json
{
    "ids": [1, 2]
}
```

响应体示例:

```json
{
    "code": 200,
    "data": {},
    "msg": "delete success"
}
```

### 3. 查询任务日志

- URL: /task/search
- Method: POST
- Auth: Y

请求体示例:

```json
{
    "page": 1, 
    "page_size": 2,
    "name": "",
    "task_id": 0, 
    "node_uuid": "",
    "sucess": true
}
```

响应体示例:

```json
{
    "code": 200,
    "data": {
        "list": [
            {
                "id": 3,
                "name": "test_fail_http_method",
                "command": "http://localhost:9898/ping",
                "script_id": [],
                "timeout": 60,
                "retry_times": 0,
                "retry_interval": 10,
                "task_type": 2,
                "http_method": 1,
                "notify_type": 0,
                "status": 0,
                "notify_to": [
                    1
                ],
                "spec": "12 15 * * *",
                "run_on": "8a893ff4-1920-11f1-b2b5-e0c264b3a993",
                "note": "for fail http test",
                "created": 1772780119,
                "updated": 1772781019,
                "host_name": "",
                "ip": "",
                "cmd": null
            },
            {
                "id": 4,
                "name": "show_gcc_version",
                "command": "gcc --version",
                "script_id": [],
                "timeout": 60,
                "retry_times": 0,
                "retry_interval": 10,
                "task_type": 1,
                "http_method": 0,
                "notify_type": 0,
                "status": 0,
                "notify_to": [
                    1
                ],
                "spec": "0 16 * * *",
                "run_on": "9e71666e-1920-11f1-82ed-e0c264b3a993",
                "note": "show gcc version",
                "created": 1772782191,
                "updated": 0,
                "host_name": "",
                "ip": "",
                "cmd": null
            }
        ],
        "total": 3,
        "page": 1,
        "page_size": 2
    },
    "msg": "search success"
}
```

### 4. 立即执行任务

- URL: /task/once
- Method: POST
- Auth: Y

请求体格式:

```json
{
    "task_id": 6,
    "node_uuid": "8a893ff4-1920-11f1-b2b5-e0c264b3a993"
}
```

响应体格式:

```json
{
    "code": 200,
    "data": {},
    "msg": "task once success"
}
```

### 5. 查询指定任务

- URL: /task/find?id=n
- Method: GET
- Auth: Y

请求体格式: 无

响应体格式:

```json
{
    "code": 200,
    "data": {
        "id": 4,
        "name": "show_gcc_version",
        "command": "gcc --version",
        "script_id": [],
        "timeout": 60,
        "retry_times": 0,
        "retry_interval": 10,
        "task_type": 1,
        "http_method": 0,
        "notify_type": 0,
        "status": 0,
        "notify_to": [
            1
        ],
        "spec": "0 16 * * *",
        "run_on": "9e71666e-1920-11f1-82ed-e0c264b3a993",
        "note": "show gcc version",
        "created": 1772782191,
        "updated": 0,
        "host_name": "",
        "ip": "",
        "cmd": null
    },
    "msg": "find success"
}
```

## 数据统计

### 1. 查询本日节点和任务运行情况

- URL: /stat/today
- Method: GET
- Auth: Y

请求体格式: 无

响应体格式:

```json
{
    "code": 200,
    "data": {
        "normal_node_count": 3,
        "fail_node_count": 4,
        "task_exc_success_count": 2,
        "task_running_count": 0,
        "task_exc_fail_count": 0
    },
    "msg": "ok"
}
```

### 2. 查询本周节点和任务运行情况

- URL: /stat/week
- Method: GET
- Auth: Y

请求体格式: 无

响应体格式:

```json
{
    "code": 200,
    "data": {
        "success_date_count": [	// 成功任务分组
            {
                "date": "2026-03-06",
                "count": "4"
            }
        ],
        "fail_date_count": []  // 失败任务分组
    },
    "msg": "ok"
}
```

### 3. 查询系统信息

- URL: /stat/sytem?uuid="xxxxxx"
- Method: GET
- Auth: Y

请求体格式: 无

响应体格式:

```json
{
    "code": 200,
    "data": {
        "os": {
            "goos": "linux",
            "numCpu": 20,
            "compiler": "gc",
            "goVersion": "go1.25.5",
            "numGoroutine": 32
        },
        "cpu": {
            "cpus": [
                15.789473684213675,
                0,
                11.11111111108304,
                0,
                21.052631578670237,
                0,
                61.90476190453504,
                0,
                16.666666666792985,
                0,
                29.99999999964757,
                0,
                22.727272727183472,
                14.999999999931788,
                15.78947368415384,
                14.285714285636947,
                5.2631578947179465,
                4.999999999977263,
                5.000000000022737,
                10.000000000045475
            ],
            "cores": 14
        },
        "ram": {
            "usedMb": 6488,
            "totalMb": 31826,
            "usedPercent": 20
        },
        "disk": {
            "usedMb": 165940,
            "usedGb": 162,
            "totalMb": 975737,
            "totalGb": 952,
            "usedPercent": 17
        }
    },
    "msg": "ok"
}
```

















