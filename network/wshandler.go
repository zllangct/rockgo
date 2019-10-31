package network

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils/UUID"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsConn struct {
	locker sync.Mutex
	wsConn *websocket.Conn
}

func (this *WsConn) Addr() string {
	return this.wsConn.RemoteAddr().String()
}

func (this *WsConn) WriteMessage(messageType uint32, data []byte) error {
	msg := make([]byte, 4)
	msg = append(msg, data...)
	binary.BigEndian.PutUint32(msg[:4], messageType)
	this.locker.Lock()
	defer this.locker.Unlock()

	return this.wsConn.WriteMessage(2, msg)
}

func (this *WsConn) Close() error {
	return this.wsConn.Close()
}

type websocketHandler struct {
	conf      *ServerConf
	ts        *Server
	numInvoke int32
	acceptNum int32
	invokeNum int32
	idleTime  time.Time
	gpool     *Pool
}

func (h *websocketHandler) Listen() error {
	conf := h.conf
	//对象池模式下，初始pool大小为20
	if conf.PoolMode && conf.MaxInvoke == 0 {
		conf.MaxInvoke = 20
	}
	h.gpool = GetGlobalPool(int(conf.MaxInvoke), conf.QueueCap)

	gin.SetMode(gin.ReleaseMode)
	//router:=gin.Default()
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/", serveHome)
	router.GET("/ws", func(ctx *gin.Context) {
		conn, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			_, _ = ctx.Writer.WriteString("server internal error")
			return
		}
		logger.Debug("TCP accept:", conn.RemoteAddr())
		atomic.AddInt32(&h.acceptNum, 1)
		sess := &Session{
			ID:         UUID.Next(),
			properties: make(map[string]interface{}),
			conn:       &WsConn{wsConn: conn},
		}
		if h.conf.OnClientConnected != nil {
			//TODO 异常处理
			h.conf.OnClientConnected(sess)
		}
		h.recv(sess, conn)
		sess.locker.Lock()
		sess.conn = nil
		sess.locker.Unlock()
		if h.conf.OnClientDisconnected != nil {
			h.conf.OnClientDisconnected(sess)
		}
		atomic.AddInt32(&h.acceptNum, -1)
		if h.gpool != nil {
			h.gpool.Release()
		}
	})

	go func() {
		logger.Info(fmt.Sprintf("Websocket server listening and serving HTTP on [ %s ]", conf.Address))
		err := router.Run(conf.Address)
		if err != nil {
			logger.Fatal("ListenAndServe: ", err)
		}
	}()
	return nil
}

func (h *websocketHandler) Handle() error {
	return nil
}

func (h *websocketHandler) recv(sess *Session, conn *websocket.Conn) {
	defer conn.Close()

	sess.SetProperty("workerID", WORKER_ID_RANDOM)

	handler := func(poolCtx []interface{},args ...interface{}) {
		if poolCtx != nil && len(poolCtx)>0 {
			args[0].(*Session).SetProperty("workerID", poolCtx[0].(int32))
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, "cid", args[0])
		if h.conf.Handler != nil {
			h.conf.Handler(args[0].(*Session), args[1].([]byte))
		} else {
			mid, mes := h.conf.PackageProtocol.ParseMessage(ctx, args[1].([]byte))
			if h.conf.NetAPI != nil && mid != nil {
				h.ts.invoke(ctx, mid[0], mes)
			} else {
				logger.Error("no message handler")
				return
			}
		}
	}
	for !h.ts.isClosed {
		_, pkg, err := conn.ReadMessage()
		if err != nil || pkg == nil {
			logger.Error(fmt.Sprintf("Close connection %s: %v", h.conf.Address, err))
			return
		}
		// use goroutine pool
		if h.conf.PoolMode {
			var wid int32
			var ok bool
			m,PropertyOk:=sess.GetProperty("workerID")
			if wid,ok=m.(int32);!PropertyOk || !ok{
				wid = WORKER_ID_RANDOM
			}
			h.gpool.AddJobFixed(handler, []interface{}{sess, pkg}, wid)
		} else {
			go handler(nil,sess, pkg)
		}
	}
}

func serveHome(ctx *gin.Context) {
	r := ctx.Request
	w := ctx.Writer
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
