package main

import (
	"ccscharts/chart"
	"ccscharts/wsconn"
)

func main() {
	r := wsconn.Request{
		Params: wsconn.TotalStatParams{
			FromDate:          "2014-12-01T00:00:00",
			ToDate:            "2014-12-03T00:00:00",
			SupplierAccountID: 1663,
		},
		JsonRPC: "2.0",
		Method:  "call_total",
	}
	pointsValue := wsconn.FuncCall("http://192.168.0.231:8888", "stat/stat", &r)
	ch := chart.New()
	ch.CreateCurrentLine(pointsValue)
	ch.CreatePredictLine(pointsValue)
	<-ch.LineDone
	if err := ch.Plot.Save(6, 3, "points.png"); err != nil {
		panic(err)
	}
}
