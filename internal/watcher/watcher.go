package watcher

import (
	"bytes"
	"fmt"
	"io"
	"log"
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
		panic(fmt.Sprintf("Ошибка: %v\n", err))
	}
	defer watcher.Close()

	// Проверка возможности добавить наблюдение за каталогом.
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Ошибка: %v\n", err))
	}
	astra_input_path := filepath.Join(cwd, "astra", "input")
	err = watcher.Add(astra_input_path)
	if err != nil {
		panic(fmt.Sprintf("Ошибка: %v\n", err))
	}

	log.Printf("Наблюдение за каталогом: %s\n", astra_input_path)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				panic("Ошибка цикла обработки событий")
			}
			// Проверяем, что это создание нового файла
			if event.Op&fsnotify.Create == fsnotify.Create {
				fileInfo, err := os.Stat(event.Name)
				if err != nil {
					log.Printf("Ошибка получения информации о файле: %v\n", err)
					continue
				}
				// Пропускаем каталоги
				if fileInfo.IsDir() {
					continue
				}
				log.Printf("Обнаружен новый файл: %s (размер: %d байт)\n", filepath.Base(event.Name), fileInfo.Size())
				reqwErr := sendAstraResponseFile(event.Name, serviceUrl)
				if reqwErr != nil {
					log.Printf("Ошибка отправки файла: %v\n", err)
					// Перемещаем файл во каталог для отложенной обработки.
					moveFile(event.Name, filepath.Join(deferredFileHandlingPath, filepath.Base(event.Name)))
				} else {
					// Удаляем файл, если он успешно обработан противоположной стороной.
					os.Remove(event.Name)
				}
			}
		case _, ok := <-watcher.Errors:
			if !ok {
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

	// Создаём HTTP-запрос
	filename := filepath.Base(astraResponsePath)
	extension := filepath.Ext(astraResponsePath)
	nameWithoutExt := filename[:len(filename)-len(extension)]
	targeURI, _ := url.JoinPath(targetURL, "/handle-response/", nameWithoutExt)
	request, err := http.NewRequest("POST", targeURI, body)
	if err != nil {
		return err
	}

	// Устанавливаем заголовок Content-Type с boundary
	request.Header.Set("Content-Type", writer.FormDataContentType())

	// Выполняем запрос
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Проверяем статус ответа
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("сервер вернул ошибку: %s", response.Status)
	}

	return nil
}

func moveFile(source, destination string) error {
	err := os.Rename(source, destination)
	if err != nil {
		return fmt.Errorf("ошибка перемещения файла: %w", err)
	}
	fmt.Printf("Файл перемещён: %s → %s\n", source, destination)
	return nil
}
