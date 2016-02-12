package client

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	printer "github.com/olekukonko/tablewriter"
	konfig "github.com/zalando-techmonkeys/chimp/conf/client"
	. "github.com/zalando-techmonkeys/chimp/types"
	"golang.org/x/crypto/ssh/terminal"
)

//Client is the struct for accessing client functionalities
type Client struct {
	Config      *konfig.ClientConfig
	AccessToken string
	Scheme      string
	Clusters    []string
}

var homeDirectories = []string{"HOME", "USERPROFILES"}

//RenewAccessToken is used to get a new Oauth2 access token
func (bc *Client) RenewAccessToken(username string) {
	if username == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your username: ")
		username, _ = reader.ReadString('\n')
	}
	fmt.Print("Enter your password: ")
	bytePassword, err := terminal.ReadPassword(0)
	fmt.Println("")
	if err != nil {
		fmt.Printf("Cannot read password\n")
		os.Exit(1)
	}
	password := strings.TrimSpace(string(bytePassword))
	u, err := url.Parse(bc.Config.OauthURL)
	if err != nil {
		fmt.Printf("ERR: Could not parse given Auth URL: %s\n", bc.Config.OauthURL)
		os.Exit(1)
	}
	authURLStr := fmt.Sprintf("https://%s%s%s%s", u.Host, u.Path, u.RawQuery, u.Fragment)
	fmt.Printf("Getting token as %s\n", username)
	client := &http.Client{}
	req, err := http.NewRequest("GET", authURLStr, nil)
	req.SetBasicAuth(username, password)
	res, err := client.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		fmt.Printf("ERR: Could not get Access Token, caused by: %s\n", err)
		os.Exit(1)
	}
	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("ERR: Can not read response body, caused by: %s\n", err)
		os.Exit(1)
	}

	if len(respBody) > 0 && res.StatusCode == 200 {
		bc.AccessToken = string(respBody)
		fmt.Printf("SUCCESS. Your access token is stored in .chimp-token in your home directory.\n")
		//store token to file
		var homeDir string
		for _, home := range homeDirectories {
			if dir := os.Getenv(home); dir != "" {
				homeDir = dir
			}
		}
		tokenFileName := fmt.Sprintf("%s/%s", homeDir, ".chimp-token")
		f, _ := os.Create(tokenFileName)
		_, _ = f.WriteString(strings.TrimSpace(bc.AccessToken)) //not important if doens't work, we'll try again next time
	} else {
		fmt.Printf("ERR: %d - %s\n", res.StatusCode, respBody)
	}
}

//GetAccessToken sets the access token inside the request
func (bc *Client) GetAccessToken(username string) {
	if bc.Config.Oauth2Enabled {
		//before trying to get the token I try to read the old one
		var homeDir string
		for _, home := range homeDirectories {
			if dir := os.Getenv(home); dir != "" {
				homeDir = dir
			}
		}
		tokenFileName := fmt.Sprintf("%s/%s", homeDir, ".chimp-token")
		data, err := ioutil.ReadFile(tokenFileName)
		var oldToken string
		if err != nil {
			fmt.Println("ERR: Could not get an AccessToken which is required. Please login again.")
			os.Exit(1)
		} else {
			oldToken = strings.TrimSpace(string(data))
		}
		bc.AccessToken = oldToken
	}
}

func (bc *Client) buildDeploymentURL(name string, params map[string]string, cluster string) string {
	u := new(url.URL)
	u.Scheme = bc.Scheme
	host := bc.Config.Clusters[cluster].IP
	port := bc.Config.Clusters[cluster].Port
	u.Host = net.JoinHostPort(host, strconv.Itoa(port))
	if bc.Scheme == "https" && port == 443 {
		u.Host = host
	}
	u.Path = path.Join("/deployments", name)
	q := u.Query()
	for k := range params {
		q.Set(k, params[k])
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (bc *Client) buildDeploymentReplicasURL(name string, replicas int, cluster string, force bool) string {
	u := new(url.URL)
	u.Scheme = bc.Scheme
	host := bc.Config.Clusters[cluster].IP
	port := bc.Config.Clusters[cluster].Port
	u.Host = net.JoinHostPort(host, strconv.Itoa(port))
	if bc.Scheme == "https" && port == 443 {
		u.Host = host
	}
	q := u.Query()
	q.Set("force", strconv.FormatBool(force))
	u.RawQuery = q.Encode()
	u.Path = path.Join("/deployments", name, "replicas", strconv.Itoa(replicas))
	return u.String()
}

//DeleteDeploy is used to delete a deployment from the cluster/server
func (bc *Client) DeleteDeploy(name string) {
	for _, clusterName := range bc.Clusters {
		url := bc.buildDeploymentURL(name, nil, clusterName)
		_, res, err := bc.makeRequest("DELETE", url, nil)
		if res != nil {
			defer res.Body.Close()
		}
		if err != nil {
			fmt.Println(errorMessageBuilder("Cannot delete deployment", err))
			continue
		}
		if checkStatusOK(res.StatusCode) {
			if checkAuthOK(res.StatusCode) {
				if res.StatusCode >= 400 && res.StatusCode <= 499 {
					e := Error{}
					unmarshalResponse(res, &e)
					fmt.Printf("Cannot delete deployment: %s\n", e.Err)
				} else {
					fmt.Println("Delete operation successful")
				}
			} else {
				handleAuthNOK(res.StatusCode)
			}
		} else {
			handleStatusNOK(res.StatusCode)
		}
	}
}

//InfoDeploy is used the get the information for a currently running deployment
func (bc *Client) InfoDeploy(name string, verbose bool) {
	for _, clusterName := range bc.Clusters {
		fmt.Println(clusterName)
		url := bc.buildDeploymentURL(name, nil, clusterName)
		_, res, err := bc.makeRequest("GET", url, nil)
		if res != nil {
			defer res.Body.Close()
		}
		if err != nil {
			fmt.Println(errorMessageBuilder("Cannot get info for deploy", err))
			continue
		}
		if checkStatusOK(res.StatusCode) {
			if checkAuthOK(res.StatusCode) {
				if res.StatusCode >= 400 && res.StatusCode <= 499 {
					e := Error{}
					unmarshalResponse(res, &e)
					fmt.Printf("Cannot get info for deployment: %s\n", e.Err)
				} else {
					artifact := Artifact{}
					unmarshalResponse(res, &artifact)
					printInfoTable(verbose, artifact)
				}
			} else {
				handleAuthNOK(res.StatusCode)
			}
		} else {
			handleStatusNOK(res.StatusCode)
		}
	}
}

//ListDeploy is used to get a list of the running deployments in the cluster
func (bc *Client) ListDeploy(all bool) {
	for _, clusterName := range bc.Clusters {
		fmt.Println(clusterName)
		var query map[string]string
		if all {
			query = map[string]string{"all": "true"}
		}
		url := bc.buildDeploymentURL("", query, clusterName)
		_, res, err := bc.makeRequest("GET", url, nil)
		if res != nil {
			defer res.Body.Close()
		}
		if err != nil {
			fmt.Println(errorMessageBuilder("Cannot list deployments", err))
			continue
		}
		if checkStatusOK(res.StatusCode) {
			if checkAuthOK(res.StatusCode) {
				if res.StatusCode >= 400 && res.StatusCode <= 499 {
					e := Error{}
					err = unmarshalResponse(res, &e)
					if err != nil {
						fmt.Printf("Cannot get list of deployments: %s\n", err.Error())
						continue
					}
					fmt.Printf("Cannot get list of deployments: %s\n", e.Err)
				} else {
					var ld ListDeployments
					unmarshalResponse(res, &ld)
					fmt.Printf("List of deployed applications: \n")
					for _, name := range ld.Deployments {
						fmt.Printf("\t%s\n", name)
					}
				}
			} else {
				handleAuthNOK(res.StatusCode)
			}
		} else {
			handleStatusNOK(res.StatusCode)
		}
	}
}

//CreateDeploy is used to deploy a new app. If an app with the same name is already deployed,
//an error will be returned.
func (bc *Client) CreateDeploy(cmdReq *CmdClientRequest) {
	//for each datacenter, create the app
	for _, clusterName := range bc.Clusters {
		fmt.Println(clusterName)
		deploy := map[string]interface{}{"Name": cmdReq.Name, "Ports": cmdReq.Ports, "Labels": cmdReq.Labels,
			"ImageURL": cmdReq.ImageURL, "Env": cmdReq.Env, "Replicas": cmdReq.Replicas, "CPULimit": cmdReq.CPULimit,
			"MemoryLimit": cmdReq.MemoryLimit, "Force": cmdReq.Force, "Volumes": cmdReq.Volumes}
		url := bc.buildDeploymentURL("", nil, clusterName)
		_, res, err := bc.makeRequest("POST", url, deploy)
		if res != nil {
			defer res.Body.Close()
		}
		if err != nil {
			fmt.Println(errorMessageBuilder("Deploy unsuccessful", err))
			continue
		}
		if checkStatusOK(res.StatusCode) {
			if checkAuthOK(res.StatusCode) {
				if res.StatusCode >= 400 && res.StatusCode <= 499 {
					e := Error{}
					unmarshalResponse(res, &e)
					fmt.Printf("Deploy unsuccessful: %s\n", e.Err)
				} else {
					fmt.Println("Application successfully deployed.")
				}
			} else {
				handleAuthNOK(res.StatusCode)
			}
		} else {
			handleStatusNOK(res.StatusCode)
		}

	}

}

//UpdateDeploy is used to update an already deployed app
func (bc *Client) UpdateDeploy(cmdReq *CmdClientRequest) {
	for _, clusterName := range bc.Clusters {
		fmt.Println(clusterName)
		deploy := map[string]interface{}{"Name": cmdReq.Name, "Ports": cmdReq.Ports, "Labels": cmdReq.Labels,
			"ImageURL": cmdReq.ImageURL, "Env": cmdReq.Env, "Replicas": cmdReq.Replicas, "CPULimit": cmdReq.CPULimit,
			"MemoryLimit": cmdReq.MemoryLimit, "Force": cmdReq.Force}
		url := bc.buildDeploymentURL(cmdReq.Name, nil, clusterName)
		_, res, err := bc.makeRequest("PUT", url, deploy)
		if res != nil {
			defer res.Body.Close()
		}
		if err != nil {
			fmt.Println(errorMessageBuilder("Deploy unsuccessful", err))
			continue
		}
		if checkStatusOK(res.StatusCode) {
			if checkAuthOK(res.StatusCode) {
				if res.StatusCode >= 400 && res.StatusCode <= 499 {
					e := Error{}
					unmarshalResponse(res, &e)
					fmt.Printf("Update unsuccessful: %s\n", e.Err)
				} else {
					fmt.Println("Application successfully updated.")
				}
			} else {
				handleAuthNOK(res.StatusCode)
			}
		} else {
			handleStatusNOK(res.StatusCode)
		}
	}
}

//Scale is used to scale an existing application to the number of replicas specified
func (bc *Client) Scale(name string, replicas int, force bool) {
	for _, clusterName := range bc.Clusters {
		fmt.Println(clusterName)
		deploy := map[string]interface{}{"Name": name, "Replicas": replicas}
		url := bc.buildDeploymentReplicasURL(name, replicas, clusterName, force)
		_, res, err := bc.makeRequest("PATCH", url, deploy)
		if res != nil {
			defer res.Body.Close()
		}
		if err != nil {
			fmt.Println(errorMessageBuilder("Cannot scale", err))
			continue
		}
		if checkStatusOK(res.StatusCode) {
			if checkAuthOK(res.StatusCode) {
				if res.StatusCode >= 400 && res.StatusCode <= 499 {
					e := Error{}
					unmarshalResponse(res, &e)
					fmt.Printf("Scale unsuccessful: %s\n", e.Err)
				} else {
					fmt.Println("Application scaled.")
				}
			} else {
				handleAuthNOK(res.StatusCode)
			}
		} else {
			handleStatusNOK(res.StatusCode)
		}
	}
}

func errorMessageBuilder(message string, err error) string {
	if strings.Contains(err.Error(), "tls: oversized") {
		return fmt.Sprintf("%s, caused by: cannot estabilish an https connection.", message)
	}
	return fmt.Sprintf("%s, caused by: %s", message, err.Error())
}

func printInfoTable(verbose bool, artifact Artifact) {
	table := printer.NewWriter(os.Stdout)
	//iterate table and print
	table.SetHeader([]string{"Name", "Status", "Endpoints", "Num Replicas", "CPUs", "Memory", "Last Message"})
	row := []string{}
	var endpoints string
	var ports string
	for _, replica := range artifact.RunningReplicas {
		endpoints = endpoints + fmt.Sprintf("%s\n", replica.Endpoints)
		for _, port := range replica.Ports {
			ports = ports + fmt.Sprintf("%d, ", port.Port)
		}
	}
	row = append(row, artifact.Name)
	row = append(row, artifact.Status)
	row = append(row, artifact.Endpoint)
	row = append(row, fmt.Sprintf("%d/%d", len(artifact.RunningReplicas), artifact.RequestedReplicas))
	cpus := strconv.FormatFloat(artifact.CPUS, 'f', 1, 64)
	memory := strconv.FormatFloat(artifact.Memory, 'f', 1, 64)
	row = append(row, cpus)
	row = append(row, memory)
	row = append(row, artifact.Message)
	table.Append(row)
	table.Render()

	//second table in case of verbose flag set
	if verbose {
		containerTable := printer.NewWriter(os.Stdout)
		containerTable.SetRowLine(true)
		containerTable.SetHeader([]string{"Container Status", "Image", "Endpoint", "Logfile"})
		for _, replica := range artifact.RunningReplicas {
			cRow := []string{}
			cRow = append(cRow, replica.Containers[0].Status)
			cRow = append(cRow, replica.Containers[0].ImageURL)
			cRow = append(cRow, replica.Endpoints[0])
			cRow = append(cRow, replica.Containers[0].LogInfo["containerName"])
			cRow = append(cRow)
			containerTable.Append(cRow)
		}
		containerTable.Render()

		settingsTable := printer.NewWriter(os.Stdout)
		settingsTable.SetRowLine(true)
		settingsTable.SetHeader([]string{"Env name", "value"})
		for k, v := range artifact.Env {
			sRow := make([]string, 0, 2)
			sRow = append(sRow, k)
			sRow = append(sRow, v)
			settingsTable.Append(sRow)
		}
		settingsTable.Render()

		labelsTable := printer.NewWriter(os.Stdout)
		labelsTable.SetRowLine(true)
		labelsTable.SetHeader([]string{"Label", "value"})
		for k, v := range artifact.Labels {
			sRow := make([]string, 0, 2)
			sRow = append(sRow, k)
			sRow = append(sRow, v)
			labelsTable.Append(sRow)
		}
		labelsTable.Render()
	}

}

func printLogs(artifact Artifact) {
	containerTable := printer.NewWriter(os.Stdout)
	containerTable.SetRowLine(false)
	containerTable.SetRowLine(false)
	containerTable.SetBorder(false)
	containerTable.SetHeader([]string{"Endpoint", "Log URL"})
	for _, replica := range artifact.RunningReplicas {
		cRow := []string{}
		cRow = append(cRow, replica.Endpoints[0])
		cRow = append(cRow, replica.Containers[0].LogInfo["remoteURL"])
		cRow = append(cRow)
		containerTable.Append(cRow)
	}
	containerTable.Render()

}
