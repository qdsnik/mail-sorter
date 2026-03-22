package main

import (
	"fmt"
	"log"
	"net/http"

	"vgb2-mail-service/internal/views"
)

func main() {

	http.HandleFunc("/checkAlive", views.CheckAliveHandler)

	fmt.Println("Веб‑сервер запущен на http://127.0.0.1:6566")
	log.Fatal(http.ListenAndServe(":6566", nil))
}
