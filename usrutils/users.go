package usrutils

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/cheggaaa/pb"
)

func DeactivateUsers(host string, users []string, token string, loggers *Loggers, verbose bool, simulation bool) {

	// Initialize the progress bar
	bar := pb.StartNew(len(users))
	bar.ShowBar = true

	for i, user := range users {
		data := url.Values{}
		// data.Set("name", user)

		resource := fmt.Sprintf("/group/user/direct?groupname=stash_users&username=%s", user)
		u, _ := url.ParseRequestURI(host)
		u.Path = resource
		apiUrl := u.String()

		count := fmt.Sprintf("[%d/%d]", i+1, len(users))
		if verbose {
			loggers.Trace.Println("-->", count, " Deactivating user", user, "...")
		}

		statusCode, status := UpdateUser("DELETE", apiUrl, token, data, loggers, verbose, simulation)
		if statusCode == 200 || statusCode == 204 {
			if verbose {
				if simulation {
					loggers.Trace.Println("--> << SIMULATION >> User", user, "deactivated")
				} else {
					loggers.Trace.Println("--> User", user, "deactivated")
				}
			}
		} else {
			loggers.Error.Println("An error ocurred while trying to deactivate the user:", status)
		}

		// Update the bar
		bar.Increment()
	}

	bar.FinishPrint("Done!")

}

func UpdateUser(mode string, url string, token string, body url.Values, loggers *Loggers, verbose bool, simulation bool) (statusCode int, status string) {

	if simulation {
		return 200, "Simulated call"
	} else {
		client := &http.Client{}

		req, _ := http.NewRequest(mode, url, bytes.NewBufferString(body.Encode()))
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", token))
		req.Header.Add("Accept", "application/json")

		res, err := client.Do(req)
		if err != nil {
			loggers.Error.Println(err)
		}
		defer res.Body.Close()

		return res.StatusCode, res.Status
	}
}

func SetUsers() map[string][]string {
	users := make(map[string][]string, 3)
	users[":none"] = make([]string, 0)
	users[":three_months"] = make([]string, 0)
	users[":six_months"] = make([]string, 0)

	return users
}
