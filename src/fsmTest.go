package main

import(
	"fmt"
   "elevDrivers"
   "elevNet"
   "comsManager"
   "elevOrders"
   "elevFSM"
)

type Elevator struct{
	driver      elevDrivers.Drivers_s
	orders      elevOrders.Orders_s
	fsm         elevFSM.Fsm_s
}

func main(){
	end := make(chan bool)

	fmt.Println("start of main")
   var drivers = elevDrivers.Init()
   var net = elevNet.Init()
   var coms = comsManager.Init(net.Ip, net.ExtComs)
   var orders = elevOrders.Init(drivers.ExtComs, coms.ExtComs)
   var fsm = elevFSM.Init(drivers.ExtComs, orders.ExtComs)
   
   var Elev = Elevator{drivers, orders, fsm}
       
   <-end
	fmt.Println(Elev)
    
}
