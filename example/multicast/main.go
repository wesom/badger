package main

import (
	"log"
	"net/http"

	"badger"

	"go.uber.org/zap"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return
	}
	gate := badger.NewWsGateWay(
		badger.WithLogger(logger),
	)
	defer gate.Close()

	gate.OnTextMessage(func(conn *badger.Connection, data []byte) {
		logger.Info("OnTextMessage: ", zap.Uint64("connID", conn.ConnID()), zap.String("data", string(data)))
		conn.WriteTextMessage(data)
		// conn.Close()
	})
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHome)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		gate.ServeHTTP(w, r)
	})
	httpsrv := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}
	err = httpsrv.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
