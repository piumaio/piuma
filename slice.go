package main

import  "strings"
import "fmt"
//import "encoding/json"

/*questa funzione controlla se immagine appartiene a tvl o no */
func isofsite(url string) int {
	if strings.Contains(url, "tvl.lotrek.it") {
     return 1
	} else {
     return 0
	}	
}
/* questa funzione spezzetta ilo parametro in url passato */ 
func slice (name string) ([]string) {

   /*dichiaro un array di ritorno con tre parametri*/	
   arr := make([]string, 3)
   /*ci separiamo la stringa passata per il carattere */
   stringSlice := strings.Split(name, "/") 
  /*ora in posizoione 0 della stringslide, avremo dimensione e qualita*/
   var dimqual = stringSlice[0]
   /* se togliamo dalla stringa originale con un replace il dimqual, ecco che abbiamo la urla*/    
   var url = strings.Replace(name, dimqual+"/", "", -1)
  
  /*prima di tutto, l'immagine appartiene a tvl ?  */
   var is=isofsite(url)
   if is != 0 {

		  /* se dimqual ha il parametro in più , tipo "_" si deve comportare in maniera diversa rispetto a non averlo */
		   if strings.Contains(dimqual, "_") {
		   	
			   	stringSlice2 := strings.Split(dimqual, "_")
			   	var  dim = stringSlice2[0]
			   	var qual = stringSlice2[1]
			    arr[0] = dim 
			    arr[1] = qual
		   
		   } else {
		    
		    arr[0] = dimqual
		    arr[1] = ""
		   
		   }
		   arr[2] = url
  
  } else {
  	arr[0] =""
  	arr[1] =""
  	arr[2] =""
    
  }
 
  return arr
}

func main() {
   var a = slice("500x500_quality/http://tvl.lotrek.it/img1.jpg");
   fmt.Printf(a[0])
   fmt.Printf(a[1])
   fmt.Printf(a[2])
}

