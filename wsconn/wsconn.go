package wsconn

import (
	"fmt"
	"github.com/gorilla/websocket"
	// "log"
	"net"
	"net/http"
	"net/url"
)

type Responce struct {
	Result  interface{} `json:"result"`
	JsonRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
}

type JsonResponce struct {
	Result []JsonResult `json:"result"`
}

type JsonResult struct {
	Param             float64 `json:"acd"`
	SupplierAccountID int     `json:"supplier_account_id"`
	FromDate          string  `json:"from_date"`
	ToDate            string  `json:"to_date"`
	CountAll          float64 `json:"count_all"`
	Minutes           float64 `json:"minutes"`
}

func CreateConn(remoteAddr, ustr string) (*websocket.Conn, error) {
	u, err := url.Parse(fmt.Sprintf("%s/websocket/%s", remoteAddr, ustr))
	if err != nil {
		return nil, err
	}

	rawConn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return nil, err
	}

	wsHeaders := http.Header{
		"Origin":                   {remoteAddr},
		"Sec-WebSocket-Extensions": {"permessage-deflate; client_max_window_bits, x-webkit-deflate-frame"},
	}

	wsConn, resp, err := websocket.NewClient(rawConn, u, wsHeaders, 1024, 1024)
	if err != nil {
		return nil, fmt.Errorf("websocket.NewClient Error: %s\nResp:%+v", err, resp)
	}
	return wsConn, nil
}

func FuncCall(remoteAddr, fpath string, msgType int, msg []byte, callsChan chan []float64) {
	wsConn, err := CreateConn(remoteAddr, fpath)
	if err != nil {
		panic(err)
	}
	defer wsConn.Close()
	go func() {
		err = wsConn.WriteMessage(msgType, msg)
		if err != nil {
			panic(err)
		}
	}()
	readDone := make(chan string)
	var resp JsonResponce
	go func() {
		err := wsConn.ReadJSON(&resp)
		if err != nil {
			panic(err)
		}
		readDone <- "json parsed"
	}()
	<-readDone
	var calls_list []float64
	for _, result := range resp.Result {
		calls_list = append(calls_list, result.CountAll)
	}
	callsChan <- calls_list
}

func ReceiveMsg(remoteAddr, fpath string, msgType int, msg []byte, answerChan chan []byte) {
	wsConn, err := CreateConn(remoteAddr, fpath)
	if err != nil {
		panic(err)
	}
	defer wsConn.Close()
	go func() {
		err = wsConn.WriteMessage(msgType, msg)
		if err != nil {
			panic(err)
		}
	}()
	_, answerMsg, err := wsConn.ReadMessage()
	if err != nil {
		panic(err)
	}
	answerChan <- answerMsg
}
