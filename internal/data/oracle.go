package data

import (
    // "fmt"
    "strconv"
)

func GenerateOracleData(cols *[]map[string]string, refData *map[string][]interface{},  numRows int) (*[][]interface{}, error) {
    // fmt.Printf("reference data: <%#v> \n", (*refData)["ISSUER_CD02"])

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

            // fmt.Printf("The column has reference data: <%s> \n",colDef["COLUMN_NAME"] )
            if (*refData)[colDef["COLUMN_NAME"]] != nil {
                // fmt.Printf("Data list: <%#v> \n", (*refData)[colDef["COLUMN_NAME"]] )
                _row[idx] = pickRandItem((*refData)[colDef["COLUMN_NAME"]])
                continue
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

