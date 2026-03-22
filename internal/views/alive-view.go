package views

import (
	"fmt"
	"net/http"
)

func CheckAliveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Alive")
}
