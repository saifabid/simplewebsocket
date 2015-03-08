package main // change to websocket when ready to ship

import (
	"errors"
	"fmt"
	"github.com/cookieo9/go-misc/slice"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}
var connections []Socket

type Socket struct {
	Ws       *websocket.Conn
	Id       string
	TTL      *time.Timer
	deployed bool
	deleted  bool
}

func main() {
}

func DeployWsDameon() {

	for {

		for index, j := range connections {
			if j.deployed == true {
				continue
			}
			go func() {
				j.deployed = true
				<-j.TTL.C
				fmt.Println(j.Id, ": Closing")
				j.Ws.Close()
				c := slice.Delete(connections, index)

				connections = c.([]Socket)

			}()

		}

	}

}

func SetReadBuffer(num int) {

	upgrader.ReadBufferSize = num

}

func SetWriteBuffer(num int) {

	upgrader.WriteBufferSize = num

}

func WsUpgrade(w http.ResponseWriter, r *http.Request, id string, timeout time.Duration) *Socket {

	//timeout is how long you want the connection to live
	Conn, err := upgrader.Upgrade(w, r, nil)
	if timeout == 0 {
		fmt.Println("No time specified")
	} else {

		if err != nil {
			fmt.Println(err)
		} else {
			timer := time.NewTimer(timeout * time.Second)
			returnConn := &Socket{Conn, id, timer, false, false}
			connections = append(connections, *returnConn)
			return returnConn
		}
	}
	timer := time.NewTimer(1 * time.Second)
	returnConn := &Socket{Conn, id, timer, true, false} // will not delete from dameon due to timer.
	return returnConn

}

func (conn *Socket) SendText(msg string) error {
	msgB := []byte(msg)

	return conn.Ws.WriteMessage(1, msgB)

}
func (conn *Socket) SendBinary(msgB []byte) error {

	return conn.Ws.WriteMessage(1, msgB)
}

func (conn *Socket) BroadcastString(msg string) error {

	for _, users := range connections {

		e := users.SendText(msg)
		if e != nil {
			return e
		}

	}

	return nil

}

func (conn *Socket) BroadcastBinary(msg []byte) error {

	for _, users := range connections {

		e := users.SendBinary(msg)
		if e != nil {
			return e
		}

	}

	return nil

}

func (conn *Socket) SendTextToUser(id string, msg string) error {

	for _, sock := range connections {

		if sock.Id == id {

			sock.SendText(msg)
			return nil

		}

	}
	return errors.New("id not found")

}

func (conn *Socket) SentBinaryToUser(id string, msg []byte) error {

	for _, sock := range connections {

		if sock.Id == id {

			sock.SendBinary(msg)
			return nil

		}

	}
	return errors.New("id not found")

}

func (conn *Socket) ReadMessage() (int, []byte, error) {

	return conn.Ws.ReadMessage()

}
