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
	// -1 means invalid status
	OrderStatus_Pending   = 0 // DON'T change!
	OrderStatus_Consuming = 5 // DON'T change!
	//OrderStatus_Ending    = 10 // DON'T change!
	OrderStatus_Ended     = 15 // DON'T change!
)

type PurchaseOrder struct {
	Id                int64      `json:"id,omitempty"`
	Order_id          string     `json:"order_id,omitempty"`
	Account_id        string     `json:"namespace,omitempty"` // accountId
	Region            string     `json:"region,omitempty"`
	Plan_id           string     `json:"plan_id,omitempty"`
	Plan_type         string     `json:"_,omitempty"`
	Start_time        time.Time  `json:"start_time,omitempty"`
	End_time          time.Time  `json:"_,omitempty"`        // po
	EndTime           *time.Time `json:"end_time,omitempty"` // vo
	Deadline_time     time.Time  `json:"deadline,omitempty"`
	Last_consume_id   int        `json:"_,omitempty"`
	Ever_payed        int        `json:"_,omitempty"`
	Num_renew_retires int        `json:"_,omitempty"`
	Status            int        `json:"_,omitempty"`      // po
	StatusLabel       string     `json:"status,omitempty"` // vo
	Creator           string     `json:"creator,omitempty"`
}

func orderPO2VO(order *PurchaseOrder) {
	order.Start_time = order.Start_time.Local()
	order.End_time = order.End_time.Local()
	order.Deadline_time = order.Deadline_time.Local()
	if order.Status == OrderStatus_Ended { // ||  order.Status == OrderStatus_Ending {
		order.EndTime = &order.End_time
	}
}

//=============================================================
//
//=============================================================

type DbOrTx interface {
        Exec(query string, args ...interface{}) (sql.Result, error)
        Prepare(query string) (*sql.Stmt, error)
        Query(query string, args ...interface{}) (*sql.Rows, error)
        QueryRow(query string, args ...interface{}) *sql.Row
}

//=============================================================
//
//=============================================================

// return the auto generated id
func CreateOrder(db *sql.DB, orderInfo *PurchaseOrder) (int64, error) {
	// zongsan: now pending orders will be kept in db.
	/*
	order, err := RetrieveOrderByID(db, orderInfo.Order_id)
	if err != nil {
		return err
	}

	if order != nil {
		if order.Status != OrderStatus_Consuming {
			// delete old order
			err = RemoveOrder(db, orderInfo.Order_id)
			if err != nil {
				return err
			}
		} else {
			// todo: change plan for order

			return fmt.Errorf("order (id=%s) already existed", orderInfo.Order_id)
		}
	}
	*/

	startTime := orderInfo.Start_time.UTC().Format("2006-01-02 15:04:05.999999")
	endTime := orderInfo.End_time.UTC().Format("2006-01-02 15:04:05.999999")
	consumeTime := orderInfo.Deadline_time.UTC().Format("2006-01-02 15:04:05.999999")
	sqlstr := fmt.Sprintf(`insert into DF_PURCHASE_ORDER (
				ORDER_ID,
				ACCOUNT_ID, REGION, 
				PLAN_ID, PLAN_TYPE, 
				START_TIME, END_TIME, DEADLINE_TIME, LAST_CONSUME_ID, EVER_PAYED,
				CREATOR, STATUS, 
				RENEW_RETRIES
				) values (
				?, 
				?, ?, 
				?, ?, 
				'%s', '%s', '%s', 0, 0,
				?, ?, 
				0
				)`, 
				startTime, endTime, consumeTime,
				)
	result, err := db.Exec(sqlstr,
				orderInfo.Order_id, 
				orderInfo.Account_id, orderInfo.Region,  
				orderInfo.Plan_id, orderInfo.Plan_type, 
				orderInfo.Creator, orderInfo.Status,  
				)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func RemoveOrder(db *sql.DB, orderId string) error {
	sqlstr := `delete from DF_PURCHASE_ORDER where ORDER_ID=?`

	_, err := db.Exec(sqlstr, orderId)

	return err
}

//=============

const DeadlineExtendedDuration_Month = time.Duration(-1)

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func extendTime(t time.Time, extended time.Duration, startTime time.Time) time.Time {
	switch extended {
	case DeadlineExtendedDuration_Month:
		y := t.Year()
		m := t.Month() + 1
		d := startTime.Day()
		if m > time.December {
			y++
			m = time.January
		}
		
		lastDay := daysInMonth(y, m)
		if d > lastDay {
			d = lastDay
		}
		next := time.Date(y, m, d, 
			startTime.Hour(), startTime.Minute(), startTime.Second(), startTime.Nanosecond(), 
			time.FixedZone(startTime.Zone()))

		//fmt.Println("=== startTime = ", startTime)
		//fmt.Println("=== t = ", t)
		//fmt.Println("=== next = ", next)

		return next
	}

	return t.Add(extended)
}

const MaxNumRenewalRetries = 100

func IncreaseOrderRenewalFails(db *sql.DB, orderAutoGenId int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	
	return func() error {
		type db chan struct{} // avoid misuing db

		order, err := RetrieveOrderByAutoGenID(tx, orderAutoGenId)
		if err != nil {
			tx.Rollback()
			return err
		}
		if order == nil {
			tx.Rollback()
			return fmt.Errorf("order (id=%s) not found", orderAutoGenId)
		}

		if order.Num_renew_retires >= MaxNumRenewalRetries {
			tx.Rollback()
			return nil
		}

		order.Num_renew_retires ++

		sqlstr := fmt.Sprintf(`update DF_PURCHASE_ORDER set
					RENEW_RETRIES=%d
					where ID=%d`, 
					order.Num_renew_retires,
					orderAutoGenId,
					)

		result, err := tx.Exec(sqlstr)
		_ = result
		if err != nil {
			tx.Rollback()
			return err
		}

		//n, _ := result.RowsAffected()
		//if n < 1 {
		//	return fmt.Errorf("order (%s) not found", orderId)
		//}

		err = tx.Commit()
		if err != nil {
			return err
		}
		
		return nil
	}()
}

/*


func IncreaseOrderRenewalFails(db *sql.DB, orderId string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	
	return func() error {
		type db chan struct{} // avoid misuing db

		order, err := RetrieveOrderByID(tx, orderId, OrderStatus_Consuming)
		if err != nil {
			tx.Rollback()
			return err
		}
		if order == nil {
			tx.Rollback()
			return fmt.Errorf("consuming order (id=%s) not found", orderId)
		}

		if order.Num_renew_retires >= MaxNumRenewalRetries {
			tx.Rollback()
			return nil
		}

		order.Num_renew_retires ++

		sqlstr := fmt.Sprintf(`update DF_PURCHASE_ORDER set
					RENEW_RETRIES=%d
					where ORDER_ID=?`, 
					order.Num_renew_retires,
					)

		result, err := tx.Exec(sqlstr,
					orderId,
					)
		_ = result
		if err != nil {
			tx.Rollback()
			return err
		}

		//n, _ := result.RowsAffected()
		//if n < 1 {
		//	return fmt.Errorf("order (%s) not found", orderId)
		//}

		err = tx.Commit()
		if err != nil {
			return err
		}
		
		return nil
	}()
}
*/

/*
func RenewOrder(db *sql.DB, orderId string, extendedDuration time.Duration) (*PurchaseOrder, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	return func() (*PurchaseOrder, error) {
		type db chan struct{} // avoid misuing db

		order, err := RetrieveOrderByID(tx, orderId, OrderStatus_Consuming)
		order, err := RetrieveOrderByID(tx, orderId, -1) // also not good
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		if order == nil {
			tx.Rollback()
			return nil, fmt.Errorf("order (Order_id=%s) not found", orderId)
		}

		// need checking this? This function should be only called when a payment was just made.
		//if order.Status != OrderStatus_Consuming && order.Status != OrderStatus_Pending {
		//	tx.Rollback()
		//	return fmt.Errorf("order (id=%s) not consumable", orderId)
		//}

		deadlineTime := extendTime(order.Deadline_time, extendedDuration, order.Start_time)
		lastConsumeId := order.Last_consume_id + 1

		onOk := func() { // why this? may be history reason
			order.Deadline_time = deadlineTime
			order.Last_consume_id = lastConsumeId
		}

		// todo: renewToTime should be larger than DEADLINE_TIME. Needed?

		timestr := deadlineTime.UTC().Format("2006-01-02 15:04:05.999999")
		sqlstr := fmt.Sprintf(`update DF_PURCHASE_ORDER set
					DEADLINE_TIME='%s', LAST_CONSUME_ID=%d, EVER_PAYED=1, RENEW_RETRIES=0, STATUS=%d
					where ORDER_ID=?`, 
					timestr, lastConsumeId,
					OrderStatus_Consuming,
					)

		result, err := tx.Exec(sqlstr,
					orderId,
					)
		_ = result
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		//n, _ := result.RowsAffected()
		//if n < 1 {
		//	return nil, fmt.Errorf("order (%s) not found", orderId)
		//}

		err = tx.Commit()
		if err != nil {
			return nil, err
		}

		onOk()

		return order, nil
	}()
}
*/

func GetDurationToRenewNextConsumingOrder(db *sql.DB) (time.Duration, error) {
	return time.Hour, nil
}

func RenewOrder(db *sql.DB, orderAutoGenId int64, lastConsumeId int, extendedDuration time.Duration) (*PurchaseOrder, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	return func() (*PurchaseOrder, error) {
		type db chan struct{} // avoid misuing db

		order, err := RetrieveOrderByAutoGenID(tx, orderAutoGenId)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		if order == nil {
			tx.Rollback()
			return nil, fmt.Errorf("order (id=%d) not found", orderAutoGenId)
		}

		if order.Last_consume_id != lastConsumeId { // already renewed?
			tx.Rollback()
			return order, nil
		}

		// need checking this? This function should be only called when a payment was just made.
		//if order.Status != OrderStatus_Consuming && order.Status != OrderStatus_Pending {
		//	tx.Rollback()
		//	return fmt.Errorf("order (id=%s) not consumable", orderId)
		//}

		deadlineTime := extendTime(order.Deadline_time, extendedDuration, order.Start_time)
		lastConsumeId := order.Last_consume_id + 1

		onOk := func() { 
			order.Deadline_time = deadlineTime
			order.Last_consume_id = lastConsumeId

			// the returned order will be used to create consumeing history.
			// so here above two feilds must be correct after the following update.
		}

		// todo: renewToTime should be larger than DEADLINE_TIME. Needed?

		timestr := deadlineTime.UTC().Format("2006-01-02 15:04:05.999999")
		sqlstr := fmt.Sprintf(`update DF_PURCHASE_ORDER set
					DEADLINE_TIME='%s', LAST_CONSUME_ID=%d, EVER_PAYED=1, RENEW_RETRIES=0, STATUS=%d
					where ID=%d`, 
					timestr, lastConsumeId, OrderStatus_Consuming,
					orderAutoGenId,
					)

		result, err := tx.Exec(sqlstr)
		_ = result
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		//n, _ := result.RowsAffected()
		//if n < 1 {
		//	return nil, fmt.Errorf("order (%s) not found", orderId)
		//}

		err = tx.Commit()
		if err != nil {
			return nil, err
		}

		onOk()

		return order, nil
	}()
}

func EndOrder(db *sql.DB, orderInfo *PurchaseOrder, endTime time.Time, lastConsume *ConsumeHistory, remainingMoney float32) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	return func() error {
		type db chan struct{} // avoid misuing db

		endTimeStr := endTime.UTC().Format("2006-01-02 15:04:05.999999")

		sqlstr := fmt.Sprintf(`update DF_PURCHASE_ORDER set
					STATUS=%d, END_TIME='%s'
					where 
					STATUS=%d and ID=%d and ORDER_ID=? and ACCOUNT_ID=?`, 
					OrderStatus_Ended, endTimeStr,
					OrderStatus_Consuming, orderInfo.Id, 
					)
		result, err := tx.Exec(sqlstr,
					orderInfo.Order_id, orderInfo.Account_id,
					)
		_ = result
		if err != nil {
			tx.Rollback()
			return err
		}

		n, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return err
		}

		removed := n > 0 // n should be 1

		if removed {
			orderInfo.Last_consume_id++
			onFailed := func() { 
				orderInfo.Last_consume_id--
			}

			err := CreateConsumeHistory(tx, orderInfo, endTime, -remainingMoney, 
							lastConsume.Plan_history_id, ConsumeExtraInfo_EndOrder)
			if err != nil {
				onFailed()

				tx.Rollback()
				return err
			}
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		return nil
	}()
}

// status argument must be a valid order status value
func RetrieveOrderByAutoGenID(db DbOrTx, orderAutoGenId int64) (*PurchaseOrder, error) {
	sqlWhere := "ID=?"
	sqlParams := make([]interface{}, 0, 1)
	sqlParams = append(sqlParams, orderAutoGenId)

	return getSingleOrder(db, sqlWhere, sqlParams...)
}

// status argument must be a valid order status value
func RetrieveOrderByStatusAndAutoGenID(db DbOrTx, orderAutoGenId int64, status int) (*PurchaseOrder, error) {
	sqlWhere := "ID=?"
	sqlParams := make([]interface{}, 0, 2)
	sqlParams = append(sqlParams, orderAutoGenId)
	if status >= 0 {
		sqlWhere += " and STATUS=?"
		sqlParams = append(sqlParams, OrderStatus_Consuming)
	}

	return getSingleOrder(db, sqlWhere, sqlParams...)
}

// status argument must be a valid order status value
func RetrieveOrderByID(db DbOrTx, orderId, region string, status int) (*PurchaseOrder, error) {
	sqlWhere := "ORDER_ID=?"
	sqlParams := make([]interface{}, 0, 3)
	sqlParams = append(sqlParams, orderId)
	if region != "" {
		sqlWhere += " and REGION=?"
		sqlParams = append(sqlParams, region)
	}
	if status >= 0 {
		sqlWhere += " and STATUS=?"
		sqlParams = append(sqlParams, status)
	}

	return getSingleOrder(db, sqlWhere, sqlParams...)
}

func QueryOrders(db DbOrTx, accountId string, region string, status int, renewalFailedOnly bool, offset int64, limit int) (int64, []*PurchaseOrder, error) {
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

	if region != "" {
		if sqlWhere == "" {
			sqlWhere = fmt.Sprintf("REGION=?")
		} else {
			sqlWhere = sqlWhere + fmt.Sprintf(" and REGION=?")
		}
		sqlParams = append(sqlParams, region)
	}

	if renewalFailedOnly {
		if sqlWhere == "" {
			sqlWhere = "RENEW_RETRIES>0"
		} else {
			sqlWhere = sqlWhere + " and RENEW_RETRIES>0"
		}
	}

	// filter out pending orders
	if sqlWhere == "" {
		sqlWhere = "EVER_PAYED=1"
	} else {
		sqlWhere = sqlWhere + " and EVER_PAYED=1"
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

// for maintaining
func QueryConsumingOrdersToRenew(db DbOrTx, limit int) (int64, []*PurchaseOrder, error) {

	sqlWhere := fmt.Sprintf("STATUS=%d", OrderStatus_Consuming)

	// filter out pending orders
	sqlWhere = sqlWhere + " and EVER_PAYED=1"

	// ...

	orderBy, sortOrder := "DEADLINE_TIME", SortOrder_Asc

	sqlSort := fmt.Sprintf("%s %s", orderBy, sortOrder)

	// ...

	return getOrderList(db, 0, limit, sqlWhere, sqlSort)
}

//=======================================================================
// 
//=======================================================================

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

func getOrderList(db DbOrTx, offset int64, limit int, sqlWhere string, sqlSort string, sqlParams ...interface{}) (int64, []*PurchaseOrder, error) {
	//if strings.TrimSpace(sqlWhere) == "" {
	//	return 0, nil, errors.New("sqlWhere can't be blank")
	//}

	count, err := queryOrdersCount(db, sqlWhere, sqlParams...)
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

func queryOrdersCount(db DbOrTx, sqlWhere string, sqlParams ...interface{}) (int64, error) {
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

func getSingleOrder(db DbOrTx, sqlWhere string, sqlParams ...interface{}) (*PurchaseOrder, error) {
	orders, err := queryOrders(db, sqlWhere, 1, 0, sqlParams...)
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

func queryOrders(db DbOrTx, sqlWhere string, limit int, offset int64, sqlParams ...interface{}) ([]*PurchaseOrder, error) {
	sqlWhere = strings.TrimSpace(sqlWhere)

	sql_where_all := ""
	if sqlWhere != "" {
		sql_where_all = fmt.Sprintf("where %s", sqlWhere)
	}

	offset_str := ""
	if offset > 0 {
		offset_str = fmt.Sprintf("offset %d", offset)
	}
	sql_str := fmt.Sprintf(`select
					ID,
					ORDER_ID, 
					ACCOUNT_ID, REGION, 
					PLAN_ID, PLAN_TYPE,
					START_TIME, END_TIME, DEADLINE_TIME, LAST_CONSUME_ID, EVER_PAYED,
					CREATOR, STATUS, 
					RENEW_RETRIES
					from DF_PURCHASE_ORDER
					%s
					limit %d
					%s
					`,
		sql_where_all,
		limit,
		offset_str)
	
	// println("sql_str = ", sql_str)

	rows, err := db.Query(sql_str, sqlParams...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*PurchaseOrder, 0, 100)
	for rows.Next() {
		order := &PurchaseOrder{}
		err := rows.Scan(
			&order.Id, 
			&order.Order_id, 
			&order.Account_id, &order.Region, 
			&order.Plan_id, &order.Plan_type, 
			&order.Start_time, &order.End_time, &order.Deadline_time, &order.Last_consume_id, &order.Ever_payed,
			&order.Creator, &order.Status, 
			&order.Num_renew_retires,
		)
		if err != nil {
			return nil, err
		}
		//>>
		orderPO2VO(order)
		//<<
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

//==================================================================
// reports
//==================================================================

func ConsumingToMoney(consuming int64) float32 {
	return 0.0001 * float32(consuming) // DON"T CHANGE
}

func MoneyToConsuming(money float32) int64 {
	return int64(money * 10000) // DON"T CHANGE
}

const (
	// low 8 bits in ConsumeHistoryExtra_info
	ConsumeExtraInfo_NewOrder    = 0x0
	ConsumeExtraInfo_RenewOrder  = 0x1
	ConsumeExtraInfo_SwitchOrder = 0x2
	ConsumeExtraInfo_EndOrder    = 0x3
)

type ConsumeHistory struct {
	Id                int64     `json:"_"`
	Order_id          string    `json:"order_id,omitempty"`
	Consume_id        int       `json:"_,omitempty"`
	Consuming         int64     `json:"_,omitempty"`       // po
	Money             float32   `json:"money,omitempty"`   // vo, Money = Consuming * 0.0001
	Consume_time      time.Time `json:"time,omitempty"`
	Deadline_time     time.Time `json:"deadline,omitempty"`
	Account_id        string    `json:"project,omitempty"` // accountId
	Region            string    `json:"region,omitempty"`
	Plan_id           string    `json:"plan_id,omitempty"`
	Plan_history_id   int64     `json:"plan_history_id,omitempty"`
	Extra_info        int       `json:"_"`
}

func consumePO2VO(consume *ConsumeHistory) {
	consume.Money = ConsumingToMoney(consume.Consuming)
	consume.Consume_time = consume.Consume_time.Local()
	consume.Deadline_time = consume.Deadline_time.Local()
}

func CreateConsumeHistory(db DbOrTx, orderInfo *PurchaseOrder, consumeTime time.Time, money float32, planHistoryId int64, extraInfo int) error {
	consuming := MoneyToConsuming(money)

	sqlstr := fmt.Sprintf(`insert into DF_CONSUMING_HISTORY (
				ID, ORDER_ID, CONSUME_ID,
				CONSUMING, CONSUME_TIME, DEADLINE_TIME,
				ACCOUNT_ID, REGION, PLAN_ID, PLAN_HISTORY_ID, 
				EXTRA_INFO
				) values (
				%d, '%s', %d, 
				%d, '%s', '%s', 
				'%s', '%s', '%s', %d,
				%d
				)`, 
				orderInfo.Id, orderInfo.Order_id, orderInfo.Last_consume_id,
				consuming, consumeTime.UTC().Format("2006-01-02 15:04:05.999999"), orderInfo.Deadline_time.UTC().Format("2006-01-02 15:04:05.999999"), 
				orderInfo.Account_id, orderInfo.Region, orderInfo.Plan_id, planHistoryId,
				extraInfo,
				)
	_, err := db.Exec(sqlstr)

	return err
}

func RetrieveConsumeHistory(db DbOrTx, orderAutoId int64, orderId string, cunsumeId int) (*ConsumeHistory, error) {
	sqlWhere := fmt.Sprintf("ID=%d and ORDER_ID=? and CONSUME_ID=%d", orderAutoId, cunsumeId)

	sqlParams := []interface{}{orderId}

	return getSingleConsuming(db, sqlWhere, sqlParams...)
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
		sqlWhere = sqlWhere + fmt.Sprintf(" and REGION=?")
		sqlParams = append(sqlParams, region)
	}

	// ...

	orderBy, sortOrder := "CONSUME_TIME", SortOrder_Desc
	orderBy2, sortOrder2 := "ID", SortOrder_Desc

	sqlSort := fmt.Sprintf("%s %s, %s %s", orderBy, sortOrder, orderBy2, sortOrder2)

	// ...

	return getConsumingList(db, offset, limit, sqlWhere, sqlSort, sqlParams...)
}

//================================================

func getConsumingList(db *sql.DB, offset int64, limit int, sqlWhere string, sqlSort string, sqlParams ...interface{}) (int64, []*ConsumeHistory, error) {
	//if strings.TrimSpace(sqlWhere) == "" {
	//	return 0, nil, errors.New("sqlWhere can't be blank")
	//}

	count, err := queryConsumingsCount(db, sqlWhere, sqlParams...)
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

func getSingleConsuming(db DbOrTx, sqlWhere string, sqlParams ...interface{}) (*ConsumeHistory, error) {
	consumings, err := queryConsumings(db, sqlWhere, 1, 0, sqlParams...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if len(consumings) == 0 {
		return nil, nil
	}

	return consumings[0], nil
}

func queryConsumings(db DbOrTx, sqlWhere string, limit int, offset int64, sqlParams ...interface{}) ([]*ConsumeHistory, error) {
	sqlWhere = strings.TrimSpace(sqlWhere)
	sql_where_all := ""
	if sqlWhere != "" {
		sql_where_all = fmt.Sprintf("where %s", sqlWhere)
	}
	offset_str := ""
	if offset > 0 {
		offset_str = fmt.Sprintf("offset %d", offset)
	}
	sql_str := fmt.Sprintf(`select
					ID, ORDER_ID, CONSUME_ID,
					CONSUMING, CONSUME_TIME, DEADLINE_TIME,
					ACCOUNT_ID, REGION, PLAN_ID, PLAN_HISTORY_ID,
					EXTRA_INFO
					from DF_CONSUMING_HISTORY
					%s
					limit %d
					%s
					`,
		sql_where_all,
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
			&consume.Id, &consume.Order_id, &consume.Consume_id, 
			&consume.Consuming, &consume.Consume_time, &consume.Deadline_time, 
			&consume.Account_id, &consume.Region, &consume.Plan_id, &consume.Plan_history_id, 
			&consume.Extra_info,
		)
		if err != nil {
			return nil, err
		}
		//>>
		consumePO2VO(consume)
		//<<
		consumings = append(consumings, consume)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return consumings, nil
}
