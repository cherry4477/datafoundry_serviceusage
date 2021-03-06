CREATE TABLE IF NOT EXISTS DF_PURCHASE_ORDER
(
   ID                 BIGINT NOT NULL AUTO_INCREMENT,
   ORDER_ID           VARCHAR(64) NOT NULL,
   ACCOUNT_ID         VARCHAR(64) NOT NULL COMMENT 'may be project',
   REGION             VARCHAR(32) NOT NULL COMMENT 'for query',
   PLAN_ID            VARCHAR(64) NOT NULL,
   PLAN_TYPE          VARCHAR(16) NOT NULL COMMENT 'for query',
   START_TIME         DATETIME,
   END_TIME           DATETIME COMMENT 'invalid when status is consuming',
   DEADLINE_TIME      DATETIME COMMENT 'time to terminate order',
   LAST_CONSUME_ID    INT DEFAULT 0 COMMENT 'charging times',
   EVER_PAYED         TINYINT DEFAULT 0 COMMENT 'LAST_CONSUME_ID > 0',
   RENEW_RETRIES      TINYINT DEFAULT 0 COMMENT 'num renew fails, most 100',
   STATUS             TINYINT NOT NULL COMMENT 'pending, consuming, ending, ended',
   CREATOR            VARCHAR(64) NOT NULL COMMENT 'who made this order',
   RESOURCE_NAME      VARCHAR(64) NOT NULL COMMENT 'volume name, bsi name, ...',
   PRIMARY KEY (ID)
)  DEFAULT CHARSET=UTF8;

CREATE TABLE IF NOT EXISTS DF_CONSUMING_HISTORY
(
   ID                 BIGINT NOT NULL COMMENT 'copied from DF_PURCHASE_ORDER.ID',
   ORDER_ID           VARCHAR(64) NOT NULL,
   CONSUME_ID         INT,
   CONSUMING          BIGINT NOT NULL COMMENT 'scaled by 10000',
   CONSUME_TIME       DATETIME,
   DEADLINE_TIME      DATETIME,
   ACCOUNT_ID         VARCHAR(64) NOT NULL COMMENT 'for query',
   REGION             VARCHAR(32) NOT NULL COMMENT 'for query',
   PLAN_ID            VARCHAR(64) NOT NULL COMMENT 'for query',
   PLAN_HISTORY_ID    BIGINT NOT NULL COMMENT 'auto gen id, important to retrieve history plan',
   EXTRA_INFO         INT COMMENT 'one bit for: new|renew|switch',
   PRIMARY KEY (ID, ORDER_ID, CONSUME_ID)
)  DEFAULT CHARSET=UTF8;

CREATE TABLE IF NOT EXISTS DF_ITEM_STAT
(
   STAT_KEY     VARCHAR(255) NOT NULL COMMENT '3*255 = 765 < 767',
   STAT_VALUE   INT NOT NULL,
   PRIMARY KEY (STAT_KEY)
)  DEFAULT CHARSET=UTF8;

