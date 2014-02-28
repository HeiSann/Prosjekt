package message


import "strings"

const SIZE = 1024

type Message struct{
	From string //ipAdr
	Msg_type string //order, deadElev, auction, connect to me
	Payload string
}

func ConstructMessage(adr string, typ string, msg string) Message{
	return Message{adr,typ,msg}
}

//not generic, could use reflect...
func Message2bytestream (m Message) []byte{
	msg := m.From +"~"+ m.Msg_type +"~"+ m.Payload
	return []byte(msg+"\x00")
}

//not generic, could use reflect..
func Bytestream2message(m []byte) Message{
	msg_string := string(m[:])
	msg_array := strings.Split(msg_string, "~")
	msg := Message{msg_array[0], msg_array[1], msg_array[2]}
	return msg
}

