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

type Drivers_s struct{
	ExtComs 	elevTypes.Drivers_ExtComs_s
}


func setLight(floor int, dir elevTypes.Direction_t){
    switch{  
    case floor == 0 && dir == elevTypes.NONE:
            Set_bit(LIGHT_COMMAND1)
    case floor == 0 && dir == elevTypes.UP:
            Set_bit(LIGHT_UP1)
 	case floor == 1 && dir == elevTypes.NONE:
        Set_bit(LIGHT_COMMAND2)
    case floor == 1 && dir == elevTypes.UP:
        Set_bit(LIGHT_UP2)
    case floor == 1 && dir == elevTypes.DOWN:
        Set_bit(LIGHT_DOWN2)
    case floor == 2 && dir == elevTypes.NONE:
        Set_bit(LIGHT_COMMAND3)
    case floor == 2 && dir == elevTypes.UP:
        Set_bit(LIGHT_UP3)
    case floor == 2 && dir == elevTypes.DOWN:
        Set_bit(LIGHT_DOWN3)        
    case floor == 3 && dir == elevTypes.NONE:
        Set_bit(LIGHT_COMMAND4)
    case floor == 3 && dir == elevTypes.DOWN:
        Set_bit(LIGHT_DOWN4)
    default:
        fmt.Println("elevDrivers.setLight: Error, Illegal floor or direction")
        fmt.Println("dir: ", dir, ", floor: ",floor)
	}
}

func clearLight(floor int, dir elevTypes.Direction_t){
    switch{  
    case floor == 0 && dir == elevTypes.NONE:
        Clear_bit(LIGHT_COMMAND1)
    case floor == 0 && dir == elevTypes.UP:
        Clear_bit(LIGHT_UP1)
 	case floor == 1 && dir == elevTypes.NONE:
        Clear_bit(LIGHT_COMMAND2)
    case floor == 1 && dir == elevTypes.UP:
        Clear_bit(LIGHT_UP2)
    case floor == 1 && dir == elevTypes.DOWN:
        Clear_bit(LIGHT_DOWN2)      
    case floor == 2 && dir == elevTypes.NONE:
        Clear_bit(LIGHT_COMMAND3)
    case floor == 2 && dir == elevTypes.UP:
        Clear_bit(LIGHT_UP3)
    case floor == 2 && dir == elevTypes.DOWN:
        Clear_bit(LIGHT_DOWN3)   
    case floor == 3 && dir == elevTypes.NONE:
        Clear_bit(LIGHT_COMMAND4)
    case floor == 3 && dir == elevTypes.DOWN:
        Clear_bit(LIGHT_DOWN4)
    default:
        fmt.Println("elevDrivers.clearLight: Error! Illegal floor or direction!")
		fmt.Println("dir: ", dir, ", floor: ",floor)
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
		Set_bit(DOOR_OPEN)
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
							fmt.Println("elevDrivers.motorCtrl: lastDir=newDir=elevTypes.NONE, problem?")
		      		default:
		         		fmt.Println("elevDrivers.motorCtrl: ERROR, illegal lastDir")
					}
				default:
		        	Write_analog(MOTOR,SPEED0)
		        	fmt.Println("elevDrivers.motorCtrl: ERROR, illegal motor direction")
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
        FLOOR_UP1:      {0, elevTypes.UP},
        FLOOR_UP2:      {1, elevTypes.UP},
        FLOOR_UP3:      {2, elevTypes.UP},
        FLOOR_DOWN2:    {1, elevTypes.DOWN},
        FLOOR_DOWN3:    {2, elevTypes.DOWN},
        FLOOR_DOWN4:    {3, elevTypes.DOWN},
    }

   	buttonList := make(map[int]bool)
    for key, _ := range buttonMap {
        buttonList[key] = Read_bit(key)
    }    
    
	for {
		for key, button := range buttonMap {
			newValue := Read_bit(key)
			if newValue && !buttonList[key] {
				fmt.Println("Drivers.listenButtonsbutton: button pressed: ", button)
           		newButton := button
            	go func() {		//why not select???
					//fmt.Println("waiting to send...")
                	buttonChan <- newButton
					//fmt.Println("button sendt!")
            	}()
			}
			buttonList[key] = newValue      
		}
		time.Sleep(time.Millisecond*elevTypes.SLOW_DOWM_MUTHA_FUKKA)
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
        floorList[key] = Read_bit(key)
    }
    
    for {
        atFloor = false
        for key, floor := range floorMap {
            //fmt.Println("drivers.listenSensors: checking key, floor: ", key, floor)
            if Read_bit(key) {
               //fmt.Println("drivers.listenSensors: got reading on floor: ", floor)
               /*
                //fmt.Println("got floor reading")
                go func() {		//why not select???
					//fmt.Println("drivers.listenSensors: waiting to send: ", floor)
                	sensorChan <- floor
					//fmt.Println("drivers.listenSensors: floor sendt!")
            	}()
                */
                // fmt.Println("drivers.listenSensors: trying to send floor on blocking: ", floor)
                sensorChan <- floor
                //fmt.Println("drivers.listenSensors: sendt floor on blocking: ", floor)
                /*
                fmt.Println("drivers.listenSensors: trying to send floor unblocking: ", floor)
                select {		//why not go?
                    case sensorChan <- floor:
                        fmt.Println("drivers.listenSensors: sendt floor unblocking: ", floor)
                    
                    
                    default:
                        fmt.Println("drivers.listenSensors: gave up sending unblocking: ", floor)
                }
                */
                atFloor = true
            }
        }
        if !atFloor {
	        select {
            case sensorChan <- -1:
            default:
            }
			
        }
        time.Sleep(time.Millisecond*elevTypes.SLOW_DOWM_MUTHA_FUKKA)
	}   
}

func Init() Drivers_s{
	
    buttonChan		:= make(chan elevTypes.Button)
    sensorChan		:= make(chan int)
    motorChan		:= make(chan elevTypes.Direction_t)
    stopButtonChan  := make(chan bool)
    obsChan			:= make(chan bool)	
	
	setLightChan	:= make(chan elevTypes.Light_t)
	setFloorIndChan := make(chan int)
	doorOpenChan	:= make(chan bool)

	if !IoInit(){
        fmt.Println("elevdriver: Driver init()... OK!")
	} else {
	    fmt.Println("elevdriver: Driver init()... FAILED!")
	}
	
	clearAllLights();
	
	go listenButtons(buttonChan)
	go listenSensors(sensorChan)
	go motorCtrl(motorChan)
	go listenCtrlSignals(setLightChan, setFloorIndChan, doorOpenChan)
	
	doorOpenChan <- false
	
	driver := elevTypes.Drivers_ExtComs_s{}

	driver.ButtonChan       = buttonChan
	driver.SensorChan       = sensorChan
	driver.MotorChan        = motorChan
	driver.StopButtonChan   = stopButtonChan
	driver.ObsChan          = obsChan 
	driver.SetLightChan     = setLightChan
	driver.SetFloorIndChan  = setFloorIndChan
	driver.DoorOpenChan     = doorOpenChan
   
	
	go func() {
    	// capture ctrl+c and stop elevator
        c := make(chan os.Signal)
        signal.Notify(c, os.Interrupt)
        s := <-c
        log.Printf("Got: %v, terminating program..", s)
		/* stop motor, no reverse and delay */
        Clear_bit(MOTORDIR)
    	Write_analog(MOTOR,SPEED0)
        clearAllLights()
        os.Exit(1)
    }()
    
    return Drivers_s{driver}
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
                    Set_bit(DOOR_OPEN)
                }else{
                    Clear_bit(DOOR_OPEN)
                }
        }
        time.Sleep(time.Millisecond*elevTypes.SLOW_DOWM_MUTHA_FUKKA)
    }
}

func setFloor(floor int) {
        switch floor {
        case 0:
                Clear_bit(FLOOR_IND1)
                Clear_bit(FLOOR_IND2)
        case 1:
                Clear_bit(FLOOR_IND1)
                Set_bit(FLOOR_IND2)
        case 2:
                Set_bit(FLOOR_IND1)
                Clear_bit(FLOOR_IND2)
        case 3:
                Set_bit(FLOOR_IND1)
                Set_bit(FLOOR_IND2)
        }
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
	
