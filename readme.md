
## Overview

一个服务配置对应若干套餐，一般一个预付费套餐和一个后付费套餐，或者只有两者之一。

用户选择一个服务配置+一个套餐，生成一个订单，开启一个服务实例。一个订单对应一个服务实例。

## 套餐属性

套餐属性：
* plan type: 目前只支持订单有效期前7天扣费
* price: cost per circle
* circle: 扣费周期类型，目前之支持自然月

订单对应的服务实例的配置和套餐不可更改。若需更改，需创建一个新的订单并终止老的订单。

## APIs

### POST /usageapi/v1/orders

管理员（创建一个服务实例的时候）创建一个订单。

Body Parameters (json):
```
project: 不可省略。
planId: 套餐Id，不可省略。 (todo: need Plan.PlanType, Plan.ConsumeType)
creator: who made this order，可省略，如果不省略，auth user必须为管理员。
```

Return Result (json):
```
code: 返回码
msg: 返回信息
orderId: 订单号。
```

### PUT /usageapi/v1/orders/{orderId}

管理员（修改一个服务实例的时候）修改一个订单。

Path Parameters:
```
orderId: 订单号。
```

Body Parameters (json):
```
action: 目前只支持cancel
project: 帐户Id，不可省略，作校验用。
```

Return Result (json):
```
code: 返回码
msg: 返回信息
```

### GET /usageapi/v1/orders/{orderId}?project={project}

(一般情况下，用户不应该调用这个接口，用户看到的应该是服务实例。一个服务实例对应一个订单)

1. 管理员查询任何一个订单详情。
1. 当前用户查询自己帐户的一个订单详情。

Path Parameters:
```
orderId: 订单号。
```

Query Parameters:
```
project: 不可省略，作校验用。
```

Return Result (json):
```
code: 返回码
msg: 返回信息
data.order
data.order.id
```

### GET /usageapi/v1/orders?project={project}&status={status}&orderby={orderby}

(一般情况下，用户不应该调用这个接口，用户看到的应该是服务实例列表。每个服务实例对应一个订单)

1. 管理员查询任何帐户的订单列表。
1. 当前用户查询自己帐户的订单列表。

Query Parameters:
```
project: 被不可省略，作校验用。
status: 订单状态。consuming|ended。可以缺省，表示所有订单。
page: 第几页。可选。最小值为1。默认为1。
size: 每页最多返回多少条数据。可选。最小为1，最大为100。默认为30。
```

Return Result (json):
```
code: 返回码
msg: 返回信息
data.total
data.orders
data.orders[0].id
...

```

### GET /usageapi/v1/usages?project={project}&order={orderId}

1. 管理员查询任何订单的历史消费记录。
1. 当前用户查询自己订单的历史消费记录。

Query Parameters:
```
project: 被查询的帐户，不可省略，作校验用。
orderId: 订单号。可省略，表示project内的所有订单。
page: 第几页。可选。最小值为1。默认为1。
size: 每页最多返回多少条数据。可选。最小为1，最大为100。默认为30。
```

Return Result (json):
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

## 数据库设计

```
CREATE TABLE IF NOT EXISTS DF_PURCHASE_ORDER
(
   ORDER_ID           VARCHAR(64) NOT NULL,
   ACCOUNT_ID         VARCHAR(64) NOT NULL COMMENT 'may be project',
   REGION             VARCHAR(4) NOT NULL COMMENT 'for query',
   QUANTITIES         INT DEFAULT 1 COMMENT 'for postpay only',
   PLAN_ID            VARCHAR(64) NOT NULL,
   PLAN_TYPE          VARCHAR(2) NOT NULL COMMENT 'for query',
   START_TIME         DATETIME,
   END_TIME           DATETIME COMMENT 'invalid when status is consuming',
   NEXT_CONSUME_TIME  DATETIME COMMENT 'when to charge next time',
   LAST_CONSUME_ID    INT DEFAULT 0 COMMENT 'charging times',
   STATUS             TINYINT NOT NULL COMMENT 'pending, consuming, ending, ended',
   CREATOR            VARCHAR(64) NOT NULL COMMENT 'who made this order',
   PRIMARY KEY (ORDER_ID)
)  DEFAULT CHARSET=UTF8;
```

对后付费，消费报表:
```
CREATE TABLE IF NOT EXISTS DF_CONSUMING_REPORT
(
   ORDER_ID           VARCHAR(64) NOT NULL,th, day, etc',
   CONSUME_ID         INT,
   CONSUME_TIME       DATETIME,
   CONSUMING          BIGINT NOT NULL COMMENT 'scaled by 10000',
   ACCOUNT_ID         VARCHAR(64) NOT NULL COMMENT 'for query',
   PLAN_ID            VARCHAR(64) NOT NULL COMMENT 'for query',
   PRIMARY KEY (ORDER_ID, CONSUME_ID)
)  DEFAULT CHARSET=UTF8;
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