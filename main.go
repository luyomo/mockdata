package main

import (
    "bufio"
    "bytes"
    "fmt"
    "gopkg.in/yaml.v3"
    "html/template"
    "io/ioutil"
    "log"
    "math/rand"
    "os"
    "runtime"
    "strconv"
    "strings"
    "sync"
    "time"
    "regexp"
    "encoding/hex"

    "archive/tar"
    "compress/gzip"
    "io"

    "os/exec"

    "errors"
    "github.com/spf13/cobra"

    "github.com/luyomo/mockdata/embed"
    "github.com/google/uuid"
)

const (
    // String
    CHARSET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
    BROWSER = [...]string{"Chrome", "IE", "Safari"}
)

type TiDBLightningConn struct {
    TiDBHost         string
    TiDBPort         int
    TiDBUser         string
    TiDBPassword     string
    PDIP             string
    DataFolder       string
    LightningVersion string
}

var rootCmd = &cobra.Command{
    Use:   "mock",
    Short: "Generate mock data",
    Long:  `Generate mock data `,
    Run: func(cmd *cobra.Command, args []string) {
        // fmt.Printf("Hello world \n")
    },
}

type MockDataStructure struct {
    Columns []struct {
        Idx        int    `yaml:"IDX"`
        Name       string `yaml:"Name"`
        DataType   string `yaml:"DataType"`
        Function   string `yaml:"Function"`
        Values     []string `yaml:"Values"`
        Max        int    `yaml:"Max"`
        Min        int    `yaml:"Min"`
        Parameters []struct {
            Key   string `yaml:"key"`
            Value string `yaml:"value"`
        } `yaml:"Parameters"`
    } `yaml:"COLUMNS"`
    Rows int `yaml:"ROWS"`
}

var MAPFunc = make(map[string]func()(string, error))

func main() {

    var threads, rows, loop, base int
    var configFile   string
    var outputFolder string
    var fileName     string
    var lightningVer string

    var dbConn TiDBLightningConn

    MAPFunc["BROWSER"] = func()(string, error){
        _value := BROWSER[rand.Intn(len(BROWSER))]
        return _value, nil
    }

    MAPFunc["IPADDR"] = func()(string, error){
        _value := fmt.Sprintf("%d.%d.%d.%d", 1+rand.Intn(254),rand.Intn(255), rand.Intn(255),rand.Intn(255)  )
        return _value, nil
    }

    MAPFunc["UUID"] = func()(string, error){
        return uuid.New().String(), nil
    }

    MAPFunc["RandomInt"] = func()(string, error){
        s1 := rand.NewSource(time.Now().UnixNano())
        r2 := rand.New(s1)
        _data := strconv.Itoa(r2.Intn(100))
        return _data, nil
    }

    MAPFunc["RandomHex"] = func()(string, error){
        rand.Seed(time.Now().UnixNano())

        // Getting random character
	_data := ""
        for _idx := 0; _idx < 10 ; _idx++ {
            c := CHARSET[rand.Intn(len(CHARSET))]
            _data += string(c)
        }

	return hex.EncodeToString([]byte(_data)), nil
    }

    //

    // Set the arguments
    rootCmd.PersistentFlags().IntVar(&threads, "threads", runtime.NumCPU(), "Threads to generate the data")
    rootCmd.PersistentFlags().IntVar(&rows, "rows", 1, "Number of rows for each thread")
    rootCmd.PersistentFlags().IntVar(&loop, "loop", 1, "TiDB Lightning loop")
    rootCmd.PersistentFlags().IntVar(&base, "base", 0, "base number of the PK")
    rootCmd.PersistentFlags().StringVar((*string)(&configFile), "config", "", "Config file for data generattion")
    rootCmd.PersistentFlags().StringVar((*string)(&outputFolder), "output", "", "Output folder for data generattion")
    rootCmd.PersistentFlags().StringVar((*string)(&fileName), "file-name", "", "file or table name for data generattion. For tidb lightning, please user schema_name.table_name.csv")
    rootCmd.PersistentFlags().StringVar((*string)(&lightningVer), "lightning-ver", "v7.5.2", "The version of the lighting")

    rootCmd.PersistentFlags().StringVar((*string)(&dbConn.TiDBHost), "host", "", "TiDB Host name")
    rootCmd.PersistentFlags().IntVar(&dbConn.TiDBPort, "port", 4000, "TiDB Port")
    rootCmd.PersistentFlags().StringVar((*string)(&dbConn.TiDBUser), "user", "root", "TiDB User")
    rootCmd.PersistentFlags().StringVar((*string)(&dbConn.TiDBPassword), "password", "", "TiDB Password")
    rootCmd.PersistentFlags().StringVar((*string)(&dbConn.PDIP), "pd-ip", "", "pd ip address")

    rootCmd.Execute()
    //fmt.Printf("The TiDB config info is <%#v> \n", dbConn)

    // Install TiDB lightning locally
    InstallTiDBLightning(lightningVer)

    // Read the data config file. (example: etc/data.config.yaml)
    yfile, err := ioutil.ReadFile(configFile)

    if err != nil {
        log.Fatal(err)
    }

    var mockDataConfig MockDataStructure

    err = yaml.Unmarshal([]byte(yfile), &mockDataConfig)
    if err != nil {
        log.Fatalf("error: %v", err)
    }

    // Prepare the folder to keep the data
    csvOutputFolder := outputFolder + "/data"
    cmd := exec.Command("mkdir", "-p", csvOutputFolder )
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
            panic(err)
            return
    }

    for _loop := 0; _loop < loop; _loop++  {
        var waitGroup sync.WaitGroup
        //errChan := make(chan error , 2)

        // fmt.Printf("The number thread is <%d> \n", threads)
        for _idx := 0; _idx < threads; _idx++ {
            waitGroup.Add(1)
            go func(_index int) {
                // fmt.Printf("The index is <%d> \n", _index + _loop*threads)

                csvFile := fmt.Sprintf("%s/%s.%03d.csv", csvOutputFolder, fileName, _index)
                GenerateDataTo(base + _index + _loop*threads, rows, mockDataConfig, csvFile)

                defer waitGroup.Done()
            }(_idx)
        }

        waitGroup.Wait()

        dbConn.DataFolder = csvOutputFolder
        lightningConfigFile := fmt.Sprintf("%s/tidb-lightning.toml", outputFolder)
        parseTemplate(dbConn, lightningConfigFile )
        // fmt.Printf("The file is <%s> \n", lightningConfigFile )
        // fmt.Printf("Starting to call %s \n", fmt.Sprintf( "/tmp/temp%d.txt", rand.Intn(100)))
        cmd = exec.Command("mockdata/bin/tidb-lightning", "--config", lightningConfigFile )
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            panic(err)
            return
        }

        csvOutputFolder := outputFolder + "/data"
        cmd := exec.Command("rm", "-f", csvOutputFolder + "/*" )
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            panic(err)
            return
        }
    }
}

type RandomUserID struct {
    min       int
    max       int
    MsgFormat string
}

func (p RandomUserID) GenerateData() string {
    s1 := rand.NewSource(time.Now().UnixNano())
    r2 := rand.New(s1)
    _data := r2.Intn(p.max)
    return fmt.Sprintf(p.MsgFormat, _data)
}

func GenerateDataTo(threads, rows int, dataConfig MockDataStructure, file string) error {
    // Prepare the file handle to output the data into
    fFile, err := os.Create(file)
    if err != nil {
        log.Fatal(err)
    }
    writer := bufio.NewWriter(fFile)

    // Generate the random number generation instance
    s1 := rand.NewSource(time.Now().UnixNano())

    // Generate the string template instance

    var mapTemplate = make(map[string]*template.Template)
    var mapFunc = make(map[string]func()(string, error))
    var mapFuncs = make(map[string]map[string]func()(string, error))

    var mapGeneration = make(map[string]interface{})

    // Define all the implementation here to improve the performance. Prepare the implementation only one time.
    for _, column := range dataConfig.Columns {
        // Define the function for template.
        if column.Function == "template" {
            var _min, _max int
            var _format, _content string
            for _, _data := range column.Parameters {
                if _data.Key == "min" {
                    _min, err = strconv.Atoi(_data.Value)
                    if err != nil {
                        return err
                    }
                }
                if _data.Key == "max" {
                    _max, err = strconv.Atoi(_data.Value)
                    if err != nil {
                        return err
                    }
                }
                if _data.Key == "format" {
                    _format = _data.Value
                }
                if _data.Key == "content" {
                    _content = _data.Value
                }
            }

            tmpl, err := template.New("").Parse(_content)

            if err != nil {
                log.Fatalf("Parse: %v", err)
            }
            mapTemplate[column.Name] = tmpl

            userGenerate := RandomUserID{_min, _max, _format}

            mapGeneration[column.Name] = &userGenerate
        }

        if column.Function == "Template" {
            var _content string
            for _, _data := range column.Parameters {
                if _data.Key == "content" {
                    _content = _data.Value
                }
            }
            // 1. Get the #IPADDR
            re := regexp.MustCompile("(?U){{\\$(.*)}}")
            ret := re.FindAllStringSubmatch(_content, -1)
            // fmt.Printf("The result is %#v \n", ret)

            var _mapFunc = make(map[string]func()(string, error))

            for _ , _match := range ret {
                _content = strings.Replace(_content, _match[0], fmt.Sprintf("{{ index .Data \"%s\"}}", _match[1]), -1 )

                _mapFunc[_match[1]] = MAPFunc[_match[1]]
            }

            mapFuncs[column.Name] = _mapFunc

            tmpl, err := template.New("").Parse(_content)
            if err != nil {
                log.Fatalf("Parse: %v", err)
            }
            mapTemplate[column.Name] = tmpl
        }

        // Generate the random date between min and max
        if column.Function == "RandomDate" {
           var _min, _max time.Time
           for _, _data := range column.Parameters {
               if _data.Key == "min" {
                   _min, err = time.Parse("2006-01-02", _data.Value)
                   if err != nil {
                       panic(err)
                       return err
                   }
               }
               if _data.Key == "max" {
                   _max, err = time.Parse("2006-01-02", _data.Value)
                   if err != nil {
                       return err
                   }
               }
           }
           _days := int(_max.Sub(_min).Hours() / 24)
           mapFunc[column.Name] = func() (string, error) {
               return _min.AddDate(0, 0, rand.Intn(_days)).Format("2006-01-02"), nil
           }
       }
    }

    for idx := 1; idx <= rows; idx++ {
        var arrData []string
        for _, column := range dataConfig.Columns {
            var _data string
            if column.Function == "sequence" {
                _data = strconv.Itoa(threads*rows + idx)
            }

            if column.Function == "uuid" {
                _data = uuid.New().String()
            }

            if column.Function == "list" {
                _data = column.Values[rand.Intn(len(column.Values))]
            }

            if column.Function == "random" {
                if column.DataType == "int" {
                    r2 := rand.New(s1)
                    _data = strconv.Itoa(r2.Intn(column.Max))
                }
            }

            if column.Function == "RandomDate" {
                 _data, err = mapFunc[column.Name]()
                 if err != nil{
                     panic(err)
                 }
            }

	    // Make the function common
            if column.Function == "RandomString" {
                var _min, _max int
                for _, _p := range column.Parameters {
                    if _p.Key == "min" {
                        _min, err = strconv.Atoi(_p.Value)
                        if err != nil {
                            return err
                        }
                    }
                    if _p.Key == "max" {
                        _max, err = strconv.Atoi(_p.Value)
                        if err != nil {
                            return err
                        }
                    }
                }
                rand.Seed(time.Now().UnixNano())

                // Getting random character
                _data = ""
                for _idx := 0; _idx < _min + rand.Intn(_max - _min + 1) ; _idx++ {
                    c := CHARSET[rand.Intn(len(CHARSET))]
                    _data += string(c)
                }
            }

            if column.Function == "UniqueRandomString" {
                var _min, _max int
                for _, _p := range column.Parameters {
                    if _p.Key == "min" {
                        _min, err = strconv.Atoi(_p.Value)
                        if err != nil {
                            return err
                        }
                    }
                    if _p.Key == "max" {
                        _max, err = strconv.Atoi(_p.Value)
                        if err != nil {
                            return err
                        }
                    }
                }
                rand.Seed(time.Now().UnixNano())

                // Getting random character
                _data = ""
                for _idx := 0; _idx < _min + rand.Intn(_max - _min + 1) ; _idx++ {
                    c := CHARSET[rand.Intn(len(CHARSET))]
                    _data += string(c)
                }
                _data += strconv.Itoa(threads*rows + idx)
            }

            if column.Function == "template" {
                userGenerate := mapGeneration[column.Name]

                tmpl := mapTemplate[column.Name]
                var test bytes.Buffer
                tmpl.Execute(&test, userGenerate)
                _data = test.String()
            }

            if column.Function == "Template" {
                type TplData struct {
                    Data map[string]string
                }
                var mapParams = make(map[string]string)

                _funcs := mapFuncs[column.Name]
                for _key, _func := range _funcs {
                    _value, err := _func()
                    if err != nil {
                        return err
                    }
                    mapParams[_key] = _value
//                    fmt.Printf("The key: <%#v>, value: <%#v>", _key, _func)
                }

                // mapParams["IPADDR"] = "192.168.1.2"
                // mapParams["BROWSER"] = "Chrome"
                var tplData TplData
                tplData.Data = mapParams

                tmpl := mapTemplate[column.Name]
                var test bytes.Buffer
                tmpl.Execute(&test, tplData)
                _data = strings.Replace(test.String(), "\"", "\\\"", -1)
            }

            if column.DataType != "int" {
                _data = "\"" + _data + "\""
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

func InstallTiDBLightning(lightningVer string) error {
    if runtime.GOOS == "windows" {
        fmt.Println("Can't Execute this on a windows machine")
    } else {
        if _, err := os.Stat("mockdata/bin/tidb-lightning"); errors.Is(err, os.ErrNotExist) {
            binFile := fmt.Sprintf("tidb-community-toolkit-%s-linux-%s.tar.gz", lightningVer, runtime.GOARCH)
            fullBinFile := fmt.Sprintf("tidb-community-toolkit-%s-linux-%s/tidb-lightning-%s-linux-%s.tar.gz", lightningVer, runtime.GOARCH, lightningVer, runtime.GOARCH)

            cmd := exec.Command("wget", "https://download.pingcap.org/"+binFile, "-O", "/tmp/"+binFile)
            cmd.Stdout = os.Stdout
            cmd.Stderr = os.Stderr
            if err := cmd.Run(); err != nil {
                return err
            }

            r, err := os.Open("/tmp/" + binFile)
            if err != nil {
                fmt.Println("error")
            }
            ExtractTarGz(r, []string{fullBinFile})

            cmd = exec.Command("rm", "-rf", "/tmp/"+binFile)
            cmd.Stdout = os.Stdout
            cmd.Stderr = os.Stderr
            if err := cmd.Run(); err != nil {
                return err
            }

            r, err = os.Open(fmt.Sprintf("tidb-community-toolkit-%s-linux-%s/tidb-lightning-%s-linux-%s.tar.gz", lightningVer, runtime.GOARCH, lightningVer, runtime.GOARCH  ))
            if err != nil {
                fmt.Println("error")
            }
            ExtractTarGz(r, []string{"tidb-lightning"})

            cmd = exec.Command("mkdir", "-p", "mockdata/bin")
            cmd.Stdout = os.Stdout
            cmd.Stderr = os.Stderr
            if err := cmd.Run(); err != nil {
                return err
            }

            cmd = exec.Command("mv", "tidb-lightning", "mockdata/bin/")
            cmd.Stdout = os.Stdout
            cmd.Stderr = os.Stderr
            if err := cmd.Run(); err != nil {
                return err
            }

            cmd = exec.Command("chmod", "755", "mockdata/bin/tidb-lightning")
            cmd.Stdout = os.Stdout
            cmd.Stderr = os.Stderr
            if err := cmd.Run(); err != nil {
                return err
            }

            cmd = exec.Command("rm", "-rf", fmt.Sprintf("tidb-community-toolkit-%s-linux-%s", lightningVer, runtime.GOARCH))
            cmd.Stdout = os.Stdout
            cmd.Stderr = os.Stderr
            if err := cmd.Run(); err != nil {
                return err
            }
        }
                // fmt.Print("Completed the file check \n")
    }
    return nil
}

func ExtractTarGz(gzipStream io.Reader, files []string) {
    uncompressedStream, err := gzip.NewReader(gzipStream)
    if err != nil {
        fmt.Printf("The error is <%#v>", err)
        log.Fatal("ExtractTarGz: NewReader failed")
    }

    tarReader := tar.NewReader(uncompressedStream)

    for true {
        header, err := tarReader.Next()

        if err == io.EOF {
            break
        }

        if err != nil {
            log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
        }

        switch header.Typeflag {
        case tar.TypeDir:
            fmt.Printf("The folder name is %s \n", header.Name)
            if err := os.Mkdir(header.Name, 0755); err != nil {
                log.Fatalf("ExtractTarGz: Mkdir() failed: %s", err.Error())
            }
        case tar.TypeReg:
            fmt.Printf("The file name is %s \n", header.Name)
            if contains(files, header.Name) {
                outFile, err := os.Create(header.Name)
                if err != nil {
                    log.Fatalf("ExtractTarGz: Create() failed: %s", err.Error())
                }
                if _, err := io.Copy(outFile, tarReader); err != nil {
                    log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
                }
                outFile.Close()
            }

        default:
            log.Fatalf(
                "ExtractTarGz: uknown type: %s in %s",
                header.Typeflag,
                header.Name)
        }

    }
}

func contains(s []string, str string) bool {
    for _, v := range s {
        if v == str {
            return true
        }
    }

    return false
}

func parseTemplate(dbConn TiDBLightningConn, configFile string) {
    data, err := embed.ReadTemplate("templates/tidb-lightning.toml.tpl")
    if err != nil {
        panic(err)
    }

    tmpl, err := template.New("").Parse(string(data))
    if err != nil {
        log.Fatalf("Parse: %v", err)
    }

    var ret bytes.Buffer
    err = tmpl.Execute(&ret, dbConn)

    fo, err := os.Create(configFile)
    if err != nil {
        panic(err)
    }

    defer func() {
        if err := fo.Close(); err != nil {
            panic(err)
        }
    }()

    if _, err := fo.Write(ret.Bytes()); err != nil {
        panic(err)
    }

}
