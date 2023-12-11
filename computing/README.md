# computing part

## HTTP Request / Response

假定计算节点的域名为：`cpgateway.com`

**Greet Request**

主要功能为“握手”相关，一般每一步返回 `[ACK]` 确认该流程走完后才会走下一步。

Route: `http://cpgateway.com/greet`

Method: GET

参数根据不同的握手流程有不同变化和含义：

| type 值 | 描述 |
| --- | --- |
| 0 | 检查租约是否能被接受 |
| 1 | 请求计算节点授权后续的访问 |
| 2 | 验证节点的访问权限 |
| 3 | 在节点上部署服务 |

传入其他的type值会返回response(400): `{"err", "unsupported msg type"}`

> msg_type = 0

example: `http://cpgateway.com/greet?type=0&input=0x1234`

| 参数 | 描述 |
| --- | --- |
| input | 传入记录租约信息的合约地址（目前未完全确定形式，所以随便传） |

response(200): `{"msg": "[ACK] the lease is acceptable"}`

目前接口只走一个流程，不会返回 `[Fail] the lease is not acceptable`。

> msg_type = 1

example: `http://cpgateway.com/greet?type=1&input=0x1234`

| 参数 | 描述 |
| --- | --- |
| input | 目前先传入用户的账户地址，方便走授权流程。 |

response(200): `{"msg": "[ACK] <input value> authorized ok"}`

response(400，如果input为空): `{"msg": "[Fail] no user address is provided"}`

response(500，授权过程中出错): `{"msg": "[Fail] failed to authorized"}`

> msg_type = 2

example: `http://cpgateway.com/greet?type=2&input=0x1234`

| 参数 | 描述 |
| --- | --- |
| input | 用户的账户地址 |

response(200): `{"msg": "[ACK] already authorized"}`

response(400，如果账户地址未完成授权): `{"msg": "[Fail] Failed to verify your account"}`

> msg_type = 3

example: `http://cpgateway.com/greet?type=3&addr=0x1234&input=http://download.yaml`

| 参数 | 描述 |
| --- | --- |
| addr | 用户的账户地址 |
| input | 服务描述文件（yaml）的下载地址（如果 `cmd/main.go` 中的 `test` 参数设为true，则可以直接换成转发的目标host，方便测试，如 `baidu.com`） |

response(200): `{"msg": "[ACK] deployed ok"}`

response(400，如果参数不完整): `{"msg": "[Fail] empty ..."}`

response(400，如果账户地址未完成授权): `{"msg": "[Fail] user is not authorized"}`

**Process Request**

主要功能为计算相关。用户对已部署的服务发起请求，在计算节点上运行的服务会完成相应计算并返回响应。

Route: `http://cpgateway.com/process`

Method: GET（后续大概率会更改成POST）

example: `http://cpgateway.com/process?addr=0x1234`

目前只是demo版测试框架运行情况，`Method`和其他参数的传入还未完全确定。目前只需要保证`addr`参数与前面已经完成授权的账户地址一致即可，后端在验证通过后会自动构建一个`http GET /`的请求访问服务的根路径。

response(200): `{"msg": "[ACK] compute ok", "response": http-response-bytecode}`

response(400): 未授权、未部署服务、计算或请求出错都有可能触发。

## 交互

为了避免任何人都可以调用计算节点的API，有一套简单的交互流程。

1. user需要先与server做一轮握手。user发送自己的地址（身份标识符）给server，server可以判断该地址是否有资格使用自己的API。这一步目的是为了让user确定自己是否能继续后续流程。（`Greet`相关）
2. 后续client发送消息除了用于转发的http request，还需要附带 `api_key` 和 `address` 用于认证。其中 `api_key` 可以只在第一次使用时根据当前时间用私钥签名生成，后续在过期时间内都是有效的，不需要重复生成。

user有两种方式与server交互：

- 通过浏览器直接用http1的方式交互。主要方式，适合大多数用户。
- 通过grpc用http2的方式交互。扩展方式，适合写程序远程调用API的用户。
