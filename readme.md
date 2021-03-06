
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

### POST /usageapi/v1/orders?drytry=[0|1]

用户创建一个订单。

Query Parameters:
```
drytry: 0|1, 如果为1，将不真正生成订单，而只是返回创建订单所需金额。
```

Body Parameters (json):
```
namespace: 可省略，默认为当前用户名称
plan_id: 套餐Id，不可省略。
parameters:
parameters.resource_name: 某些类型的套餐需要。pvc name / bsi name / etc.
```

Return Result (json):
```
code: 返回码
msg: 返回信息
data.money: 购买金额
data.order.id: （数据库中的key字段，整型，非uuid）
data.order.order_id: uuid
data.order.namespace
data.order.region
data.order.quantities
data.order.plan_id
data.order.start_time
data.order.endTime: 只有订单已经被终止的时候存在
data.order.status: "pending" | "consuming" | "ended"。
data.order.creator
data.order.resource_name: 订单对应的资源名称，并非所有订单都此字段。对于volume，其为volume_name；对于bsi，其为bsi_name。
```

### PUT /usageapi/v1/orders/{order_id}

用户修改一个订单。

Path Parameters:
```
order_id: 订单号（数据库中的key字段，整型，非uuid）。
```

Body Parameters (json):
```
action: 目前只支持cancel
namespace: 可省略，默认为当前用户名称
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
order_id: 订单号（数据库中的key字段，整型，非uuid）。
```

Query Parameters:
```
namespace: 可省略，默认为当前用户名称。
```

Return Result (json):
```
code: 返回码
msg: 返回信息
data.order.id
data.order.order_id
data.order.namespace
data.order.region
data.order.quantities
data.order.plan_id
data.order.start_time
data.order.end_time: 只有订单已经被终止的时候存在
data.order.status: "pending" | "consuming" | "ended"
data.order.creator
```

### GET /usageapi/v1/orders?namespace={namespace}&status={status}&region={region}&resource_name={resource_name}&page={page}&size={size}

用户查询订单列表

Query Parameters:
```
namespace: 可省略，默认为当前用户名称。
status: 订单状态。"consuming" | "ended" | "renewalfailed"。可以缺省，表示consuming。
region: 区标识，不可缺省。
resource_name: 资源名，可以省略。一般指定此字段是为了根据resource_name找到订单号。
page: 第几页。可选。最小值为1。默认为1。
size: 每页最多返回多少条数据。可选。最小为1，最大为100。默认为30。
```

Return Result (json):
```
code: 返回码
msg: 返回信息
data.total
data.results
data.results[0].order
data.results[0].order.id
data.results[0].order.order_id
data.results[0].order.namespace
data.results[0].order.region
data.results[0].order.quantities
data.results[0].order.plan_id
data.results[0].order.start_time
data.results[0].order.end_time: 只有订单已经被终止的时候存在
data.results[0].order.status: "consuming" | "ended"
data.results[0].order.creator
...

```

### GET /usageapi/v1/usages?namespace={namespace}&order={order}&region={region}&page={page}&size={size}

当前用户查询历史消费记录。

Query Parameters:
```
namespace: 可省略，默认为当前用户名称。
order: 订单号。可省略，表示namespace内的所有订单。
region: 区标识，不可缺省。
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

see `_db/initdb_v001.sql`

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
    -e  MYSQL_CONFIG_DONT_UPGRADE_TABLES="false" \
    -e  LOG_LEVEL="debug" \
    \
    -e  ENV_NAME_MYSQL_ADDR="BSI_MYSQL_MYSQLFORSERVICEUSAGE_HOST" \
    -e  ENV_NAME_MYSQL_PORT="BSI_MYSQL_MYSQLFORSERVICEUSAGE_PORT" \
    -e  ENV_NAME_MYSQL_DATABASE="BSI_MYSQL_MYSQLFORSERVICEUSAGE_NAME" \
    -e  ENV_NAME_MYSQL_USER="BSI_MYSQL_MYSQLFORSERVICEUSAGE_USERNAME" \
    -e  ENV_NAME_MYSQL_PASSWORD="BSI_MYSQL_MYSQLFORSERVICEUSAGE_PASSWORD"



```
以下3个环境变量已废止：
    -e  DATAFOUNDRY_HOST_ADDR="xxx" \
    -e  DATAFOUNDRY_ADMIN_USER="xxx" \
    -e  DATAFOUNDRY_ADMIN_PASS="xxx" \
使用以下取而代之：
    -e  DATAFOUNDRY_INFO_CN_NORTH_1="addr user pass" \
    -e  DATAFOUNDRY_INFO_CN_NORTH_2="addr user pass" \
```

```
以下3个环境变量已废止：
    PAYMENT_SERVICE_API_SERVER="xxx" 
    PLAN_SERVICE_API_SERVER="xxx" 
    RECHARGE_SERVICE_API_SERVER="xxx" 
使用以下6个取而代之：
    ENV_NAME_DATAFOUNDRYPAYMENT_SERVICE_HOST="DATAFOUNDRYPAYMENT_SERVICE_HOST"
    ENV_NAME_DATAFOUNDRYPAYMENT_SERVICE_PORT="DATAFOUNDRYPAYMENT_SERVICE_PORT"
    ENV_NAME_DATAFOUNDRYPLAN_SERVICE_HOST="DATAFOUNDRYPLAN_SERVICE_HOST"
    ENV_NAME_DATAFOUNDRYPLAN_SERVICE_PORT="DATAFOUNDRYPLAN_SERVICE_PORT"
    ENV_NAME_DATAFOUNDRYRECHARGE_SERVICE_HOST="DATAFOUNDRYRECHARGE_SERVICE_HOST"
    ENV_NAME_DATAFOUNDRYRECHARGE_SERVICE_PORT="DATAFOUNDRYRECHARGE_SERVICE_PORT_8090_TCP"
```

oc bind MysqlForServiceUsage datafoundryserviceusage

oc expose service datafoundryserviceusage --hostname=datafoundry-serviceusage.app.dataos.io

oc start-build datafoundryserviceusage

```

## test

> go test -v -cover $(go list ./... | grep -v /vendor/)
