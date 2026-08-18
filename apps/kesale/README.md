# KESale

## 定位

KESale 是 CoreKG 平台的**支付与订单管理库**。提供订单创建、支付渠道对接（微信支付）、支付回调处理、订单状态验证和超时订单清理。通过 `Init()` 函数被宿主服务初始化，不独立运行。

## 核心能力

| 能力 | 说明 |
|------|------|
| 订单创建 | Snowflake 风格订单号生成 |
| 支付对接 | 微信支付实现（支付宝预留） |
| 支付回调 | Redis 分布式锁防重复处理 |
| 订单验证 | 每小时 Cron 任务查询真实支付状态 |
| 业务回调 | 解耦的业务完成处理（GlobalHandlers） |

## 代码架构

```
apps/kesale/
├── app.go                   # Init() 入口：初始化支付客户端、注册回调、启动定时任务
├── sale_manager.go          # SaleManager: 创建/查询/回调/验证
├── verify_job.go            # 每小时验证待处理订单
├── callbacks/               # 回调 Handler 接口 + 全局注册
├── client/
│   ├── client.go            # PaymentClient 接口
│   ├── wechat_client.go     # 微信支付实现
│   └── config.go            # 支付配置
├── models/
│   ├── sale/                # 订单 CRUD
│   └── saletype/            # 订单类型 + DB 初始化
├── services/                # 订单业务服务
├── notify/                  # 通知服务
└── utils/                   # 订单号生成
```

## 与其他服务的关系

- 被 `apps/corekg` 和 `apps/kecore` 初始化调用
- 依赖 `yg-go/job` - Cron 定时任务
- 依赖 `yg-go/dbtools/redispool` - Redis 分布式锁
- 微信支付 SDK
