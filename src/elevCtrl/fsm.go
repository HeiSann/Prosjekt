package elevCtrl

import (
	"fmt"
	"elevdriver"
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
	stop
	obst
)

/* 	FSM Actions */
func (elev *Elevator)action_start_down(){
	elevdriver.MotorDown(elev.motorChan)
	elev.state = MOVING_DOWN
	elev.lastDir = elevdriver.DOWN
	fmt.Println("fsm: MOVING_DOWN\n")
}

func (elev *Elevator)action_start_up(){
	elevdriver.MotorUp(elev.motorChan)
	elev.state = MOVING_UP
	elev.lastDir = elevdriver.UP
	fmt.Println("fsm: MOVING_UP\n")
}

func (elev *Elevator)action_exec(){
	elevdriver.OpenDoor()
	elevdriver.SetLight(elev.lastFloor, elev.lastDir)
	//start_timer()	
	//order_executed()
	elev.state = DOORS_OPEN 
	fmt.Println("fsm: DOORS_OPEN\n")
}

func (elev *Elevator)action_halt_n_exec(){
	elevdriver.MotorStop(elev.motorChan)
	elevdriver.OpenDoor()
	elevdriver.SetLight(elev.lastFloor, elev.lastDir)
	//start_timer()	
	//order_executed()
	elev.state = DOORS_OPEN 
	fmt.Println("fsm: DOORS_OPEN\n")
}

func (elev *Elevator)action_done(){
	elevdriver.CloseDoor()
	// stop_timer()
	elev.lastDir = elev.get_nearest_order()
	elev.state = IDLE
	fmt.Println("fsm: IDLE\n")
	elev.ready
}

func (elev *Elevator)action_next(){
    elev.handle_new_order()
}

func (elev *Elevator)action_stop(){
	elevdriver.MotorStop(elev.motorChan)
	elev.state = EMG
}

func (elev *Elevator)action_pause(){
	elev.state = OBST
}

func action_dummy(){
	fmt.Println("fsm: dummy!\n")
}

/* Finite State Machine */
func (elev *Elevator)fsm_init(){
	elev.fsm_table = [][]func(){
/*STATES:	  \	EVENTS:	//start_down			   //start_up              //exec_order			//timeout			   //ready
/*IDLE       */  []func(){elev.action_start_down, elev.action_start_up,elev.action_exec,	      action_dummy,       elev.action_next},
/*DOORS_OPEN */  []func(){elev.action_start_down, elev.action_start_up,elev.action_exec,	      elev.action_done,   action_dummy},  
/*MOVING_UP  */  []func(){action_dummy, 		      action_dummy,        elev.action_halt_n_exec,action_dummy,       action_dummy},
/*MOVING_DOWN*/  []func(){action_dummy, 		      action_dummy,			elev.action_halt_n_exec,action_dummy,       action_dummy},
/*EMG_STOP   */  []func(){action_dummy, 		      action_dummy,			action_dummy,           action_dummy,       action_dummy},  
/*OBST       */  []func(){action_dummy, 		      action_dummy,			action_dummy,           action_dummy,       action_dummy}, 
	}
}

/* FSM help functions */
func (elev *Elevator)fsm_update(event Event_t){
	elev.fsm_table[elev.state][event]()
}

func (elev *Elevator)should_stop(floor int) bool{
	//communicate with orders and check if current floor got pending orders in correct direction
	stop := true	
	return stop
}

func (elev *Elevator)get_nearest_order() elevdriver.Direction_t{
	//communicate with orders and get next direction
	return elevdriver.UP
}

func (elev *Elevator)handle_new_order(){
    if elev.state == IDLE{
	    switch(elev.get_nearest_order()){
	    case elevdriver.UP:
	       elev.fsm_update(start_up)
	    case elevdriver.DOWN:
	       elev.fsm_update(start_down)
	    case elevdriver.NONE:
	       elev.state = IDLE
	       fmt.Println("fsm: IDLE")
	    default:
	       fmt.Println("fsm: ERROR, undefined elev.lastDir in execute_next_order")
	    }
	}else{
	    fmt.Println("fsm: new order registered, will be executed when I'm ready")
}

func (elev *Elevator)fsm_generate_n_handle_events(){
	for{
	   select{
	   case <- elev.stopButtonChan:
		   elev.fsm_update(stop)
	   case <- elev.obsChan:
		   elev.fsm_update(obst)
	   case floor:=<- elev.floorChan:
		   if floor != -1{
			   elev.lastFloor = floor
			   //set floor_light
			   if elev.should_stop(floor){
				   elev.fsm_update(exec_order)
			   }
		   }
	   case <- elev.newOrder:
		   elev.handle_new_order()
       case <-elev.timer:
	       elev.fsm_update(timeout)	
	   case <-elev.ready:
	       elev.fsm_update(ready)
	   }
	}
}
