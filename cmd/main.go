package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"path/filepath"
	"mail-sorter/internal/views"
	"mail-sorter/internal/watcher"
)

func isDirectoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func main() {
	var targetUrl string
	flag.StringVar(&targetUrl, "targetUrl", "", "URL сервиса Джанго")
	flag.Parse()

	if targetUrl == "" {
		fmt.Fprintln(os.Stderr, "Ошибка: флаг -targetUrl обязателен")
		fmt.Fprintln(os.Stderr, "Использование:")
		flag.Usage()
		os.Exit(1)
	}

	// Инициализация каталогов.
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	baseAstraPath := filepath.Join(cwd, "astra")

	// Путь к каталогу, для временного хранения файла в случае ошибки.
	deferredFileHandlingPath := filepath.Join(baseAstraPath, "deferred")

	isExists, _ := isDirectoryExists(baseAstraPath)
	if !isExists {
		err := os.Mkdir(baseAstraPath, 0755)
		if err != nil {
			log.Printf("ошибка создания директории %s: %v", baseAstraPath, err)
			panic(fmt.Errorf("ошибка создания директории %s: %w", baseAstraPath, err))
		}
		inputAstraPath := filepath.Join(baseAstraPath, "input")
		err = os.Mkdir(inputAstraPath, 0755)
		if err != nil {
			panic(fmt.Errorf("ошибка создания директории %s: %w", inputAstraPath, err))
		}
		outputAstraPath := filepath.Join(baseAstraPath, "output")
		err = os.Mkdir(outputAstraPath, 0755)
		if err != nil {
			panic(fmt.Errorf("ошибка создания директории %s: %w", outputAstraPath, err))
		}
		err = os.Mkdir(deferredFileHandlingPath, 0755)
		if err != nil {
			panic(fmt.Errorf("ошибка создания директории %s: %w", deferredFileHandlingPath, err))
		}
	}

	// Настройка логирования в файл.
	now := time.Now()
	logFilename := fmt.Sprintf("app %s.log", now.Format("2006-01-02_150405"))
	file, err := os.OpenFile(logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// Устанавливаем файл как вывод для логгера
	log.SetOutput(file)

	go watcher.WatchDirectory(targetUrl, deferredFileHandlingPath)
	fmt.Println(fmt.Sprintf("URL сервиса для обратной связи: %s", targetUrl))

	http.HandleFunc("/checkAlive", views.CheckAliveHandler)
	http.HandleFunc("/sendAstraRequest", views.ReceiveAstraRequestFileHandler)
	http.HandleFunc("/checkDeferredAstraRequest", views.CheckDeferredFilesHandler)

	fmt.Println("Веб‑сервер запущен на http://127.0.0.1:6566")
	log.Fatal(http.ListenAndServe(":6566", nil))
}
