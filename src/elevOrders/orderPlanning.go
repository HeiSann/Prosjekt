package elevOrders

import(
	"fmt"
	"math"
	"elevTypes"
)


func orderPlanning_getNextOrder(queue [elevTypes.N_FLOORS][elevTypes.N_DIR]bool,order elevTypes.Order_t) elevTypes.Order_t{
	fmt.Println("			order.get_next_order: with order: ", order)
	fmt.Println("			order.get_next_order: queue is now: ", queue)
	switch(order.Direction){
		case elevTypes.UP:
			orderUpAbove := nextOrderAbove(order.Floor, queue)
			if orderUpAbove.Active { 
				fmt.Println("			order.get_next_order: returning orderUpAbove: ", orderUpAbove)
				return orderUpAbove }
			
			orderDown := nextOrderBelow(elevTypes.N_FLOORS-1, queue)
			if orderDown.Active { 
				fmt.Println("			order.get_next_order: returning orderDown: ", orderDown)
				return orderDown }
			
			orderUpBelow := nextOrderAbove(0, queue)
			if orderUpBelow.Active { 
				fmt.Println("			order.get_next_order: returning orderUpBelow: ", orderUpBelow)
				return orderUpBelow }
			
			/*  no orders left  */
			return elevTypes.Order_t{}
			
		case elevTypes.DOWN:
			orderDown_below := nextOrderBelow(order.Floor, queue)
			if orderDown_below.Active { 
				fmt.Println("			order.get_next_order: returning orderDown_below: ", orderDown_below)
				return orderDown_below }
			
			orderUpBelow := nextOrderAbove(0, queue)
			if orderUpBelow.Active{ 
				fmt.Println("			order.get_next_order: returning orderUpBelow: ", orderUpBelow)
				return orderUpBelow }
			
			orderDownAbove := nextOrderBelow(elevTypes.N_FLOORS-1, queue)
			if orderDownAbove.Active{ 
				fmt.Println("			order.get_next_order: returning orderDownAbove: ", orderDownAbove)
				return orderDownAbove}
			/*  no orders left  */
			return elevTypes.Order_t{}

		case elevTypes.NONE:
			orderUpAbove := nextOrderAbove(order.Floor, queue)
			if orderUpAbove.Active { 
				fmt.Println("			order.get_next_order: returning orderUpAbove: ", orderUpAbove)
				return orderUpAbove }

			orderDown_below := nextOrderBelow(order.Floor, queue)
			if orderDown_below.Active { 
				fmt.Println("			order.get_next_order: returning orderDown_below: ", orderDown_below)
				return orderDown_below }

			orderDownAbove := nextOrderBelow(elevTypes.N_FLOORS-1, queue)
			if orderDownAbove.Active{ 
				fmt.Println("			order.get_next_order: returning orderDownAbove: ", orderDownAbove)
				return orderDownAbove}

			orderUpBelow := nextOrderAbove(0, queue)
			if orderUpBelow.Active { 
				fmt.Println("			order.get_next_order: returning orderUpBelow: ", orderUpBelow)
				return orderUpBelow }
			return elevTypes.Order_t{}
		default:
			fmt.Println("			order.get_next_order: dir on request = NONE. this when first order is same floor")
			return order
	}
}  

func nextOrderAbove(thisFloor int, queue[elevTypes.N_FLOORS][elevTypes.N_DIR]bool) elevTypes.Order_t{
	orderOut:= false
	orderUp:= false
	for floor := thisFloor; floor < elevTypes.N_FLOORS; floor++{
			orderOut = queue[floor][elevTypes.NONE]
			orderUp = queue[floor][elevTypes.UP]
			if orderOut{
				fmt.Println("			next_order_above: returning next: ", elevTypes.Order_t{floor, elevTypes.NONE, true}) 
				return elevTypes.Order_t{floor, elevTypes.NONE, true}
			}
			if orderUp{
				fmt.Println("			next_order_above returning next: ", elevTypes.Order_t{floor, elevTypes.UP, true}) 
				return elevTypes.Order_t{floor, elevTypes.UP, true}
			}
		}
	fmt.Println("			next_order_above found nothing, returning empty:  ", elevTypes.Order_t{}) 
	return elevTypes.Order_t{}
}

func nextOrderBelow(thisFloor int, queue[elevTypes.N_FLOORS][elevTypes.N_DIR]bool) elevTypes.Order_t{
		orderOut	:= false
		orderDown   := false
		for floor := thisFloor; floor >= 0; floor--{
			orderOut = queue[floor][elevTypes.NONE]
			orderDown = queue[floor][elevTypes.DOWN]
			if orderOut{
				fmt.Println("			next_order_below: returning next: ", elevTypes.Order_t{floor, elevTypes.NONE, true}) 
				return elevTypes.Order_t{floor, elevTypes.NONE, true}
			}
			if orderDown{
				fmt.Println("			next_order_below: returning next: ", elevTypes.Order_t{floor, elevTypes.DOWN, true}) 
				return elevTypes.Order_t{floor, elevTypes.DOWN, true}
			}
		}
	fmt.Println("			next_order_below found nothing, returning empty:  ", elevTypes.Order_t{}) 
	return elevTypes.Order_t{}
}


func orderPlanning_getScore(order elevTypes.Order_t, elev elevTypes.ElevPos_t, queue [elevTypes.N_FLOORS][elevTypes.N_DIR]bool) int{
	order_already_added := queue[order.Floor][order.Direction] || queue[order.Floor][elevTypes.NONE]
	n_order := countOrders(queue)

	//Empty queue
	if n_order == 0{
		ansFloat := float64(order.Floor) - float64(elev.Floor)
		return 0 + int(math.Abs(ansFloat))

	//Existing identical order
	}else if order_already_added{
		ansFloat := float64(order.Floor) - float64(elev.Floor)
		return elevTypes.N_FLOORS + int(math.Abs(ansFloat))
		
	//Order in same direction, infront of the elevator
	}else if (elev.Direction == order.Direction) && (order.Floor>elev.Floor) && (order.Direction==elevTypes.UP){
		return 2*elevTypes.N_FLOORS + order.Floor-elev.Floor + 2*n_order
	}else if (elev.Direction == order.Direction) && (order.Floor<elev.Floor) && (order.Direction==elevTypes.DOWN){
		return 2*elevTypes.N_FLOORS + elev.Floor-order.Floor + 2*n_order
		
	//Order in opposite direction infront of elevator
	}else if  (elev.Direction != order.Direction) && (order.Floor>elev.Floor) && (order.Direction==elevTypes.UP){
		return 5*elevTypes.N_FLOORS + order.Floor-elev.Floor + 2*n_order
	}else if (elev.Direction != order.Direction) && (order.Floor<elev.Floor) && (order.Direction==elevTypes.DOWN){
		return 5*elevTypes.N_FLOORS + elev.Floor-order.Floor + 2*n_order
		
	//Order in opposite direction behind the elevator
	}else if (elev.Direction != order.Direction) && (order.Floor>elev.Floor) && (order.Direction==elevTypes.UP){
		return 8*elevTypes.N_FLOORS + order.Floor-elev.Floor + 2*n_order
	}else if (elev.Direction != order.Direction) && (order.Floor<elev.Floor) && (order.Direction==elevTypes.DOWN){
		return 8*elevTypes.N_FLOORS + elev.Floor-order.Floor + 2*n_order
		
	//Order in same direction behind the elevator
	}else if (elev.Direction == order.Direction) && (order.Floor<elev.Floor) && (order.Direction==elevTypes.UP){
		return 11*elevTypes.N_FLOORS + elev.Floor-order.Floor + 2*n_order
	}else if  (elev.Direction == order.Direction) && (order.Floor>elev.Floor) && (order.Direction==elevTypes.DOWN){
		return 11*elevTypes.N_FLOORS + order.Floor-elev.Floor + 2*n_order
	}
	//default 
	return 255
}

func countOrders(queue [elevTypes.N_FLOORS][elevTypes.N_DIR]bool) int{
	count := 0
	for floor:=0; floor< elevTypes.N_FLOORS; floor++{
		for dir:=0; dir< elevTypes.N_DIR; dir++{
			if	queue[floor][dir] == true{
				count ++
			}
		}
	}
	return count
}
