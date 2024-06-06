package migrate

import (
	"os"
	"path/filepath"
)

func createFile(path string) {
	absolutepath, _ := filepath.Abs(path)
    // Create directories recursively
    dir := filepath.Dir(absolutepath)
    if err := os.MkdirAll(dir, os.ModePerm); err != nil {
        // Handle error
        panic(err)
    }

    // Check if file exists if not create it
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		file.Close()
	}
}

func saveFile(path string, content string) {
    createFile(path)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString(content)
}