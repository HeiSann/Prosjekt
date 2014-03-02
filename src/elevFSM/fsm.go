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
	NEW_ORDER Event_t = iota
	FLOOR_REACHED
	EXEC_ORDER 
	TIMEOUT
	OBSTRUCTION
	EMG
)

type intComs_s struct{
   eventChan		chan Event_t
	startTimerChan chan bool
	timeoutChan    chan bool
	orderChan      chan Order_t
	newOrderChan	chan Order_t
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
func (self *Fsm_s)action_start(){
		
}

func (self *Fsm_s)action_check_order(){
	//get_floor
	//update_lights
	//send_stop_request order
	current := Order_t{floor, fsm.lastDir}
	self.ExtComs.RequestStopChan <- current
}

func (self *Fsm_s)action_exec_same(){
	self.ExtComs.DoorOpenChan <- true
	self.ExtComs.SetLightChan <- elevTypes.Light_t{self.lastFloor, self.lastDir, false}
	self.intComs.startTimerChan <- true
	self.ExtComs.OrderExdChan <- elevTypes.Order_t{self.lastFloor, self.lastDir, false}
	self.state = DOORS_OPEN 
	fmt.Println("fsm: DOORS_OPEN\n")
}

func (self *Fsm_s)action_exec(){
	self.ExtComs.MotorChan <- elevTypes.NONE
	self.ExtComs.DoorOpenChan <- true
	self.ExtComs.SetLightChan <- elevTypes.Light_t{self.lastFloor, self.lastDir, false}
	self.intComs.startTimerChan <- true	
	self.state = DOORS_OPEN 
	fmt.Println("fsm: DOORS_OPEN\n")
}

func (self *Fsm_s)action_done(){
	self.ExtComs.DoorOpenChan <- false
	self.state = IDLE
	fmt.Println("fsm: IDLE\n")
	self.ExtComs.OrderExdChan <- elevTypes.Order_t{self.lastFloor, self.lastDir, false}
}   
	
func (self *Fsm_s)start){
    self.handle_next_order()
}

func (self *Fsm_s)action_stop(){
	self.ExtComs.MotorChan <- elevTypes.NONE
	self.state = EMG_STOP
}

func (self *Fsm_s)action_pause(){
	self.state = OBSTRUCTED
}

func (self *Fsm_s)action_discard(){
	<-self.intComs.orderChan
	fmt.Println("New order will be handled later!\n")
}

func action_dummy(){
	fmt.Println("fsm: dummy!\n")
}

/* Finite State Machine initializations */
func (fsm *Fsm_s)init_fsm_table2(){
	fsm.table = [][]func(){
/*STATES:	  \	EVENTS:	//NewOrder			//FloorReached        	//Exec  					//TimerOut			//Obst			//EmgPressed
/*IDLE       */  []func(){fsm.action_start,	action_dummy,				fsm.action_exec_same,action_dummy,		action_pause,	action_stop},
/*DOORS_OPEN */  []func(){action_discard,		action_dummy,   			action_dummy,  		fsm.action_done,	action_pause,  action_stop},  
/*MOVING_UP  */  []func(){action_discard,		fsm.action_check_order, fsm.action_exec,		action_dummy,		action_pause,  action_stop},
/*MOVING_DOWN*/  []func(){action_discard,		fsm.action_check_order, fsm.action_exec,		action_dummy,		action_pause,  action_stop},
/*EMG_STOP   */  []func(){action_discard,		action_dummy,           action_dummy,     	action_dummy,		action_pause,  action_stop},  
/*OBST       */  []func(){action_discard,		action_dummy,           action_dummy,     	action_dummy,		action_dummy,  action_stop}, 
/*OBST+EMG	 */  []func(){action_discard,		action_dummy,           action_dummy,     	action_dummy,		action_dummy,  action_stop}, 
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
func (self *Fsm_s)start_down(){
   self.ExtComs.MotorChan <- elevTypes.DOWN
	self.state = MOVING_DOWN
	self.lastDir = elevTypes.DOWN
	fmt.Println("fsm: MOVING_DOWN\n")
}

func (self *Fsm_s)start_up(){
	self.ExtComs.MotorChan <- elevTypes.UP
	self.state = MOVING_UP
	self.lastDir = elevTypes.UP
	fmt.Println("fsm: MOVING_UP\n")
}

func (self *Fsm_s)fsm_update(){
	var event Event_t
	for{
		event =<- self.intComs.eventChan
		self.fsm_table[self.state][event]()
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
				fsm.intComs.eventChan <- FLOOR_REACHED
		   }
	   case newOrder:=<- fsm.ExtComs.NewOrdersChan:
		   fsm.intComs.eventChan <- NEW_ORDER
			fsm.intComs.orderChan <- newOrder
			}
		case <-fsm.intComs.execChan:
			fsm.intComs.eventChan <- EXEC_ORDER
      case <-fsm.intComs.timeoutChan:
			 fsm.intComs.eventChan <- TIMEOUT				
	}
}

