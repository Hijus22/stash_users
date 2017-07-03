package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"stash/usr"
	"strings"
	"time"
)

var simulation, verbose, help bool
var logPath, host, token string

// Parses the arguments
func init() {
	const (
		uHelp    = "\tShow this help message and exit"
		uVerbose = "\tExecute in verbose mode"
		uSim     = "\tExecute in simulation mode"
		uLogPath = "\tThe log path where the loggers will write to.\n\t\t\tExample: /var/etc/log/stash"
		uHost    = "\tThe host url.\n\t\t\tExample: https://host.com"
		uToken   = "\tThe token needed for the Crowd Api.\n\t\t\tExample: 'Basic xfg23gfjsf2=='"
	)

	flag.BoolVar(&help, "help", false, uHelp)
	flag.BoolVar(&help, "h", false, uHelp+"(shorthand)")

	flag.BoolVar(&verbose, "verbose", false, uVerbose)
	flag.BoolVar(&verbose, "v", false, uVerbose+"(shorthand)")

	flag.BoolVar(&simulation, "simulation", true, uSim)      // CARE IT DEFAULTS TO TRUE!!
	flag.BoolVar(&simulation, "s", true, uSim+"(shorthand)") // CARE IT DEFAULTS TO TRUE!!

	flag.StringVar(&logPath, "log", "", uLogPath)
	flag.StringVar(&logPath, "l", "", uLogPath+"(shorthand)")

	flag.StringVar(&host, "api", "", uHost)
	flag.StringVar(&host, "a", "", uHost+"(shorthand)")

	flag.StringVar(&token, "token", "", uToken)
	flag.StringVar(&token, "t", "", uToken+"(shorthand)")

	flag.Parse()

}

// Check if the required arguments are present
func checkArgs(host string, token string, logPath string, verbose bool, help bool) {
	if len(os.Args) == 1 || help {
		fmt.Println("usage: stash_users [--help|-h] [--verbose|-v] [--simulation|-s] --api|-a HOST --token|-t TOKEN --log|-l LOGPATH\n")
		flag.PrintDefaults()

		os.Exit(0)
	}

	if host == "" || token == "" || logPath == "" {
		fmt.Println("Please provide a host, token and logPath\n")
		fmt.Println("usage: stash_users [--help|-h] [--verbose|-v] [--simulation|-s] --api|-a HOST --token|-t TOKEN --log|-l LOGPATH\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {

	checkArgs(host, token, logPath, verbose, help)

	// Sets up the different logging levels
	Info, Trace, Warning, Error, file := usr.SetLoggers(logPath, simulation)

	defer file.Close()
	defer fmt.Println("Logs available at:", file.Name())

	if simulation {
		fmt.Println(" << Simulation mode ON >>")
	}

	Info.Println("-------- START ", time.Now().Local().Format("2006-01-02_10:01"), "--------")

	isLastPage := false
	start, processed, pageSize, errors := 0, 0, 100, 0
	now := time.Now()

	// Initialize the user arrays
	users := usr.SetUsers()

	// Request the User and Password for bitbucket stash query
	cred := usr.GetCredentials()

	client := &http.Client{}

	if verbose {
		Trace.Println("Retrieving stash users")
	}

	fmt.Println("Retrieving stash users")

	for !isLastPage {
		url := fmt.Sprintf("/bitbucket/rest/api/1.0/admin/groups/more-members?context=stash-users&limit=%d&start=%d", pageSize, start)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", host, url), nil)
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", cred))

		fmt.Println("--> Requesting", pageSize, "users...")

		res, err := client.Do(req)
		if err != nil {
			Error.Println(err)
		}
		defer res.Body.Close()

		if res.StatusCode == 200 {

			if verbose {
				Trace.Println("--> Status 200 OK")
			}
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				panic(err.Error())
			}

			var r = usr.Response{}
			err = json.Unmarshal(body, &r)
			isLastPage = r.IsLastPage
			start += pageSize

			// Iterate through all the retrieved users
			for _, user := range r.Values {

				// Sets up the pretty user JSON for the output
				userJSON, err := json.MarshalIndent(user, " ", "\t")
				if err != nil {
					Error.Println(err)
				}

				// Clasify the users whether they have logged in the last X days
				if user.LastAuthenticationTimestamp == 0 {
					users[":none"] = append(users[":none"], user.Name)
					if verbose {
						Trace.Println("USER NEVER LOGGED IN: \n", string(userJSON))
					}

				} else {
					if user.LastAuthenticationTimestamp/1000 < uint64(now.AddDate(0, 0, -180).Unix()) {
						users[":six_months"] = append(users[":six_months"], user.Name)
						if verbose {
							Trace.Println("USER DIDNT LOG FOR 180 DAYS: \n", string(userJSON))
						}
					}
					if user.LastAuthenticationTimestamp/1000 < uint64(now.AddDate(0, 0, -90).Unix()) {
						users[":three_months"] = append(users[":three_months"], user.Name)
						if verbose {
							Trace.Println("USER DIDNT LOG FOR 90 DAYS: \n", string(userJSON))
						}
					}

				}

			}
			processed += r.Size
		} else {
			errors += 1
			Warning.Println("Request failed: ", res.Status)
			if errors > 5 {
				Error.Println("Too many errors with Status Code: ", res.Status)
				fmt.Println("An error ocurred, check the logs for more info")
				os.Exit(1)
			}
		}
	}
	fmt.Println("Done")

	Info.Println("Fetched a total of ", processed, " users")
	if verbose {
		fmt.Println("Fetched a total of ", processed, " users")
	}
	Info.Println("##############################")
	if verbose {
		fmt.Println(("##############################"))
	}
	Info.Println("--> Never logged in: ", len(users[":none"]))
	if verbose {
		fmt.Println("--> Never logged in [", len(users[":none"]), "]: ", strings.Join(users[":none"], ", "))
	}
	Info.Println("--> Not Logged in the last 90 days: ", len(users[":three_months"]))
	if verbose {
		fmt.Println("--> Not Logged in the last 90 days: [", len(users[":three_months"]), "]: ", strings.Join(users[":three_months"], ", "))
	}
	Info.Println("--> Not Logged in the last 180 days: ", len(users[":six_months"]))
	if verbose {
		fmt.Println("--> Not Logged in the last 180 days: [", len(users[":six_months"]), "]: ", strings.Join(users[":six_months"], ", "))
	}

	// Deactivate crowd users
	fmt.Println("Deactivating users in Crowd")
	Info.Println("Deactivating users in Crowd")

	loggers := usr.Loggers{Info, Trace, Warning, Error}

	usr.DeactivateUsers(host, users[":none"], token, &loggers, verbose, simulation)

	Info.Println("Done")

	Info.Println("-------- FINISHED ", time.Now().Local().Format("2006-01-02"), "--------")

}
