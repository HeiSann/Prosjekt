package elevCtrl

import (
	"fmt"
	"elevTypes"
)

type intComs_s struct{
   eventChan		chan Event_t
	startTimerChan chan bool
	timeoutChan    chan bool
	readyChan      chan bool
}

type Fsm_s struct{
   fsm_table	   [][]func()
	state 		   State_t
	lastDir 	      elevTypes.Direction_t
	lastFloor      int 
	intComs			intComs_s
	ExtComs        elevTypes.Fsm_ExtComs_s
}

func Init(driver elevTypes.Drivers_ExtComs_s, orders elevTypes.Orders_ExtComs_s)Fsm_s{

   fmt.Println("elevCtrl.init()...")
	
	fmt.Println("OK")
   
   return Fsm_s{}
}
