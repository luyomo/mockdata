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
	"strconv"
	"strings"
	"sync"
	"time"

	//    "errors"

	//	"github.com/luyomo/mockdata/pkg/tui"
	"github.com/spf13/cobra"
)

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
		Max        int    `yaml:"Max"`
		Min        int    `yaml:"Min"`
		Parameters []struct {
			Key   string `yaml:"key"`
			Value string `yaml:"value"`
		} `yaml:"Parameters"`
	} `yaml:"COLUMNS"`
	Rows int `yaml:"ROWS"`
}

func main() {

	// fmt.Printf("This is the test \n")
	var threads, rows int
	var configFile string
	var outputFile string

	rootCmd.PersistentFlags().IntVar(&threads, "threads", 1, "Threads to generate the data")
	rootCmd.PersistentFlags().IntVar(&rows, "rows", 1, "Number of rows for each thread")
	rootCmd.PersistentFlags().StringVar((*string)(&configFile), "config", "", "Config file for data generattion")
	rootCmd.PersistentFlags().StringVar((*string)(&outputFile), "output", "", "Output file for data generattion")
	rootCmd.Execute()
	// fmt.Printf("The threads are %d \n", threads)
	// fmt.Printf("The config file are %s \n", configFile)
	// fmt.Printf("The config file are %s \n", outputFile)

	yfile, err := ioutil.ReadFile(configFile)

	if err != nil {
		log.Fatal(err)
	}

	var mockDataConfig MockDataStructure

	err = yaml.Unmarshal([]byte(yfile), &mockDataConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	//fmt.Printf("--- t:\n%v\n\n", mockDataConfig)
	//fmt.Printf("The number of rows is <%d> \n", mockDataConfig.Rows)
	// for _, _columnCfg := range mockDataConfig.Columns {
	//     fmt.Printf("Column name is : %s and data type: %s, Function: %s, Max: %d, Min: %d \n", _columnCfg.Name, _columnCfg.DataType, _columnCfg.Function, _columnCfg.Min, _columnCfg.Max)
	// }

	var waitGroup sync.WaitGroup
	//errChan := make(chan error , 2)

	for _idx := 0; _idx < threads; _idx++ {
		waitGroup.Add(1)
		go func() {
			// fmt.Printf("Starting to call %s \n", fmt.Sprintf( "/tmp/temp%d.txt", rand.Intn(100)))
			defer waitGroup.Done()
			GenerateDataTo(rows, mockDataConfig, outputFile)
		}()
	}
	waitGroup.Wait()
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

func GenerateDataTo(rows int, dataConfig MockDataStructure, file string) error {
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

	var mapGeneration = make(map[string]interface{})

	for _, column := range dataConfig.Columns {
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
	}

	for idx := 1; idx <= rows; idx++ {
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
				userGenerate := mapGeneration[column.Name]

				tmpl := mapTemplate[column.Name]
				var test bytes.Buffer
				tmpl.Execute(&test, userGenerate)
				_data = test.String()
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
