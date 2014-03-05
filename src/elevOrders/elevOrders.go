package elevOrders

import(
   "fmt"
   "elevTypes"
)

type Orders_s struct{
   table          [][]bool
   ToFsm_new      chan elevTypes.Order_t  
   ToComs_new     chan elevTypes.Order_t     
   ToComs_del     chan elevTypes.Order_t
}

func Init() Orders_s{
   fmt.Println("elevOrders.init()...")
   
   var table [][]bool
   
   toFsm_new   := make(chan elevTypes.Order_t)
   toComs_new  := make(chan elevTypes.Order_t)
   toFsm_del   := make(chan elevTypes.Order_t)
   
   return Orders_s{table, toFsm_new, toComs_new, toFsm_del}
}

func compare(elevTypes.Order_t) bool{
	
}
