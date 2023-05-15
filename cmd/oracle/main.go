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

    // schemas, err := oracleDB.GetOracleSchemas()
    // if err != nil {
    //     panic(err)
    // }

    // fmt.Printf("The schemas : <%#v> \n", schemas)

    // fmt.Printf("The tables are : <%s> \n", cfg.Tables)

    for _, _entry := range cfg.Tables {
        tableInfo, err := oracleDB.GetTableInfo(_entry["schema"], _entry["table"])
        if err != nil {
            panic(err)
        }
        PrintTableInfo(tableInfo)

        GenTableData(oracleDB, tableInfo, cfg.NumOfRows)

        // tableDef, err := oracleDB.getTableColDef(_entry["schema"], _entry["table"])
        // if err != nil {
        //     panic(err)
        // }
        // for _, def := range *tableDef {
        //     fmt.Printf("The table definition: <%#v> \n", def)
        // }   

        // Look for reference table
        // mapRefTableInfo, err := oracleDB.GetReferenceTableColDef(_entry["schema"], _entry["table"])
        // if err != nil {
        //     panic(err)
        // }

        // fmt.Printf("The config info is: <%#v> \n", mapRefTableInfo)

        // fmt.Printf("reference table: <%#v> \n", arrRefTableInfo)
        // for _, _refTable := range *arrRefTableInfo {
        //     data, err := dt.GenerateOracleData(_refTable.Columns, cfg.NumOfRows)
        //     if err != nil {
        //         panic(err)
        //     }

        //     // fmt.Printf("Schema name : <%#v> \n", _refTable)

        //     numOfRows, err := oracleDB.InsertData(_refTable.SchemaName, _refTable.TableName, _refTable.Columns, data)
        //     if err != nil {
        //         panic(err)
        //     }
        //     fmt.Printf("%d rows have been instered into table(%s.%s) \n", numOfRows, _refTable.SchemaName, _refTable.TableName)
        // }

        // continue

        // data, err := dt.GenerateOracleData(tableDef, cfg.NumOfRows)
        // if err != nil {
        //     panic(err)
        // }
        // // fmt.Printf("Generated data: %#v \n", *data)

        // numOfRows, err := oracleDB.InsertData(_entry["schema"], _entry["table"], tableDef, data)
        // if err != nil {
        //     panic(err)
        // }
        // fmt.Printf("%d rows have been instered into table(%s.%s) \n", numOfRows, _entry["schema"], _entry["table"])
    }

	// // 初始化日志 logger
	// logger.NewZapLogger(cfg)
	// config.RecordAppVersion("transferdb", cfg)

	// go func() {
	// 	if err := http.ListenAndServe(cfg.AppConfig.PprofPort, nil); err != nil {
	// 		zap.L().Fatal("listen and serve pprof failed", zap.Error(errors.Cause(err)))
	// 	}
	// 	os.Exit(0)
	// }()

	// // 信号量监听处理
	// signal.SetupSignalHandler(func() {
	// 	os.Exit(1)
	// })

	// // 程序运行
	// ctx := context.Background()
	// if err := server.Run(ctx, cfg); err != nil {
	// 	zap.L().Fatal("server run failed", zap.Error(errors.Cause(err)))
	// }
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

func GenTableData(oracleDB *oracle.Oracle, tableInfo *oracle.TableInfo, numRows int){
    if tableInfo == nil {
        return
    }

    if tableInfo.RefTables != nil {
        for _, _subTableInfo := range *tableInfo.RefTables {
            GenTableData(oracleDB, &_subTableInfo, numRows)
        }
    }

    data, err := dt.GenerateOracleData(tableInfo.Columns, tableInfo.RefTables, numRows)
    if err != nil {
        panic(err)
    }

    numOfRows, err := oracleDB.InsertData(tableInfo.SchemaName, tableInfo.TableName, tableInfo.Columns, data)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%d rows have been instered into table(%s.%s) \n", numOfRows, tableInfo.SchemaName, tableInfo.TableName)
}
