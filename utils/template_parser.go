package utils

import (
	"strings"
	"errors"
	"os"
	"bufio"
)

func RequestTemplateParse(hwTmplPath string, key string, delimStart string, delimEnd string) (string, error) {
	template := ""
	openDelim := true

	hwTmplFile, err := os.Open(hwTmplPath)
	if err != nil {
		return "", err
	}
	defer hwTmplFile.Close()

	scanner := bufio.NewScanner(hwTmplFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, delimStart) && strings.Contains(line, key) {
			for scanner.Scan() {
				line = scanner.Text()
				if strings.HasPrefix(line, delimEnd) {
					openDelim = false
					break
				}
				if strings.HasSuffix(line, "\\n") {
					line = strings.Replace(line, "\\n", "\n", 1)
				}
				template += line
			}
			break
		}
	}

	if openDelim {
		return "", errors.New("Open delimiter")
	}

	return template, err;
}
