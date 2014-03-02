package elevFSM

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

	fsm := Fsm_s{}
	fsm.init_fsm_table()
	fsm.init_intComs()
	fsm.init_ExtComs()

	fsm.start()
	
	fmt.Println("fsm.Init: OK!")
	return fsm
}

/* Finite State Machine Actions */
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

func (fsm *Fsm_s)init_fsm_table2(){
	fsm.table = [][]func(){
/*STATES:	  \	EVENTS:	//NewOrder			//FloorReached        	//Exec  				//TimerOut			//Obst			//EmgPressed
/*IDLE       */  []func(){fsm.action_start,	action_dummy,				fsm.action_exec,  action_dummy,		action_dummy,	action_dummy},
/*DOORS_OPEN */  []func(){action_dummy,		action_dummy,   			action_dummy,  	fsm.action_done,	action_dummy,  action_dummy},  
/*MOVING_UP  */  []func(){action_dummy,		fsm.action_check_order, fsm.action_exec,	action_dummy,		action_dummy,  action_dummy},
/*MOVING_DOWN*/  []func(){action_dummy,		fsm.action_check_order, fsm.action_exec,	action_dummy,		action_dummy,  action_dummy},
/*EMG_STOP   */  []func(){action_dummy,		action_dummy,           action_dummy,     action_dummy,		action_dummy,  action_dummy},  
/*OBST       */  []func(){action_dummy,		action_dummy,           action_dummy,     action_dummy,		action_dummy,  action_dummy}, 
	}
}

func(self *Fsm_s)init_intComs(){
	self.intComs.eventChan 			= make(chan Event_t)
	self.intComs.startTimerChan 	= make(chan bool)
	self.intComs.timeoutChan		= make(chan bool)
	self.intComs.readyChan  		= make(chan bool)
}

func(self *Fsm_s)init_ExtComs(driver elevTypes.Drivers_ExtComs_s, orders elevTypes.Orders_ExtComs_s){
	self.OrderExdChan 	= make(chan Order_t)  //fsm -> orders	
	self.ButtonChan 		= driver.ButtonChan
	self.FloorChan 		= driver.SensorChan
	self.StopButtonChan	= driver.StopButtonChan
	self.ObsChan			= driver.ObsChan	
	self.MotorChan			= driver.MotorChan
	self.DoorOpenChan		= driver.DoorOpenChan
	self.SetLightChan		= driver.SetLightChan
	self.FloorIndChan		= driver.SetFloorIndChan
	self.NewOrdersChan	= orders.OrderUpdatedChan
	self.EmgTriggerChan	= orders.EMG2Fsm
}

func (self *Fsm_s)start(){
	if not_at_floor
		self.ExtComs.MotorChan <- elevTypes.DOWN
	for not_at_floor{
		//wait until floor reached
	}	
	//stop motor
	self.ExtComs.MotorChan <- elevTypes.NONE
	//update fsm vars
	self.state = IDLE
	self.lastfloor = cur_floor
	self.lastDir = NONE
	//start the fsm routunes
	go fsm_update()
	go fsm.generate_events()
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

func (fsm *Fsm_s)handle_next_order(){
    if fsm.state == IDLE{
	    switch(fsm.get_nearest_order()){
	    case elevTypes.UP:
	       fsm.intComs.eventChan <- START_UP
	    case elevTypes.DOWN:
			 fsm.intComs.eventChan <- START_DOWN
	    case elevTypes.NONE:
	       // no new orders: do nothing
	    default:
	       fmt.Println("fsm: ERROR, undefined elev.lastDir in execute_next_order")
	    }
	}else{
	    fmt.Println("fsm: new order registered, will be executed when I'm ready")
   }
}

func (fsm *Fsm_s)handle_new_order(){
    if self.state == IDLE{
	    switch(fsm.get_nearest_order()){
	    case elevTypes.UP:
	       fsm.intComs.eventChan <- START_UP
	    case elevTypes.DOWN:
			 fsm.intComs.eventChan <- START_DOWN
	    case elevTypes.NONE:
	       fsm.intComs.eventChan <- EXEC_ORDER
	    default:
	       fmt.Println("fsm: ERROR, undefined elev.lastDir in execute_next_order")
	    }
	}else{
	    fmt.Println("fsm: new order registered, will be executed when I'm ready")
   }
}

func (fsm *Fsm_s)generate_events(){
	for{
	   select{
	   case <- fsm.ExtComs.StopButtonChan:
			fsm.intComs.eventChan <- EMG
	   case <- fsm.ExtComs.ObsChan:
			fsm.intComs.eventChan <- OBSTRUCTION
	   case floor:=<- self.ExtComs.FloorChan:
		   if floor != -1 && floor != fsm.lastFloor{
			   fsm.lastFloor = floor
			   //set floor_light
			   if self.should_stop(floor){
					self.intComs.eventChan <- EXEC_ORDER
			   }
		   }
	   case <- fsm.ExtComs.NewOrdersChan:
		   fsm.handle_new_order()
      case <-fsm.intComs.timeoutChan:
			 fsm.intComs.eventChan <- TIMEOUT				
	   case <-fsm.intComs.readyChan:
			 fsm.intComs.eventChan <- READY
	   }
	}
}
