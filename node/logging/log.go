package log

import (
	"log"
	"os"
)

var (
	Info  *log.Logger
	Error *log.Logger
)

func init() {
	file, err := os.OpenFile("InverseTransparency.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		log.Fatal(err)
	}

	Info = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
	Error = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
}
