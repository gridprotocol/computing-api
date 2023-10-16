# computing part

## 交互

为了避免任何人都可以调用计算节点的API，有一套简单的交互流程。

1. client需要先与server做一轮握手。client发送自己的地址（身份标识符）给server，server可以判断该地址是否有资格使用自己的API，如果没有则返回 `no permission` 的结果，如果有则返回 `ok` 的结果。这一步目的是为了让client确定自己是否能继续后续流程。
2. 后续client发送消息除了模型的输入，还需要附带 `api_key` 和 `address` 用于认证。其中 `api_key` 可以只在第一次使用时根据当前时间用私钥签名生成，后续在过期时间内都是有效的，不需要重复生成。
3. 对于server，在根据 `address` 对应的公钥验证了 `api_key` 后（包括有效时间），会根据地址cache这个 `api_key`。然后运行模型，根据输入给出输出结果。

其他扩展的部分就落在额外的模型参数设置，如 `temperature` 等，以及返回结果时附带的一些额外的 `metric` 值。