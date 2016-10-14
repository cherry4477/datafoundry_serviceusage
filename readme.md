
## Overview

一个服务配置对应若干套餐，一般一个预付费套餐和一个后付费套餐，或者只有两者之一。

用户选择一个服务配置+一个套餐，生成一个订单，开启一个服务实例。一个订单对应一个服务实例。

## 套餐属性

套餐属性：
* plan type: cpuqouta | bsi | ...
* consume type: 扣费方式
* price: cost per circle
* circle: 扣费周期类型，目前之支持自然月，或许可以包含在consume type中

一个套餐或许需要若干参数，在Order中记录这些参数值。

订单对应的服务实例的配置和套餐不可更改。若需更改，需创建一个新的订单并终止老的订单。

## APIs

### POST /usageapi/v1/orders

用户创建一个订单。

Body Parameters (json):
```
namespace: 可省略，默认为当前用户名称
plan_id: 套餐Id，不可省略。
```

Return Result (json):
```
code: 返回码
msg: 返回信息
data.order_id
data.project
data.region
data.quantities
data.plan_id
data.start_time
data.endTime: 只有订单已经被终止的时候存在
data.status: "pending" | "consuming" | "ending" | "ended"。
data.creator
```

### PUT /usageapi/v1/orders/{order_id}

用户修改一个订单。

Path Parameters:
```
order_id: 订单号。
```

Body Parameters (json):
```
action: 目前只支持cancel
namespace: 不可省略，作校验用
```

Return Result (json):
```
code: 返回码
msg: 返回信息
```

### GET /usageapi/v1/orders/{order_id}?namespace={namespace}

用户查询某个订单（status=consuming）

Path Parameters:
```
order_id: 订单号。
```

Query Parameters:
```
namespace: 不可省略，作校验用。
```

Return Result (json):
```
code: 返回码
msg: 返回信息
data.order_id
data.project
data.region
data.quantities
data.plan_id
data.start_time
data.end_time: 只有订单已经被终止的时候存在
data.status: "pending" | "consuming" | "ending" | "ended"
data.creator
```

### GET /usageapi/v1/orders?namespace={namespace}&status={status}&region={region}&page={page}&size={size}

用户查询订单列表

Query Parameters:
```
namespace: 不可省略，作校验用。
status: 订单状态。"consuming" | "ending" | "ended" | "renewalfailed"。可以缺省，表示consuming。
region: 区标识。
page: 第几页。可选。最小值为1。默认为1。
size: 每页最多返回多少条数据。可选。最小为1，最大为100。默认为30。
```

Return Result (json):
```
code: 返回码
msg: 返回信息
data.total
data.results
data.results[0].order_id
data.results[0].namespace
data.results[0].region
data.results[0].quantities
data.results[0].plan_id
data.results[0].start_time
data.results[0].end_time: 只有订单已经被终止的时候存在
data.results[0].status: "consuming" | "ending" | "ended"
data.results[0].creator
...

```

### GET /usageapi/v1/usages?namespace={namespace}&order={order}&region={region}&page={page}&size={size}

当前用户查询历史消费记录。

Query Parameters:
```
namespace: 不可省略，作校验用。
order: 订单号。可省略，表示namespace内的所有订单。
region: 区标识。
page: 第几页。可选。最小值为1。默认为1。
size: 每页最多返回多少条数据。可选。最小为1，最大为100。默认为30。
```

Return Result (json):
```
code: 返回码
msg: 返回信息
data.total
data.results
data.results[0].order_id
data.results[0].namespace
data.results[0].region
data.results[0].time
data.results[0].money
data.results[0].plan_id
...
```

## 数据库设计

订单：
```
CREATE TABLE IF NOT EXISTS DF_PURCHASE_ORDER
(
   ID                 BIGINT NOT NULL AUTO_INCREMENT,
   ORDER_ID           VARCHAR(64) NOT NULL,
   ACCOUNT_ID         VARCHAR(64) NOT NULL COMMENT 'may be project',
   REGION             VARCHAR(4) NOT NULL COMMENT 'for query',
   PLAN_HISTORY_ID    BIGINT NOT NULL COMMENT 'important to retrieve history plan',
   PLAN_ID            VARCHAR(64) NOT NULL,
   PLAN_TYPE          VARCHAR(2) NOT NULL COMMENT 'for query',
   START_TIME         DATETIME,
   END_TIME           DATETIME COMMENT 'invalid when status is consuming',
   DEADLINE_TIME      DATETIME COMMENT 'time to terminate order',
   LAST_CONSUME_ID    INT DEFAULT 0 COMMENT 'charging times',
   EVER_PAYED         TINYINT DEFAULT 0 COMMENT 'LAST_CONSUME_ID > 0',
   RENEW_RETRIES      TINYINT DEFAULT 0 COMMENT 'num renew fails, most 100',
   STATUS             TINYINT NOT NULL COMMENT 'pending, consuming, ending, ended',
   CREATOR            VARCHAR(64) NOT NULL COMMENT 'who made this order',
   PRIMARY KEY (ORDER_ID)
)  DEFAULT CHARSET=UTF8;
```
    ID
    PLAN_HISTORY_ID

消费报表：
```
CREATE TABLE IF NOT EXISTS DF_CONSUMING_HISTORY
(
   ID                 BIGINT NOT NULL COMMENT 'copied from DF_PURCHASE_ORDER.Id',
   ORDER_ID           VARCHAR(64) NOT NULL,
   CONSUME_ID         INT,
   CONSUMING          BIGINT NOT NULL COMMENT 'scaled by 10000',
   CONSUME_TIME       DATETIME,
   DEADLINE_TIME      DATETIME,
   ACCOUNT_ID         VARCHAR(64) NOT NULL COMMENT 'for query',
   REGION             VARCHAR(4) NOT NULL COMMENT 'for query',
   PLAN_ID            VARCHAR(64) NOT NULL COMMENT 'for query',
   PLAN_HISTORY_ID    BIGINT NOT NULL COMMENT 'important to retrieve history plan',
   PRIMARY KEY (ID, ORDER_ID, CONSUME_ID)
)  DEFAULT CHARSET=UTF8;
```
    ID
    DEADLINE_TIME
    PLAN_HISTORY_ID


type Plan struct {
	Plan_id        string    `json:"plan_id,omitempty"`
	Plan_name      string    `json:"plan_name,omitempty"`
	Plan_type      string    `json:"plan_type,omitempty"`
	Specification1 string    `json:"specification1,omitempty"`
	Specification2 string    `json:"specification2,omitempty"`
	Price          float32   `json:"price,omitempty"`
	Cycle          string    `json:"cycle,omitempty"`
	Region         string    `json:"region,omitempty"`
	Create_time    time.Time `json:"creation_time,omitempty"`
	Status         string    `json:"status,omitempty"`
}

## 部署

```
oc new-instance MysqlForServiceUsage --service=Mysql --plan=NoCase

oc new-app --name datafoundryserviceusage https://github.com/asiainfoLDP/datafoundry_serviceusage.git#develop \
    -e  CLOUD_PLATFORM="dataos" \
    \
    -e  DATAFOUNDRY_HOST_ADDR="xxx" \
    -e  DATAFOUNDRY_ADMIN_USER="xxx" \
    -e  DATAFOUNDRY_ADMIN_PASS="xxx" \
    \
    -e  PAYMENT_SERVICE_API_SERVER="xxx" \
    -e  PLAN_SERVICE_API_SERVER="xxx" \
    -e  RECHARGE_SERVICE_API_SERVER="xxx" \
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

## test

> go test -v -cover $(go list ./... | grep -v /vendor/)
