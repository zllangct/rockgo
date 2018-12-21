package network

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils/UUID"
	"github.com/zllangct/RockGO/utils/current"
	"github.com/zllangct/RockGO/utils/gpool"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func (r *http.Request) bool {
		return true
	},
}

type WsConn struct {
	wsConn *websocket.Conn
}

func (this *WsConn)Addr()string  {
	return this.wsConn.RemoteAddr().String()
}

func (this *WsConn)WriteMessage(messageType uint32, data []byte) error{
	msg := make([]byte, 4)
	msg = append(msg, data...)
	binary.BigEndian.PutUint32(msg[:4], messageType)
	return this.wsConn.WriteMessage(2,msg)
}

func (this *WsConn)Close() error {
	return this.wsConn.Close()
}

type websocketHandler struct {
	conf *ServerConf
	ts   *Server
	numInvoke 	int32
	acceptNum   int32
	invokeNum   int32
	idleTime    time.Time
	gpool       *gpool.Pool
}

func (h *websocketHandler) Listen() error {
	cfg := h.conf
	gin.SetMode(gin.ReleaseMode)
	//router:=gin.Default()
	router:=gin.New()
	router.Use(gin.Recovery())
	router.GET("/", serveHome)
	router.GET("/ws", func(ctx *gin.Context) {
		conn,err:=upGrader.Upgrade(ctx.Writer,ctx.Request,nil)
		if err!=nil {
			_,_=ctx.Writer.WriteString("server internal error")
			return
		}
		logger.Debug("TCP accept:", conn.RemoteAddr())
		atomic.AddInt32(&h.acceptNum, 1)
		sess:=&Session{
			ID:UUID.Next(),
			properties:make( map[string]interface{}),
			conn:&WsConn{wsConn:conn},
		}
		if h.conf.OnClientConnected!=nil {
			h.conf.OnClientConnected(sess)
		}
		h.recv(sess, conn)
		if h.conf.OnClientDisconnected!=nil{
			h.conf.OnClientDisconnected(sess)
		}
		atomic.AddInt32(&h.acceptNum, -1)
	})

	go func() {
		logger.Info(fmt.Sprintf("Websocket server listening and serving HTTP on [ %s ]", cfg.Address))
		err := router.Run(cfg.Address)
		if err != nil {
			logger.Fatal("ListenAndServe: ", err)
		}
	}()

	return nil
}

func (h *websocketHandler) Handle() error {
	return nil
}

func (h *websocketHandler) recv(sess *Session,conn *websocket.Conn) {
	defer conn.Close()

	sess.SetProperty("workerID",-1)

	for !h.ts.isClosed {
		_, message, err := conn.ReadMessage()
		if err != nil {
			logger.Error(fmt.Sprintf("Close connection %s: %v", h.conf.Address, err))
			return
		}
		handler := func() {
			ctx := context.Background()
			remoteAddr := conn.RemoteAddr().String()
			ipPort := strings.Split(remoteAddr, ":")
			ctx = current.ContextWithCurrent(ctx)
			ok := current.SetClientIPWithContext(ctx, ipPort[0])
			ctx =context.WithValue(ctx,"sess",sess)
			if !ok {
				logger.Error("Failed to set context with client ip")
			}
			ok = current.SetClientPortWithContext(ctx, ipPort[1])
			if !ok {
				logger.Error("Failed to set context with client port")
			}
			mid,data:= h.conf.PackageProtocol.ParseMessage(ctx,message)
			h.ts.invoke(ctx,mid[0],data)
		}

		cfg := h.conf
		if cfg.MaxInvoke > 0 { // use goroutine pool
			if h.gpool == nil {
				h.gpool = gpool.NewPool(int(cfg.MaxInvoke), cfg.QueueCap)
			}
			job:=h.gpool.JobPool.Get().(*gpool.Job)
			if id,ok:= sess.GetProperty("workerID");ok{
				job.WorkerID =id.(int32)
			}
			job.Job = handler
			job.Callback= func(w int32){
				sess.SetProperty("workerID",w)
			}
			h.gpool.JobQueue <-job
		}else{
			go handler()
		}

	}
	if h.gpool != nil {
		h.gpool.Release()
	}
}


func serveHome(ctx *gin.Context) {
	r:=ctx.Request
	w:=ctx.Writer
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Api not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//http.ServeFile(w, r, "home.html")
}