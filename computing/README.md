# computing part

## 启动参数

- `--version`：打印版本后直接退出。make编译后才会有git的版本信息。
- `--test`：启用测试模式。测试模式下：（1）原本的badgerDB数据库会用内存map代替。（2）deploy中传入的`input`参数将直接作为转发的入口host，而不再是yaml的下载链接。

## HTTP Request / Response

假定计算节点的域名为：`cpgateway.com`

**Greet Request**

主要功能为“握手”相关，一般每一步返回 `[ACK]` 确认该流程走完后才会走下一步。

Route: `http://cpgateway.com/greet`

Method: GET

参数根据不同的握手流程有不同变化和含义：

| type 值 | 描述 |
| --- | --- |
| 0 | 检查租约是否能被接受（链上未实现，暂时为`true`） |
| 1 | 请求计算节点记录用户地址，授权后续的访问（链上未实现，暂时直接传用户地址） |
| 2 | 为有权访问的节点授予cookie，后续将用cookie验证节点身份保证正常使用 |
| 3 | 在节点上部署服务（url的yaml部署或者本地测试部署，打开`--test`选项后url可直接传入口地址如`https://baidu.com`） |

传入其他的type值会返回response(400): `{"err", "unsupported msg type"}`

> msg_type = 0

example: `http://cpgateway.com/greet?type=0&input=0x1234`

| 参数 | 描述 |
| --- | --- |
| input | 传入记录租约信息的合约地址（目前未完全确定形式，所以随便传） |

response(200): `{"msg": "[ACK] the lease is acceptable"}`

目前接口只走一个流程，不会返回 `[Fail] the lease is not acceptable`。

> msg_type = 1

example: `http://cpgateway.com/greet?type=1&input=0x0d2897e7e3ad18df4a0571a7bacb3ffe417d3b06`

| 参数 | 描述 |
| --- | --- |
| input | 目前先传入用户的账户地址，方便后续走授权流程。 |

response(200): `{"msg": "[ACK] <input value> authorized ok"}`

response(400，如果input为空): `{"msg": "[Fail] no user address is provided"}`

response(500，授权过程中出错): `{"msg": "[Fail] failed to authorized"}`

> msg_type = 2

这块与backend的方式比较像。注意传入的时间戳参数有效期目前为1分钟，超过1分钟会被视作无效的时间戳。

example: `http://cpgateway.com/greet?type=2&input=1703757313&addr=0x0d2897e7e3ad18df4a0571a7bacb3ffe417d3b06&sig=0x1fd59e72cd4811467ebb0dfc09d8897c31ad2bde6acf6a1d7fb164f09f69bf127c2f676552d7be0a69ea6b5e8f6012394424f9c0e07e9f22ac8e6d18a62b8a3b01`

| 参数 | 描述 |
| --- | --- |
| input | Unix时间戳（生成签名时的时间） |
| addr | 用户的账户地址 |
| sig | 用户对时间戳的哈希签名 |

response(200): `{"msg": "[ACK] already authorized"}`

response(400，如果账户地址未完成授权或传入了无效签名、过期时间戳): `{"msg": "[Fail] Failed to verify your account"}`

测试白名单码：只需要传入`input=cheat`即可跳过验证，直接获得cookie。

> msg_type = 3

需要携带cookie才能使用。通过`type=2`的流程获取cookie。

example: `http://cpgateway.com/greet?type=3&input=http://download.yaml`

| 参数 | 描述 |
| --- | --- |
| input | 服务描述文件（yaml）的下载地址（如果设置了启动参数`--test`，则可以直接换成转发的目标host，方便测试，如`https://www.baidu.com`） |

response(200): `{"msg": "[ACK] deployed ok"}`

response(400，如果参数不完整): `{"msg": "[Fail] empty ..."}`

response(400，如果账户地址未完成授权): `{"msg": "[Fail] user is not authorized"}`

**Process Request**

主要功能为计算相关。用户对已部署的服务发起请求，在计算节点上运行的服务会完成相应计算并返回响应。需要提前获取cookie，并完成服务的部署。

Route: `http://cpgateway.com/`

Method: 只做请求转发，无论什么`method`都支持。一般可以先用`GET`获取根页面。

example: `http://cpgateway.com/`

访问gateway不需要传参数，会通过cookie确认身份和对应的入口。后续传的其他参数均与部署的服务有关，gateway不会干涉。

response(200): 服务页面或者服务返回的响应数据。

response(400): 未授权、未部署服务、计算或请求出错都有可能触发。

## 交互

为了避免任何人都可以调用计算节点的API，有一套简单的交互流程。

1. user需要先与server做一轮握手。user发送自己的地址（身份标识符）给server，server可以判断该地址是否有资格使用自己的API。这一步目的是为了让user确定自己是否能继续后续流程。（`Greet`相关）
2. 后续client发送消息除了用于转发的http request，还需要附带 `cookie` 用于认证。cookie过期后需要重新申请。

user有两种方式与server交互：

- 通过浏览器直接用http1的方式交互。主要方式，适合大多数用户。
- 通过grpc用http2的方式交互。扩展方式，适合写程序远程调用API的用户。
