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
	OrderStatus_Pending   = 0 // DON'T change!
	OrderStatus_Consuming = 5 // DON'T change!
	OrderStatus_Ending    = 10 // DON'T change!
	OrderStatus_Ended     = 15 // DON'T change!
)

type PurchaseOrder struct {
	Order_id          string     `json:"orderId,omitempty"`
	Account_id        string     `json:"project,omitempty"` // accountId
	Region            string     `json:"region,omitempty"`
	Quantities        int        `json:"quantities,omitempty"`
	Plan_id           string     `json:"planId,omitempty"`
	Plan_type         string     `json:"_,omitempty"`
	Start_time        time.Time  `json:"startTime,omitempty"`
	End_time          time.Time  `json:"_,omitempty"`       // po 
	EndTime           *time.Time `json:"endTime,omitempty"` // vo
	Deadline_time     time.Time  `json:"deadline,omitempty"`
	Last_consume_id   int        `json:"_,omitempty"`
	Status            int        `json:"status,omitempty"`
	Creator           string     `json:"creator,omitempty"`
}

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
	consumeTime := orderInfo.Deadline_time.Format("2006-01-02 15:04:05.999999")
	sqlstr := fmt.Sprintf(`insert into DF_PURCHASE_ORDER (
				ORDER_ID,
				ACCOUNT_ID, REGION, 
				QUANTITIES, PLAN_ID, PLAN_TYPE, 
				START_TIME, END_TIME, DEADLINE_TIME, LAST_CONSUME_ID, 
				CREATOR, STATUS
				) values (
				?, 
				?, ?, 
				?, ?, ?, 
				'%s', '%s', '%s', 0,
				?
				)`, 
				startTime, endTime, consumeTime,
				)
	_, err = db.Exec(sqlstr,
				orderInfo.Order_id, 
				orderInfo.Account_id, orderInfo.Region,  
				orderInfo.Quantities, orderInfo.Plan_id, orderInfo.Plan_type, 
				orderInfo.Creator, orderInfo.Status,  
				)

	return err
}

/*
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

	// todo: renewToTime should be larger than DEADLINE_TIME

	timestr := renewToTime.Format("2006-01-02 15:04:05.999999")
	sqlstr := fmt.Sprintf(`update DF_PURCHASE_ORDER set
				DEADLINE_TIME='%s'
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
*/

func EndOrder(db *sql.DB, orderId string, accountId string) error {
	order, err := RetrieveOrderByID(db, orderId)
	if err != nil {
		return err
	}
	if order != nil {
		return fmt.Errorf("order (id=%s) already existed", orderId)
	}
	if order.Account_id != accountId {
		return fmt.Errorf("account id of order (id=%s) and input account id (%s) not match", orderId, accountId)
	}
	if order.Status == OrderStatus_Ending || order.Status == OrderStatus_Ended {
		return fmt.Errorf("order (id=%s) already ended", orderId)
	}

	sqlstr := fmt.Sprintf(`update DF_PURCHASE_ORDER set
				STATUS=%d
				where 
				ORDER_ID=?`, 
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

func QueryOrders(db *sql.DB, accountId string, status int, offset int64, limit int) (int64, []*PurchaseOrder, error) {
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

	orderBy, sortOrder := "", ""
	
	switch strings.ToLower(orderBy) {
	case "consumetime":
		orderBy = "DEADLINE_TIME"
		sortOrder = SortOrder_Desc
	case "endtime":
		orderBy = "END_TIME"
		sortOrder = SortOrder_Desc
	// case "starttime":
	default:
		orderBy = "START_TIME"
		sortOrder = SortOrder_Desc
	}

	sqlSort := fmt.Sprintf("%s %s", orderBy, sortOrder)

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
					ORDER_ID, 
					ACCOUNT_ID, REGION, 
					QUANTITIES, PLAN_ID, PLAN_TYPE,
					START_TIME, END_TIME, DEADLINE_TIME, LAST_CONSUME_ID, 
					CREATOR, STATUS
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
			&order.Order_id, 
			&order.Account_id, &order.Region, 
			&order.Quantities, &order.Plan_id, &order.Plan_type, 
			&order.Start_time, &order.End_time, &order.Deadline_time, &order.Last_consume_id,
			&order.Creator, &order.Status, 
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

//==================================================================
// auto renew
//==================================================================

// TODO: extend order deadline

//==================================================================
// reports
//==================================================================

func ConsumingToMoney(consuming int64) float64 {
	return 0.0001 * float64(consuming) // DON"T CHANGE
}

func MoneyToConsuming(money float64) int64 {
	return int64(money * 10000) // DON"T CHANGE
}

type ConsumeHistory struct {
	Order_id          string    `json:"orderId,omitempty"`
	Consume_id        int       `json:"_,omitempty"`
	Consume_time      time.Time `json:"time,omitempty"`
	Consuming         int64     `json:"_,omitempty"`       // po
	Money             float64   `json:"money,omitempty"`   // vo, Money = Consuming * 0.0001
	Account_id        string    `json:"project,omitempty"` // accountId
	Region            string    `json:"region,omitempty"`
	Plan_id           string    `json:"planId,omitempty"`
}

func CreateConsumeHistory(db *sql.DB, orderInfo *PurchaseOrder, consumeId int, consumeTime time.Time, money float64) error {
	order, err := RetrieveOrderByID(db, orderInfo.Order_id)
	if err != nil {
		return err
	}
	if order != nil {
		return fmt.Errorf("order (id=%s) already existed", orderInfo.Order_id)
	}

	consuming := MoneyToConsuming(money)

	sqlstr := fmt.Sprintf(`insert into DF_CONSUMING_HISTORY (
				ORDER_ID, CONSUME_ID,
				CONSUMING, CONSUME_TIME,
				ACCOUNT_ID, REGION, PLAN_ID
				) values (
				%d, %d, 
				%d, '%s', 
				'%s', '%s', '%s'
				)`, 
				orderInfo.Order_id, consumeId,
				consuming, consumeTime.Format("2006-01-02 15:04:05.999999"), 
				orderInfo.Account_id, orderInfo.Region, orderInfo.Plan_id, 
				)
	_, err = db.Exec(sqlstr)

	return err
}

func QueryConsumeHistories(db *sql.DB, accountId string, orderId string, region string, offset int64, limit int) (int64, []*ConsumeHistory, error) {
	sqlParams := make([]interface{}, 0, 4)
	
	// ...

	sqlWhere := "ACCOUNT_ID=?"
	sqlParams = append(sqlParams, accountId)
	
	if orderId != "" {
		sqlWhere = sqlWhere + " and ORDER_ID=?"
		sqlParams = append(sqlParams, orderId)
	}
	if region != "" {
		sqlWhere = sqlWhere + " and REGION=?"
		sqlParams = append(sqlParams, region)
	}

	// ...

	orderBy, sortOrder := "CONSUME_TIME", SortOrder_Desc

	sqlSort := fmt.Sprintf("%s %s", orderBy, sortOrder)

	// ...

	return getConsumingList(db, offset, limit, sqlWhere, sqlSort, sqlParams...)
}

//================================================

func getConsumingList(db *sql.DB, offset int64, limit int, sqlWhere string, sqlSort string, sqlParams ...interface{}) (int64, []*ConsumeHistory, error) {
	//if strings.TrimSpace(sqlWhere) == "" {
	//	return 0, nil, errors.New("sqlWhere can't be blank")
	//}

	count, err := queryConsumingsCount(db, sqlWhere)
	if err != nil {
		return 0, nil, err
	}
	if count == 0 {
		return 0, []*ConsumeHistory{}, nil
	}
	validateOffsetAndLimit(count, &offset, &limit)

	subs, err := queryConsumings(db,
		fmt.Sprintf(`%s order by %s`, sqlWhere, sqlSort),
		limit, offset, sqlParams...)

	return count, subs, err
}

func queryConsumingsCount(db *sql.DB, sqlWhere string, sqlParams ...interface{}) (int64, error) {
	sqlWhere = strings.TrimSpace(sqlWhere)
	sql_where_all := ""
	if sqlWhere != "" {
		sql_where_all = fmt.Sprintf("where %s", sqlWhere)
	}

	count := int64(0)
	sql_str := fmt.Sprintf(`select COUNT(*) from DF_CONSUMING_HISTORY %s`, sql_where_all)
	err := db.QueryRow(sql_str, sqlParams...).Scan(&count)

	return count, err
}

func queryConsumings(db *sql.DB, sqlWhereAll string, limit int, offset int64, sqlParams ...interface{}) ([]*ConsumeHistory, error) {
	offset_str := ""
	if offset > 0 {
		offset_str = fmt.Sprintf("offset %d", offset)
	}
	sql_str := fmt.Sprintf(`select
					ORDER_ID, CONSUME_ID,
					CONSUMING, CONSUME_TIME,
					ACCOUNT_ID, REGION, PLAN_ID
					from DF_CONSUMING_HISTORY
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

	consumings := make([]*ConsumeHistory, 0, 100)
	for rows.Next() {
		consume := &ConsumeHistory{}
		err := rows.Scan(
			&consume.Order_id, &consume.Consume_id, 
			&consume.Consuming, &consume.Consume_time, 
			&consume.Account_id, &consume.Region, &consume.Plan_id, 
		)
		if err != nil {
			return nil, err
		}
		consume.Money = ConsumingToMoney(consume.Consuming)
		consumings = append(consumings, consume)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return consumings, nil
}

