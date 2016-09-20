
## Overview

一个服务配置对应若干套餐，一般一个预付费套餐和一个后付费套餐，或者只有两者之一。

用户选择一个服务配置+一个套餐，生成一个订单，开启一个服务实例。一个订单对应一个服务实例。

订单对应的服务实例的配置和套餐可能更改。

本模块不维护服务配置、套餐和服务实例信息。
 * 套餐模块需提供一个接口供本模块获取所有套餐信息。
 * 本模块将每隔数小时查询更新一次套餐信息。
 * 套餐的价格变动最好提前至少一个计费周期发布。

本模块负责消费报表（消费记录）的生成，即每个订单（服务实例）每个计费周期的消费额。

本模块也将提供一个帐户消费速度查询接口。

概念：
* 订单计费步长: 消费阶梯递增时间单位。
* 最短消费时间: 必须为消费步长的整数倍数。
* 消费报表的时间间隔: 必须为消费步长的整数倍数，必须大于等于最短消费时间。

## APIs

### GET /usageapi/v1/usages?order={orderId}&timestep={timeStep}&starttime={startTime}&endtime={endTime}

1. 管理员查询任何订单的历史消费记录。
1. 当前用户查询自己订单的历史消费记录。

Query Parameters:
```
orderId: 订单号。可省略，表示所有订单。
timeStep: day|month，至少支持月，是否支持天视订单计费步长大小而定。
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
data.results[0].orders
data.results[0].orders[0].consuming
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

### POST /usageapi/v1/orders

管理员（创建一个服务实例的时候）创建一个订单。

Body Parameters (json):
```
accountId
serviceId
planId
startTime
usageDuration: 使用时长，只对预付费付费有效。
```

### PUT /usageapi/v1/orders/{orderId}

管理员（修改一个服务实例的时候）修改一个订单。

Path Parameters:
```
orderId: 订单号。
```

Body Parameters (json):
```
action: stop | renew | modify
planId: for modify only
endTime: for renew only
```

### GET /usageapi/v1/orders/{orderId}

1. 管理员查询任何一个订单详情。
1. 当前用户查询自己帐户的一个订单详情。

Path Parameters:
```
orderId: 订单号。
```

Return Result:
```
code: 返回码
msg: 返回信息
data.order
data.order.id
```

### GET /usageapi/v1/orders?account={accountId}

1. 管理员查询任何帐户的订单列表。
1. 当前用户查询自己帐户的订单列表。

Query Parameters:
```
accountId: 被查询的帐户。不可省略。
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

## 数据库设计

```
CREATE TABLE IF NOT EXISTS DF_PURCHASE_ORDER
(
   ORDER_ID           VARCHAR(64) NOT NULL,
   TYPE               TINYINT NOT NULL COMMENT 'prepay, postpay. etc',
   ACCOUNT_ID         VARCHAR(64) NOT NULL,
   SERVICE_ID         VARCHAR(64) NOT NULL,
   QUANTITIES         INT DEFAULT 1,
   PLAN_ID            VARCHAR(64) NOT NULL,
   START_TIME         DATETIME,
   END_TIME           DATETIME,
   CANCEL_TIME        DATETIME,
   STATUS             TINYINT NOT NULL,
   PRIMARY KEY (ORDER_ID)
)  DEFAULT CHARSET=UTF8;

CREATE TABLE IF NOT EXISTS DF_CONSUMING_HISTORY
(
   ORDER_ID           VARCHAR(64) NOT NULL,
   START_TIME         VARCHAR(20) NOT NULL,
   DURATION           TINYINT NOT NULL,
   CONSUMING          BIGINT NOT NULL,
   PRIMARY KEY (USAGE_ID, START_TIME, DURATION)
)  DEFAULT CHARSET=UTF8;
```

## 套餐在内存中的表示

```golang
PlanPrice {
    UniqueID  string
    Money     float64
    Duration  time.Duration
    StartTime time.Time
}

Plan {
    PlanID string
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