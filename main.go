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
//	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"archive/tar"
	"compress/gzip"
	"io"

	"os/exec"

	"errors"

	//	"github.com/luyomo/mockdata/pkg/tui"
	"github.com/spf13/cobra"

        "github.com/luyomo/mockdata/embed"
)


type TiDBLightningConn struct {
    TiDBHost string
    TiDBPort int
    TiDBUser string
    TiDBPassword string
    PDIP string
    DataFolder string
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
	InstallTiDBLightning()

	// fmt.Printf("This is the test \n")
	var threads, rows int
	var configFile string
	var outputFolder string
        var fileName string

        var dbConn TiDBLightningConn

	rootCmd.PersistentFlags().IntVar(&threads, "threads", runtime.NumCPU(), "Threads to generate the data")
	rootCmd.PersistentFlags().IntVar(&rows, "rows", 1, "Number of rows for each thread")
	rootCmd.PersistentFlags().StringVar((*string)(&configFile), "config", "", "Config file for data generattion")
	rootCmd.PersistentFlags().StringVar((*string)(&outputFolder), "output", "", "Output folder for data generattion")
	rootCmd.PersistentFlags().StringVar((*string)(&fileName), "file-name", "", "file or table name for data generattion. For tidb lightning, please user schema_name.table_name.csv")

	rootCmd.PersistentFlags().StringVar((*string)(&dbConn.TiDBHost), "host", "", "TiDB Host name")
	rootCmd.PersistentFlags().IntVar(&dbConn.TiDBPort, "port", 4000, "TiDB Port")
	rootCmd.PersistentFlags().StringVar((*string)(&dbConn.TiDBUser), "user", "", "TiDB User")
	rootCmd.PersistentFlags().StringVar((*string)(&dbConn.TiDBPassword), "password", "", "TiDB Password")
	rootCmd.PersistentFlags().StringVar((*string)(&dbConn.PDIP), "pd-ip", "", "pd ip address")

	rootCmd.Execute()
	// fmt.Printf("The threads are %d \n", threads)
	// fmt.Printf("The config file are %s \n", configFile)
	// fmt.Printf("The config file are %s \n", outputFile)
        fmt.Printf("The TiDB config info is <%#v> \n", dbConn)


	yfile, err := ioutil.ReadFile(configFile)

	if err != nil {
		log.Fatal(err)
	}

	var mockDataConfig MockDataStructure

	err = yaml.Unmarshal([]byte(yfile), &mockDataConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
        csvOutputFolder := outputFolder + "/data"
	cmd := exec.Command("mkdir", "-p", csvOutputFolder )
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
            panic(err)
            return
	}

        for _loop := 0; _loop < 100; _loop++  {
	var waitGroup sync.WaitGroup
	//errChan := make(chan error , 2)

        fmt.Printf("The number thread is <%d> \n", threads)
	for _idx := 0; _idx < threads; _idx++ {
		waitGroup.Add(1)
		go func(_index int) {
			fmt.Printf("The index is <%d> \n", _index + _loop*threads)
			//outputFile := fmt.Sprintf("%s%03d.csv", retValue[1], _index)

                        // Make the folder to populate the data


                        csvFile := fmt.Sprintf("%s/%s.%03d.csv", csvOutputFolder, fileName, _index)
			GenerateDataTo(_index + _loop*threads, rows, mockDataConfig, csvFile)

			defer waitGroup.Done()
		}(_idx)
	}

	waitGroup.Wait()

        dbConn.DataFolder = csvOutputFolder
        lightningConfigFile := fmt.Sprintf("%s/tidb-lightning.toml", outputFolder)
        parseTemplate(dbConn, lightningConfigFile )
	fmt.Printf("The file is <%s> \n", lightningConfigFile )
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
				_data = strconv.Itoa(threads*rows + idx)
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

func InstallTiDBLightning() error {
	if runtime.GOOS == "windows" {
		fmt.Println("Can't Execute this on a windows machine")
	} else {
                fmt.Printf("Starting to check the data \n")
		if _, err := os.Stat("mockdata/bin/tidb-lightning"); errors.Is(err, os.ErrNotExist) {
			// file does not exist
                        fmt.Printf("Started to check mockdata \n")

			fmt.Printf("The os is <%s> \n", runtime.GOOS)
			fmt.Printf("The os is <%s> \n", runtime.GOARCH)
			binFile := fmt.Sprintf("tidb-community-toolkit-%s-linux-%s.tar.gz", "v6.2.0", runtime.GOARCH)
			fullBinFile := fmt.Sprintf("tidb-community-toolkit-%s-linux-%s/tidb-lightning-%s-linux-%s.tar.gz", "v6.2.0", runtime.GOARCH, "v6.2.0", runtime.GOARCH)

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

			r, err = os.Open(fmt.Sprintf("tidb-community-toolkit-%s-linux-%s/tidb-lightning-%s-linux-%s.tar.gz", "v6.2.0", runtime.GOARCH, "v6.2.0", runtime.GOARCH  ))
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

			cmd = exec.Command("rm", "-rf", fmt.Sprintf("tidb-community-toolkit-%s-linux-%s", "v6.2.0", runtime.GOARCH))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
                fmt.Print("Completed the file check \n")
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
    fmt.Printf("This is the testing data \n")
    data, err := embed.ReadTemplate("templates/tidb-lightning.toml.tpl")
    if err != nil {
        panic(err)
    }
 //   fmt.Printf("This is the template %s \n",data)

    tmpl, err := template.New("").Parse(string(data))
    if err != nil {
        log.Fatalf("Parse: %v", err)
    }

    var ret bytes.Buffer
    err = tmpl.Execute(&ret, dbConn)
//    fmt.Printf("The data is %s \n", ret.String())

//    fo, err := os.Create("/tmp/tidb-lightning.toml")
    fo, err := os.Create(configFile)
    if err != nil {
        panic(err)
    }

    // close fo on exit and check for its returned error
    defer func() {
        if err := fo.Close(); err != nil {
            panic(err)
        }
    }()

    if _, err := fo.Write(ret.Bytes()); err != nil {
        panic(err)
    }

}
