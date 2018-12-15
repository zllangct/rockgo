// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"bufio"
	"encoding/gob"
	"errors"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/timer"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	RPC_CALL_TYPE_NONE = iota
	RPC_CALL_TYPE_NORMAL
	RPC_CALL_TYPE_WITHOUTREPLY
)

var isDebug = true

// ServerError represents an error that has been returned from
// the remote side of the RPC connection.
type ServerError string

func (e ServerError) Error() string {
	return string(e)
}

var ErrShutdown = errors.New("connection is shut down")
var ErrTimeout = errors.New("rpc timeout")
var ErrConnClosing = errors.New("conn is closing")
var ErrErrorBody = errors.New("reading error body")

// Call represents an active RPC.
type Call struct {
	ServiceMethod string      // The name of the service and method to call.
	Args          interface{} // The argument to the function (*struct).
	Reply         interface{} // The reply from the function (*struct).
	Error         error       // After completion, the error status.
	Done          chan *Call  // Strobes when call is complete.
	Type          int
}

// TcpClient represents an RPC TcpClient.
// There may be multiple outstanding Calls associated
// with a single TcpClient, and a TcpClient may be used by
// multiple goroutines simultaneously.
type TcpClient struct {
	codec ClientCodec

	network    string
	conn       net.Conn

	reqMutex sync.Mutex // protects following
	request  Request

	mutex         sync.Mutex // protects following
	seq           uint64
	pending       map[uint64]*Call
	closing       bool // user has called Close
	shutdown      bool // server has told us to stop
	CloseCallback func(event string, data ...interface{})
}

// A ClientCodec implements writing of RPC requests and
// reading of RPC responses for the client side of an RPC session.
// The client calls WriteRequest to write a request to the connection
// and calls ReadResponseHeader and ReadResponseBody in pairs
// to read responses. The client calls Close when finished with the
// connection. ReadResponseBody may be called with a nil
// argument to force the body of the response to be read and then
// discarded.
// See NewClient's comment for information about concurrent access.
type ClientCodec interface {
	WriteRequest(*Request, interface{}) error
	ReadResponseHeader(*Response) error
	ReadResponseBody(interface{}) error

	Close() error
}

func (client *TcpClient) IsClosed()bool{
	client.mutex.Lock()
	defer client.mutex.Unlock()
	return client.closing || client.shutdown
}

func (client *TcpClient) send(call *Call) {
	client.reqMutex.Lock()
	defer client.reqMutex.Unlock()

	// Register this call.
	client.mutex.Lock()
	if client.shutdown || client.closing {
		client.mutex.Unlock()
		call.Error = ErrShutdown
		call.done()
		return
	}
	seq := client.seq
	client.seq++
	client.pending[seq] = call
	client.mutex.Unlock()

	// Encode and send the request.
	client.request.Seq = seq
	client.request.ServiceMethod = call.ServiceMethod
	client.request.Type = call.Type
	err := client.codec.WriteRequest(&client.request, call.Args)
	if err != nil {
		client.mutex.Lock()
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}
func (client *TcpClient) sendWithoutReply(call *Call) {
	client.reqMutex.Lock()
	defer client.reqMutex.Unlock()

	// Register this call.
	if client.shutdown || client.closing {
		call.Error = ErrShutdown
		call.done()
		return
	}
	// Encode and send the request.
	client.request.ServiceMethod = call.ServiceMethod
	client.request.Type = call.Type
	err := client.codec.WriteRequest(&client.request, call.Args)
	if err != nil {
		if call != nil {
			call.Error = err
			call.done()
		}
	}
	if call != nil {
		call.Error = nil
		call.done()
	}
}
func (client *TcpClient) input() {
	var err error
	var response Response
	for err == nil {
		response = Response{}
		err = client.codec.ReadResponseHeader(&response)
		if err != nil {
			break
		}
		seq := response.Seq
		client.mutex.Lock()
		call := client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()

		switch {
		case call == nil:
			// We've got no pending call. That usually means that
			// WriteRequest partially failed, and call was already
			// removed; response is a server telling us about an
			// error reading request body. We should still attempt
			// to read error body, but there's no one to give it to.
			err = client.codec.ReadResponseBody(nil)
			if err != nil {
				err = errors.New("reading error body: " + err.Error())
			}
		case response.Error != "":
			// We've got an error response. Give this to the request;
			// any subsequent requests will get the ReadResponseBody
			// error if there is one.
			call.Error = ServerError(response.Error)
			err = client.codec.ReadResponseBody(nil)
			if err != nil {
				err = ErrErrorBody
			}
			call.done()
		default:
			if call.Reply == nil {
				//TODO 待测试，call调用，但无返回值
				err = client.codec.ReadResponseBody(&struct{}{})
				if err != nil {
					err = errors.New("reading error body: " + err.Error())
				}
			} else {
				err = client.codec.ReadResponseBody(call.Reply)
				if err != nil {
					call.Error = errors.New("reading body " + err.Error())
				}
			}
			call.done()
		}
	}
	// Terminate pending calls.
	client.reqMutex.Lock()
	client.mutex.Lock()
	client.shutdown = true
	closing := client.closing
	if err == io.EOF {
		if closing {
			err = ErrShutdown
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
	client.CloseCallback("close", client.conn.RemoteAddr().String())
	client.mutex.Unlock()
	client.reqMutex.Unlock()
	if debugLog && err != io.EOF && !closing {
		log.Println("rpc: client protocol error:", err)
	}
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		// We don't want to block here. It is the caller's responsibility to make
		// sure the channel has enough buffer space. See comment in Go().
		if debugLog {
			log.Println("rpc: discarding Call reply due to insufficient Done chan capacity")
		}
	}
}

func (client *TcpClient) StartHeartBeat() {
	if !DebugMode{
		pingCount := 0
		res := &HeartBeatReuslt{}
		c:=time.NewTicker(HeartInterval)
		for {
			res.Result = 0
			<-c.C
			if client.closing || client.shutdown {
				return
			}
			err := client.Call("InnerResponse.HeartBeat", struct {}{}, res)
			if err != nil || res.Result != 1 {
				pingCount++
				if pingCount > 3 && !DebugMode{
					logger.Error("tcp client timeout")
					c.Stop()
					client.Close()
				}
			} else {
				pingCount = 0
			}
		}
	}
}


func (client *TcpClient) Reconnect() error {
	if !client.closing && !client.shutdown {
		return nil
	}
	if client.closing && !client.shutdown {
		return ErrConnClosing
	}
	conn, err := net.Dial(client.conn.RemoteAddr().Network(), client.conn.RemoteAddr().String())
	if err != nil {
		return err
	}
	*client = *NewClientWithConn(conn)
	return nil
}
// NewClient returns a new TcpClient to handle requests to the
// set of services at the other end of the connection.
// It adds a buffer to the write side of the connection so
// the header and payload are sent as a unit.
//
// The read and write halves of the connection are serialized independently,
// so no interlocking is required. However each half may be accessed
// concurrently so the implementation of conn should protect against
// concurrent reads or concurrent writes.
func NewClient(conn net.Conn) *TcpClient {
	return NewClientWithConn(conn)
}

func NewTcpClient(network string, remoteAddr string, callback ...func(event string, data ...interface{})) (*TcpClient, error) {
	conn, err := net.Dial(network, remoteAddr)
	if err != nil {
		return nil, err
	}
	return NewClientWithConn(conn,callback...),nil
}

func NewClientWithConn(conn net.Conn, callback ...func(event string, data ...interface{})) *TcpClient {
	encBuf := bufio.NewWriter(conn)
	codec := &gobClientCodec{conn, gob.NewDecoder(conn), gob.NewEncoder(encBuf), encBuf}
	client := &TcpClient{
		conn:conn,
		codec:         codec,
		pending:       make(map[uint64]*Call),
	}
	go client.input()
	go client.StartHeartBeat()
	if len(callback) > 0 {
		client.CloseCallback = callback[0]
		client.CloseCallback("connected", client)
	}
	return client
}

type gobClientCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
}

func (c *gobClientCodec) WriteRequest(r *Request, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		return
	}
	if err = c.enc.Encode(body); err != nil {
		return
	}
	return c.encBuf.Flush()
}

func (c *gobClientCodec) ReadResponseHeader(r *Response) error {
	return c.dec.Decode(r)
}

func (c *gobClientCodec) ReadResponseBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *gobClientCodec) Close() error {
	c.encBuf.Flush()
	return c.rwc.Close()
}

// DialHTTP connects to an HTTP RPC server at the specified network address
// listening on the default HTTP RPC path.
func DialHTTP(network, address string) (*TcpClient, error) {
	return DialHTTPPath(network, address, DefaultRPCPath)
}

// DialHTTPPath connects to an HTTP RPC server
// at the specified network address and path.
func DialHTTPPath(network, address, path string) (*TcpClient, error) {
	var err error
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	io.WriteString(conn, "CONNECT "+path+" HTTP/1.0\n\n")

	// Require successful HTTP response
	// before switching to RPC protocol.
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == connected {
		return NewClient(conn), nil
	}
	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	conn.Close()
	return nil, &net.OpError{
		Op:   "dial-http",
		Net:  network + " " + address,
		Addr: nil,
		Err:  err,
	}
}

// Dial connects to an RPC server at the specified network address.
func Dial(network, address string, callback ...func(event string, data ...interface{})) (*TcpClient, error) {
	c, err := NewTcpClient(network, address, callback...)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Close calls the underlying codec's Close method. If the connection is already
// shutting down, ErrShutdown is returned.
func (client *TcpClient) Close() error {
	client.mutex.Lock()
	if client.closing {
		client.mutex.Unlock()
		return ErrShutdown
	}
	client.closing = true
	client.mutex.Unlock()
	err := client.codec.Close()
	if err == nil {
		//if client.CloseCallback!=nil && !client.shutdown{
		//	go client.CloseCallback("close",client)
		//}
	}
	return err
}

// Go invokes the function asynchronously. It returns the Call structure representing
// the invocation. The done channel will signal when the call is complete by returning
// the same Call object. If done is nil, Go will allocate a new channel.
// If non-nil, done must be buffered or Go will deliberately crash.
func (client *TcpClient) Go(serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call {
	call := new(Call)
	call.ServiceMethod = serviceMethod
	call.Args = args
	call.Reply = reply
	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		// If caller passes done != nil, it must arrange that
		// done has enough buffer for the number of simultaneous
		// RPCs that will be using that channel. If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}
	call.Done = done
	if reply == nil {
		call.Type = RPC_CALL_TYPE_WITHOUTREPLY
		client.sendWithoutReply(call)
	} else {
		call.Type = RPC_CALL_TYPE_NORMAL
		client.send(call)
	}
	return call
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client *TcpClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	call := client.Go(serviceMethod, args, reply, make(chan *Call, 1))
	select {
	case <-timer.After(CallTimeout):
		call.Error = ErrTimeout
		return ErrTimeout
	case <-call.Done:
		return call.Error
	}
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client *TcpClient) CallWithoutReply(serviceMethod string, args interface{}) error {
	call := <-client.Go(serviceMethod, args, nil, make(chan *Call, 1)).Done
	return call.Error
}
