package badger

import (
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func tickWriter(logger *zap.Logger) {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws://127.0.0.1:8080/test", nil)
	if nil != err {
		logger.Sugar().Error(err)
		return
	}
	defer conn.Close()

	tC := time.After(time.Second * 10)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-tC:
			logger.Sugar().Info("trick writer")
			break loop
		case <-ticker.C:
			logger.Sugar().Info("trick C")
			err := conn.WriteMessage(websocket.BinaryMessage, []byte("I am a client"))
			if nil != err {
				logger.Sugar().Error(err)
				break loop
			}
		}
	}
	logger.Info("quit trick writer")
}

type EchoServer struct {
	logger *zap.Logger
}

func (e *EchoServer) OnConnect(c Conn) {
	e.logger.Info("new connection comming",
		zap.String("localAddr", c.LocalAddr().String()),
		zap.String("remoteAddr", c.RemoteAddr().String()),
	)
}

func (e *EchoServer) OnDisconnect(c Conn) {
	e.logger.Sugar().Infof("disconnect connID: %d", c.ConnID())
}

func (e *EchoServer) OnMessage(c Conn, data []byte) {
	e.logger.Sugar().Infof("receive data, connID: %d data: %v", c.ConnID(), string(data))
}

func TestWsServer(t *testing.T) {
	// logger := zap.NewExample()
	logger, err := zap.NewDevelopment()
	if err != nil {
		return
	}
	echo := &EchoServer{
		logger: logger,
	}
	b := NewWsServer(
		WithLogger(logger),
		WithEventHandler(echo),
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/test", b.Handle)
	httpsrv := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}

	go func() {
		for i := 0; i < 200; i++ {
			go tickWriter(logger)
		}
	}()

	httpsrv.ListenAndServe()

}
