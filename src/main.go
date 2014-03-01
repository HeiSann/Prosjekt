package main

import "runtime"
import "fmt"
import "elevNet"
import "comsManager"
import "elevdriver"
import "elevCtrl"
import "os/exec"


const BCAST_IP = "129.241.187.255"
const LISTEN_PORT = "20022"
const TARGET_PORT = "20011"


func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) 
	//sleepChan :=make(chan int)
	//test :=coms.ConstructPckg("129.241.187.255","PING", "cake or death?")
	coms.ComsChanInit()

	//go coms.SendPckgTo(BCAST_IP,TARGET_PORT ,test)
	
	go coms.ListenToBroadcast(LISTEN_PORT,coms.ComsChan )
	go network.DeliverPckg(coms.ComsChan)
	fmt.Println("Coms OK")

	buttonChan := make(chan elevdriver.Button)
    floorChan := make(chan int)
    motorChan := make(chan elevdriver.Direction_t)
    stopButtonChan := make(chan bool)
    obsChan := make(chan bool)

	var elev elevCtrl.Elevator

	elev.ElevInit(
		buttonChan, 
		floorChan, 
		motorChan,
		stopButtonChan, 
		obsChan)

	elevdriver.MotorDown(motorChan)
	sensor := -1
	for{
		select{
		case sensor=<-floorChan:
			if sensor == 1{
				elevdriver.MotorUp(motorChan)
			}
			if sensor == 4{
				elevdriver.MotorDown(motorChan)
			}
		}	
	}
}
