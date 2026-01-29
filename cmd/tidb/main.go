package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"gopkg.in/yaml.v3"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"archive/tar"
	"compress/gzip"
	"io"

	"os/exec"
	"path/filepath"

	"errors"
	"github.com/spf13/cobra"

	"github.com/google/uuid"
	"github.com/luyomo/mockdata/embed"
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
		// fmt.Printf("opts: %#v \n", opts)
		// fmt.Printf("dbConn: %#v \n", dbConn)
		err := generateData(&opts, &dbConn)
		if err != nil {
			panic(err)
		}
	},
}

type MockDataStructure struct {
	Columns []struct {
		Idx             int      `yaml:"IDX"`
		Name            string   `yaml:"Name"`
		DataType        string   `yaml:"DataType"`
		Function        string   `yaml:"Function"`
		Values          []string `yaml:"Values"`
		Max             int      `yaml:"Max"`
		Min             int      `yaml:"Min"`
		NullProbability float64  `yaml:"NullProbability"`
		Parameters      []struct {
			Key   string      `yaml:"key"`
			Value interface{} `yaml:"value"`
		} `yaml:"Parameters"`
	} `yaml:"COLUMNS"`
	Rows int `yaml:"ROWS"`
}

type CmdOpts struct {
	Threads      int
	Rows         int
	Loop         int
	Base         int
	ConfigFile   string
	OutputFolder string
	FileName     string
	LightningVer string
	DataGenOnly  bool
}

var opts CmdOpts
var dbConn TiDBLightningConn

var MAPFunc = make(map[string]func(*rand.Rand, interface{}) (string, error))

func main() {

	// Set the arguments
	rootCmd.PersistentFlags().IntVar(&opts.Threads, "threads", runtime.NumCPU(), "Threads to generate the data")
	rootCmd.PersistentFlags().IntVar(&opts.Rows, "rows", 1, "Number of rows for each thread")
	rootCmd.PersistentFlags().IntVar(&opts.Loop, "loop", 1, "TiDB Lightning loop")
	rootCmd.PersistentFlags().IntVar(&opts.Base, "base", 0, "base number of the PK")
	rootCmd.PersistentFlags().StringVar((*string)(&opts.ConfigFile), "config", "", "Config file for data generattion")
	rootCmd.PersistentFlags().StringVar((*string)(&opts.OutputFolder), "output", "", "Output folder for data generattion")
	rootCmd.PersistentFlags().StringVar((*string)(&opts.FileName), "file-name", "", "file or table name for data generattion. For tidb lightning, please user schema_name.table_name.csv")
	rootCmd.PersistentFlags().StringVar((*string)(&opts.LightningVer), "lightning-ver", "v7.5.2", "The version of the lighting")

	rootCmd.PersistentFlags().StringVar((*string)(&dbConn.TiDBHost), "host", "", "TiDB Host name")
	rootCmd.PersistentFlags().IntVar(&dbConn.TiDBPort, "port", 4000, "TiDB Port")
	rootCmd.PersistentFlags().StringVar((*string)(&dbConn.TiDBUser), "user", "root", "TiDB User")
	rootCmd.PersistentFlags().StringVar((*string)(&dbConn.TiDBPassword), "password", "", "TiDB Password")
	rootCmd.PersistentFlags().StringVar((*string)(&dbConn.PDIP), "pd-ip", "", "pd ip address")
	rootCmd.PersistentFlags().BoolVar(&opts.DataGenOnly, "data-only", false, "Data generation only")

	rootCmd.Execute()

	return

}

func generateData(opts *CmdOpts, dbConn *TiDBLightningConn) error {
	// s1 := rand.NewSource(time.Now().UnixNano())
	// r2 := rand.New(s1)
	// var mu sync.Mutex

	MAPFunc["BROWSER"] = func(r2 *rand.Rand, intf interface{}) (string, error) {
		_value := BROWSER[r2.Intn(len(BROWSER))]
		return _value, nil
	}

	MAPFunc["IPADDR"] = func(r2 *rand.Rand, intf interface{}) (string, error) {
		_value := fmt.Sprintf("%d.%d.%d.%d", 1+r2.Intn(254), r2.Intn(255), r2.Intn(255), r2.Intn(255))
		return _value, nil
	}

	MAPFunc["UUID"] = func(r2 *rand.Rand, intf interface{}) (string, error) {
		return uuid.New().String(), nil
	}

	MAPFunc["RandomInt"] = func(r2 *rand.Rand, intf interface{}) (string, error) {
		_data := strconv.Itoa(r2.Intn(100))
		return _data, nil
	}

	MAPFunc["RandomDecimal"] = func(r2 *rand.Rand, intf interface{}) (string, error) {
		intPart := r2.Intn(100)
		decimalPart := r2.Intn(100)
		_data := fmt.Sprintf("%d.%02d", intPart, decimalPart)
		return _data, nil
	}

	MAPFunc["RandomHex"] = func(r2 *rand.Rand, intf interface{}) (string, error) {
		// Getting random character
		_data := ""
		for _idx := 0; _idx < 10; _idx++ {
			c := CHARSET[r2.Intn(len(CHARSET))]
			_data += string(c)
		}

		return hex.EncodeToString([]byte(_data)), nil
	}

	//fmt.Printf("The TiDB config info is <%#v> \n", dbConn)

	// If it's data only generation, the tidb lighting is not required.
	if opts.DataGenOnly == false {
		// Install TiDB lightning locally
		InstallTiDBLightning(opts.LightningVer)
	}

	// Read the data config file. (example: etc/data.config.yaml)
	yfile, err := ioutil.ReadFile(opts.ConfigFile)

	if err != nil {
		log.Fatal(err)
	}

	var mockDataConfig MockDataStructure

	err = yaml.Unmarshal([]byte(yfile), &mockDataConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// fmt.Printf("The mock data config is <%#v> \n", mockDataConfig)
	// return nil

	// Prepare the folder to keep the data
	csvOutputFolder := opts.OutputFolder + "/data"
	cmd := exec.Command("mkdir", "-p", csvOutputFolder)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
		return err
	}

	for _loop := 0; _loop < opts.Loop; _loop++ {
		var waitGroup sync.WaitGroup
		//errChan := make(chan error , 2)

		// fmt.Printf("The number thread is <%d> \n", _loop)
		for _idx := 0; _idx < opts.Threads; _idx++ {
			waitGroup.Add(1)
			go func(_index int) {
				// fmt.Printf("The index is <%d> \n", _index + _loop*threads)

				csvFile := fmt.Sprintf("%s/%s.%03d.csv", csvOutputFolder, opts.FileName, _index+_loop*opts.Threads)
				err := GenerateDataTo(opts.Base+_index+_loop*opts.Threads, opts.Rows, mockDataConfig, csvFile)
				if err != nil {
					panic(err)
				}

//				if err := pushCSV2S3(csvFile); err != nil {
//                    panic(err)
//				}

				defer waitGroup.Done()
			}(_idx)
		}

		waitGroup.Wait()

		// If it's data only generation, the tidb lighting is not required.
		if opts.DataGenOnly == false {
			dbConn.DataFolder = csvOutputFolder
			lightningConfigFile := fmt.Sprintf("%s/tidb-lightning.toml", opts.OutputFolder)
			parseTemplate(dbConn, lightningConfigFile)
			// fmt.Printf("The file is <%s> \n", lightningConfigFile )
			// fmt.Printf("Starting to call %s \n", fmt.Sprintf( "/tmp/temp%d.txt", rand.Intn(100)))
			cmd = exec.Command("mockdata/bin/tidb-lightning", "--config", lightningConfigFile)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				panic(err)
				return err
			}

			csvOutputFolder := opts.OutputFolder + "/data"
			cmd := exec.Command("rm", "-f", csvOutputFolder+"/*")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				panic(err)
				return err
			}
		}
	}
	return nil
}

type RandomUserID struct {
	min       int
	max       int
	MsgFormat string
}

func (p RandomUserID) GenerateData() string {
	// fmt.Printf("Starting to generate data for random User id \n")
	s1 := rand.NewSource(time.Now().UnixNano())
	r2 := rand.New(s1)
	_data := r2.Intn(p.max)
	return fmt.Sprintf(p.MsgFormat, _data)
}

func GenerateDataTo(threads, rows int, dataConfig MockDataStructure, file string) error {
	// fmt.Printf("Starting to generate data for %d rows \n", rows)
	// Prepare the file handle to output the data into
	fFile, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	writer := bufio.NewWriter(fFile)

	// Generate the random number generation instance
	s1 := rand.NewSource(time.Now().UnixNano())
	r2 := rand.New(s1)
	
	// Generate the string template instance
    var mapTemplate = make(map[string]*template.Template)
	var mapFunc = make(map[string]func(interface{}) (string, error))
	var mapFuncs = make(map[string]map[string]func(*rand.Rand, interface{}) (string, error))

	var mapGeneration = make(map[string]interface{})

	// Define all the implementation here to improve the performance. Prepare the implementation only one time.
	for _, column := range dataConfig.Columns {
		// fmt.Printf("The column is <%#v> \n", column)

		// Processes a column configured with the "template" function to enable dynamic data generation.
		// This logic pre-processes and caches templates and generators upfront for two key purposes:
		//   1. Performance: Parsing templates and initializing generators in advance avoids
		//      repeated setup costs during bulk data generation.
		//   2. Reusability: Compiled templates and configured generators are stored in global
		//      maps (mapTemplate/mapGeneration) for efficient reuse across records.
		//
		// Implementation details:
		//   - Extracts parameters (min, max, format, content) from column configuration
		//   - Compiles the 'content' as a Go text template for later execution
		//   - Initializes a RandomUserID generator with min/max bounds and format rules
		//   - Validates all inputs early to fail fast if configuration is invalid
		//
		// Note: Failures during parsing (e.g., invalid templates or parameters) will
		// immediately return/log errors to surface issues during initialization.
		if column.Function == "template" {
			// fmt.Printf("The data is staring to generate for template\n")
			var _min, _max int
			var _format, _content string
			for _, _data := range column.Parameters {
				if _data.Key == "min" {
					_min, err = strconv.Atoi(_data.Value.(string))
					if err != nil {
						return err
					}
				}
				if _data.Key == "max" {
					_max, err = strconv.Atoi(_data.Value.(string))
					if err != nil {
						return err
					}
				}
				if _data.Key == "format" {
					_format = _data.Value.(string)
				}
				if _data.Key == "content" {
					_content = _data.Value.(string)
				}
			}
			// fmt.Printf("min: %d, max: %d \n", _min, _max)
			// fmt.Printf("format: %s, content: %s \n", _format, _content)

			tmpl, err := template.New("").Parse(_content)

			if err != nil {
				log.Fatalf("Parse: %v", err)
			}
			mapTemplate[column.Name] = tmpl

			userGenerate := RandomUserID{_min, _max, _format}

			mapGeneration[column.Name] = &userGenerate
		}

		if column.Function == "WeightedList" {
			var values []interface{}
			var weights []int

			// Extract values and weights from parameters
			for _, param := range column.Parameters {
				if param.Key == "values" {
					values = param.Value.([]interface{})
				}
				if param.Key == "weights" {
					// Convert interface{} slice to []int for weights
					if weightsInterface, ok := param.Value.([]interface{}); ok {
						weights = make([]int, len(weightsInterface))
						for i, w := range weightsInterface {
							// Handle numeric interface{} conversion
							switch v := w.(type) {
							case int:
								weights[i] = v
							case float64:
								weights[i] = int(v)
							default:
								return fmt.Errorf("weight must be numeric, got %T", w)
							}
						}
					} else {
						return fmt.Errorf("weights must be array, got %T", param.Value)
					}
				}
			}

			// Validate parameters
			if len(values) != len(weights) {
				return fmt.Errorf("values and weights must have same length")
			}

			// Calculate total weight
			totalWeight := 0
			for _, w := range weights {
				totalWeight += w
			}

			mapFunc[column.Name] = func(_after interface{}) (string, error) {
				// Generate random number between 0 and total weight
				r := r2.Intn(totalWeight)

				// Find the selected value based on weights
				currentWeight := 0
				for i, w := range weights {
					currentWeight += w
					if r < currentWeight {
						return fmt.Sprintf("%v", values[i]), nil
					}
				}
				return "", nil
			}
		}

if column.Function == "WeightedIntRange" {
    var ranges []string
    var weights []int

    // Extract ranges and weights from parameters
    for _, param := range column.Parameters {
        if param.Key == "values" {
            if rangesInterface, ok := param.Value.([]interface{}); ok {
                ranges = make([]string, len(rangesInterface))
                for i, r := range rangesInterface {
                    ranges[i] = r.(string)
                }
            }
        }
        if param.Key == "weights" {
            if weightsInterface, ok := param.Value.([]interface{}); ok {
                weights = make([]int, len(weightsInterface))
                for i, w := range weightsInterface {
                    switch v := w.(type) {
                    case int:
                        weights[i] = v
                    case float64:
                        weights[i] = int(v)
                    }
                }
            }
        }
    }

    // Validate parameters
    if len(ranges) != len(weights) {
        return fmt.Errorf("ranges and weights must have same length")
    }

    // Calculate total weight
    totalWeight := 0
    for _, w := range weights {
        totalWeight += w
    }

	mapValue := make(map[int][]int)
	for _i, _v := range ranges {
		// Parse the selected range (e.g. "1-10")
		rangeParts := strings.Split(_v, "-")
		if len(rangeParts) != 2 {
			return fmt.Errorf("invalid range format: %s", _v)
		}
		
		min, err := strconv.Atoi(rangeParts[0])
		if err != nil {
			return err
		}
		
		max, err := strconv.Atoi(rangeParts[1]) 
		if err != nil {
			return err
		}
		mapValue[_i] = []int{min, max}
	}

    mapFunc[column.Name] = func(_after interface{}) (string, error) {
        // Generate random number between 0 and total weight
        r := r2.Intn(totalWeight)

        // Find the selected range based on weights
        currentWeight := 0
        for i, w := range weights {
            currentWeight += w
            if r < currentWeight {
				min, max := mapValue[i][0], mapValue[i][1]
                // Generate random number within the selected range
                return strconv.Itoa(min + r2.Intn(max-min+1)), nil
            }
        }
        return "", nil
    }
}

		if column.Function == "Template" {
			var _content string
			for _, _data := range column.Parameters {
				if _data.Key == "content" {
					_content = _data.Value.(string)
				}
			}
			// 1. Get the #IPADDR
			re := regexp.MustCompile("(?U){{\\$(.*)}}")
			ret := re.FindAllStringSubmatch(_content, -1)
			// fmt.Printf("The result is %#v \n", ret)

			var _mapFunc = make(map[string]func(*rand.Rand, interface{}) (string, error))

			for _, _match := range ret {
				_content = strings.Replace(_content, _match[0], fmt.Sprintf("{{ index .Data \"%s\"}}", _match[1]), -1)

				_mapFunc[_match[1]] = MAPFunc[_match[1]]
			}

			mapFuncs[column.Name] = _mapFunc

			tmpl, err := template.New("").Parse(_content)
			if err != nil {
				log.Fatalf("Parse: %v", err)
			}
			mapTemplate[column.Name] = tmpl
		}

		// ----- Generate the random date between min date and max
		if column.Function == "RandomDate" {
			var _min, _max time.Time
			_minDays := 0
			_maxDays := 7
			_after := ""
			ok := false
			for _, _data := range column.Parameters {
				if _data.Key == "min" {
					minVal, ok := _data.Value.(time.Time)
					if !ok {
						return fmt.Errorf("min value must be time.Time, got %T", _data.Value)
					}
					_min = minVal
				}
				if _data.Key == "max" {
					maxVal, ok := _data.Value.(time.Time)
					if !ok {
						return fmt.Errorf("max value must be time.Time, got %T", _data.Value)
					}
					_max = maxVal
				}
				if _data.Key == "minDays" {
					_minDays, ok = _data.Value.(int)
					if !ok {
						return fmt.Errorf("min value must be int, got %T", _data.Value)
					}
				}
				if _data.Key == "maxDays" {
					_maxDays, ok = _data.Value.(int)
					if !ok {
						return fmt.Errorf("min value must be int, got %T", _data.Value)
					}
				}
				if _data.Key == "after" {
					_after, ok = _data.Value.(string)
					if !ok {
                        return fmt.Errorf("after value must be string, got %T", _data.Value)
					}
				}
			}
			_days := int(_max.Sub(_min).Hours() / 24)
			// fmt.Printf("max: %d, min: %d, _days: %d \n", _max, _min, _days)
			mapFunc[column.Name] = func(_dataMap interface{}) (string, error) {
				if _after == "" {
					return _min.AddDate(0, 0, r2.Intn(_days)).Format("2006-01-02"), nil
				}
				dataMap, ok := _dataMap.(map[string]string)
				if !ok {
					return "", fmt.Errorf("dataMap must be map[string]string, got %T", _dataMap)
				}

				afterStr := strings.Trim(dataMap[_after], "\"")

				afterDate, err := time.Parse("2006-01-02", afterStr)
				if err != nil {
					return "", fmt.Errorf("failed to parse after date: %v", err)
				}
				daysDiff := r2.Intn(_maxDays-_minDays+1) + _minDays
				return afterDate.AddDate(0, 0, daysDiff).Format("2006-01-02"), nil
			}
		}
		// ----------
	}

	insUUID := uuid.New()
	// fmt.Printf("The data is staring to generate \n")
	for idx := 1; idx <= rows; idx++ {
		var arrData []string
		mapData := make(map[string]string)
		// originTime := time.Now()
		for _, column := range dataConfig.Columns {
			// passTime := time.Now().Sub(originTime)
			// originTime = time.Now()
			// fmt.Printf("[%s] The data is staring to generate for %d \n", passTime, colIdx)
			var _data string

			if column.Function == "sequence" {
				_data = strconv.Itoa(threads*rows + idx)
			}

			if column.Function == "uuid" {
				_data = insUUID.String()
			}

			if column.Function == "list" {
				_data = column.Values[r2.Intn(len(column.Values))]
			}

			if column.Function == "random" {
				_data = strconv.Itoa(r2.Intn(column.Max))
			}

			if column.Function == "RandomDecimal" {
				_data, err = MAPFunc["RandomDecimal"](r2, nil)
				if err != nil {
					panic(err)
				}
			}

			if column.Function == "RandomHex" {
				_data, err = MAPFunc["RandomHex"](r2, nil)
				if err != nil {
					panic(err)
				}
			}

			if column.Function == "WeightedList" {
				_data, err = mapFunc[column.Name](nil)
				if err != nil {
					panic(err)
				}
			}

			if column.Function == "WeightedIntRange" {
				_data, err = mapFunc[column.Name](nil)
				if err != nil {
					panic(err)
				}
			}

			if column.Function == "RandomDate" {
				_data, err = mapFunc[column.Name](mapData)
				if err != nil {
					panic(err)
                }
			}

			// Make the function common
			if column.Function == "RandomString" {
				var _min, _max int
				for _, _p := range column.Parameters {
					if _p.Key == "min" {
						minVal, ok := _p.Value.(int)
						if !ok {
							return fmt.Errorf("min value must be integer, got %T", _p.Value)
						}
						_min = minVal
					}
					if _p.Key == "max" {
						maxVal, ok := _p.Value.(int)
						if !ok {
							return fmt.Errorf("max value must be integer, got %T", _p.Value)
						}
						_max = maxVal
					}
				}

				// Getting random character
				_data = ""
				for _idx := 0; _idx < _min+r2.Intn(_max-_min+1); _idx++ {
					c := CHARSET[r2.Intn(len(CHARSET))]
					_data += string(c)
				}
			}

			if column.Function == "UniqueRandomString" {
				var _min, _max int
				for _, _p := range column.Parameters {
					if _p.Key == "min" {
						_min, err = strconv.Atoi(_p.Value.(string))
						if err != nil {
							return err
						}
					}
					if _p.Key == "max" {
						_max, err = strconv.Atoi(_p.Value.(string))
						if err != nil {
							return err
						}
					}
				}

				// Getting random character
				_data = ""
				for _idx := 0; _idx < _min+r2.Intn(_max-_min+1); _idx++ {
					c := CHARSET[r2.Intn(len(CHARSET))]
					_data += string(c)
				}
				_data += strconv.Itoa(threads*rows + idx)
			}

			if column.Function == "template" {
				fmt.Printf("Starting to generate the data [%s] vs [%#v] -> [%#v] \n", column.Name, mapGeneration[column.Name], mapTemplate[column.Name])
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
					_value, err := _func(r2, nil)
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

			if !strings.Contains("int,decimal,tinyint,bigint,smallint,mediumint,float,double", column.DataType) {
				_data = "\"" + _data + "\""
			}

			mapData[column.Name] = _data
			arrData = append(arrData, _data)
		}
		// fmt.Printf("The data is <%#v> and <%s> \n", arrData, strings.Join(arrData, ","))
		// fmt.Printf("The data is <%#v> \n", mapData)
		_, err := writer.WriteString(strings.Join(arrData, ",") + "\n")
		if err != nil {
			return err
		}
		// passTime := time.Now().Sub(originTime)
		// originTime = time.Now()
		// fmt.Printf("[%s] WriteString  \n", passTime)
		if idx%10000 == 0 {
			writer.Flush()
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

			r, err = os.Open(fmt.Sprintf("tidb-community-toolkit-%s-linux-%s/tidb-lightning-%s-linux-%s.tar.gz", lightningVer, runtime.GOARCH, lightningVer, runtime.GOARCH))
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

func parseTemplate(dbConn *TiDBLightningConn, configFile string) {
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

func pushCSV2S3(fileName string) error {
    // Create AWS session
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-east-1"), // Replace with your desired region
    })
    if err != nil {
        return fmt.Errorf("failed to create AWS session: %v", err)
    }

    // Create S3 service client
    s3Client := s3.New(sess)

    // Open the file
    file, err := os.Open(fileName)
    if err != nil {
        return fmt.Errorf("failed to open file %s: %v", fileName, err)
    }
    defer file.Close()

    // Get file size and create buffer
    fileInfo, _ := file.Stat()
    size := fileInfo.Size()
    buffer := make([]byte, size)
    
    // Read file content
    file.Read(buffer)

    // Create input parameters for S3 upload
    input := &s3.PutObjectInput{
        Bucket:        aws.String("jay-data"), // Replace with your bucket name
        Key:           aws.String(fmt.Sprintf("mockdata/%s", filepath.Base(fileName))),
        Body:          bytes.NewReader(buffer),
        ContentType:   aws.String("text/csv"),
        ContentLength: aws.Int64(size),
    }

    // Upload to S3
    _, err = s3Client.PutObject(input)
    if err != nil {
        return fmt.Errorf("failed to upload file to S3: %v", err)
    }

    // Delete the local file after successful S3 upload
    if err := os.Remove(fileName); err != nil {
        return fmt.Errorf("failed to delete local file %s: %v", fileName, err)
    }

    return nil
}
