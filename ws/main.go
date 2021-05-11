package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/robfig/cron"
	"github.com/thisissc/awsclient"
	"github.com/thisissc/chattingroom"
	"github.com/thisissc/config"
	"github.com/thisissc/rest"
	"github.com/thisissc/timezone"
	"github.com/thisissc/uuid"
)

var (
	globalTaskId = uuid.UUID4()

	cronObj = cron.NewWithLocation(timezone.PRC)
)

func main() {
	app := rest.NewApplication()
	config.SetConfigFile(app.ConfigFile)
	config.LoadConfig("AWS", &awsclient.Config{})

	hub := chattingroom.GetHub()

	app.Router.HandleFunc("/ws/{class:lectures|lives}/{id}", func(w http.ResponseWriter, r *http.Request) {
		//vars := mux.Vars(r)
		//lectureId := vars["id"]
		// TODO: auth

		room := r.URL.Path
		hub.Handle(room, w, r)
	})

	app.AddRoute("/api/v1/hub/broadcast", &BroadcastMessageHandler{}, "POST")

	// Save clients count every minute
	go func() {
		cronObj.AddFunc("@every 1m", func() {
			for roomName, conns := range hub.Rooms() {
				count := len(conns)

				// TODO: log online-count
				log.Println(globalTaskId, roomName, count)

			}
		})
		cronObj.Start()
	}()

	go hub.Run()

	app.Run()
	cronObj.Stop()
}

type BroadcastMessageHandler struct {
	rest.BaseHandler
}

func (h *BroadcastMessageHandler) Handle() error {
	room := h.R.FormValue("room")

	data, _ := ioutil.ReadAll(h.R.Body)
	defer h.R.Body.Close()

	msg := chattingroom.NewMessage(room, data)
	hub := chattingroom.GetHub()
	hub.Broadcast(*msg)

	return h.RenderMsg(201, 201, "OK")
}
