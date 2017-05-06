package view

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jcaberio/go-cimd/util"
	"github.com/spf13/viper"
)

const (
	writeWait = 60 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	DRCountChan = make(chan uint64)
	SMCountChan = make(chan uint64)
	homeTempl   = template.Must(template.New("").Parse(homeHTML))
	upgrader    = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type MTCount struct {
	DRCount uint64
	SMCount uint64
}

func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(1024)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func writer(ws *websocket.Conn) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case drCount := <-DRCountChan:
			log.Println("CURRENT DR DOUNT: ", drCount)
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteJSON(MTCount{DRCount: drCount, SMCount: util.SubmitCount}); err != nil {
				return
			}
		case smCount := <-SMCountChan:
			log.Println("CURRENT SM DOUNT: ", smCount)
			if err := ws.WriteJSON(MTCount{DRCount: util.DeliveryCount, SMCount: smCount}); err != nil {
				return
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
	go writer(ws)
	reader(ws)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var v = struct {
		Host    string
		DRCount uint64
		SMCount uint64
	}{
		r.Host,
		util.DeliveryCount,
		util.SubmitCount,
	}
	homeTempl.Execute(w, &v)
}

func injectMO(w http.ResponseWriter, r *http.Request) {
	moMsg := r.PostFormValue("mo_message")
	util.MoMsgChan <- moMsg
}

func Render() {
	addr := fmt.Sprint(":", viper.GetInt("http_port"))
	r := mux.NewRouter()
	r.HandleFunc("/", serveHome)
	r.HandleFunc("/ws", serveWs)
	r.HandleFunc("/mo", injectMO)
	log.Fatal(http.ListenAndServe(addr, r))
}

const homeHTML = `<!DOCTYPE html>
<html lang="en">
    <head>
        <title>CIMD</title>
        <meta charset="utf-8">
        <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css">
        <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap-theme.min.css">
        <script src="https://ajax.googleapis.com/ajax/libs/jquery/3.2.1/jquery.min.js"></script>
        <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js"></script>
    </head>
    <body><div class="container">
        <h3>MT</h3>
        <ul class="list-group">
            <li class="list-group-item">
                <span id="sm_count" class="badge">{{.SMCount}}</span>
                Submitted Messages
            </li>
            <li class="list-group-item">
                <span id="dr_count" class="badge">{{.DRCount}}</span>
                Delivered Messages
            </li>
        </ul>
        <h3>MO</h3>
        <div class="form-group">
	    <form method="post" action="/mo" id="mo">
	        <textarea name="mo_message" class="form-control" cols="35" wrap="soft"></textarea><br>
                <button type="submit" class="btn btn-primary">Send</button>
            </form>

        </div>
        <script type="text/javascript">
            (function() {
                var dr = document.getElementById("dr_count");
                var sm = document.getElementById("sm_count");
                var conn = new WebSocket("ws://{{.Host}}/ws");
                conn.onclose = function(evt) {
                    alert("Connection closed");
                }
                conn.onmessage = function(evt) {
                    dr.textContent = JSON.parse(evt.data).DRCount;
                    sm.textContent = JSON.parse(evt.data).SMCount;
                }
                $('#mo').submit(function(e){
                    e.preventDefault();
                    $.ajax({
                        url:"/mo",
                        type:"post",
                        data:$('#mo').serialize(),
                        success:function(){
                            alert("Success");
                        }
                    });
                });
            })();
        </script>
    </div></body>
</html>`
