package main

import (
	"ccscharts/chart"
	"ccscharts/wsconn"
	// "fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

func DrawChart(currentDayData []float64) *wsconn.Responce {
	ch := chart.New("Calls data", "Minutes", "Calls")
	// ch.CreatePreviousDayLine(prevdata, "gray")
	ch.CreateCurrentDayLine(currentDayData, "blue")
	// ch.CreatePredictLine(curentData, "yellow")
	<-ch.LineDone

	rawData := map[string]string{"img_raw_data": ch.GetRawDataImg(600, 300)}
	return &wsconn.Responce{
		Result:  rawData,
		JsonRPC: "2.0",
	}
}

func WSConnection(rw http.ResponseWriter, req *http.Request) *websocket.Conn {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ws, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		panic(err)
	}
	return ws
}

func HandleStatsRequest(rw http.ResponseWriter, req *http.Request) {
	// wspath1 := mux.Vars(req)["wspath1"]
	// wspath2 := mux.Vars(req)["wspath2"]

	ws := WSConnection(rw, req)
	defer ws.Close()
	msgType, msg, err := ws.ReadMessage()
	if err != nil {
		log.Println(err)
	}
	// answerChan := make(chan []byte)
	currentDayData := make(chan []float64)
	// go wsconn.ReceiveMsg("http://192.168.0.231:8888", fmt.Sprintf("%s/%s", wspath1, wspath2), msgType, msg, answerChan)
	// go wsconn.ReceiveMsg("http://192.168.0.231:8888", "stat/stat", msgType, msg, answerChan)
	go wsconn.FuncCall("http://192.168.0.231:8888", "stat/stat", msgType, msg, currentDayData)
	resp := DrawChart(<-currentDayData)
	// log.Println(string(<-answerChan))
	err = ws.WriteJSON(resp)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	r := mux.NewRouter().StrictSlash(true)
	w := r.PathPrefix("/websocket").Subrouter()
	// w.HandleFunc("/{wspath1}/{wspath2}", HandleStatsRequest)
	w.HandleFunc("/stat/stat", HandleStatsRequest)
	log.Println("Starting server on localhost:8889")
	err := http.ListenAndServe(":8889", r)
	if err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
}
