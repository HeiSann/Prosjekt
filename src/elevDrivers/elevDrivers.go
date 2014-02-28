package elevDrivers

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "time"
    "elevTypes"
)

const SPEED0 = 2048
const SPEED1 = 4024
const REV_TIME = 10 * time.Millisecond


func SetLight(floor int, dir elevTypes.Direction_t){
    switch{  
    case floor == 1 && dir == elevTypes.NONE:
            Set_bit(LIGHT_COMMAND1)
    case floor == 1 && dir == elevTypes.UP:
            Set_bit(LIGHT_UP1)
 	case floor == 2 && dir == elevTypes.NONE:
        Set_bit(LIGHT_COMMAND2)
    case floor == 2 && dir == elevTypes.UP:
        Set_bit(LIGHT_UP2)
    case floor == 2 && dir == elevTypes.DOWN:
        Set_bit(LIGHT_DOWN2)
    case floor == 3 && dir == elevTypes.NONE:
        Set_bit(LIGHT_COMMAND3)
    case floor == 3 && dir == elevTypes.UP:
        Set_bit(LIGHT_UP3)
    case floor == 3 && dir == elevTypes.DOWN:
        Set_bit(LIGHT_DOWN4)        
    case floor == 4 && dir == elevTypes.NONE:
        Set_bit(LIGHT_COMMAND4)
    case floor == 4 && dir == elevTypes.DOWN:
        Set_bit(LIGHT_DOWN4)
    default:
        fmt.Println("Error: Illegal floor or direction")
	}
}

func ClearLight(floor int, dir elevTypes.Direction_t){
    switch{  
    case floor == 1 && dir == elevTypes.NONE:
        Clear_bit(LIGHT_COMMAND1)
    case floor == 1 && dir == elevTypes.UP:
        Clear_bit(LIGHT_UP1)
 	case floor == 2 && dir == elevTypes.NONE:
        Clear_bit(LIGHT_COMMAND2)
    case floor == 2 && dir == elevTypes.UP:
        Clear_bit(LIGHT_UP2)
    case floor == 2 && dir == elevTypes.DOWN:
        Clear_bit(LIGHT_DOWN2)      
    case floor == 3 && dir == elevTypes.NONE:
        Clear_bit(LIGHT_COMMAND3)
    case floor == 3 && dir == elevTypes.UP:
        Clear_bit(LIGHT_UP3)
    case floor == 3 && dir == elevTypes.DOWN:
        Clear_bit(LIGHT_DOWN4)   
    case floor == 4 && dir == elevTypes.NONE:
        Clear_bit(LIGHT_COMMAND4)
    case floor == 4 && dir == elevTypes.DOWN:
        Clear_bit(LIGHT_DOWN4)
    default:
        fmt.Println("elevdriver: Error! Illegal floor or direction!")
		fmt.Println("dir: ", dir, ", floor: ",floor)
	}
}

func ClearAllLights(){
        ClearLight(1, elevTypes.UP)
        ClearLight(2, elevTypes.UP)
        ClearLight(3, elevTypes.UP)
        ClearLight(2, elevTypes.DOWN)
        ClearLight(3, elevTypes.DOWN)
        ClearLight(4, elevTypes.DOWN)
        ClearLight(1, elevTypes.NONE)
        ClearLight(2, elevTypes.NONE)
        ClearLight(3, elevTypes.NONE)
        ClearLight(4, elevTypes.NONE)
		CloseDoor()
        ClearStopButton()
}

func motorCtrl(motorChan chan elevTypes.Direction_t){
    lastDir := elevTypes.NONE
	newDir := elevTypes.NONE

    for {
        newDir=<-motorChan
		fmt.Println("motorCtrl recv newDir=", newDir)
		switch newDir{
        case elevTypes.UP:
           	Clear_bit(MOTORDIR)
            Write_analog(MOTOR,SPEED1)
        case elevTypes.DOWN:
            Set_bit(MOTORDIR)
            Write_analog(MOTOR,SPEED1)
        case elevTypes.NONE:
			/* Reverse direction before stopping*/	
            switch lastDir{
			case elevTypes.DOWN:
				/* Reverse */				
				Clear_bit(MOTORDIR)
        	    Write_analog(MOTOR,SPEED1)
				time.Sleep(REV_TIME)
				/* Stop */
    	        Clear_bit(MOTORDIR)
    	        Write_analog(MOTOR,SPEED0)
            case elevTypes.UP:
				/* Reverse */
				Set_bit(MOTORDIR)
    	        Write_analog(MOTOR,SPEED1)
				time.Sleep(REV_TIME)
				/* Stop */
    	        Set_bit(MOTORDIR)
    	        Write_analog(MOTOR,SPEED0)
			case elevTypes.NONE:
				fmt.Println("elevdriver: lastDir=newDir=elevTypes.NONE, problem?")
            default:
            fmt.Println("elevdriver: ERROR, illegal lastDir")
			}
		default:
            Write_analog(MOTOR,SPEED0)
            fmt.Println("elevdriver: ERROR, illegal motor direction")
        }
        lastDir = newDir
	}
}

func listenButtons(buttonChan chan elevTypes.Button){
    var buttonMap = map[int]elevTypes.Button{
        FLOOR_COMMAND1: {1, elevTypes.NONE},
        FLOOR_COMMAND2: {2, elevTypes.NONE},
        FLOOR_COMMAND3: {3, elevTypes.NONE},
        FLOOR_COMMAND4: {4, elevTypes.NONE},
        FLOOR_UP1:      {1, elevTypes.UP},
        FLOOR_UP2:      {2, elevTypes.UP},
        FLOOR_UP3:      {3, elevTypes.UP},
        FLOOR_DOWN2:    {2, elevTypes.DOWN},
        FLOOR_DOWN3:    {3, elevTypes.DOWN},
        FLOOR_DOWN4:    {4, elevTypes.DOWN},
    }

   	buttonList := make(map[int]bool)
    for key, _ := range buttonMap {
        buttonList[key] = Read_bit(key)
    }    
    
     for key, button := range buttonMap {
         newValue := Read_bit(key)
         if newValue && !buttonList[key] {
            newButton := button
            go func() {		//why not select???
                buttonChan <- newButton
            }()
         }
      buttonList[key] = newValue
      }
}

func listenSensors(sensorChan chan int){
    var floorMap = map[int]int{
        SENSOR1: 1,
        SENSOR2: 2,
        SENSOR3: 3,
        SENSOR4: 4,
    }
    
    atFloor := false
    
    floorList := make(map[int]bool)
    for key, _ := range floorMap {
        floorList[key] = Read_bit(key)
    }
    
    for {
        time.Sleep(25 * time.Millisecond)
        atFloor = false
        for key, floor := range floorMap {
            if Read_bit(key) {
                select {		//why not go?
                    case sensorChan <- floor:
                    default:
                }
                atFloor = true
            }
        }
        if !atFloor {
	        select {
            case sensorChan <- -1:
            default:
            }
        }
	}   
}

func InitElev(	) elevTypes.Drivers_s{
	
   buttonChan  := make(chan elevTypes.Button)
   sensorChan   := make(chan int)
   motorChan   := make(chan elevTypes.Direction_t)
   stopButtonChan := make(chan bool)
   obsChan     := make(chan bool)		

	if !IoInit(){
        fmt.Println("elevdriver: Driver init()... OK!")
	} else {
	    fmt.Println("elevdriver: Driver init()... FAILED!")
	}
	
	ClearAllLights();
	
	go listenButtons(buttonChan)
	go listenSensors(sensorChan)
	go motorCtrl(motorChan)
	
	drivers := elevTypes.Drivers_s{
	   buttonChan,
	   sensorChan,
	   motorChan,
	   stopButtonChan,
	   obsChan,
   }
	
	go func() {
    	// capture ctrl+c and stop elevator
        c := make(chan os.Signal)
        signal.Notify(c, os.Interrupt)
        s := <-c
        log.Printf("Got: %v, terminating program..", s)
		/* stop motor, no reverse and delay */
        Clear_bit(MOTORDIR)
    	Write_analog(MOTOR,SPEED0)
        ClearAllLights()
        os.Exit(1)
    }()
    
    return drivers
}

func OpenDoor(){
    	Set_bit(DOOR_OPEN)
}

func CloseDoor(){
    	Clear_bit(DOOR_OPEN)
}

func GetStopButton() bool{
    	return Read_bit(STOP)
}

func SetStopButton(){
    	Set_bit(STOP)
}

func ClearStopButton(){
    	Clear_bit(STOP)
}

func GetObstruction() bool{
	return Read_bit(OBSTRUCTION)
}
	
