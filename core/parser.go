package core
import  "strings"
import "strconv"
import "fmt"


func Parser (name string) (uint,uint,uint,error) {

    stringSlice := strings.Split(name, "/")

    var dimqual string = stringSlice[0]

    dimQualityArray := strings.Split(dimqual, "_")
    fmt.Println(dimQualityArray)
    arrayOfInt := make([]uint, 3)

    var err error = nil
    var tmpr int

    arrayOfInt[2] = 100
    arrayOfInt[0] = 0
    arrayOfInt[1] = 0
    for i := 0; i <len(dimQualityArray); i++ {
        tmpr,err = strconv.Atoi(dimQualityArray[i])
        arrayOfInt[i] = uint(tmpr)
        fmt.Println(arrayOfInt[i])
        if err != nil {
            return 0, 0, 0, err
        }
    }
    if err != nil {
        fmt.Println(err)
    }
    return arrayOfInt[0], arrayOfInt[1], arrayOfInt[2], nil
}

