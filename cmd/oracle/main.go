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
	//"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()
    fmt.Printf("Paramter : <%#v> \n", cfg)
	if err := cfg.Parse(os.Args[1:]); err != nil {
		log.Fatalf("start meta failed. error is [%s], Use '--help' for help.", err)
	}
    fmt.Printf("---------- <%#v> \n", cfg)

    oracleDB, err := oracle.NewOracleDBEngine(context.Background(), cfg.OracleConfig)
    if err != nil {
        panic(err)
    }

    schemas, err := oracleDB.GetOracleSchemas()
    if err != nil {
        panic(err)
    }

    fmt.Printf("The schemas : <%#v> \n", schemas)

    fmt.Printf("The tables are : <%s> \n", cfg.Tables)

    for _, _entry := range cfg.Tables {
        tableDef, err := oracleDB.GetTableColDef(_entry["schema"], _entry["table"])
        if err != nil {
            panic(err)
        }
        for _, def := range *tableDef {
            fmt.Printf("The table definition: <%#v> \n", def)
        }   

        data := make([][]interface{}, 5)
        for index := 1; index<=5; index++{
            row := make([]interface{}, 2)
            row[0] = index
            row[1] = index
            // row = append(row, index)
            // row = append(row, index)

            //data = append(data, row)
            data[index - 1] = row
        }
        fmt.Printf("data is: <%#v> \n", data)

        if err := oracleDB.InsertData(_entry["schema"], _entry["table"], tableDef, data); err != nil {
            panic(err)
        }
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
