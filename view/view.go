package view

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
	"sync/atomic"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jcaberio/go-cimd/util"
	"github.com/spf13/viper"
	"github.com/jcaberio/go-cimd/cimd"
	"encoding/json"
)

const (
	writeWait = 60 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	homeTempl   = template.Must(template.New("").Parse(homeHTML))
	upgrader    = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type MTCount struct {
	TpsCount int64
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
	writeTicker := time.NewTicker(time.Millisecond * 1000)
	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case  <- writeTicker.C:
			if err := ws.WriteJSON(MTCount{TpsCount: atomic.LoadInt64(&util.TpsCount), SMCount: atomic.LoadUint64(&util.SubmitCount)}); err != nil {
				return
			}
			atomic.StoreInt64(&util.TpsCount, 0)
			msgListStr := cimd.RedisClient.LRange(cimd.MsgList, 0, -1).Val()
			msgList := make([]util.Message, 0)
			var msgObj util.Message
			for _, msgStr := range msgListStr {
				json.Unmarshal([]byte(msgStr), &msgObj)
				msgList = append(msgList, msgObj)
			}
			if err := ws.WriteJSON(struct {
				Messages []util.Message
			}{Messages: msgList}); err != nil {
				log.Println("Error in writing latest messages to web socket")
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

	// var v = struct {
	// 	Host    string
	// 	DRCount uint64
	// 	SMCount uint64
	// }{
	// 	r.Host,
	// 	util.DeliveryCount,
	// 	util.SubmitCount,
	// }
	homeTempl.Execute(w, nil)
}

func injectMO(w http.ResponseWriter, r *http.Request) {
	moMsg := r.PostFormValue("mo_message")
	select {
	case util.MoMsgChan <- moMsg:
		fmt.Println("SENT: ", moMsg)
	default:
		fmt.Println("NOT SENT: ", moMsg)
	}

}

func reset(w http.ResponseWriter, r *http.Request) {
	atomic.StoreUint64(&util.SubmitCount, 0)
}

func Render() {
	addr := fmt.Sprint("0.0.0.0:", viper.GetInt("http_port"))
	r := mux.NewRouter()
	r.HandleFunc("/", serveHome)
	r.HandleFunc("/ws", serveWs)
	r.HandleFunc("/mo", injectMO)
	r.HandleFunc("/reset", reset)
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
                <span id="sm_count" class="badge">0</span>
                Submitted Messages
            </li>
			<li class="list-group-item">
                <span id="tps" class="badge">0</span>
                Request-per-second
            </li>
		</ul>
		<div class="form-group">
	    <form method="post" action="/reset" id="reset">
                <button type="submit" class="btn btn-warning">Reset</button>
        </form>
        </div>

    <h4>Latest Messages</h4>	
	<table id="arrmsg" class="table table-hover">
		<thead>
		<tr>
		<th>source</th>
		<th>destination</th>
		<th>message</th>
		</tr>
		</thead>

		<tbody></tbody>
	</table>

        <h3>MO</h3>
        <div class="form-group">
	    <form method="post" action="/mo" id="mo">
	        <textarea name="mo_message" class="form-control" cols="35" wrap="soft"></textarea><br>
                <button type="submit" class="btn btn-primary">Send</button>
            </form>

        </div>
        <script type="text/javascript">
            (function() {
							var sm = document.getElementById("sm_count");
							var tps =document.getElementById("tps");
							var conn = new WebSocket("ws://" + window.location.host + "/ws");
							conn.onclose = function(evt) {
								alert("Connection closed");
							}
							conn.onmessage = function(evt) {
								console.log(evt.data)
								sm.textContent = JSON.parse(evt.data).SMCount;
								tps.textContent = JSON.parse(evt.data).TpsCount;
								var elems = JSON.parse(evt.data).Messages;		
								if(elems !== undefined && elems.length > 0) {
									var tbody = $('#arrmsg tbody');
									tbody.empty();
									var props = ["sender", "recipient", "message"];
    			$.each(elems, function(i, elem) {
      				var tr = $('<tr>');
      				$.each(props, function(i, prop) {
        				$('<td>').html(elem[prop]).appendTo(tr);
      				});
      				tbody.append(tr);
    			});
			}
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

							$('#reset').submit(function(e){
								e.preventDefault();
								$.ajax({
									url:"/reset",
									type:"post",
									success:function(){
										alert("Reset Submitted Messages");
									}
								});
							});

            })();
        </script>
    </div></body>
</html>`
