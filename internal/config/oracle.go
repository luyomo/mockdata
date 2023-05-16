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
package config

import (
    //"encoding/json"
    "flag"
    "fmt"
    //"github.com/wentaojin/transferdb/common"
    "os"
    "strings"

    "github.com/BurntSushi/toml"
)

type Config struct {
     *flag.FlagSet `json:"-"`
//     AppConfig     AppConfig     `toml:"app" json:"app"`
//     ReverseConfig ReverseConfig `toml:"reverse" json:"reverse"`
//     CheckConfig   CheckConfig   `toml:"check" json:"check"`
//     FullConfig    FullConfig    `toml:"full" json:"full"`
//     CSVConfig     CSVConfig     `toml:"csv" json:"csv"`
//     AllConfig     AllConfig     `toml:"all" json:"all"`
     OracleConfig  OracleConfig  `toml:"oracle" json:"oracle"`
//     MySQLConfig   MySQLConfig   `toml:"mysql" json:"mysql"`
//     MetaConfig    MetaConfig    `toml:"meta" json:"meta"`
//     LogConfig     LogConfig     `toml:"log" json:"log"`
//     DiffConfig    DiffConfig    `toml:"compare" json:"compare"`
    ConfigFile    string        `json:"config-file"`
    TablesArg     string        `json:"tables"`
    Tables        []map[string]string
    NumOfRows     int           `json:"number-of-rows"`
    Prompt        bool          `json:"prompt"`
    
//     PrintVersion  bool
//     TaskMode      string `json:"task-mode"`
//     DBTypeS       string `json:"db-type-s"`
//     DBTypeT       string `json:"db-type-t"`
}
// 
// type AppConfig struct {
//     InsertBatchSize  int    `toml:"insert-batch-size" json:"insert-batch-size"`
//     SlowlogThreshold int    `toml:"slowlog-threshold" json:"slowlog-threshold"`
//     PprofPort        string `toml:"pprof-port" json:"pprof-port"`
// }
// 
// type DiffConfig struct {
//     ChunkSize         int           `toml:"chunk-size" json:"chunk-size"`
//     DiffThreads       int           `toml:"diff-threads" json:"diff-threads"`
//     OnlyCheckRows     bool          `toml:"only-check-rows" json:"only-check-rows"`
//     EnableCheckpoint  bool          `toml:"enable-checkpoint" json:"enable-checkpoint"`
//     IgnoreStructCheck bool          `toml:"ignore-struct-check" json:"ignore-struct-check"`
//     FixSqlDir         string        `toml:"fix-sql-dir" json:"fix-sql-dir"`
//     TableConfig       []TableConfig `toml:"table-config" json:"table-config"`
// }
// 
// type ReverseConfig struct {
//     ReverseThreads   int    `toml:"reverse-threads" json:"reverse-threads"`
//     DirectWrite      bool   `toml:"direct-write" json:"direct-write"`
//     DDLReverseDir    string `toml:"ddl-reverse-dir" json:"ddl-reverse-dir"`
//     DDLCompatibleDir string `toml:"ddl-compatible-dir" json:"ddl-compatible-dir"`
// }
// 
// type CheckConfig struct {
//     CheckThreads int    `toml:"check-threads" json:"check-threads"`
//     CheckSQLDir  string `toml:"check-sql-dir" json:"check-sql-dir"`
// }
// 
// type TableConfig struct {
//     SourceTable string `toml:"source-table" json:"source-table"`
//     IndexFields string `toml:"index-fields" json:"index-fields"`
//     Range       string `toml:"range" json:"range"`
// }
// 
// type CSVConfig struct {
//     Header           bool   `toml:"header" json:"header"`
//     Separator        string `toml:"separator" json:"separator"`
//     Terminator       string `toml:"terminator" json:"terminator"`
//     Delimiter        string `toml:"delimiter" json:"delimiter"`
//     EscapeBackslash  bool   `toml:"escape-backslash" json:"escape-backslash"`
//     Charset          string `toml:"charset" json:"charset"`
//     Rows             int    `toml:"rows" json:"rows"`
//     OutputDir        string `toml:"output-dir" json:"output-dir"`
//     TaskThreads      int    `toml:"task-threads" json:"task-threads"`
//     TableThreads     int    `toml:"table-threads" json:"table-threads"`
//     SQLThreads       int    `toml:"sql-threads" json:"sql-threads"`
//     EnableCheckpoint bool   `toml:"enable-checkpoint" json:"enable-checkpoint"`
// }
// 
// type FullConfig struct {
//     ChunkSize        int  `toml:"chunk-size" json:"chunk-size"`
//     TaskThreads      int  `toml:"task-threads" json:"task-threads"`
//     TableThreads     int  `toml:"table-threads" json:"table-threads"`
//     SQLThreads       int  `toml:"sql-threads" json:"sql-threads"`
//     ApplyThreads     int  `toml:"apply-threads" json:"apply-threads"`
//     EnableCheckpoint bool `toml:"enable-checkpoint" json:"enable-checkpoint"`
// }
// 
// type AllConfig struct {
//     LogminerQueryTimeout int `toml:"logminer-query-timeout" json:"logminer-query-timeout"`
//     FilterThreads        int `toml:"filter-threads" json:"filter-threads"`
//     ApplyThreads         int `toml:"apply-threads" json:"apply-threads"`
//     WorkerQueue          int `toml:"worker-queue" json:"worker-queue"`
//     WorkerThreads        int `toml:"worker-threads" json:"worker-threads"`
// }

type OracleConfig struct {
    Username      string   `toml:"username" json:"username"`
    Password      string   `toml:"password" json:"password"`
    Host          string   `toml:"host" json:"host"`
    Port          int      `toml:"port" json:"port"`
    ServiceName   string   `toml:"service-name" json:"service-name"`
    PDBName       string   `toml:"pdb-name" json:"pdb-name"`
    LibDir        string   `toml:"lib-dir" json:"lib-dir"`
    ConnectParams string   `toml:"connect-params" json:"connect-params"`
    SessionParams []string `toml:"session-params" json:"session-params"`
    SchemaName    string   `toml:"schema-name" json:"schema-name"`
    IncludeTable  []string `toml:"include-table" json:"include-table"`
    ExcludeTable  []string `toml:"exclude-table" json:"exclude-table"`
}

// type MySQLConfig struct {
//     DBType        string `toml:"db-type" json:"db-type"`
//     Username      string `toml:"username" json:"username"`
//     Password      string `toml:"password" json:"password"`
//     Host          string `toml:"host" json:"host"`
//     Port          int    `toml:"port" json:"port"`
//     ConnectParams string `toml:"connect-params" json:"connect-params"`
//     MetaSchema    string `toml:"meta-schema" json:"meta-schema"`
//     SchemaName    string `toml:"schema-name" json:"schema-name"`
//     TableOption   string `toml:"table-option" json:"table-option"`
//     Overwrite     bool   `toml:"overwrite" json:"overwrite"`
// }

// type MetaConfig struct {
//     Username   string `toml:"username" json:"username"`
//     Password   string `toml:"password" json:"password"`
//     Host       string `toml:"host" json:"host"`
//     Port       int    `toml:"port" json:"port"`
//     MetaSchema string `toml:"meta-schema" json:"meta-schema"`
// }
// 
// type LogConfig struct {
//     LogLevel   string `toml:"log-level" json:"log-level"`
//     LogFile    string `toml:"log-file" json:"log-file"`
//     MaxSize    int    `toml:"max-size" json:"max-size"`
//     MaxDays    int    `toml:"max-days" json:"max-days"`
//     MaxBackups int    `toml:"max-backups" json:"max-backups"`
// }

func NewConfig() *Config {
    cfg := &Config{}
    cfg.FlagSet = flag.NewFlagSet("mock data to oracle", flag.ContinueOnError)
    fs := cfg.FlagSet
    fs.Usage = func() {
        fmt.Fprintln(os.Stderr, "Usage of oracle mock data:")
        fs.
            PrintDefaults()
    }
    // fs.BoolVar(&cfg.PrintVersion, "V", false, "print version information and exit")
    fs.StringVar(&cfg.ConfigFile, "config", "./config.toml", "path to the configuration file")
    // fs.StringVar(&cfg.TaskMode, "mode", "", "specify the program running mode: [prepare assess reverse full csv all check compare]")
    // fs.StringVar(&cfg.DBTypeS, "source", "oracle", "specify the source db type")
    // fs.StringVar(&cfg.DBTypeT, "target", "mysql", "specify the target db type")
    fs.StringVar(&cfg.TablesArg, "tables", "", "specify the tables to generate data into(schema01.table01,schema02.table02)")
    fs.IntVar(&cfg.NumOfRows, "num-of-rows", 10, "specify the number of rows to be inserted into db")
    fs.BoolVar(&cfg.Prompt, "prompt", true, "Prompt before data generation")
    return cfg
}

 func (c *Config) Parse(args []string) error {
     err := c.FlagSet.Parse(args)
     switch err {
     case nil:
     case flag.ErrHelp:
         os.Exit(0)
     default:
         os.Exit(2)
     }
 
//     if c.PrintVersion {
//         fmt.Println(GetRawVersionInfo())
//         os.Exit(0)
//     }
 
     if c.ConfigFile != "" {
         if err = c.configFromFile(c.ConfigFile); err != nil {
             return err
         }
     } else {
         return fmt.Errorf("no config file")
     }
 
     c.AdjustConfig()
 
     return nil
 }
 
func (c *Config) configFromFile(file string) error {
    if _, err := toml.DecodeFile(file, c); err != nil {
        return fmt.Errorf("failed decode toml config file %s: %v", file, err)
    }
    return nil
}
 
func (c *Config) AdjustConfig() error {
    arrTables := strings.Split(c.TablesArg, ",")
    for _, _table := range arrTables{
        _schemaTable := strings.Split(_table, ".")
        mapTable := make(map[string]string)

        mapTable["schema"] = _schemaTable[0]
        mapTable["table"] = _schemaTable[1]

        c.Tables = append(c.Tables, mapTable)
    }
//     c.DBTypeS = common.StringUPPER(c.DBTypeS)
//     c.DBTypeT = common.StringUPPER(c.DBTypeT)
//     c.TaskMode = common.StringUPPER(c.TaskMode)
//     c.OracleConfig.SchemaName = common.StringUPPER(c.OracleConfig.SchemaName)
//     c.OracleConfig.PDBName = common.StringUPPER(c.OracleConfig.PDBName)
//     c.MySQLConfig.SchemaName = common.StringUPPER(c.MySQLConfig.SchemaName)
// 
//     err := c.adjustCSVConfig()
//     if err != nil {
//         return err
//     }
// 
     return nil
}
 
// func (c *Config) adjustCSVConfig() error {
//     if c.CSVConfig.Separator == "" {
//         c.CSVConfig.Separator = ","
//     }
//     if c.CSVConfig.Terminator == "" {
//         c.CSVConfig.Terminator = "\r\n"
//     }
//     isSupport := false
//     if c.CSVConfig.Charset != "" {
//         switch strings.ToUpper(c.CSVConfig.Charset) {
//         case common.UTF8CharacterSetCSV:
//             isSupport = true
//         case common.GBKCharacterSetCSV:
//             isSupport = true
//         default:
//             isSupport = false
//         }
//     }
//     if !isSupport {
//         return fmt.Errorf("target db character is not support: [%s]", c.CSVConfig.Charset)
//     }
//     return nil
// }
// 
// func (c *Config) String() string {
//     cfg, err := json.Marshal(c)
//     if err != nil {
//         return "<nil>"
//     }
//     return string(cfg)
// }
