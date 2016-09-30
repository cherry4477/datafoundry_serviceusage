package usage

import (
	"database/sql"
	//"errors"
	"fmt"
	"time"
	//"bytes"
	"strings"
	//"io/ioutil"
	//"path/filepath"s
	//stat "github.com/asiainfoLDP/datafoundry_serviceusage/statistics"
	//"github.com/asiainfoLDP/datahub_commons/log"
)

//=============================================================
//
//=============================================================

const (
	OrderMode_Prepay  = 0 // DON'T change!
	OrderMode_Postpay = 1 // DON'T change!
)

const (
	OrderStatus_Consuming = 0 // DON'T change!
	OrderStatus_Ending    = 1 // DON'T change!
	OrderStatus_Ended     = 2 // DON'T change!
)

type PurchaseOrder struct {
	Order_id          string     `json:"orderId,omitempty"`
	Mode              int        `json:"mode,omitempty"`
	Account_id        string     `json:"accountId,omitempty"`
	Region            string     `json:"region,omitempty"`
	Quantities        int        `json:"quantities,omitempty"`
	Plan_id           string     `json:"planId,omitempty"`
	Start_time        time.Time  `json:"startTime,omitempty"`
	End_time          time.Time  `json:"_,omitempty"` // po 
	EndTime           *time.Time `json:"endTime,omitempty"` // vo
	Last_consume_time time.Time  `json:"_,omitempty"`
	Last_consume_id   int        `json:"_,omitempty"`
	Status            int        `json:"status,omitempty"`
}

const (
	ReportStep_Prepayed = 0 // DON'T change!
	ReportStep_Month    = 1 // DON'T change!
	ReportStep_Day      = 2 // DON'T change!
)

type ConsumingReport struct {
	Order_id          string    `json:"orderId,omitempty"`
	Consume_id        int       `json:"_,omitempty"`
	Start_time        time.Time `json:"startTime,omitempty"`
	Duration          int       `json:"duration,omitempty"`
	Time_tag          string    `json:"_,omitempty"`
	Step_tag          string    `json:"_,omitempty"`
	Consuming         int64     `json:"_,omitempty"`        
	Money             float64   `json:"consuming,omitempty"` // vo, Money = Consuming * 0.0001
	Account_id        string    `json:"_,omitempty"`
	Plan_id           string    `json:"_,omitempty"`
}

/*
CREATE TABLE IF NOT EXISTS DF_PURCHASE_ORDER
(
   ORDER_ID           VARCHAR(64) NOT NULL,
   MODE               TINYINT NOT NULL COMMENT 'prepay, postpay. etc',
   ACCOUNT_ID         VARCHAR(64) NOT NULL,
   REGION             VARCHAR(4) NOT NULL COMMENT 'for query',
   SERVICE_ID         VARCHAR(64) NOT NULL,
   QUANTITIES         INT DEFAULT 1 COMMENT 'for postpay only',
   PLAN_ID            VARCHAR(64) NOT NULL,
   START_TIME         DATETIME,
   END_TIME           DATETIME,
   LAST_CONSUME_TIME  DATETIME COMMENT 'already payed to this time',
   LAST_CONSUME_ID    INT COMMENT 'for report',
   STATUS             TINYINT NOT NULL COMMENT 'consuming, ended, paused',
   PRIMARY KEY (ORDER_ID)
)  DEFAULT CHARSET=UTF8;
*/

//=============================================================
//
//=============================================================

func CreateOrder(db *sql.DB, orderInfo *PurchaseOrder) error {
	order, err := RetrieveOrderByID(db, orderInfo.Order_id)
	if err != nil {
		return err
	}
	if order != nil {
		return fmt.Errorf("order (id=%s) already existed", orderInfo.Order_id)
	}

	startTime := orderInfo.Start_time.Format("2006-01-02 15:04:05.999999")
	endTime := orderInfo.End_time.Format("2006-01-02 15:04:05.999999")
	consumeTime := orderInfo.Last_consume_time.Format("2006-01-02 15:04:05.999999")
	sqlstr := fmt.Sprintf(`insert into DF_PURCHASE_ORDER (
				ORDER_ID, MODE,
				ACCOUNT_ID, REGION, 
				QUANTITIES, PLAN_ID,
				START_TIME, END_TIME, LAST_CONSUME_TIME, LAST_CONSUME_ID, 
				STATUS
				) values (
				?, ?, 
				?, ?, 
				?, ?,
				'%s', '%s', '%s', 0,
				%d
				)`, 
				startTime, endTime, consumeTime,
				OrderStatus_Consuming,
				)
	_, err = db.Exec(sqlstr,
				orderInfo.Order_id, orderInfo.Mode,  
				orderInfo.Account_id, orderInfo.Region,  
				orderInfo.Quantities, orderInfo.Plan_id, 
				)

	return err
}

func RenewOrder(db *sql.DB, orderId string, renewToTime time.Time) error {
	order, err := RetrieveOrderByID(db, orderId)
	if err != nil {
		return err
	}
	if order != nil {
		return fmt.Errorf("order (id=%s) already existed", orderId)
	}
	if order.Status != OrderStatus_Consuming {
		return fmt.Errorf("order (id=%s) not in consuming status", orderId)
	}

	// todo: renewToTime should be larger than LAST_CONSUME_TIME

	timestr := renewToTime.Format("2006-01-02 15:04:05.999999")
	sqlstr := fmt.Sprintf(`update DF_PURCHASE_ORDER set
				LAST_CONSUME_TIME='%s'
				where ORDER_ID=?`, 
				timestr,
				)
	result, err := db.Exec(sqlstr,
				orderId,
				)
	_ = result
	if err != nil {
		return err
	}

	//n, _ := result.RowsAffected()
	//if n < 1 {
	//	return fmt.Errorf("order (%s) not found", orderId)
	//}

	return nil
}

func EndOrder(db *sql.DB, orderId string) error {
	order, err := RetrieveOrderByID(db, orderId)
	if err != nil {
		return err
	}
	if order != nil {
		return fmt.Errorf("order (id=%s) already existed", orderId)
	}
	if order.Status == OrderStatus_Ended {
		return fmt.Errorf("order (id=%s) already ended", orderId)
	}

	sqlstr := fmt.Sprintf(`update DF_PURCHASE_ORDER set
				STATUS=%d
				where ORDER_ID=?`, 
				OrderStatus_Ended,
				)
	result, err := db.Exec(sqlstr,
				orderId,
				)
	_ = result
	if err != nil {
		return err
	}

	//n, _ := result.RowsAffected()
	//if n < 1 {
	//	return fmt.Errorf("order (%s) not found", orderId)
	//}

	return nil
}

func RetrieveOrderByID(db *sql.DB, orderId string) (*PurchaseOrder, error) {
	return getSingleOrder(db, fmt.Sprintf("where ORDER_ID='%s'", orderId))
}

func getSingleOrder(db *sql.DB, sqlWhere string) (*PurchaseOrder, error) {
	orders, err := queryOrders(db, sqlWhere, 1, 0)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if len(orders) == 0 {
		return nil, nil
	}

	return orders[0], nil
}

func QueryOrders(db *sql.DB, accountId string, status int, orderBy string, sortOrder bool, offset int64, limit int) (int64, []*PurchaseOrder, error) {
	sqlParams := make([]interface{}, 0, 4)
	
	// ...

	sqlWhere := ""
	if accountId != "" {
		if sqlWhere == "" {
			sqlWhere = "ACCOUNT_ID=?"
		} else {
			sqlWhere = sqlWhere + " and ACCOUNT_ID=?"
		}
		sqlParams = append(sqlParams, accountId)
	}
	if status >= 0 {
		if sqlWhere == "" {
			sqlWhere = fmt.Sprintf("STATUS=%d", status)
		} else {
			sqlWhere = sqlWhere + fmt.Sprintf(" and STATUS=%d", status)
		}
	}

	// ...
	
	switch strings.ToLower(orderBy) {
	default:
		orderBy = "START_TIME"
		sortOrder = false
	case "consumetime":
		orderBy = "LAST_CONSUME_TIME"
	case "endtime":
		orderBy = "END_TIME"
	}

	sqlSort := fmt.Sprintf("%s %s", orderBy, sortOrderText[sortOrder])

	// ...

	return getOrderList(db, offset, limit, sqlWhere, sqlSort, sqlParams...)
}

//================================================

func validateOffsetAndLimit(count int64, offset *int64, limit *int) {
	if *limit < 1 {
		*limit = 1
	}
	if *offset >= count {
		*offset = count - int64(*limit)
	}
	if *offset < 0 {
		*offset = 0
	}
	if *offset + int64(*limit) > count {
		*limit = int(count - *offset)
	}
}

const (
	SortOrder_Asc  = "asc"
	SortOrder_Desc = "desc"
)

// true: asc
// false: desc
var sortOrderText = map[bool]string{true: "asc", false: "desc"}

func ValidateSortOrder(sortOrder string, defaultOrder bool) bool {
	switch strings.ToLower(sortOrder) {
	case SortOrder_Asc:
		return true
	case SortOrder_Desc:
		return false
	}

	return defaultOrder
}

func getOrderList(db *sql.DB, offset int64, limit int, sqlWhere string, sqlSort string, sqlParams ...interface{}) (int64, []*PurchaseOrder, error) {
	//if strings.TrimSpace(sqlWhere) == "" {
	//	return 0, nil, errors.New("sqlWhere can't be blank")
	//}

	count, err := queryOrdersCount(db, sqlWhere)
	if err != nil {
		return 0, nil, err
	}
	if count == 0 {
		return 0, []*PurchaseOrder{}, nil
	}
	validateOffsetAndLimit(count, &offset, &limit)

	subs, err := queryOrders(db,
		fmt.Sprintf(`%s order by %s`, sqlWhere, sqlSort),
		limit, offset, sqlParams...)

	return count, subs, err
}

func queryOrdersCount(db *sql.DB, sqlWhere string, sqlParams ...interface{}) (int64, error) {
	sqlWhere = strings.TrimSpace(sqlWhere)
	sql_where_all := ""
	if sqlWhere != "" {
		sql_where_all = fmt.Sprintf("where %s", sqlWhere)
	}

	count := int64(0)
	sql_str := fmt.Sprintf(`select COUNT(*) from DF_PURCHASE_ORDER %s`, sql_where_all)
	err := db.QueryRow(sql_str, sqlParams...).Scan(&count)

	return count, err
}

func queryOrders(db *sql.DB, sqlWhereAll string, limit int, offset int64, sqlParams ...interface{}) ([]*PurchaseOrder, error) {
	offset_str := ""
	if offset > 0 {
		offset_str = fmt.Sprintf("offset %d", offset)
	}
	sql_str := fmt.Sprintf(`select
					ORDER_ID, MODE, 
					ACCOUNT_ID, REGION, 
					QUANTITIES, PLAN_ID,
					START_TIME, END_TIME, LAST_CONSUME_TIME, LAST_CONSUME_ID, 
					STATUS
					from DF_PURCHASE_ORDER
					%s
					limit %d
					%s
					`,
		sqlWhereAll,
		limit,
		offset_str)
	rows, err := db.Query(sql_str, sqlParams...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*PurchaseOrder, 0, 100)
	for rows.Next() {
		order := &PurchaseOrder{}
		err := rows.Scan(
			&order.Order_id, &order.Mode, 
			&order.Account_id, &order.Region, 
			&order.Quantities, &order.Plan_id,
			&order.Start_time, &order.End_time, &order.Last_consume_time, &order.Last_consume_id,
			&order.Status, 
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

