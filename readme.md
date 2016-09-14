
## 

### 模式一

预付费，包时长。

### 模式二

后付费，订阅默认，按时长收费。

```
自动根据使用的services统计费用？

DF_PURCHASE_ORDER表并不需要
```

## 数据库设计

```
CREATE TABLE IF NOT EXISTS DF_PURCHASE_ORDER
(
   ORDER_ID           VARCHAR(32) NOT NULL,
   TYPE                  COMMENT 'prepay, postpay',
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

modify order (cancel, upgrade)

### GET /usageapi/v1/purchaseorders/{orderId}

get order

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
