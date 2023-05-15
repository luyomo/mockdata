package data

import (
    //"fmt"
    "strconv"
    "github.com/luyomo/mockdata/internal/database/oracle"
)

func GenerateOracleData(cols *[]map[string]string, refData *map[string]oracle.TableInfo , numRows int) (*[][]interface{}, error) {
    data := make([][]interface{}, numRows)
    for index := 0; index < numRows; index++{
        _row := make([]interface{}, len(*cols))
        for idx, colDef := range *cols {
            // fmt.Printf("Column definition: %d,  %#v \n", idx, colDef)

            dataLen, err := strconv.Atoi(colDef["DATA_LENGTH"])
            if err != nil {
                return nil, err
            }

            dataPrecision := 0
            if colDef["DATA_PRECISION"] != "NULLABLE" {
                dataPrecision, err = strconv.Atoi(colDef["DATA_PRECISION"])
                if err != nil {
                    return nil, err
                }
            }

            dataScale := 0
            if colDef["DATA_PRECISION"] != "NULLABLE" {
                dataScale, err = strconv.Atoi(colDef["DATA_SCALE"])
                if err != nil {
                    return nil, err
                }
            }

            switch colDef["DATA_TYPE"] {
            case "VARCHAR2":
                if colDef["NULL"] == "Y" {
                    _row[idx] = generateString(dataLen, false, false, true)     // data length, includKanji, isChar, isNullable
                }else {
                    _row[idx] = generateString(dataLen, false, false, false)
                }
            case "CHAR":
                if colDef["NULL"] == "Y" {
                    _row[idx] = generateString(dataLen, false, true, true)
                } else {
                    _row[idx] = generateString(dataLen, false, true, false)
                }
            case "TIMESTAMP(6)":
                _row[idx] = generateTimestamp(dataScale, "yyyy-mm-dd HH:MI:SS.ssssss")
            case "NUMBER":
                _row[idx] = generateNumber(dataLen,  dataPrecision, dataScale)
            }
        }
        data[index] = _row
    }

    return &data, nil
}

