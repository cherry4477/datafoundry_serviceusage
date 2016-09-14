package usage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	//"bytes"
	"strings"
	//"io/ioutil"
	//"path/filepath"s
	//stat "github.com/asiainfoLDP/datafoundry_appmarket/statistics"
	//"github.com/asiainfoLDP/datahub_commons/log"
)

//=============================================================
//
//=============================================================

type SaasApp struct {
	App_id      string    `json:"appId,omitempty"`
	Provider    string    `json:"provider,omitempty"`
	Url         string    `json:"url,omitempty"`
	Name        string    `json:"name,omitempty"`
	Version     string    `json:"version,omitempty"`
	Category    string    `json:"category,omitempty"`
	Description string    `json:"description,omitempty"`
	Icon_url    string    `json:"iconUrl,omitempty"`
	Create_time time.Time `json:"createTime,omitempty"`
	Hotness     int       `json:"-"`
	// Price_plans
	// Usage_readme
}

//=============================================================
//
//=============================================================

func CreateApp(db *sql.DB, appInfo *SaasApp) error {
	app, err := RetrieveAppByID(db, appInfo.App_id)
	if err != nil {
		return err
	}
	if app != nil {
		return fmt.Errorf("app (id=%s) already existed", appInfo.App_id)
	}

	nowstr := time.Now().Format("2006-01-02 15:04:05.999999")
	sqlstr := fmt.Sprintf(`insert into DF_SAAS_APP_2 (
				APP_ID, 
				PROVIDER, URL, NAME, VERSION, 
				CATEGORY, DESCRIPTION, ICON_URL,
				CREATE_TIME, HOTNESS
				) values (
				'%s', 
				?, ?, ?, ?,
				?, ?, ?,
				'%s', 0
				)`, 
				appInfo.App_id,
				nowstr,
				)
	_, err = db.Exec(sqlstr,
				appInfo.Provider, appInfo.Url, appInfo.Name, appInfo.Version,
				appInfo.Category, appInfo.Description, appInfo.Icon_url,
				)

	return err
}

func ModifyApp(db *sql.DB, appInfo *SaasApp) error {
	sqlstr := fmt.Sprintf(`update DF_SAAS_APP_2 set
				PROVIDER=?, URL=?, NAME=?, VERSION=?, 
				CATEGORY=?, DESCRIPTION=?, ICON_URL=?,
				where APP_ID='%s'`, 
				appInfo.App_id,
				)
	result, err := db.Exec(sqlstr,
				appInfo.Provider, appInfo.Url, appInfo.Name, appInfo.Version,
				appInfo.Category, appInfo.Description, appInfo.Icon_url,
				)
	
	if err != nil {
		return err
	}

	n, _ := result.RowsAffected()
	if n < 1 {
		return fmt.Errorf("app (%s) not found", appInfo.App_id)
	}

	return nil
}

func DeleteApp(db *sql.DB, appId string) error {
	sqlstr := fmt.Sprintf(`delete from DF_SAAS_APP_2
				where APP_ID='%s'`, 
				appId,
				)
	result, err := db.Exec(sqlstr)
	if err != nil {
		return err
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return errors.New ("failed to delete")
	}

	return nil
}

func RetrieveAppByID(db *sql.DB, appId string) (*SaasApp, error) {
	return getSingleApp(db, fmt.Sprintf("where App_ID='%s'", appId))
}

func getSingleApp(db *sql.DB, sqlWhere string) (*SaasApp, error) {
	apps, err := queryApps(db, sqlWhere, 1, 0)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if len(apps) == 0 {
		return nil, nil
	}

	return apps[0], nil
}

func QueryApps(db *sql.DB, provider, category, orderBy string, sortOrder bool, offset int64, limit int) (int64, []*SaasApp, error) {
	sqlParams := make([]interface{}, 0, 4)
	
	// ...

	sqlWhere := ""
	provider = strings.ToLower(provider)
	if provider != "" {
		if sqlWhere == "" {
			sqlWhere = "PROVIDER=?"
		} else {
			sqlWhere = sqlWhere + " and PROVIDER=?"
		}
		sqlParams = append(sqlParams, provider)
	}
	if category != "" {
		if sqlWhere == "" {
			sqlWhere = "CATEGORY=?"
		} else {
			sqlWhere = sqlWhere + " and CATEGORY=?"
		}
		sqlParams = append(sqlParams, category)
	}

	// ...
	
	switch strings.ToLower(orderBy) {
	default:
		orderBy = "CREATE_TIME"
		sortOrder = false
	case "createtime":
		orderBy = "CREATE_TIME"
	case "hotness":
		orderBy = "HOTNESS"
	}

	sqlSort := fmt.Sprintf("%s %s", orderBy, sortOrderText[sortOrder])

	// ...

	return getAppList(db, offset, limit, sqlWhere, sqlSort, sqlParams...)
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

func ValidateOrderBy(orderBy string) string {
	switch orderBy {
	case "createtime":
		return "CREATE_TIME";
	case "hotness":
		return "HOTNESS";
	}

	return ""
}

func getAppList(db *sql.DB, offset int64, limit int, sqlWhere string, sqlSort string, sqlParams ...interface{}) (int64, []*SaasApp, error) {
	//if strings.TrimSpace(sqlWhere) == "" {
	//	return 0, nil, errors.New("sqlWhere can't be blank")
	//}

	count, err := queryAppsCount(db, sqlWhere)
	if err != nil {
		return 0, nil, err
	}
	if count == 0 {
		return 0, []*SaasApp{}, nil
	}
	validateOffsetAndLimit(count, &offset, &limit)

	subs, err := queryApps(db,
		fmt.Sprintf(`%s order by %s`, sqlWhere, sqlSort),
		limit, offset, sqlParams...)

	return count, subs, err
}

func queryAppsCount(db *sql.DB, sqlWhere string, sqlParams ...interface{}) (int64, error) {
	sqlWhere = strings.TrimSpace(sqlWhere)
	sql_where_all := ""
	if sqlWhere != "" {
		sql_where_all = fmt.Sprintf("where %s", sqlWhere)
	}

	count := int64(0)
	sql_str := fmt.Sprintf(`select COUNT(*) from DF_SAAS_APP_2 %s`, sql_where_all)
	err := db.QueryRow(sql_str, sqlParams...).Scan(&count)

	return count, err
}

func queryApps(db *sql.DB, sqlWhereAll string, limit int, offset int64, sqlParams ...interface{}) ([]*SaasApp, error) {
	offset_str := ""
	if offset > 0 {
		offset_str = fmt.Sprintf("offset %d", offset)
	}
	sql_str := fmt.Sprintf(`select
					APP_ID, 
					PROVIDER, URL, NAME, VERSION, 
					CATEGORY, DESCRIPTION, ICON_URL,
					CREATE_TIME, HOTNESS
					from DF_SAAS_APP_2
					%s
					limit %d
					%s
					`,
		sqlWhereAll,
		limit,
		offset_str)
	rows, err := db.Query(sql_str, sqlParams...)

fmt.Println(">>> ", sql_str)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps := make([]*SaasApp, 0, 100)
	for rows.Next() {
		app := &SaasApp{}
		err := rows.Scan(
			&app.App_id,
			&app.Provider, &app.Url, &app.Name, &app.Version,
			&app.Category, &app.Description, &app.Icon_url,
			&app.Create_time, &app.Hotness,
		)
		if err != nil {
			return nil, err
		}
		//validateApp(s) // already done in scanAppWithRows
		apps = append(apps, app)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return apps, nil
}
