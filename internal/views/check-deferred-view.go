package views

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Перемещает файл из каталога А в каталог Б.
func moveFile(source, destination string) error {
	err := os.Rename(source, destination)
	if err != nil {
		log.Println("ERROR: ошибка перемещения файла: ", err)
		return err
	}
	log.Printf("INFO: Файл перемещён: %s → %s\n", source, destination)
	return nil
}

// Возвращает количество файлов в указанном каталоге.
func countFiles(path string) (int, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}
	return count, nil
}

// Представление для проверки отложенных файлов.
func CheckDeferredFilesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("INFO: [<--] Получен запрос на проверку отложенных файлов")
	// Инициализация каталогов.
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	baseAstraPath := filepath.Join(cwd, "astra")

	// Путь к каталогу, для временного хранения файла в случае ошибки.
	deferredFileHandlingPath := filepath.Join(baseAstraPath, "deferred")

	// Проверка наличия отложенных файлов.
	filesQuantity, err := countFiles(deferredFileHandlingPath)
	if err != nil {
		log.Printf("ERROR: Ошибка при проверки отложенных файлов")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Копирование в каталог "input", если есть отложенные файлы.
	if filesQuantity > 0 {
		entries, err := os.ReadDir(deferredFileHandlingPath)
		if err != nil {
			log.Printf("ERROR: Ошибка проверки каталога ./astra/deferred/")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, entry := range entries {
			srcFilePath := filepath.Join(deferredFileHandlingPath, entry.Name())
			dstFilePath := strings.Replace(srcFilePath, "/deferred/", "/input/", 1)
			moveFile(srcFilePath, dstFilePath)
		}
	} else {
		log.Printf("INFO: Каталог ./astra/deferred/ пуст")
	}
}
