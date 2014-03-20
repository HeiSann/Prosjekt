package comsManager

import("elevTypes"
	"strconv")


func constructUpdateMsg(myIp string, order elevTypes.Order_t, actionElev string)elevTypes.Message{
	msg:=elevTypes.Message{}
	msg.From=myIp
	msg.Type="UPDATE_BACKUP"
	msg.Payload = actionElev
	msg.Order = order	 
	return msg
}

func constructNewOrderMsg(ToIpadr string, myIp string, order elevTypes.Order_t)elevTypes.Message{
	msg:=elevTypes.Message{} 
	msg.To=ToIpadr
	msg.From = myIp
	msg.Type = "ADD_ORDER" 
   	msg.Order= order
	return msg
}



func constructCostMsg(myIp string, toIp string, order elevTypes.Order_t, cost int)elevTypes.Message{
	msg:=elevTypes.Message{}
	msg.From = myIp
	msg.To = toIp
	msg.Type="COST"
	msg.Payload = strconv.Itoa(cost)
	msg.Order = order	 
	return msg
}

func constructNeedCostMsg(myIP string, order elevTypes.Order_t)elevTypes.Message{
	msg:=elevTypes.Message{}
	msg.From = myIP
	msg.Type = "NEED_COST"
	msg.Order = order
	return msg
}
