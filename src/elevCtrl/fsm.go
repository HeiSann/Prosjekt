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
	EMG
	OBST
	OBST_EMG
)

type Event_t int
const(
	start_down Event_t = iota
	start_up
	exec_order 
	timeout
	ready
	stop
	obst
)

/* 	FSM Actions */
func (self *Fsm_s)action_start_down(){
   self.Coms.motorChan <- elevTypes.DOWN
	self.state = MOVING_DOWN
	self.lastDir = elevTypes.DOWN
	fmt.Println("fsm: MOVING_DOWN\n")
}

func (self *Fsm_s)action_start_up(){
	self.Coms.motorChan <- elevTypes.UP
	self.state = MOVING_UP
	self.lastDir = elevTypes.UP
	fmt.Println("fsm: MOVING_UP\n")
}

func (self *Fsm_s)action_exec(){
	self.Coms.doorOpenChan <- true
	self.Coms.setLightChan <- elevTypes.Light_t{self.lastFloor, self.lastDir, false}
	self.startTimerChan <- true
	self.Coms.orderExdChan <- elevTypes.Order_t{self.lastFloor, self.lastDir, false}
	self.state = DOORS_OPEN 
	fmt.Println("fsm: DOORS_OPEN\n")
}

func (self *Fsm_s)action_halt_n_exec(){
	self.Coms.motorChan <- elevTypes.NONE
	self.Coms.doorOpenChan <- true
	self.Coms.setLightChan <- elevTypes.Light_t{self.lastFloor, self.lastDir, false}
	self.startTimerChan <- true	
	self.Coms.orderExdChan <- elevTypes.Order_t{self.lastFloor, self.lastDir, false}
	self.state = DOORS_OPEN 
	fmt.Println("fsm: DOORS_OPEN\n")
}

func (self *Fsm_s)action_done(){
	self.Coms.doorOpenChan <- false
	self.lastDir = self.get_nearest_order()
	self.state = IDLE
	fmt.Println("fsm: IDLE\n")
	self.eventChan <- ready
}   
	
func (self *Fsm_s)action_next(){
    self.handle_next_order()
}

func (self *Fsm_s)action_stop(){
	self.Coms.motorChan <- elevTypes.NONE
	self.state = EMG
}

func (self *Fsm_s)action_pause(){
	self.state = OBST
}

func action_dummy(){
	fmt.Println("fsm: dummy!\n")
}

/* Finite State Machine */
func (elev *Fsm_s)fsm_init(){
	elev.fsm_table = [][]func(){
/*STATES:	  \	EVENTS:	//start_down			   //start_up              //exec_order			   //timeout			   //ready
/*IDLE       */  []func(){elev.action_start_down,  elev.action_start_up,   elev.action_exec,       action_dummy,       elev.action_next},
/*DOORS_OPEN */  []func(){elev.action_start_down,  elev.action_start_up,   elev.action_exec,       elev.action_done,   action_dummy},  
/*MOVING_UP  */  []func(){action_dummy,            action_dummy,           elev.action_halt_n_exec,action_dummy,       action_dummy},
/*MOVING_DOWN*/  []func(){action_dummy,            action_dummy,           elev.action_halt_n_exec,action_dummy,       action_dummy},
/*EMG_STOP   */  []func(){action_dummy,            action_dummy,           action_dummy,           action_dummy,       action_dummy},  
/*OBST       */  []func(){action_dummy,            action_dummy,           action_dummy,           action_dummy,       action_dummy}, 
	}
}

/* FSM help functions */
func (self *Fsm_s)fsm_update(event Event_t){
	self.fsm_table[self.state][event]()
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
	       elev.eventChan <- start_up
	    case elevTypes.DOWN:
			 elev.eventChan <- start_down
	    case elevTypes.NONE:
	       // do nothing
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
	       elev.eventChan <- start_up
	    case elevTypes.DOWN:
			 elev.eventChan <- start_down
	    case elevTypes.NONE:
	       elev.eventChan <- exec_order
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
	   case <- elev.Coms.stopButtonChan:
			elev.eventChan <- stop
	   case <- elev.Coms.obsChan:
			elev.eventChan <- obst
	   case floor:=<- elev.Coms.floorChan:
		   if floor != -1{
			   elev.lastFloor = floor
			   //set floor_light
			   if elev.should_stop(floor){
					elev.eventChan <- exec_order
			   }
		   }
	   case <- elev.Coms.newOrdersChan:
		   elev.handle_new_order()
      case <-elev.timeoutChan:
			 elev.eventChan <- timeout				
	   case <-elev.readyChan:
			 elev.eventChan <- ready
	   }
	}
}
