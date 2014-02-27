package elevCtrl

import (
	"fmt"
	"elevdriver"
)


type Elevator struct{
	fsm_table	[][]func()
	state 		State_t
	lastDir 	elevdriver.Direction_t
	lastFloor 	int
	orders		[][]bool
	newOrder	chan bool   // ser dir?
	timer       chan bool
	buttonChan 	chan elevdriver.Button
	floorChan 	chan int
	motorChan 	chan elevdriver.Direction_t
	stopButtonChan chan bool
	obsChan 	chan bool
}

func (elev *Elevator)ElevInit(	
		buttonChan chan elevdriver.Button,
		sensorChan chan int,
		motorChan chan elevdriver.Direction_t,
		stopButtonChan chan bool,
		obsChan chan bool){

    fmt.Println("elevCtrl: Initializng...")
    elevdriver.InitElev(
	buttonChan, 
	sensorChan, 
	motorChan,
	stopButtonChan, 
	obsChan)

    elev = &Elevator{
        state:          IDLE,
        lastDir:      	elevdriver.NONE,
        lastFloor:      -1,
    }
       
    //go emergWatcher()
    //go obsWatcher()
    //go floorWatcher()
    fmt.Println("elevCtrl: Init OK!")
}
