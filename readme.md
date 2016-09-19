
## Overview

一个服务配置对应若干套餐，一般一个预付费套餐和一个后付费套餐，或者只有两值之一。

用户选择一个服务配置+一个套餐，生成一个订单，开启一个服务实例。一个订单对应一个服务实例。

订单对应的服务实例的配置和套餐可能更改。

本模块不维护服务配置、套餐和服务实例信息。
 * 套餐模块需提供一个接口供本模块获取所有套餐信息。
 * 本模块将每隔数小时查询更新一次套餐信息。
 * 套餐的价格变动最好提前至少一个计费周期发布。

本模块负责消费报表的生成，包括每个服务实例每个计费周期的开销和每个帐户每个计费周期的开销。

本模块也将提供一个帐户消费速度查询接口。

## API设计

### GET /usageapi/v1/usages?user={userId}&order={orderId}&starttime={startTime}&endtime={endTime}

```
order参数将压制user参数
```

### GET /usageapi/v1/speed?user={userId}&order={orderId}&time={time}

```
order参数将压制user参数
```

### POST /usageapi/v1/orders

```
create order

{
    userId
    serviceId
    planId
    startTime
    endTime
}
```

### PUT /usageapi/v1/orders/{orderId}

```
create order

{
    action: cancel | renew | modify
    planId: for modify only
    endTime: for renew only
}
```

### GET /usageapi/v1/orders/{orderId}

get order

### GET /usageapi/v1/orders

get orders


## 数据库设计

```
CREATE TABLE IF NOT EXISTS DF_PURCHASE_ORDER
(
   ORDER_ID           VARCHAR(32) NOT NULL,
   TYPE               TINYINT NOT NUL COMMENT 'prepay, postpay. etc',
   USER               VARCHAR(120) NOT NULL,
   SERVICE_ID         ,
   QUANTITIES ,
   PLAN_ID            ,
   START_TIME         ,
   END_TIME           ,
   STATUS             ,
   PRIMARY KEY (ORDER_ID)
)  DEFAULT CHARSET=UTF8;

CREATE TABLE IF NOT EXISTS DF_USAGE_HISTORY
(
   USAGE_ID           BIGINT NOT NULL AUTO_INCREMENT,
   ORDER_ID           VARCHAR(32) NOT NULL,
   ORDER_DETAILS      ,
   PLAN_DETAILS       ,
   START_TIME           ,
   DURATION
   CONSUMING          ,
   PRIMARY KEY (USAGE_ID)
)  DEFAULT CHARSET=UTF8;
```

## 套餐在内存中的表示

```
PlanPrice {
    UniqueID
    Money
    Duration
    StartTime
}
Plan {
    PlanID
    Name
    Info
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