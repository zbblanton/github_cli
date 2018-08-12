package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
  "os"
  "net/http"
  "encoding/json"
  "bufio"
  "io/ioutil"
  "io"
  "strconv"
  "bytes"
)

type releaseListRespItem struct {
  Id   int
  Name string
}

func callGitAPI(method, url, token string, body io.Reader) (respBody []byte, err error) {
  req, err := http.NewRequest(method, url, body)
  if err != nil {
    fmt.Println(err)
    return
  }
  req.Header.Set("Authorization", "token " + token)
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    fmt.Println(err)
    return
  }
	reader := bufio.NewReader(resp.Body)
	respBody, _ = ioutil.ReadAll(reader)
	resp.Body.Close()

	return respBody, nil
}

func checkReqFlags(c *cli.Context) (owner, repo, token string){
  owner = c.String("owner")
  repo = c.String("repo")
  token = c.String("token")

	if owner == "" {
		panic("Must provide a repo owner.")
	}

	if repo == "" {
		panic("Must provide a repo name.")
	}

  if token == "" {
		panic("Must provide an API token.")
	}

  return owner, repo, token
}

func createRelease(c *cli.Context){
  owner, repo, token := checkReqFlags(c)

  tag := c.String("tag")
  release := c.String("release")
  prerelease := c.Bool("prerelease")

  if tag == "" {
    panic("Must provide a tag name.")
  }

  if release == "" {
    panic("Must provide a release name.")
  }

  type createReleaseReq struct {
    TagName   string `json:"tag_name"`
    Name       string `json:"name"`
    Prerelease bool   `json:"prerelease"`
  }

  j := createReleaseReq{TagName: tag, Name: release, Prerelease: prerelease}
  b := new(bytes.Buffer)
  json.NewEncoder(b).Encode(j)

  url := "https://api.github.com/repos/" + owner + "/" + repo + "/releases"
  resp, err := callGitAPI("POST", url, token, b)
  if err != nil {
    panic(err)
  }
  fmt.Println(string(resp))
}

func getReleaseList(c *cli.Context){
  owner, repo, token := checkReqFlags(c)

  url := "https://api.github.com/repos/" + owner + "/" + repo + "/releases"
	resp, err := callGitAPI("GET", url, token, nil)
	if err != nil {
		panic(err)
	}
  fmt.Println(string(resp))
}

func getReleaseIDByTag(c *cli.Context) {
  owner, repo, token := checkReqFlags(c)

  tag := c.Args().First()
	if tag == "" {
		panic("Must provide a tag name.")
	}

  url := "https://api.github.com/repos/" + owner + "/" + repo + "/releases/tags/" + tag
  resp, err := callGitAPI("GET", url, token, nil)
  respBody := releaseListRespItem{}
  err = json.Unmarshal(resp, &respBody)
  if err != nil {
    panic(err)
  }

  fmt.Println(strconv.Itoa(respBody.Id))
}

func deleteReleaseByTag(c *cli.Context) {
  owner, repo, token := checkReqFlags(c)
  tag := c.Args().First()
	if tag == "" {
		panic("Must provide a tag name.")
	}
  url := "https://api.github.com/repos/" + owner + "/" + repo + "/releases/tags/" + tag
  resp, err := callGitAPI("GET", url, token, nil)
  respBody := releaseListRespItem{}
  err = json.Unmarshal(resp, &respBody)
  if err != nil {
    panic(err)
  }

  id := respBody.Id
  url = "https://api.github.com/repos/" + owner + "/" + repo + "/releases/" + strconv.Itoa(id)
	_, err = callGitAPI("DELETE", url, token, nil)
  if err != nil{
    panic(err)
  }
}

func deleteTag(c *cli.Context) {
  owner, repo, token := checkReqFlags(c)
  tag := c.Args().First()
	if tag == "" {
		panic("Must provide a tag name.")
	}
  url := "https://api.github.com/repos/" + owner + "/" + repo + "/git/refs/tags/" + tag
  _, err := callGitAPI("DELETE", url, token, nil)
  if err != nil{
    panic(err)
  }
}

func cliFeatureNotImplemented(c *cli.Context){
  fmt.Println("Feature not implemented.")
}

var (
	flagOwner = cli.StringFlag{
		Name:  "owner",
		Usage: "Owner of the repo.",
	}

	flagRepo = cli.StringFlag{
		Name:  "repo",
		Usage: "Name of the repo.",
	}

	flagToken = cli.StringFlag{
		Name:  "token",
		Usage: "API token.",
	}

  flagTag = cli.StringFlag{
		Name:  "tag",
		Usage: "Tag name.",
	}

  flagRelease = cli.StringFlag{
		Name:  "release",
		Usage: "Release name.",
	}

  flagPreRelease = cli.BoolFlag{
		Name:  "prerelease",
		Usage: "Boolean to mark as pre-release.",
	}
)

func main(){
  app := cli.NewApp()
	app.Name = "GitHub API CLI"
	app.Usage = "CLI for GitHub API"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		{
			Name:  "release",
			Usage: "Helper commands for releases.",
			Subcommands: []cli.Command{
				{
					Name:   "delete",
					Usage:  "Delete a release.",
          Flags:  []cli.Flag{flagOwner, flagRepo, flagToken},
					Action: deleteReleaseByTag,
				},
        {
					Name:   "ls",
					Usage:  "List all releases.",
          Flags:  []cli.Flag{flagOwner, flagRepo, flagToken},
					Action: getReleaseList,
				},
        {
					Name:   "create",
					Usage:  "Create a release.",
          Flags:  []cli.Flag{flagOwner, flagRepo, flagToken, flagTag, flagRelease, flagPreRelease},
					Action: createRelease,
				},
        {
					Name:   "id",
					Usage:  "Get the ID of a release by tag name.",
          Flags:  []cli.Flag{flagOwner, flagRepo, flagToken},
					Action: getReleaseIDByTag,
				},
        {
					Name:   "upload",
					Usage:  "Upload an asset to a release.",
          Flags:  []cli.Flag{flagOwner, flagRepo, flagToken},
					Action: cliFeatureNotImplemented,
				},
			},
		},
    {
			Name:  "tag",
			Usage: "Helper commands for releases.",
			Subcommands: []cli.Command{
				{
					Name:   "delete",
					Usage:  "Delete a tag.",
          Flags:  []cli.Flag{flagOwner, flagRepo, flagToken},
					Action: deleteTag,
				},
        {
					Name:   "ls",
					Usage:  "List all tags.",
          Flags:  []cli.Flag{flagOwner, flagRepo, flagToken},
					Action: cliFeatureNotImplemented,
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
