package views

import (
	"log"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Представление для получения файла для отправки.
func ReceiveAstraRequestFileHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("INFO: [<--] Получен файл для отправки")
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		log.Println("ERROR: [-->] Ожидается файл до 100 MB")
		http.Error(w, "Ожидается файл до 100 МБ", http.StatusRequestEntityTooLarge)
		return
	}

	file, formHeader, err := r.FormFile("file")
	if err != nil {
		log.Printf("ERROR: [-->] Ошибка при приёме файла: %v\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("ERROR: [-->] Ошибка при приёме файла: %v\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dst, err := os.Create(filepath.Join(cwd, "astra", "output", formHeader.Filename))
	if err != nil {
		log.Printf("ERROR: [-->] Ошибка при приёме файла: %v\n", err.Error())
		http.Error(w, "Ошибка создания файла", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("ERROR: [-->] Ошибка при приёме файла: %v\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("INFO: [-->] Ok")
	fmt.Fprintf(w, "Ok")
}
