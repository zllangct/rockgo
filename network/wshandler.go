package network

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zllangct/RockGO/logger"
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
	router:=gin.Default()
	router.GET("/", serveHome)
	router.GET("/ws", func(ctx *gin.Context) {
		conn,err:=upGrader.Upgrade(ctx.Writer,ctx.Request,nil)
		if err!=nil {
			ctx.Writer.WriteString("server internal error")
			return
		}
		logger.Debug("TCP accept:", conn.RemoteAddr())
		atomic.AddInt32(&h.acceptNum, 1)
		h.recv(conn)
		atomic.AddInt32(&h.acceptNum, -1)
	})

	err := router.Run(cfg.Address)
	if err != nil {
		logger.Fatal("ListenAndServe: ", err)
	}
	return err
}

func (h *websocketHandler) Handle() error {
	return nil
}

func (h *websocketHandler) recv(conn *websocket.Conn) {
	defer conn.Close()
	var workerID = int32(-1)
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
			if !ok {
				logger.Error("Failed to set context with client ip")
			}
			ok = current.SetClientPortWithContext(ctx, ipPort[1])
			if !ok {
				logger.Error("Failed to set context with client port")
			}
			h.ts.invoke(ctx, message)
		}

		cfg := h.conf
		if cfg.MaxInvoke > 0 { // use goroutine pool
			if h.gpool == nil {
				h.gpool = gpool.NewPool(int(cfg.MaxInvoke), cfg.QueueCap)
			}
			job:=h.gpool.JobPool.Get().(*gpool.Job)
			job.WorkerID =atomic.LoadInt32(&workerID)
			job.Job = handler
			job.Callback= func(w int32){
				atomic.StoreInt32(&workerID,w)
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
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//http.ServeFile(w, r, "home.html")
}