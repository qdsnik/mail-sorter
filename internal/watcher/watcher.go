package watcher

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func WatchDirectory(serviceUrl string, deferredFileHandlingPath string) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Ошибка: %v\n", err)
		panic(fmt.Sprintf("Ошибка: %v\n", err))
	}
	defer watcher.Close()

	// Проверка возможности добавить наблюдение за каталогом.
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Ошибка: %v\n", err)
		panic(fmt.Sprintf("Ошибка: %v\n", err))
	}
	astra_input_path := filepath.Join(cwd, "astra", "input")
	err = watcher.Add(astra_input_path)
	if err != nil {
		log.Printf("Ошибка: %v\n", err)
		panic(fmt.Sprintf("Ошибка: %v\n", err))
	}

	log.Printf("INFO: Наблюдение за каталогом: %s\n", astra_input_path)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("ERROR: Ошибка цикла обработки событий")
				panic("Ошибка цикла обработки событий")
			}
			// Проверяем, что это создание нового файла
			if event.Op&fsnotify.Create == fsnotify.Create {
				fileInfo, err := os.Stat(event.Name)
				if err != nil {
					log.Printf("ERROR: Ошибка получения информации о файле: %v\n", err)
					continue
				}
				// Пропускаем каталоги
				if fileInfo.IsDir() {
					continue
				}
				log.Printf("INFO: Обнаружен новый файл: %s (размер: %d байт)\n", filepath.Base(event.Name), fileInfo.Size())
				reqwErr := sendAstraResponseFile(event.Name, serviceUrl)
				if reqwErr != nil {
					// Перемещаем файл во каталог для отложенной обработки.
					moveFile(event.Name, filepath.Join(deferredFileHandlingPath, filepath.Base(event.Name)))
				} else {
					// Удаляем файл, если он успешно обработан противоположной стороной.
					log.Printf("INFO: Файл: %s успешно обработан и удален.\n", event.Name)
					os.Remove(event.Name)
				}
			}
		case _, ok := <-watcher.Errors:
			if !ok {
				log.Printf("ERROR: Ошибка наблюдателя: %v\n", err)
				panic(fmt.Sprintf("Ошибка наблюдателя: %v\n", err))
			}
		}
	}
}

// Отправляет файл ответа для обработки в sysorg.
func sendAstraResponseFile(astraResponsePath string, targetURL string) error {
	// Открываем файл для чтения
	file, err := os.Open(astraResponsePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Создаём буфер для тела запроса
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Добавляем файл в форму
	part, err := writer.CreateFormFile("file", astraResponsePath)
	if err != nil {
		return err
	}

	// Копируем содержимое файла в форму
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	// Закрываем writer — завершаем формирование multipart-данных
	err = writer.Close()
	if err != nil {
		return err
	}

	filename := filepath.Base(astraResponsePath)
	extension := filepath.Ext(astraResponsePath)
	nameWithoutExt := filename[:len(filename)-len(extension)]
	targeURI, _ := url.JoinPath(targetURL, "/automation/astra/handle-response/", nameWithoutExt)
	log.Printf("INFO: [-->] POST %s", targeURI)
	request, err := http.NewRequest("POST", targeURI, body)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())

	// TODO: нужно предусмотреть отключение таймаута для режима отладки.
	client := &http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("ERROR: [-->] %v", err)
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		errorMessage := fmt.Sprintf("ERROR: [<--]: %s, %s", response.Status, response.Status)
		log.Printf(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	return nil
}

func moveFile(source, destination string) error {
	err := os.Rename(source, destination)
	if err != nil {
		log.Printf("ERROR: ошибка перемещения файла: %v", err)
		return err
	}
	log.Printf("INFO: Файл перемещён: %s → %s\n", source, destination)
	return nil
}
