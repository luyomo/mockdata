package data

import (
    "time"
    "math"
    "math/rand"
    //"math/big"
    // crand "crypto/rand"
    // "fmt"
)

const (
    CHARSET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func pickRandItem(data []interface{}) interface{} {
    rand.Seed(time.Now().UnixNano())
    numOfEle := len(data)

    return data[rand.Intn(numOfEle)]
    // return data[0]
} 

func generateNumber(length, precision, scale int) int{
    rand.Seed(time.Now().UnixNano())
    numLen :=  length - scale
    if precision > 0 {
        numLen = precision - scale
    }

    // retRand, err := crand.Int(crand.Reader, big.NewInt(int64(math.Pow(2, float64(numLen)))))
    // if err != nil {
    //     panic(err)
    // }
    // return retRand
    // fmt.Printf("Random data: %#v and %#v \n", rand.Intn( int(math.Pow(2, 22) ) ), float64(numLen) )
    return rand.Intn(int(math.Pow(2, float64(numLen))))
}

func generateString(length int, includKanji, isChar, isNullable bool) string {
    rand.Seed(time.Now().UnixNano())
    dataLength := length
    if isChar == false{
        if isNullable == true {
            rand.Seed(time.Now().UnixNano())
            dataLength = rand.Intn(length)
        } else {
            if length == 1 {
                dataLength = 1
            } else {
                rand.Seed(time.Now().UnixNano())
                dataLength = rand.Intn(length - 1) + 1
            }
        }
    }

	_data := ""
    for _idx := 0; _idx <  dataLength; _idx++ {
        rand.Seed(time.Now().UnixNano())
        c := CHARSET[rand.Intn(len(CHARSET))]
        _data += string(c)
    }
    return _data
}

func generateTimestamp(scale int, format string) string {
    randomTime := rand.Int63n(time.Now().Unix() - 94608000) + 94608000

    randomNow := time.Unix(randomTime, 0)
    return randomNow.Format("02-JAN-06 03:04:05.000000 PM") // String()
}
