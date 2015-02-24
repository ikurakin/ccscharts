package wsconn

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"net/url"
)

type Request struct {
	Params  TotalStatParams `json:"params"`
	JsonRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
}

type TotalStatParams struct {
	Param             string `json:"param,omitempty"`
	FromDate          string `json:"from_date,omitempty"`
	ToDate            string `json:"to_date,omitempty"`
	SupplierAccountID int    `json:"supplier_account_id,omitempty"`
	CustomerAccountID int    `json:"customer_account_id,omitempty"`
	SupplierPartnerID int    `json:"supplier_partner_id,omitempty"`
	CustomerPartnerID int    `json:"customer_partner_id,omitempty"`
	DestinationID     int    `json:"destination_id,omitempty"`
	EquipmentID       int    `json:"equipment_id,omitempty"`
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

func FuncCall(remoteAddr, fpath string, r *Request) []float64 {
	wsConn, err := CreateConn(remoteAddr, fpath)
	if err != nil {
		panic(err)
	}
	defer wsConn.Close()
	err = wsConn.WriteJSON(r)
	if err != nil {
		panic(err)
	}
	answerChan := make(chan string)
	callsChan := make(chan []float64)
	var resp JsonResponce
	go func() {
		err := wsConn.ReadJSON(&resp)
		if err != nil {
			panic(err)
		}
		answerChan <- "json parsed"
	}()
	go func() {
		<-answerChan
		var calls_list []float64
		for _, result := range resp.Result {
			calls_list = append(calls_list, result.CountAll)
		}
		callsChan <- calls_list
	}()

	return <-callsChan
}
