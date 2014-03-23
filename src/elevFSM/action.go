package elevFSM

import (
	"elevTypes"	
)


/* Finite State Machine Actions */
func (self *Fsm_s)action_start(){
	//fmt.Println("				fsm.action_start") 
	order := <- self.intComs.newOrderChan
	curr_floor:= self.lastFloor	
	switch {
		case order.Floor == curr_floor:
			//fmt.Println("				fsm.action_start: trying to send exec_order on int.eventChan")
			self.intComs.eventChan <- EXEC_ORDER
			//fmt.Println("				fsm.action_start: exec_order sendt int.eventChan")
		case order.Floor < curr_floor:
			self.startDown()
		case order.Floor > curr_floor:
			self.startUp()
	}
}

func (self *Fsm_s)action_checkOrder(){
	//fmt.Println("				fsm.action_checkOrder")
	floor:= <-self.intComs.floorChan
	self.lastFloor = floor
	self.ExtComs.SetFloorIndChan <- self.lastFloor
	
	current := elevTypes.ElevPos_t{self.lastFloor, self.lastDir, true}
	self.ExtComs.ExecRequestChan <- current
}


func (self *Fsm_s)action_execSame(){
	//fmt.Println("				fsm.action_execSame")
	self.ExtComs.DoorOpenChan <- true
	go startTimer(self.intComs.timeoutChan, elevTypes.DOOR_OPEN_TIME)
	self.lastDir = elevTypes.NONE
	self.state = DOORS_OPEN 
}


func (self *Fsm_s)action_exec(){
	self.ExtComs.MotorChan <- elevTypes.NONE
	self.ExtComs.DoorOpenChan <- true
	go startTimer(self.intComs.timeoutChan, elevTypes.DOOR_OPEN_TIME)
	self.state = DOORS_OPEN 
}


func (self *Fsm_s)action_done(){
	//fmt.Println("				fsm.action_done")
	self.ExtComs.DoorOpenChan <- false
	self.state = IDLE
	self.ExtComs.ExecdOrderChan <- elevTypes.ElevPos_t{self.lastFloor, self.lastDir, false}
}   


func (self *Fsm_s)action_stop(){
	self.ExtComs.MotorChan <- elevTypes.NONE
	self.state = EMG_STOP
}


func (self *Fsm_s)action_pause(){
	self.state = OBSTRUCTED
}


func action_dummy(){
}

