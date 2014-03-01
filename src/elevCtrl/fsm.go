package elevCtrl

import (
	"fmt"
	"elevTypes"
)

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

/* Finite State Machine */
func (elev *Fsm_s)fsm_init(){
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
