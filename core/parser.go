package core
import  "strings"
import "strconv"
import "fmt"
//import "encoding/json"

/* questa funzione spezzetta il parametro in url passato */
func Parser (name string) (uint,uint,uint,error) { //return integer, integer , integer, error

    /*ci separiamo la stringa passata per il carattere */
    stringSlice := strings.Split(name, "/")
    /*ora in posizoione 0 della stringslide, avremo dimensione e qualita*/
    var dimqual string = stringSlice[0]
    /* se togliamo dalla stringa originale con un replace il dimqual, ecco che abbiamo la urla*/
    //var url = strings.Replace(name, dimqual+"/", "", -1)

    dimQualityArray := strings.Split(dimqual, "_")
    fmt.Println(dimQualityArray)
    arrayOfInt := make([]uint, 3)
    var err error=nil

    var tmpr int
    arrayOfInt[2] = 100
    arrayOfInt[0] = 0
    arrayOfInt[1] = 0
    for i := 0; i <len(dimQualityArray); i++ {
        tmpr,err=strconv.Atoi(dimQualityArray[i])
        arrayOfInt[i]=uint(tmpr)
        fmt.Println(arrayOfInt[i])
        if err != nil { fmt.Println(err) }
    }
    if err != nil {
        fmt.Println(err)
    }
    return arrayOfInt[0],arrayOfInt[1],arrayOfInt[2],nil
}

