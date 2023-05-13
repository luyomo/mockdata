/*
Copyright © 2020 Marvin

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package oracle

import (
	"context"
	"database/sql"
//    "database/sql/driver"
	"fmt"
	"github.com/godror/godror"
	"github.com/godror/godror/dsn"
	"github.com/wentaojin/transferdb/common"
    //"github.com/luyomo/mockdata/internal/database/oracle"
    "github.com/luyomo/mockdata/internal/config"
	"runtime"
	"strconv"
	"strings"
)

type Oracle struct {
	Ctx      context.Context
	OracleDB *sql.DB
}

// 创建 oracle 数据库引擎
func NewOracleDBEngine(ctx context.Context, oraCfg config.OracleConfig) (*Oracle, error) {
	// https://pkg.go.dev/github.com/godror/godror
	// https://github.com/godror/godror/blob/db9cd12d89cdc1c60758aa3f36ece36cf5a61814/doc/connection.md
	// https://godror.github.io/godror/doc/connection.html
	// You can specify connection timeout seconds with "?connect_timeout=15" - Ping uses this timeout, NOT the Deadline in Context!
	// For more connection options, see [Godor Connection Handling](https://godror.github.io/godror/doc/connection.html).
	var (
		connString string
		oraDSN     dsn.ConnectionParams
		err        error
	)

	// https://github.com/godror/godror/pull/65
	connString = fmt.Sprintf("oracle://@%s/%s?connectionClass=POOL_CONNECTION_CLASS00&%s",
		common.StringsBuilder(oraCfg.Host, ":", strconv.Itoa(oraCfg.Port)),
		oraCfg.ServiceName, oraCfg.ConnectParams)
	oraDSN, err = godror.ParseDSN(connString)
	if err != nil {
		return nil, err
	}

	oraDSN.Username, oraDSN.Password = oraCfg.Username, godror.NewPassword(oraCfg.Password)

	if !strings.EqualFold(oraCfg.PDBName, "") {
		oraCfg.SessionParams = append(oraCfg.SessionParams, fmt.Sprintf(`ALTER SESSION SET CONTAINER = %s`, oraCfg.PDBName))
	}

	if !strings.EqualFold(oraCfg.Username, oraCfg.SchemaName) && !strings.EqualFold(oraCfg.SchemaName, "") {
		oraCfg.SessionParams = append(oraCfg.SessionParams, fmt.Sprintf(`ALTER SESSION SET CURRENT_SCHEMA = %s`, oraCfg.SchemaName))
	}

	// 关闭外部认证
	oraDSN.ExternalAuth = false
	oraDSN.OnInitStmts = oraCfg.SessionParams

	// libDir won't have any effect on Linux for linking reasons to do with Oracle's libnnz library that are proving to be intractable.
	// You must set LD_LIBRARY_PATH or run ldconfig before your process starts.
	// This is documented in various places for other drivers that use ODPI-C. The parameter works on macOS and Windows.
	if !strings.EqualFold(oraCfg.LibDir, "") {
		switch runtime.GOOS {
		case "windows", "darwin":
			oraDSN.LibDir = oraCfg.LibDir
		}
	}

	// godror logger 日志输出
	// godror.SetLogger(zapr.NewLogger(zap.L()))

	sqlDB := sql.OpenDB(godror.NewConnector(oraDSN))
	sqlDB.SetMaxIdleConns(0)
	sqlDB.SetMaxOpenConns(0)
	sqlDB.SetConnMaxLifetime(0)

	err = sqlDB.Ping()
	if err != nil {
		return nil, fmt.Errorf("error on ping oracle database connection:%v", err)
	}
	return &Oracle{
		Ctx:      ctx,
		OracleDB: sqlDB,
	}, nil
}

// Only Used for ALL Mode
func NewOracleLogminerEngine(ctx context.Context, oraCfg config.OracleConfig) (*Oracle, error) {
	// https://pkg.go.dev/github.com/godror/godror
	// https://github.com/godror/godror/blob/db9cd12d89cdc1c60758aa3f36ece36cf5a61814/doc/connection.md
	// https://godror.github.io/godror/doc/connection.html
	// You can specify connection timeout seconds with "?connect_timeout=15" - Ping uses this timeout, NOT the Deadline in Context!
	// For more connection options, see [Godor Connection Handling](https://godror.github.io/godror/doc/connection.html).
	var (
		connString string
		oraDSN     dsn.ConnectionParams
		err        error
	)

	// https://github.com/godror/godror/pull/65
	connString = fmt.Sprintf("oracle://@%s/%s?connectionClass=POOL_CONNECTION_CLASS01&%s",
		common.StringsBuilder(oraCfg.Host, ":", strconv.Itoa(oraCfg.Port)),
		oraCfg.ServiceName, oraCfg.ConnectParams)
	oraDSN, err = godror.ParseDSN(connString)
	if err != nil {
		return nil, err
	}

	oraDSN.Username, oraDSN.Password = oraCfg.Username, godror.NewPassword(oraCfg.Password)

	// 关闭外部认证
	oraDSN.ExternalAuth = false
	oraDSN.OnInitStmts = oraCfg.SessionParams

	// libDir won't have any effect on Linux for linking reasons to do with Oracle's libnnz library that are proving to be intractable.
	// You must set LD_LIBRARY_PATH or run ldconfig before your process starts.
	// This is documented in various places for other drivers that use ODPI-C. The parameter works on macOS and Windows.
	if !strings.EqualFold(oraCfg.LibDir, "") {
		switch runtime.GOOS {
		case "windows", "darwin":
			oraDSN.LibDir = oraCfg.LibDir
		}
	}

	// godror logger 日志输出
	// godror.SetLogger(zapr.NewLogger(zap.L()))

	sqlDB := sql.OpenDB(godror.NewConnector(oraDSN))
	sqlDB.SetMaxIdleConns(0)
	sqlDB.SetMaxOpenConns(0)
	sqlDB.SetConnMaxLifetime(0)

	err = sqlDB.Ping()
	if err != nil {
		return nil, fmt.Errorf("error on ping oracle database connection:%v", err)
	}
	return &Oracle{
		Ctx:      ctx,
		OracleDB: sqlDB,
	}, nil
}

func Query(ctx context.Context, db *sql.DB, querySQL string) ([]string, []map[string]string, error) {
	var (
		cols []string
		res  []map[string]string
	)
	rows, err := db.QueryContext(ctx, querySQL)
	if err != nil {
		return cols, res, fmt.Errorf("general sql [%v] query failed: [%v]", querySQL, err.Error())
	}
	defer rows.Close()

	//不确定字段通用查询，自动获取字段名称
	cols, err = rows.Columns()
	if err != nil {
		return cols, res, fmt.Errorf("general sql [%v] query rows.Columns failed: [%v]", querySQL, err.Error())
	}

	values := make([][]byte, len(cols))
	scans := make([]interface{}, len(cols))
	for i := range values {
		scans[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			return cols, res, fmt.Errorf("general sql [%v] query rows.Scan failed: [%v]", querySQL, err.Error())
		}

		row := make(map[string]string)
		for k, v := range values {
			// Oracle/Mysql 对于 'NULL' 统一字符 NULL 处理，查询出来转成 NULL,所以需要判断处理
			// 查询字段值 NULL
			// 如果字段值 = NULLABLE 则表示值是 NULL
			// 如果字段值 = "" 则表示值是空字符串
			// 如果字段值 = 'NULL' 则表示值是 NULL 字符串
			// 如果字段值 = 'null' 则表示值是 null 字符串
			if v == nil {
				row[cols[k]] = "NULLABLE"
			} else {
				// 处理空字符串以及其他值情况
				// 数据统一 string 格式显示
				row[cols[k]] = string(v)
			}
		}
		res = append(res, row)
	}

	if err = rows.Err(); err != nil {
		return cols, res, fmt.Errorf("general sql [%v] query rows.Next failed: [%v]", querySQL, err.Error())
	}
	return cols, res, nil
}

// func BulkInsert(db *sql.DB, querySQL string) error {
// 	// conn, err := go_ora.NewConnection(databaseUrl)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// err = conn.Open()
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// defer func() {
// 	// 	err = conn.Close()
// 	// 	if err != nil {
// 	// 		fmt.Println("Can't close connection: ", err)
// 	// 	}
// 	// }()
// 	// t := time.Now()
// 	// sqlText := `INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, NAME, VAL, VISIT_DATE) VALUES(:1, :2, :3, :4)`
// 	rowNum := 100
// 	col01 := make([]driver.Value, rowNum)
// 	col02 := make([]driver.Value, rowNum)
// 	for index := 0; index < rowNum; index++ {
//         col01[index] = index
//         col02[index] = index * index
// 	}
// 	result, err := db.BulkInsert(querySQL, col01, col02)
// 	if err != nil {
// 		return err
// 	}
// 	rowsAffected, _ := result.RowsAffected()
// 	fmt.Printf("%d rows inserted: \n", rowsAffected )
// 	return nil
// }

func (o *Oracle) InsertData(schemaName, tableName string, cols *[]map[string]string, data *[][]interface{} ) error {
    querySQL := prepareInsert(schemaName + "." + tableName, cols)
    fmt.Printf("inser query: [%s] \n", querySQL)

	stmt, err := o.OracleDB.Prepare(querySQL)
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()

    for _, row := range *data{
	    _, err = stmt.Exec(row...)
	    if err != nil {
	    	return err
	    }
    }
	return nil
}

func (o *Oracle) GetOracleSchemas() ([]string, error) {
	var (
		schemas []string
		err     error
	)
	cols, res, err := Query(o.Ctx, o.OracleDB, `SELECT DISTINCT username FROM DBA_USERS`)
	if err != nil {
		return schemas, err
	}
	for _, col := range cols {
		for _, r := range res {
			schemas = append(schemas, common.StringUPPER(r[col]))
		}
	}
	return schemas, nil
}

func (o *Oracle) GetOracleSchemaTable(schemaName string) ([]string, error) {
	var (
		tables []string
		err    error
	)
	_, res, err := Query(o.Ctx, o.OracleDB, fmt.Sprintf(`SELECT table_name AS TABLE_NAME FROM DBA_TABLES WHERE UPPER(owner) = UPPER('%s') AND (IOT_TYPE IS NUll OR IOT_TYPE='IOT')`, schemaName))
	if err != nil {
		return tables, err
	}
	for _, r := range res {
		tables = append(tables, strings.ToUpper(r["TABLE_NAME"]))
	}

	return tables, nil
}

func (o *Oracle) GetTableColDef(schemaName, tableName string) (*[]map[string]string, error){
	_, res, err := Query(o.Ctx, o.OracleDB, fmt.Sprintf(`SELECT COLUMN_NAME, DATA_TYPE, DATA_LENGTH, DATA_PRECISION, DATA_SCALE, NULLABLE, CHARACTER_SET_NAME, COLLATION, USER_GENERATED FROM  DBA_TAB_COLS WHERE UPPER(owner) = UPPER('%s') AND UPPER(TABLE_NAME) = UPPER('%s') ORDER BY COLUMN_ID`, schemaName, tableName))
	if err != nil {
		return nil, err
	}

    return &res, nil
}


func prepareInsert(fullTableName string, cols *[]map[string]string) string {
    var builder strings.Builder
    columnNames := make([]string, 0, len(*cols))
    for _, col := range *cols {
        if col == nil || col["USER_GENERATED"] == "NO" {
            continue
        }
        columnNames = append(columnNames, col["COLUMN_NAME"])
    }
    colList := "(" + buildColumnList(columnNames) + ")" 

    builder.WriteString("INSERT INTO " + fullTableName + colList + " VALUES ") 

    builder.WriteString("(" + placeHolder(len(columnNames)) + ")")

    return builder.String()
}

func buildColumnList(names []string) string {
	var b strings.Builder
	for i, name := range names {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(name)

	}

	return b.String()
}

func placeHolder(n int) string {
    var builder strings.Builder
    builder.Grow((n-1)*2 + 1)
    for i := 0; i < n; i++ {
        if i > 0 {
            builder.WriteString(",")
        }
        builder.WriteString(fmt.Sprintf(":%d", i+1))
    }
    return builder.String()
}
