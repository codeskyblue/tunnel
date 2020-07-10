## 请求的URLlujk

```bash
SERVER_URL=localhost:5000

$ http GET $SERVER_URL/ktunnel/status
# Output: IMOR

# 申请identifier，开放本地的3000端口
$ http GET $SERVER_URL/ktunnel/add localPort==3000
{
    "host": "30.10.88.201",
    "identifier": "sFinQvJrYJ",
    "localPort": 3000,
    "port": 56664
}

# 连接请求
# URL $SERVER_URL/ktunnel/_controlPath/
$ ./client -ident sFinQvJrYJ
```