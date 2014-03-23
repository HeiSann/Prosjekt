package elevDrivers

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
	"elevTypes"
)

const SPEED_STOP = 2048
const SPEED_MOVING = 4024
const REV_TIME = 10 * time.Millisecond

type Drivers_s struct{
	ExtComs 	elevTypes.Drivers_ExtComs_s
}


func Init() Drivers_s{
	//Driver will send on these channels
	buttonChan		:= make(chan elevTypes.Button)
	sensorChan		:= make(chan int)
	motorChan		:= make(chan elevTypes.Direction_t)
	stopButtonChan  := make(chan bool)
	obsChan			:= make(chan bool)	
	//Driver will recieve from these channels
	setLightChan	:= make(chan elevTypes.Light_t)
	setFloorIndChan := make(chan int)
	doorOpenChan	:= make(chan bool)

	if !IoInit(){
		fmt.Println("		elevdriver: Driver init()... OK!")
	} else {
		fmt.Println("		elevdriver: Driver init()... FAILED!")
	}
		
	go listenButtons(buttonChan)
	go listenSensors(sensorChan)
	go motorCtrl(motorChan)
	go listenCtrlSignals(setLightChan, setFloorIndChan, doorOpenChan)
	go captureCtrlC()
	
	doorOpenChan <- false
	clearAllLights()
	
	//Set external communication channels
	driver := elevTypes.Drivers_ExtComs_s{}
	driver.ButtonChan	   = buttonChan
	driver.SensorChan	   = sensorChan
	driver.MotorChan		= motorChan
	driver.StopButtonChan   = stopButtonChan
	driver.ObsChan		  = obsChan 
	driver.SetLightChan	 = setLightChan
	driver.SetFloorIndChan  = setFloorIndChan
	driver.DoorOpenChan	 = doorOpenChan
	  
	return Drivers_s{driver}
}


func motorCtrl(motorChan chan elevTypes.Direction_t){
		lastDir := elevTypes.NONE
		newDir := elevTypes.NONE

		for {
		   newDir=<-motorChan
			switch newDir{
				case elevTypes.UP:
					ClearBit(MOTORDIR)
					WriteAnalog(MOTOR,SPEED_MOVING)
			 	case elevTypes.DOWN:
					SetBit(MOTORDIR)
					WriteAnalog(MOTOR,SPEED_MOVING)
			 	case elevTypes.NONE:
					//Reverse direction before stopping
					switch lastDir{
						case elevTypes.DOWN:
							//Reverse 				
							ClearBit(MOTORDIR)
				 		  	WriteAnalog(MOTOR,SPEED_MOVING)
							time.Sleep(REV_TIME)
							//Stop 
			 	   			ClearBit(MOTORDIR)
			 	   			WriteAnalog(MOTOR,SPEED_STOP)
					 	case elevTypes.UP:
							//Reverse
							SetBit(MOTORDIR)
			 	   			WriteAnalog(MOTOR,SPEED_MOVING)
							time.Sleep(REV_TIME)
							//Stop
			 	   		 	SetBit(MOTORDIR)
			 		  	  	WriteAnalog(MOTOR,SPEED_STOP)
						case elevTypes.NONE:
							
				  		default:
					 		fmt.Println("		elevDrivers.motorCtrl: ERROR, illegal lastDir")
					}
				default:
					WriteAnalog(MOTOR,SPEED_STOP)
					fmt.Println("		elevDrivers.motorCtrl: ERROR, illegal motor direction")
			}
			lastDir = newDir
		}
}

func listenButtons(buttonChan chan elevTypes.Button){
	var buttonMap = map[int]elevTypes.Button{
		FLOOR_COMMAND1: {0, elevTypes.NONE},
		FLOOR_COMMAND2: {1, elevTypes.NONE},
		FLOOR_COMMAND3: {2, elevTypes.NONE},
		FLOOR_COMMAND4: {3, elevTypes.NONE},
		FLOOR_UP1:	  {0, elevTypes.UP},
		FLOOR_UP2:	  {1, elevTypes.UP},
		FLOOR_UP3:	  {2, elevTypes.UP},
		FLOOR_DOWN2:	{1, elevTypes.DOWN},
		FLOOR_DOWN3:	{2, elevTypes.DOWN},
		FLOOR_DOWN4:	{3, elevTypes.DOWN},
	}

   	buttonList := make(map[int]bool)
	for key, _ := range buttonMap {
		buttonList[key] = ReadBit(key)
	}	
	
	for {
		for key, button := range buttonMap {
			newValue := ReadBit(key)
			if newValue && !buttonList[key] {
		   		newButton := button
				go func() {	
					buttonChan <- newButton
				}()
			}
			buttonList[key] = newValue	  
		}
		time.Sleep(time.Millisecond*elevTypes.SELECT_SLEEP_MS)
	}
}

func listenSensors(sensorChan chan int){
	var floorMap = map[int]int{
		SENSOR1: 0,
		SENSOR2: 1,
		SENSOR3: 2,
		SENSOR4: 3,
	}
	
	atFloor := false
	
	floorList := make(map[int]bool)
	for key, _ := range floorMap {
		floorList[key] = ReadBit(key)
	}
	
	for {
		atFloor = false
		for key, floor := range floorMap {
			if ReadBit(key) {
				sensorChan <- floor
				atFloor = true
			}
		}
		if !atFloor {
			select {
			case sensorChan <- -1:
			default:
			}
			
		}
		time.Sleep(time.Millisecond*elevTypes.SELECT_SLEEP_MS)
	}   
}


func listenCtrlSignals(setLightChan chan elevTypes.Light_t, setFloorIndChan chan int, doorOpenChan chan bool){
	for{
		select{
			case light := <-setLightChan:
				if light.Set{
					setLight(light.Floor, light.Direction)
				}else{
					clearLight(light.Floor, light.Direction)
				}
			case floor := <-setFloorIndChan:
				setFloor(floor)
			case open := <-doorOpenChan:
				if open{
					SetBit(DOOR_OPEN)
				}else{
					ClearBit(DOOR_OPEN)
				}
		}
		time.Sleep(time.Millisecond*elevTypes.SELECT_SLEEP_MS)
	}
}


func captureCtrlC() {
		// capture ctrl+c and stop elevator
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		s := <-c
		log.Printf("Got: %v, terminating program..", s)
		/* stop motor, no reverse and delay */
		ClearBit(MOTORDIR)
		WriteAnalog(MOTOR,SPEED_STOP)
		clearAllLights()
		os.Exit(1)
}


func setFloor(floor int) {
		switch floor {
		case 0:
				ClearBit(FLOOR_IND1)
				ClearBit(FLOOR_IND2)
		case 1:
				ClearBit(FLOOR_IND1)
				SetBit(FLOOR_IND2)
		case 2:
				SetBit(FLOOR_IND1)
				ClearBit(FLOOR_IND2)
		case 3:
				SetBit(FLOOR_IND1)
				SetBit(FLOOR_IND2)
		}
}


func setLight(floor int, dir elevTypes.Direction_t){
	switch{  
	case floor == 0 && dir == elevTypes.NONE:
			SetBit(LIGHT_COMMAND1)
	case floor == 0 && dir == elevTypes.UP:
			SetBit(LIGHT_UP1)
 	case floor == 1 && dir == elevTypes.NONE:
		SetBit(LIGHT_COMMAND2)
	case floor == 1 && dir == elevTypes.UP:
		SetBit(LIGHT_UP2)
	case floor == 1 && dir == elevTypes.DOWN:
		SetBit(LIGHT_DOWN2)
	case floor == 2 && dir == elevTypes.NONE:
		SetBit(LIGHT_COMMAND3)
	case floor == 2 && dir == elevTypes.UP:
		SetBit(LIGHT_UP3)
	case floor == 2 && dir == elevTypes.DOWN:
		SetBit(LIGHT_DOWN3)		
	case floor == 3 && dir == elevTypes.NONE:
		SetBit(LIGHT_COMMAND4)
	case floor == 3 && dir == elevTypes.DOWN:
		SetBit(LIGHT_DOWN4)
	default:
		fmt.Println("		elevDrivers.setLight: Error, Illegal floor or direction")
		fmt.Println("		dir: ", dir, ", floor: ",floor)
	}
}

func clearLight(floor int, dir elevTypes.Direction_t){
	switch{  
	case floor == 0 && dir == elevTypes.NONE:
		ClearBit(LIGHT_COMMAND1)
	case floor == 0 && dir == elevTypes.UP:
		ClearBit(LIGHT_UP1)
 	case floor == 1 && dir == elevTypes.NONE:
		ClearBit(LIGHT_COMMAND2)
	case floor == 1 && dir == elevTypes.UP:
		ClearBit(LIGHT_UP2)
	case floor == 1 && dir == elevTypes.DOWN:
		ClearBit(LIGHT_DOWN2)	  
	case floor == 2 && dir == elevTypes.NONE:
		ClearBit(LIGHT_COMMAND3)
	case floor == 2 && dir == elevTypes.UP:
		ClearBit(LIGHT_UP3)
	case floor == 2 && dir == elevTypes.DOWN:
		ClearBit(LIGHT_DOWN3)   
	case floor == 3 && dir == elevTypes.NONE:
		ClearBit(LIGHT_COMMAND4)
	case floor == 3 && dir == elevTypes.DOWN:
		ClearBit(LIGHT_DOWN4)
	default:
		fmt.Println("		elevDrivers.clearLight: Error! Illegal floor or direction!")
		fmt.Println("		dir: ", dir, ", floor: ",floor)
	}
}

func clearAllLights(){
		clearLight(0, elevTypes.UP)
		clearLight(1, elevTypes.UP)
		clearLight(2, elevTypes.UP)
		clearLight(1, elevTypes.DOWN)
		clearLight(2, elevTypes.DOWN)
		clearLight(3, elevTypes.DOWN)
		clearLight(0, elevTypes.NONE)
		clearLight(1, elevTypes.NONE)
		clearLight(2, elevTypes.NONE)
		clearLight(3, elevTypes.NONE)
		ClearBit(DOOR_OPEN)
		ClearBit(LIGHT_STOP)
}


