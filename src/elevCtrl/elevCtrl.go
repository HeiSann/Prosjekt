package elevCtrl

import (
	"fmt"
	"elevTypes"
)

type Coms_s struct{
   buttonChan        chan int elevTypes.Button
   floorChan         chan int
   stopButtonChan    chan bool
   obsChan           chan bool
   
   motorChan         chan elevTypes.Direction_t
   doorOpenChan      chan bool
   setLightChan      chan elevTypes.Order_t
   floorIndChan      chan int 
   
   orderExdChan     chan elevTypes.Order_t  //fsm -> orders
   newOrdersChan    chan elevTypes.Order_t  //orders -> orders
   emgTriggerChan   chan bool               //orders -> fsm
}

type Fsm_s struct{
   fsm_table	   [][]func()
	state 		   State_t
	lastDir 	      elevTypes.Direction_t
	lastFloor      int 
	startTimerChan chan bool
	timeoutChan    chan bool
	readyChan      chan bool
	Coms           Coms_s
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
