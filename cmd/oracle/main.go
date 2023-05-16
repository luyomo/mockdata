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
package main

import (
	"context"
	// "github.com/wentaojin/transferdb/signal"
	"log"
	//"net/http"
	_ "net/http/pprof"
	"os"
    "fmt"

	//"github.com/pkg/errors"
	"github.com/luyomo/mockdata/internal/config"
	"github.com/luyomo/mockdata/internal/database/oracle"
	dt "github.com/luyomo/mockdata/internal/data"
	//"go.uber.org/zap"
    "github.com/pingcap/tiup/pkg/tui"
)

func main() {
	cfg := config.NewConfig()
    // fmt.Printf("Paramter : <%#v> \n", cfg)
	if err := cfg.Parse(os.Args[1:]); err != nil {
		log.Fatalf("start meta failed. error is [%s], Use '--help' for help.", err)
	}
    // fmt.Printf("---------- <%#v> \n", cfg)

    oracleDB, err := oracle.NewOracleDBEngine(context.Background(), cfg.OracleConfig)
    if err != nil {
        panic(err)
    }


    for _, _entry := range cfg.Tables {
        if _entry["table"] == "*" {
            schemaTables, err := oracleDB.GetOracleSchemaTable(_entry["schema"])
            if err != nil {
                panic(err)
            }
            for _, _table := range schemaTables {
                ret := false 
                if cfg.Prompt == true {
                    ret, _ = tui.PromptForConfirmNo(fmt.Sprintf("Do you want to generate the data for %s.%s?", _entry["schema"] , _table))
                }
                if ret == false {
                    fmt.Printf("Starting to generate data for %s.%s ... ... \n", _entry["schema"], _table)
                    tableInfo, err := oracleDB.GetTableInfo(_entry["schema"], _table)
                    if err != nil {
                        fmt.Printf("Failed to generate data for %s.%s", _entry["schema"], _table)
                        continue
                    }
                    GenTableData(oracleDB, tableInfo, nil, cfg.NumOfRows)
                    fmt.Println("")
                } else {
                    fmt.Printf("Skip the data generation ... ...  \n")
                }
            }
        } else {
            tableInfo, err := oracleDB.GetTableInfo(_entry["schema"], _entry["table"])
            if err != nil {
                panic(err)
            }
            // PrintTableInfo(tableInfo)

            GenTableData(oracleDB, tableInfo, nil, cfg.NumOfRows)
        }
    }
}

func PrintTableInfo(tableInfo *oracle.TableInfo){
    if tableInfo == nil {
        return
    }

    if tableInfo.RefTables != nil {
        for _col, _subTableInfo := range *tableInfo.RefTables {
            fmt.Printf("----- %s \n", _col)
            PrintTableInfo(&_subTableInfo)
        }
    }

    fmt.Printf("Schema Name: %s, Table Name: %s \n", tableInfo.SchemaName, tableInfo.TableName)
}

func GenTableData(oracleDB *oracle.Oracle, tableInfo *oracle.TableInfo, parentTable *string, numRows int){
    if tableInfo == nil {
        return
    }

    if tableInfo.RefTables != nil {
        for _, _subTableInfo := range *tableInfo.RefTables {
            _parentTable := fmt.Sprintf("%s.%s", tableInfo.SchemaName, tableInfo.TableName)
            GenTableData(oracleDB, &_subTableInfo, &_parentTable,  numRows)
        }
    }

    _count, err := oracleDB.CountTableRows(tableInfo.SchemaName, tableInfo.TableName)
    if err != nil {
        panic(err)
    }

    // Fetch the reference data
    // err = oracleDB.QueryRefTables(tableInfo.SchemaName, tableInfo.TableName)
    // if err != nil {
    //     panic(err)
    // }

    // return

    if _count > 0 && parentTable != nil {
        fmt.Printf("Table(%s.%s) depended by table(%s) has data(%d rows). Skip data generation \n", tableInfo.SchemaName, tableInfo.TableName, *parentTable, _count)
        return
    }

    // Fetch reference data map[string][]interface{}
    refData := make(map[string][]interface{})
    for _, col := range *tableInfo.Columns {
        if col["REF_QUERY"] != "" {
            // fmt.Printf("Column definition: <%#v> \n", col)
            refColData, err := oracleDB.QueryRefData(col["REF_QUERY"])
            if err != nil {
                panic(err)
            }
            refData[col["COLUMN_NAME"]] = *refColData
            // fmt.Printf("Column data: <%#v> \n", refData)
        }
    }

    // fmt.Printf("Columns: <%#v> \n", tableInfo.Columns)
    data, err := dt.GenerateOracleData(tableInfo.Columns, &refData, numRows)
    if err != nil {
        fmt.Printf("Failed to generate data (%s.%s) \n", tableInfo.SchemaName, tableInfo.TableName)
        panic(err)
    }

    numOfRows, err := oracleDB.InsertData(tableInfo.SchemaName, tableInfo.TableName, tableInfo.Columns, data)
    if err != nil {
        fmt.Printf("Failed to insert data (%s.%s) \n", tableInfo.SchemaName, tableInfo.TableName)
        panic(err)
    }
    if parentTable == nil {
        fmt.Printf("%d rows have been instered into table(%s.%s) \n", numOfRows, tableInfo.SchemaName, tableInfo.TableName)
    }else {
        fmt.Printf("%d rows have been instered into table(%s.%s) depended by table(%s) \n", numOfRows, tableInfo.SchemaName, tableInfo.TableName, *parentTable)
    }
}
