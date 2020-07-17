

客户端请求

```
GET /hijack HTTP/1.1
Host: localhost:8100
Connection: Upgrade
Pragma: no-cache
Cache-Control: no-cache
Upgrade: websocket
Sec-WebSocket-Version: 13
Sec-WebSocket-Key: irY1r2rL2oHMaokjREDMRg==
```

收到回复

```
HTTP/1.1 101 Switching Protocols
Server: TornadoServer/6.0.4
Date: Thu, 16 Jul 2020 11:11:28 GMT
Upgrade: websocket
Connection: Upgrade
Sec-Websocket-Accept: sMcuNsRG3799vV4QHkKYS/71RsE=
```