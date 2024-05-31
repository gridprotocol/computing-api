# 合约相关定义

当前设计版本：1.0

## 激励合约

通过质押获取质押释放收益，用于吸引用户和资金进入。

可以完全使用Memo的那一套质押合约逻辑，不需要改动。后期可以与节点服务次数结合来实现控制收益或释放速度。

开发需求：

- evm类开发需求无。仅需要测试使用。
- 非evm类需要做出相同的功能：质押池完整逻辑与支付mint合约。

工作量：

- evm类仅有部署、测试工作量。预计3天内。
- 非evm链，需要熟悉实现与部署相关知识，实现与evm版本相同的逻辑，并部署测试。预计2周。

## 核心功能合约

为了保证计算框架的正确运作，需要在区块链层部署具有核心功能的合约（或类似逻辑）。

链这一层的主要要求就是：

1. 能有方式去记录与读取内容。
2. 有办法做到代币随时间释放的执行逻辑。
3. （后续扩展）能适配一些门槛机制与信用机制，一般能达到evm功能水准的链理论上都可以做到。

目前EVM合约能够支持这些要求。如果其他链也可以达到类似的效果，那也可以使用其他链接入。（其他非evm链有学习与适应成本）

### 注册合约

（值得注意的是，该合约是为诚实的节点设计与使用的，保证其正常功能的运作，而不至于超出自己的服务能力外。对于恶意节点是没有机制筛选过滤的，恶意节点可以设置任意的资源量，设置错误的IP地址或域名。但合约没有必要对这些节点做复杂且不实际的过滤操作）

计算节点（computing provider）注册、登记节点的资源。后续也可以增加评分等信用机制。

是一个总合约，记录所有的计算节点。用一个`map`形式的结构记录，地址为`key`：

- 价格定位
- 资源量
- 其他扩展（如评分）

该合约不需要有复杂的执行逻辑，只需要承担记录的作用。`value`值不一定为具体的数据，也可以采用其他合约或拓展到其他存储网络（如MEFS/IPFS）上，用`URI`指向具体的数据。

计算节点需要发交易到链上完成注册。

开发需求：只需要有一个map记录，用`set`和`get`的方式能设置、获取记录即可。有一定的权限管理，仅允许`msg.sender`设置自己的条目。

工作量：实现、部署、测试，预计5天内。（非evm需要额外考虑适配时间）

### 租赁合约

约束用户与CP节点行为的租赁订单（Lease）。目前可以先做单份，后续再考虑扩展或多份。

需要记录：

- 租用的资源量
- 质押的代币
- 租期长度/试用期长度

其中后两者结合后，可做释放相关的执行逻辑。保证双方的权益。

> 时间释放逻辑：用户质押代币后，用户和计算节点都随时可以调用withdraw取出自己的份额，在withdraw内会先做时间释放的判断和执行。时间判断可用时间戳或区块高度来做，有一个试用期的判断，试用期内不做分配释放，试用期后根据时间跨度将质押代币逐渐释放给payee。在withdraw的时候会自动调用这部分逻辑，且任何人都无法任意修改余额分配，而是合约内规定好的机制。

另外还需要额外考虑的一点：为了避免女巫攻击（gas费很低的时候该攻击成本也会很低），或者用户无限制地滥用试用期，还得增加一些门槛机制，具体方式后续再考虑（可能用小额质押、信用机制等）。

开发需求：除了与数据记录相关的逻辑，还需要有时间释放的执行逻辑和质押、取回代币的逻辑（如ERC20）。其中质押`deposit`和取回`withdraw`会与一个合约内部记录的`balance map`以及代币合约交互。

工作量：实现、部署、测试，预计2-3周。（非evm需要额外考虑适配时间）

### 工厂合约

工厂合约可以用于生成模板化的租赁合约，方便验证。

而且有两种方式实现：

1. 用户生成合约，cp节点验证。
2. cp节点生成合约，用户登记使用。

不过还需要考虑一些使用过程中的问题。预计会用方式2，也可以减少合约的生成量并配合信用机制。

开发需求：需要先实现租赁合约的逻辑才能增加工厂合约用于生成对应的模板合约。工厂合约理论上没有复杂的逻辑需要实现，只需要传参数完成合约的构造与初始参数的配置。

工作量：实现、部署、测试，预计5天内。（非evm需要额外考虑适配时间）