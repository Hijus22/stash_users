package usr

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/howeyc/gopass"
)

type Loggers struct {
	Info    *log.Logger
	Trace   *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

type User struct {
	Name                        string
	EmailAddress                string
	Id                          int
	Active                      bool
	Type                        string
	deleteable                  bool
	LastAuthenticationTimestamp uint64
}

type Response struct {
	Size       int
	Limit      int
	IsLastPage bool
	Values     []User
	Start      int
}

func GetCredentials() (encoded string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter username: ")
	username, _ := reader.ReadString('\n')

	fmt.Printf("Enter password: ")
	pass, err := gopass.GetPasswd()

	if err == nil {
		fmt.Println("An error occurred with the password. Probably its Windows...")
	}
	password := string(pass)

	encoded = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",
		strings.TrimSpace(username), strings.TrimSpace(password))))

	return encoded
}

func SetLoggers(logPath string, simulation bool) (*log.Logger, *log.Logger, *log.Logger, *log.Logger, *os.File) {
	var logName string

	_ = os.MkdirAll(logPath, 0666)

	// Different log name if it is simulation
	if simulation {
		logName = fmt.Sprintf("%s/sim%s.log", logPath, time.Now().Local().Format("2006-01-02"))
	} else {
		logName = fmt.Sprintf("%s/%s.log", logPath, time.Now().Local().Format("2006-01-02"))
	}

	f, err := os.OpenFile(logName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("error opening file: %v", err)
		os.Exit(1)
	}

	Info := log.New(f, "[INFO]:  ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	Trace := log.New(f, "[TRACE]: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	Warning := log.New(f, "[WARN]:  ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	Error := log.New(f, "[ERROR]: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	return Info, Trace, Warning, Error, f
}
