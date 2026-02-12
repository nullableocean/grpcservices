package configmanager

import (
	"bufio"
	"os"
	"strings"
)

// если filename пустой ("") парсим переменные окружения ОС с помощью os.Environ()
func GetEnvLoader(filename string) Loader {
	return func() (Config, error) {
		// []string "key=val"
		var envs []string

		if filename == "" {
			envs = os.Environ()
		} else {
			f, err := os.Open(filename)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			scan := bufio.NewScanner(f)
			scan.Split(bufio.ScanLines)

			for scan.Scan() {
				envs = append(envs, scan.Text())
			}
		}

		cnf := make(Config, len(envs))
		for _, env := range envs {
			parts := strings.SplitN(env, "=", 2)
			cnf[parts[0]] = parts[1]
		}

		return cnf, nil
	}
}
