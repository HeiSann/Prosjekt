package elevFSM

import("elevTypes")


/*	Declaration of states, events and fsm_structs */
type State_t int
const(
	IDLE State_t = iota
	DOORS_OPEN
	MOVING_UP
	MOVING_DOWN
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
	timeoutChan		chan bool
	newOrderChan 	chan elevTypes.Order_t
	floorChan		chan int
}

type Fsm_s struct{
    table		[][]func()
	state 		State_t
	lastDir		elevTypes.Direction_t
	lastFloor   int 
	intComs		intComs_s
	ExtComs     elevTypes.Fsm_ExtComs_s
}

