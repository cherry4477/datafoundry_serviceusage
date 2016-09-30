
## Overview

一个服务配置对应若干套餐，一般一个预付费套餐和一个后付费套餐，或者只有两者之一。

用户选择一个服务配置+一个套餐，生成一个订单，开启一个服务实例。一个订单对应一个服务实例。

## 预付费模式（当前采用的模式）

概念：
* 订单终止时间: 服务已付费至日期。
* 预警时长: 距订单终止时间前多长时间开始预警。
* 可透支时长: 不支持。

对于预付费模式，系统不会自动创建天/月扣费记录。

## 后付费模式（目前未使用）

概念：
* 订单计费步长: 消费阶梯递增时间单位，推荐1小时。
* 消费报表间隔: 推荐1天。
* 最短消费时间: 必须为消费步长的整数倍数。
* 阶梯计价。
* 时变套餐: 套餐中的服务数量或者大小随时间不断变化。

订单对应的服务实例的配置和套餐不可更改。若需更改，需创建一个新的订单。

本模块不维护服务配置、套餐和服务实例信息。后付费模式需要缓存所有套餐信息。
 * 套餐模块需提供一个接口供本模块获取所有套餐信息。
 * 本模块将每隔数小时查询更新一次套餐信息。
 * 套餐的价格变动最好提前至少一个最小消费报表间隔（天）发布，
   套餐的开始时间必须为一个最小消费报表间隔的开始时间。

后付费模式下，本模块将负责消费报表（消费记录）的生成，即每个订单（服务实例）每个计费周期的消费额。

后付费模式下，在系统有效订单量不大的情况下(一万以下)，本模块也（可能）将提供一个帐户当前消费速度查询接口。

## APIs

### POST /usageapi/v1/orders

管理员（创建一个服务实例的时候）创建一个订单。

Body Parameters (json):
```
mode: prepay | postpay
accountId: 
planId: 
duration: 这是一个整数。
```

Return Result:
```
orderId: 订单号。
```

### PUT /usageapi/v1/orders/{orderId}

管理员（修改一个服务实例的时候）修改一个订单。

对于预付费订单：
* renew: 续费
* end: 终止一个订单，订单状态将转为ended。

对于后付费订单：
* cancel: 取消订单，EndTime-StartTime将被确定为订单计费步长的整数倍，并且大于等于最短消费时间，订单状态将转为ending。
* changePlan: 修改套餐，将立即使用老套餐生成一个消费记录。

Path Parameters:
```
orderId: 订单号。
```

Body Parameters (json):
```
action: cancel | changePlan | end | renew
planId: for action==changePlan only, 只对后付费模式有效
endTime: for action==renew only， 只对预付费模式有效，格式为RFC3339。
```

Return Result:
```
orderId: 订单号。如果action==changePlan，可能和输入的订单号不同（老订单stopped，并创建一个新订单）。
```

### GET /usageapi/v1/orders/{orderId}?account={accountId}

1. 管理员查询任何一个订单详情。
1. 当前用户查询自己帐户的一个订单详情。

Path Parameters:
```
orderId: 订单号。
```

Query Parameters:
```
accountId: 被查询的帐户。不可省略。
```

Return Result:
```
code: 返回码
msg: 返回信息
data.order
data.order.id
```

### GET /usageapi/v1/orders?account={accountId}&status={status}&orderby={orderby}

1. 管理员查询任何帐户的订单列表。
1. 当前用户查询自己帐户的订单列表。

Query Parameters:
```
accountId: 被查询的帐户。不可省略。
status: 订单状态。consuming|ended。可以缺省，表示所有订单。
orderby: 排序依据。可选。合法值包括hotness|createtime，默认为hotness。
sortOrder: 排序方向。可选。合法值包括asc|desc，默认为desc。
page: 第几页。可选。最小值为1。默认为1。
size: 每页最多返回多少条数据。可选。最小为1，最大为100。默认为30。
```

Return Result:
```
code: 返回码
msg: 返回信息
data.total
data.orders
data.orders[0].id
...

```

### GET /usageapi/v1/usages?account={accountId}&order={orderId}&timestep={timeStep}&starttime={startTime}&endtime={endTime}

1. 管理员查询任何订单的历史消费记录。
1. 当前用户查询自己订单的历史消费记录。

Query Parameters:
```
accountId: 被查询的帐户。不可省略。
orderId: 订单号。可省略，表示accountId的所有订单。
timeStep: day|month。
startTime: 开始时间
endTime: 结束时间
```

Return Result:
```
code: 返回码
msg: 返回信息
data.total
data.results
data.results[0].time
data.results[0].reports
data.results[0].reports[0].consuming
......
...
```

### GET /usageapi/v1/speed?account={accountId}

1. 管理员查询任何帐户的当前消费速度。
1. 当前用户查询自己帐户的当前消费速度。

当前消费中（尚未终止）的订单的速度总和。

Query Parameters:
```
accountId: 被查询的帐户。不可省略。
```

Return Result:
```
code: 返回码
msg: 返回信息
data.moeny
data.duration
```

## 数据库设计

```
CREATE TABLE IF NOT EXISTS DF_PURCHASE_ORDER
(
   ORDER_ID           VARCHAR(64) NOT NULL,
   MODE               TINYINT NOT NULL COMMENT 'prepay, postpay. etc',
   ACCOUNT_ID         VARCHAR(64) NOT NULL,
   REGION             VARCHAR(4) NOT NULL COMMENT 'for query',
   QUANTITIES         INT DEFAULT 1 COMMENT 'for postpay only',
   PLAN_ID            VARCHAR(64) NOT NULL,
   START_TIME         DATETIME,
   END_TIME           DATETIME,
   LAST_CONSUME_TIME  DATETIME COMMENT 'already payed to this time',
   LAST_CONSUME_ID    INT COMMENT 'for report',
   STATUS             TINYINT NOT NULL COMMENT 'consuming, ended, ending',
   PRIMARY KEY (ORDER_ID)
)  DEFAULT CHARSET=UTF8;
```

对后付费，消费报表:
```
CREATE TABLE IF NOT EXISTS DF_CONSUMING_REPORT
(
   ORDER_ID           VARCHAR(64) NOT NULL,th, day, etc',
   CONSUME_ID         INT,
   START_TIME         DATETIME,
   DURATION           INT NOT NULL COMMENT 'consuming time, in seconds', 
   TIME_TAG           VARCHAR(16) NOT NULL COMMENT '2016-02, 2016-02-28, 2016-02-28-15, etc',
   STEP_TAG           TINYINT NOT NULL COMMENT 'prepayed, mon
   CONSUMING          BIGINT NOT NULL COMMENT 'scaled by 10000',
   ACCOUNT_ID         VARCHAR(64) NOT NULL COMMENT 'for query',
   PLAN_ID            VARCHAR(64) NOT NULL COMMENT 'plan id at the report time',
   PRIMARY KEY (ORDER_ID, CONSUME_ID)
)  DEFAULT CHARSET=UTF8;

```

## 消费记录生成（只对后付费有效）

查找需要生成消费记录的订单: select * from DF_PURCHASE_ORDER where TYPE=postpay and LAST_CONSUME_TIME<'%s'

一个订单被修改的时候，将根据订单的老的属性立即产生一个消费记录
（从而一个时段的消费记录将分裂为多个。可以(推荐)在修改订单的时候重新创建一个订单避免这种情况）。

套餐价格变化最好在消费记录的分割点实施，以避免将一个时段的消费记录分裂为两个。

## 消费速度（只对后付费有效）

假设没有时变套餐。

AccountNumPlans = map[Account]map[Plan]int{}: 对每个帐户维护着一个每个套餐的数量的map。
注意订单的QUANTITIES可能大于1。

速度改变时机:
* 套餐价格变化: 遍历AccountNumPlans
* 产生了新的订单: select * from DF_PURCHASE_ORDER where START_TIME>'%s', 或许使用MQ更好
* 有订单失效了: select * from DF_PURCHASE_ORDER where END_TIME<'%s', 或许使用MQ更好

## 套餐在内存中的表示（只针对后付费）

```golang
PlanPrice {
    UniqueID  string
    Money     float64 // consuming / quantity
    Duration  time.Duration
    StartTime time.Time
}

Plan {
    PlanID string
    Type   string
    Name   string
    Info   string
    Prices []PlanPrice // sorted by StartTime
}
```

## 部署

```
oc new-instance MysqlForServiceUsage --service=Mysql --plan=NoCase

oc new-app --name datafoundryserviceusage https://github.com/asiainfoLDP/datafoundry_serviceusage.git#develop \
    -e  CLOUD_PLATFORM="dataos" \
    \
    -e  DATAFOUNDRY_HOST_ADDR="xxx" \
    \
    -e  ENV_NAME_MYSQL_ADDR="BSI_MYSQL_MYSQLFORSERVICEUSAGE_HOST" \
    -e  ENV_NAME_MYSQL_PORT="BSI_MYSQL_MYSQLFORSERVICEUSAGE_PORT" \
    -e  ENV_NAME_MYSQL_DATABASE="BSI_MYSQL_MYSQLFORSERVICEUSAGE_NAME" \
    -e  ENV_NAME_MYSQL_USER="BSI_MYSQL_MYSQLFORSERVICEUSAGE_USERNAME" \
    -e  ENV_NAME_MYSQL_PASSWORD="BSI_MYSQL_MYSQLFORSERVICEUSAGE_PASSWORD" \
    \
    -e  MYSQL_CONFIG_DONT_UPGRADE_TABLES="false" \
    -e  LOG_LEVEL="debug"

oc bind MysqlForServiceUsage datafoundryserviceusage

oc expose service datafoundryserviceusage --hostname=datafoundry-serviceusage.app.dataos.io

oc start-build datafoundryserviceusage

```