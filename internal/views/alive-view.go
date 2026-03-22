package views

import (
	"fmt"
	"net/http"
)

// Представление для проверки работоспособности сервиса.
func CheckAliveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Alive")
}
