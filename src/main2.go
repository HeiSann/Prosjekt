package main

import(
   "fmt"
   "elevTypes"
   "messages"
   "elevNet"
   "comsManager"
   "elevDrivers"
   "elevOrders"
   "elevCtrl"
)

type Elevator struct{
	driver      Drivers_s
	net         Net_s
	coms        ComsManager_s
	orders      Orders_s
	fsm         Fsm_s
}

func main(){
   var net = elevNet.init()
   var coms = comsManager(net.ExternalChannels)
   var drivers = elevDrivers.init()
   var orders = elevOrders.init(driver)
   var fsm = elevCtrl.init(driver, orders)
   
   var Elev = elevTypes.Elevator{net, coms, driver, orders, fsm}
       

   for{}
    
}
