package elevCtrl

import (
	"fmt"
	"elevTypes"
)

/*	Declaration of states, events and fsm_structs */
type State_t int
const(
	IDLE State_t = iota
	DOORS_OPEN
	MOVING_DOWN
	MOVING_UP
	EMG_STOP
	OBSTRUCTED
	OBST_EMG
)

type Event_t int
const(
	START_DOWN Event_t = iota
	START_UP
	EXEC_ORDER 
	TIMEOUT
	READY
	EMG
	OBSTRUCTION
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

/* FSM_init */
func Init(driver elevTypes.Drivers_ExtComs_s, orders elevTypes.Orders_ExtComs_s)Fsm_s{
   fmt.Println("elevCtrl.init()...")

	/* Make internal channels	*/
	

	/* Make external channel*/
	Ord_executed:= make(chan Order_t)  //fsm -> orders	
	
	ExternalComs:= elevTypes.Fsm_ExtComs_s{
		driver.ButtonChan
		driver.SensorChan
		driver.StopButtonChan
		driver.
		driver.MotorChan
		driver.	
	}

	/* Make FSM	*/
	self := Fsm_s{}
	self.int_fsm_table()
	self.intComs = internalComs
	self.ExtComs = externalComs

	self.go_to_defined_state()
		
	fmt.Println("OK")
   
   return fsm
}

/* 	FSM Actions */
func (self *Fsm_s)action_start_down(){
   self.ExtComs.MotorChan <- elevTypes.DOWN
	self.state = MOVING_DOWN
	self.lastDir = elevTypes.DOWN
	fmt.Println("fsm: MOVING_DOWN\n")
}

func (self *Fsm_s)action_start_up(){
	self.ExtComs.MotorChan <- elevTypes.UP
	self.state = MOVING_UP
	self.lastDir = elevTypes.UP
	fmt.Println("fsm: MOVING_UP\n")
}

func (self *Fsm_s)action_exec(){
	self.ExtComs.DoorOpenChan <- true
	self.ExtComs.SetLightChan <- elevTypes.Light_t{self.lastFloor, self.lastDir, false}
	self.intComs.startTimerChan <- true
	self.ExtComs.OrderExdChan <- elevTypes.Order_t{self.lastFloor, self.lastDir, false}
	self.state = DOORS_OPEN 
	fmt.Println("fsm: DOORS_OPEN\n")
}

func (self *Fsm_s)action_halt_n_exec(){
	self.ExtComs.MotorChan <- elevTypes.NONE
	self.ExtComs.DoorOpenChan <- true
	self.ExtComs.SetLightChan <- elevTypes.Light_t{self.lastFloor, self.lastDir, false}
	self.intComs.startTimerChan <- true	
	self.ExtComs.OrderExdChan <- elevTypes.Order_t{self.lastFloor, self.lastDir, false}
	self.state = DOORS_OPEN 
	fmt.Println("fsm: DOORS_OPEN\n")
}

func (self *Fsm_s)action_done(){
	self.ExtComs.DoorOpenChan <- false
	self.lastDir = self.get_nearest_order()
	self.state = IDLE
	fmt.Println("fsm: IDLE\n")
	self.intComs.eventChan <- READY
}   
	
func (self *Fsm_s)action_next(){
    self.handle_next_order()
}

func (self *Fsm_s)action_stop(){
	self.ExtComs.MotorChan <- elevTypes.NONE
	self.state = EMG_STOP
}

func (self *Fsm_s)action_pause(){
	self.state = OBSTRUCTED
}

func action_dummy(){
	fmt.Println("fsm: dummy!\n")
}

/* Finite State Machine initializations */
func (elev *Fsm_s)init_fsm_table(){
	elev.fsm_table = [][]func(){
/*STATES:	  \	EVENTS:	//START_DOWN			   //START_UP              //EXEC_ORDER			   //TIMEOUT			   //READY
/*IDLE       */  []func(){elev.action_start_down,  elev.action_start_up,   elev.action_exec,       action_dummy,       elev.action_next},
/*DOORS_OPEN */  []func(){elev.action_start_down,  elev.action_start_up,   elev.action_exec,       elev.action_done,   action_dummy},  
/*MOVING_UP  */  []func(){action_dummy,            action_dummy,           elev.action_halt_n_exec,action_dummy,       action_dummy},
/*MOVING_DOWN*/  []func(){action_dummy,            action_dummy,           elev.action_halt_n_exec,action_dummy,       action_dummy},
/*EMG_STOP   */  []func(){action_dummy,            action_dummy,           action_dummy,           action_dummy,       action_dummy},  
/*OBST       */  []func(){action_dummy,            action_dummy,           action_dummy,           action_dummy,       action_dummy}, 
	}
}

func(self *Fsm_s)init_intComs(){
	self.eventChan 		= make(chan Event_t)
	self.startTimerChan 	= make(chan bool)
	self.timeoutChan		= make(chan bool)
	self.readyChan  		= make(chan bool)
}

func(self *Fsm_s)init_ExtComs(driver elevTypes.Drivers_ExtComs_s, orders elevTypes.Orders_ExtComs_s){
	self.OrderExdChan = make(chan Order_t)  //fsm -> orders	

	self.ButtonChan 		= driver.ButtonChan
	self.FloorChan 		= driver.SensorChan
	self.StopButtonChan	= driver.StopButtonChan
	self.ObsChan			= driver.ObsChan	
	self.MotorChan			= driver.MotorChan
	self.SetLightChan
	self.FloorIndChan
	}
}

/* FSM help functions */
func (self *Fsm_s)fsm_update(){
	var event Event_t
	for{
		event =<- self.intComs.eventChan
		self.fsm_table[self.state][event]()
	}
}

func (self *Fsm_s)should_stop(floor int) bool{     
	//TODO: communicate with orders and check if current floor got pending orders in correct direction    
	stop := true	
	return stop
}

func (self *Fsm_s)get_nearest_order() elevTypes.Direction_t{
	//TODO: communicate with orders and get next direction
	return elevTypes.UP
}

func (elev *Fsm_s)handle_next_order(){
    if elev.state == IDLE{
	    switch(elev.get_nearest_order()){
	    case elevTypes.UP:
	       elev.intComs.eventChan <- START_UP
	    case elevTypes.DOWN:
			 elev.intComs.eventChan <- START_DOWN
	    case elevTypes.NONE:
	       // no new orders: do nothing
	    default:
	       fmt.Println("fsm: ERROR, undefined elev.lastDir in execute_next_order")
	    }
	}else{
	    fmt.Println("fsm: new order registered, will be executed when I'm ready")
   }
}

func (elev *Fsm_s)handle_new_order(){
    if elev.state == IDLE{
	    switch(elev.get_nearest_order()){
	    case elevTypes.UP:
	       elev.intComs.eventChan <- START_UP
	    case elevTypes.DOWN:
			 elev.intComs.eventChan <- START_DOWN
	    case elevTypes.NONE:
	       elev.intComs.eventChan <- EXEC_ORDER
	    default:
	       fmt.Println("fsm: ERROR, undefined elev.lastDir in execute_next_order")
	    }
	}else{
	    fmt.Println("fsm: new order registered, will be executed when I'm ready")
   }
}

func (elev *Fsm_s)fsm_generate_events(){
	for{
	   select{
	   case <- elev.ExtComs.StopButtonChan:
			elev.intComs.eventChan <- EMG
	   case <- elev.ExtComs.ObsChan:
			elev.intComs.eventChan <- OBSTRUCTION
	   case floor:=<- elev.ExtComs.FloorChan:
		   if floor != -1{
			   elev.lastFloor = floor
			   //set floor_light
			   if elev.should_stop(floor){
					elev.intComs.eventChan <- EXEC_ORDER
			   }
		   }
	   case <- elev.ExtComs.NewOrdersChan:
		   elev.handle_new_order()
      case <-elev.intComs.timeoutChan:
			 elev.intComs.eventChan <- TIMEOUT				
	   case <-elev.intComs.readyChan:
			 elev.intComs.eventChan <- READY
	   }
	}
}
