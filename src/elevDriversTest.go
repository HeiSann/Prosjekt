package main

import( 
	"runtime"
	"fmt"
   "elevTypes"
   "elevDrivers"
)

type Elevator struct{
	driver      elevTypes.Drivers_s
}

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU()) 
	
	fmt.Println("elevDrivers.init()...")
   var drivers = elevDrivers.Init()
   
   var elevator = Elevator{}
	elevator.driver = drivers
	fmt.Println("OK!")

	elevator.driver.MotorChan <- elevTypes.DOWN
	sensor := -1

	for{
		select{
		case sensor=<-elevator.driver.SensorChan:
			if sensor == 1{
				fmt.Println("at first floor, going up!")
				elevator.driver.MotorChan <- elevTypes.UP
			}
			if sensor == 4{
				fmt.Println("at fourth floor, going down!")
				elevator.driver.MotorChan <- elevTypes.DOWN
			}
		}	
	}
}
