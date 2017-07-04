package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hijus22/stash_users/usrutils"
)

var simulation, verbose, help bool
var logPath, chost, bhost, token, uCrowdUsr, uCrowdPass, uBitUsr, uBitPass, uCredFile string

// Parses the arguments
func init() {
	const (
		uHelp      = "\tShow this help message and exit"
		uVerbose   = "\tExecute in verbose mode"
		uSim       = "\tExecute in simulation mode"
		uLogPath   = "\tThe log path where the loggers will write to.\n\t\t\tExample: /var/etc/log/stash"
		uBHost     = "\tThe Bitbucket host url.\n\t\t\tExample: https://host.com"
		uCHost     = "\tThe Crowd host url.\n\t\t\tExample: https://host.com"
		uBitUsr    = "\tThe Bitbucket user\n\t\t\tIf missing, and no credentials file is provided, it will be requested at runtime"
		uBitPass   = "\tThe password for Bitbucket\n\t\t\tIf missing, and no credentials file is provided, it will be requested at runtime"
		uCrowdUsr  = "\tThe Crowd user\n\t\t\tIf missing, and no credentials file is provided, it will be requested at runtime"
		uCrowdPass = "\tThe password for Crowd\n\t\t\tIf missing, and no credentials file is provided, it will be requested at runtime"
		uCredFile  = "\tA file containing the credentials for Bitbucket and Crowd\n\t\t\tEvery line must have the following structure (2 columns): {Bitbucket | Crowd} user:password\n\t\t\tIf there are any parameters missing, they will be requested at runtime"
	)

	flag.BoolVar(&help, "help", false, uHelp)

	flag.BoolVar(&verbose, "v", false, uVerbose)

	flag.BoolVar(&simulation, "s", true, uSim) // CARE IT DEFAULTS TO TRUE!!

	flag.StringVar(&logPath, "log", "", uLogPath)

	flag.StringVar(&bhost, "host", "", uBHost)
	flag.StringVar(&chost, "host", "", uCHost)

	flag.StringVar(&host, "cuser", "", uCrowdUsr)
	flag.StringVar(&host, "cpass", "", uCrowdPass)

	flag.StringVar(&host, "buser", "", uBitUsr)
	flag.StringVar(&host, "bpass", "", uBitPass)

	flag.StringVar(&host, "cred", "", uCredFile)

	flag.Parse()

}

// Check if the required arguments are present
func checkArgs(bhost string, chost string, token string, logPath string, verbose bool, help bool) {
	usage := "usage: stash_users [--help|-h] [-v] [-s] --log LOGPATH\n\t\t[ [--cuser CROWD_USER] [--cpass CROWD_PASS] [--buser CROWD_USER] [--bpass CROWD_PASS] | [--cred CRED_FILE] ]"

	if len(os.Args) == 1 || help {
		fmt.Println(usage)
		flag.PrintDefaults()

		os.Exit(0)
	}

	if chost == "" || bhost == "" || token == "" || logPath == "" {
		fmt.Println("Please provide the bhost, chost, and logPath\n")
		fmt.Println(usage)
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {

	checkArgs(bhost, chost, token, logPath, verbose, help)

	// Sets up the different logging levels
	Info, Trace, Warning, Error, file := usrutils.SetLoggers(logPath, simulation)

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
	users := usrutils.SetUsers()

	// Request the User and Password for bitbucket stash query and crowd group removal
	bitbucketCred, crowdCred := usrutils.GetCredentials(uBitUsr, uBitPass, uCrowdUsr, uCrowdPass, uCredFile)

	client := &http.Client{}

	if verbose {
		Trace.Println("Retrieving stash users")
	}

	fmt.Println("Retrieving stash users")

	for !isLastPage {
		url := fmt.Sprintf("/bitbucket/rest/api/1.0/admin/groups/more-members?context=stash-users&limit=%d&start=%d", pageSize, start)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", bhost, url), nil)
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", bitbucketCred))

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

			var r = usrutils.Response{}
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

	loggers := usrutils.Loggers{Info, Trace, Warning, Error}

	usrutils.DeactivateUsers(chost, users[":none"], crowdCred, &loggers, verbose, simulation)

	Info.Println("Done")

	Info.Println("-------- FINISHED ", time.Now().Local().Format("2006-01-02"), "--------")

}
