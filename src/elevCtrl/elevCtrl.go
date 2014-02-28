package elevCtrl

import (
	"fmt"
	"elevTypes"
)

type Coms_s struct{
   buttonChan        chan int elevTypes.Button
   sensorChan        chan int int
   stopButtonChan    chan bool
   obsChan           chan bool
   
   motorChan         chan elevTypes.Direction_t
   doorOpenChan      chan bool
   lightChan         chan elevTypes.Order_t
   floorIndChan      chan int 
   
   OrderExecuted     chan elevTypes.Order_t  //fsm -> orders
   NewOrders         chan elevTypes.Order_t  //orders -> orders
   emgTrigger        chan bool               //orders -> fsm
}

type Fsm_s struct{
   fsm_table	[][]func()
	state 		State_t
	lastDir 	   elevTypes.Direction_t
	lastFloor   int 
	Coms        Coms_s
}


func Init(
   buttonChan     chan elevTypes.Button,
   sensorChan     chan int,
   motorChan      chan elevTypes.Direction_t,
   stopButtonChan chan bool,
   obsChan        chan bool,
   OrderExecuted  chan elevTypes.Order_t)Fsm_s{


   fmt.Println("elevCtrl.init()...")
   
   //use function for this
   var table [][]func()
	
	fmt.Println("OK")
   
   return Fsm_s{table, IDLE, elevTypes.NONE, 1}
}
