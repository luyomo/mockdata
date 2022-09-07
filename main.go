package main

import (
    "fmt"
    "gopkg.in/yaml.v3"
    "io/ioutil"
    "log"
    "bufio"
    "os"
    "strconv"
    "strings"
    "math/rand"
    "time"
    "html/template"
    "bytes"
    "sync"
)

type MockDataStructure struct {
    Columns []struct {
        Idx int `yaml:"IDX"`
        Name string `yaml:"Name"`
        DataType string `yaml:"DataType"`
        Function string `yaml:"Function"`
        Max int `yaml:"Max"`
        Min int `yaml:"Min"`
        Parameters []struct {
          Key string `yaml:"key"`
          Value string `yaml:"value"`
        } `yaml:"Parameters"`
    } `yaml:"COLUMNS"`
    Rows int `yaml:"ROWS"`
}

func main() {
    fmt.Printf("Hello world \n")
    yfile, err := ioutil.ReadFile("./etc/data.config.yaml")

    if err != nil {
         log.Fatal(err)
    }

    var mockDataConfig MockDataStructure

    err = yaml.Unmarshal([]byte(yfile), &mockDataConfig)
    if err != nil {
        log.Fatalf("error: %v", err)
    }
    fmt.Printf("--- t:\n%v\n\n", mockDataConfig)
    fmt.Printf("The number of rows is <%d> \n", mockDataConfig.Rows)
    for _, _columnCfg := range mockDataConfig.Columns {
        fmt.Printf("Column name is : %s and data type: %s, Function: %s, Max: %d, Min: %d \n", _columnCfg.Name, _columnCfg.DataType, _columnCfg.Function, _columnCfg.Min, _columnCfg.Max)
    }

    var waitGroup sync.WaitGroup
    //errChan := make(chan error , 2)

    for _idx:=0; _idx < 2; _idx++ {
        waitGroup.Add(1)
        go func(){
            fmt.Printf("Starting to call %s \n", fmt.Sprintf( "/tmp/temp%d.txt", rand.Intn(100)))
            defer waitGroup.Done()
            GenerateDataTo(mockDataConfig, fmt.Sprintf( "/tmp/temp%d.txt", rand.Intn(100)) )
        } ()
    }
    waitGroup.Wait()
}


type RandomUserID struct{
    min int
    max int
    MsgFormat string
}

func (p RandomUserID) Generate() string {
    s1 := rand.NewSource(time.Now().UnixNano())
    r2 := rand.New(s1)
    _data := r2.Intn(p.max)
    return fmt.Sprintf(p.MsgFormat,  _data)
}

func GenerateDataTo(dataConfig MockDataStructure, file string) (error){
    fFile, err := os.Create(file)
    if err != nil {
        log.Fatal(err)
    }
    writer := bufio.NewWriter(fFile)

    s1 := rand.NewSource(time.Now().UnixNano())

    tmpl, err := template.New("").Parse(`{{.Generate}}`)
    if err != nil {
        log.Fatalf("Parse: %v", err)
    }
    userGenerate := RandomUserID{1, 100, "user%010d"}

    for  idx := 1; idx <= dataConfig.Rows; idx++ {
        var arrData []string
        for _, column := range dataConfig.Columns {
            var _data string
            if column.Function == "sequence" {
                _data = strconv.Itoa(idx)
            }

            if column.Function == "random" {
                if column.DataType == "int" {
                    r2 := rand.New(s1)
                    _data = strconv.Itoa(r2.Intn(column.Max))
                }
            }
            if column.Function == "template" {
                //fmt.Printf("The template parametes are <%#v> \n", column.Parameters)
                //for _, param := range column.Parameters {
                //    fmt.Printf("The data is <%#v> \n", param)
                //}
                var test bytes.Buffer
                // tmpl.Execute(&test, Person("Bob"))
                // tmpl.Execute(&test, RandomUserID{1, 100, "user%010d"})
                tmpl.Execute(&test, userGenerate )
                _data = test.String()
                // fmt.Printf("\nThe test data is <%s>\n\n", test.String())
            }

            if column.DataType != "int" {
                _data = "'" + _data + "'"
            }

            arrData = append(arrData, _data)
        }
        //fmt.Printf("The data is <%#v> and <%s> \n", arrData, strings.Join(arrData, ","))
        _, err := writer.WriteString(strings.Join(arrData, ",") + "\n")
        if err != nil {
            return err
        }
    }
    writer.Flush()
    return nil
}
