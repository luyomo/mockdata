package data

import (
    "time"
    "math"
    "math/rand"
)

const (
    CHARSET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func generateNumber(length, precision, scale int) int{
    rand.Seed(time.Now().UnixNano())
    numLen := precision - scale

    return rand.Intn(int(math.Pow(2, float64(numLen))))
}

func generateString(length int, includKanji, isChar, isNullable bool) string {
    rand.Seed(time.Now().UnixNano())
    dataLength := length
    if isChar == false{
        if isNullable == true {
            dataLength = rand.Intn(length)
        } else {
            dataLength = rand.Intn(length - 1) + 1
        }
    }

	_data := ""
    for _idx := 0; _idx <  dataLength; _idx++ {
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
