
## 

套餐价格的变更必须提前（数个轮询周期）同步到usage模块。

使用的services自动视为订单？ DF_PURCHASE_ORDER表并不需要？

```
套餐在内存中的表示：
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

订单类型：
- 预付费，包时长 (在终止时间之前，服务都有效)
- 后付费，按使用时长一段一段收费（定时轮询）

订单属性：
- 对应的service instance
- 开始时间
- 结束时间（对后付费类型，只在订单结束后有效，表示服务终止时间）
- 类型(预扣费，后付费)

历史帐单：
- 开始时间
- 时长
- TotalMoney
- OrderId
- PlanUniqueId

```

## 数据库设计

```
CREATE TABLE IF NOT EXISTS DF_PURCHASE_ORDER
(
   ORDER_ID           VARCHAR(32) NOT NULL,
   TYPE                  COMMENT 'prepay, postpay. etc',
   USER               VARCHAR(120) NOT NULL,
   SERVICE_ID         ,
   SERVICE_QUANTITIES ,
   PLAN_ID            ,
   BEGIN_TIME         ,
   END_TIME           ,
   STATUS             ,
   PRIMARY KEY (ORDER_ID)
)  DEFAULT CHARSET=UTF8;

CREATE TABLE IF NOT EXISTS DF_PLAN_DETAILS
(
   UNIQUE_ID          ,
   PLAN_ID               COMMENT 'not unique',
   BEGIN_TIME         ,
   NAME               ,
   DETAILS            ,
)  DEFAULT CHARSET=UTF8;

CREATE TABLE IF NOT EXISTS DF_USAGE_HISTORY
(
   USAGE_ID           BIGINT NOT NULL AUTO_INCREMENT,
   ORDER_ID           VARCHAR(32) NOT NULL,
   ORDER_DETAILS      ,
   BEGIN_TIME         ,
   END_TIME           ,
   PAYMENT            ,
   PRIMARY KEY (USAGE_ID)
)  DEFAULT CHARSET=UTF8;
```

## API设计

### GET /usageapi/v1/amount?user={userId}&order={orderId}&starttime={startTime}&endtime={endTime}

order参数可选

### GET /usageapi/v1/speed?user={userId}&order={orderId}&time={time}

order参数可选

### POST /usageapi/v1/purchaseorders

create order

### PUT /usageapi/v1/purchaseorders/{orderId}

modify order (cancel, upgrade, renew, price changes)

### GET /usageapi/v1/purchaseorders/{orderId}

get order

### GET /usageapi/v1/purchaseorders

get orders

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
