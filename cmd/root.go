package cmd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"../helpFunctions"

	"github.com/spf13/cobra"
)

var (
	spaceKey    string
	parentPage  string
	releasePage string
	svnPath     string
	version     string
	command     *exec.Cmd
)

type Log struct {
	Entries []LogEntry `xml:"logentry"`
}

type LogEntry struct {
	Author string `xml:"author"`
	Date   string `xml:"date"`
	Msg    string `xml:"msg"`
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "releaseNotesTool",
	Short: "Update release notes page on Confluence",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		tagsPath := "/standard/imp/tags"
		trunkPath := "/standard/imp/trunk"
		if parentPage == "Icy Wilds" {
			tagsPath = "/standard/tags"
			trunkPath = "/standard/trunk"
		}

		revision, date := getRevision(svnPath+tagsPath, version)
		if revision != "" {
			command = exec.Command("svn", "log", "-r", revision, "--xml", svnPath+trunkPath)
		} else {
			command = exec.Command("svn", "log", "--xml", svnPath+trunkPath)
		}
		var out bytes.Buffer
		command.Stdout = &out
		err := command.Run()
		if err != nil {
			log.Fatal(err)
		}

		var result Log
		xml.Unmarshal(out.Bytes(), &result)

		content := createContent(result, version, date)
		parent, err := helpFunctions.GetPageByName(parentPage, spaceKey, "")
		if err != nil {
			log.Fatal(err)
		}
		child, err := helpFunctions.GetPageByName(parentPage+" Releases", spaceKey, "")
		if err != nil {
			log.Fatal(err)
		}
		newContent := content + child.Body.Storage.Value
		err = helpFunctions.UpdateConfluencePageWithParentPageID(parentPage+" Releases", parent.Id, newContent, spaceKey)
		if err != nil {
			log.Fatal(err)
		}

	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.Flags().StringVarP(&spaceKey, "spaceKey", "k", "GAMBG", "Confluence space key.")
	RootCmd.Flags().StringVarP(&parentPage, "parentPage", "p", "", "Parent page name.")
	RootCmd.Flags().StringVarP(&svnPath, "svnPath", "s", "", "Path to svn location.")
	RootCmd.Flags().StringVarP(&version, "version", "v", "", "New release version.")
}

func createContent(data Log, version string, date string) string {
	reg := regexp.MustCompile("[A-Z]+-[0-9]+")

	content := "<table><tbody><tr><th>Version</th><th>Date</th></tr>"
	content += "<tr><td>" + version + "</td><td>" + date + "</td></tr>"
	content += "</tbody></table>"
	content += "<table><tbody>"
	content += "<tr><th>JIRA</th><th>Description</th><th>Author</th></tr>"
	for _, entry := range data.Entries {

		if entry.Msg != "" {
			msg := strings.Replace(entry.Msg, "\n", "", -1)
			msg = strings.Replace(msg, "&", " and ", -1)
			issue := reg.FindString(msg)
			description := msg[len(issue):]

			if entry.Author != "ci_games_belgrade" {
				content += `
					<tr>
						<td>
							<ac:structured-macro ac:name="jiraissues">
								<ac:parameter ac:name="url"><ri:url ri:value="https://jira.g2-networks.net/browse/` + issue + `"/></ac:parameter>
								<ac:parameter ac:name="columns">key,summary</ac:parameter>
							</ac:structured-macro>
						</td>
						<td>` + description + `</td>
						<td>` + entry.Author + `</td>
					</tr>
					`
			}
		}
	}

	content += "</tbody></table><hr />"
	return content
}

func getRevision(svnPath string, version string) (string, string) {
	patchNum := strings.Split(version, ".")[2]
	patchNumInt, err := strconv.Atoi(patchNum)
	if err != nil {
		log.Fatal(err)
	}

	if patchNum == "0" {
		_, date := getSvnInfo(svnPath + "/" + version)
		return "", date
	}

	lastRev, date := getSvnInfo(svnPath + "/" + version)
	lastOldRev := ""
	for {
		oldVersion := strings.Split(version, ".")[0] + "." + strings.Split(version, ".")[1] + "." + strconv.Itoa(patchNumInt-1)
		lastOldRev, _ = getSvnInfo(svnPath + "/" + oldVersion)
		if lastOldRev != "" {
			break
		}
		patchNumInt = patchNumInt - 1
	}

	return lastRev + ":" + lastOldRev, date
}

type Info struct {
	XMLName xml.Name `xml:info`
	Entry   Entry    `xml:"entry"`
}

type Entry struct {
	Revision string `xml:"revision,attr"`
	Commit   Commit `xml:"commit"`
}

type Commit struct {
	Revision string `xml:"revision,attr"`
	Date     string `xml:"date"`
}

func getSvnInfo(svnPath string) (string, string) {
	command := exec.Command("svn", "info", "--xml", svnPath)
	var out bytes.Buffer
	command.Stdout = &out
	err := command.Run()
	if err != nil {
		log.Fatal(err)
		return "", ""
	}
	var result Info
	xml.Unmarshal(out.Bytes(), &result)
	const longForm = "2006-01-02T15:04:05.000000Z"
	t, err := time.Parse(longForm, result.Entry.Commit.Date)

	if err != nil {
		log.Fatal(err)
	}
	formatedTime := t.Format("2006-01-02 15:04:05 (Mon, 02 Jan 2006)")
	return result.Entry.Commit.Revision, formatedTime
}
