## 安裝



## 說明

URL 格式: /v1/<type>/

type 可為任意字串，目前暫定為 vm, kube

## API

- GET /v1/vm/

列出 type 為 vm 的 queue，其 queue 裡所有的 job

Response:
```json
{
  "data": [
    {
      "ID": 7,
      "CreatedAt": "2018-08-27T11:27:08.564319181+08:00",
      "UpdatedAt": "2018-08-27T11:30:02.250197186+08:00",
      "DeletedAt": null,
      "RequestID": "d",
      "Type": "vm",
      "OwnerID": "y",
      "Data": "{\"host_aggregate\":\"host1\",\"instances\":[{\"Disk\":1,\"Memory\":1,\"VCPU\":1}],\"project\":\"project id\"}",
      "Callback": "abc",
      "Priority": 2756167942488154,
      "Status": "running"
    }
  ],
  "ok": true
}
```

- GET /v1/vm/?owner_id=x&status=queued

列出 type 為 vm 的 queue，其 owner_id 為 x，且 status 為 queued 的所有 job


```json
{
  "data": [
    {
      "ID": 7,
      "CreatedAt": "2018-08-27T11:27:08.564319181+08:00",
      "UpdatedAt": "2018-08-27T11:30:02.250197186+08:00",
      "DeletedAt": null,
      "RequestID": "d",
      "Type": "vm",
      "OwnerID": "x",
      "Data": "{\"host_aggregate\":\"host1\",\"instances\":[{\"Disk\":1,\"Memory\":1,\"VCPU\":1}],\"project\":\"project id\"}",
      "Callback": "abc",
      "Priority": 2756167942488154,
      "Status": "queue"
    }
  ],
  "ok": true
}
```

- GET /v1/vm/?owner_id=y

列出 type 為 vm 的 queue，其 owner_id 為 y 的所有 job
```json
{
  "data": [
    {
      "ID": 7,
      "CreatedAt": "2018-08-27T11:27:08.564319181+08:00",
      "UpdatedAt": "2018-08-27T11:30:02.250197186+08:00",
      "DeletedAt": null,
      "RequestID": "d",
      "Type": "vm",
      "OwnerID": "x",
      "Data": "{\"host_aggregate\":\"host1\",\"instances\":[{\"Disk\":1,\"Memory\":1,\"VCPU\":1}],\"project\":\"project id\"}",
      "Callback": "abc",
      "Priority": 2756167942488154,
      "Status": "queued"
    }
  ],
  "ok": true
}
```

- POST /v1/vm/

在 vm queue 裡，建立新的 job。request_id 為 client 提供，必須不可重複。
owner_id 在 openstack 裡為 project_id。callback 為不可重複 url ，當資源滿足
job 的需求時，則會呼叫 callback url 通知 client。data 是一個字典，其內容因
type 而異。在 OpenStack 情境裡，data 有兩個 key，"instances" 和 "project"。
project 存放 project_id ，而 instances 是一個陣列，存放各個 instance 所需的資源

request:
```json
{
  "request_id": "d",
  "owner_id": "y",
  "data": {
    "instances": [
      {
        "VCPU": 1,
        "Memory": 1,
        "Disk": 1
      }
    ],
    "project": "project id"
  },
  "callback": "abc"
}
```
response:
```json
{
  "data": {
    "ID": 8,
    "CreatedAt": "2018-08-29T14:49:03.288817904+08:00",
    "UpdatedAt": "2018-08-29T14:49:03.288817904+08:00",
    "DeletedAt": null,
    "RequestID": "d",
    "Type": "vm",
    "OwnerID": "y",
    "Data": "{\"host_aggregate\":\"host1\",\"instances\":[{\"Disk\":1,\"Memory\":1,\"VCPU\":1}],\"project\":\"project id\"}",
    "Callback": "abc",
    "Priority": 2941082666988554,
    "Status": "queued"
  },
  "ok": true
}
```

- GET /v1/vm/d/

在 type 為 vm 的 queue 裡，顯示 request_id 為 d 的 job 資訊

response:
```json
{
  "data": {
    "ID": 7,
    "CreatedAt": "2018-08-27T11:27:08.564319181+08:00",
    "UpdatedAt": "2018-08-27T11:30:02.250197186+08:00",
    "DeletedAt": null,
    "RequestID": "d",
    "Type": "vm",
    "OwnerID": "y",
    "Data": "{\"host_aggregate\":\"host1\",\"instances\":[{\"Disk\":1,\"Memory\":1,\"VCPU\":1}],\"project\":\"project id\"}",
    "Callback": "abc",
    "Priority": 2756167942488154,
    "Status": "running"
  },
  "ok": true
}
```

- POST /v1/vm/a

在 type 為 vm 的 queue 裡，修改 request_id 為 a 的 job 資訊。
將 priority 的值 83334。如果需要將 job 移到 queue 的最前面，
則將 priority 設為比 queue 裡面最小的值減1。如果需要將 job 移到
queue 的最後面，則將 priority 設為比 queue 裡面最大的值加1。
如果需要將 job 移到兩個 job 之間，則取兩個 job 的平均值。

request:
```json
{
  "priority": 83334
}
```
response:
```json
{
  "data": {},
  "ok": true
}
```
