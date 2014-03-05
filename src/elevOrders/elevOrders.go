package elevOrders

import(
   "fmt"
   "elevTypes"
)

type Orders_s struct{
   table          [][]bool
   elevTypes.Orders_ExtComs_s
}

func Init() Orders_s{
   fmt.Println("elevOrders.init()...")
   
   var table [N_FLOORS][N_DIR]bool
	var extcoms = elevTypes.Orders_ExtComs_s{}

	NewOrdersChan    	make(chan elevTypes.Order_t) 
   OrderUpdatedChan	make(chan elevTypes.Order_t)
	OrderExecdChan  	make(chan elevTypes.Order_t)
	StopRequestChan  	make(chan elevTypes.Order_t)		
   EmgTriggerdChan  	make(chan bool)
 
   
   return Orders_s{}
}
