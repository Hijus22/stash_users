package usrutils

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

// Retrieves the credentials for Bitbucket and Crowd encoded in Base64
func GetCredentials(busr string, bpass string, cusr string, cpass string, file string) (encbit string, enccrowd string) {
	reader := bufio.NewReader(os.Stdin)

	bitbucket, crowd := GetCredentialsFromFile(file)

	var aux []string

	//	Check if the usr:password retrieved from bitbucket is correct and asign it if no one was provided
	//	as a command line argument
	if strings.Contains(bitbucket, ":") {
		aux = strings.Split(bitbucket, ":")

		if busr == "" {
			busr = aux[0]
		}
		if bpass == "" {
			bpass = aux[1]
		}
	}

	//	Check if the usr:password retrieved from crowd is correct and asign it if no one was provided
	//	as a command line argument
	if strings.Contains(crowd, ":") {
		aux = strings.Split(crowd, ":")

		if cusr == "" {
			cusr = aux[0]
		}
		if cpass == "" {
			cpass = aux[1]
		}
	}

	// If user:password for bitbucket is still missing, then request it through command line
	if busr == "" {
		fmt.Printf("Enter Bitbucket username: ")
		usr, _ := reader.ReadString('\n')
		busr = usr
	}

	if bpass == "" {
		fmt.Printf("Enter Bitbucket password: ")
		bytes, err := gopass.GetPasswd()

		if err != nil {
			fmt.Println("An error occurred with the password...")
			fmt.Println(err)
			os.Exit(1)
		}
		bpass = string(bytes)
	}

	bitbucketBytes := []byte(fmt.Sprintf("%s:%s", strings.TrimSpace(busr), strings.TrimSpace(bpass)))
	encbit = base64.StdEncoding.EncodeToString(bitbucketBytes)

	// If user:password for bitbucket is still missing, then request it through command line
	if cusr == "" {
		fmt.Printf("Enter Crowd username: ")
		usr, _ := reader.ReadString('\n')
		cusr = usr
	}

	if cpass == "" {
		fmt.Printf("Enter Crowd password: ")
		bytes, err := gopass.GetPasswd()

		if err != nil {
			fmt.Println("An error occurred with the password...")
			fmt.Println(err)
			os.Exit(1)
		}
		cpass = string(bytes)
	}

	crowdBytes := []byte(fmt.Sprintf("%s:%s", strings.TrimSpace(cusr), strings.TrimSpace(cpass)))
	enccrowd = base64.StdEncoding.EncodeToString(crowdBytes)

	return encbit, enccrowd
}

// Retrieves credetials from a file and returns plain user:password
func GetCredentialsFromFile(file string) (bitbucket string, crowd string) {
	f, err := os.Open(file)
	if err != nil {
		return "", ""
	}
	defer f.Close()

	bitbucket, crowd = "", ""

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		key, value := line[0], line[1]
		if key == "Bitbucket" || key == "bitbucket" {
			bitbucket = value
		} else if key == "Crowd" || key == "crowd" {
			crowd = value
		}
	}

	return bitbucket, crowd
}

// Sets up the different logging levels
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
