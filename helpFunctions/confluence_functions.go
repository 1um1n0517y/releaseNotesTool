package helpFunctions

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	BaseURL  = "https://intranet.gtechg2.com"
	Username = "ci_games_belgrade"
	Password = "ST6ou/Ka"
)

func CreateNewConfluencePageWithParentPageID(title string, parentPageID string, content string, spaceKey string) error {

	// first check if page exists
	exitIfPageExists(title, spaceKey)

	method := "POST"
	pageURL := BaseURL + "/rest/api/content/"
	page := ConfluencePage{Type: "page", Title: title}

	page.Space.Key = spaceKey

	page.Ancestors = make([]Ancestor, 1)
	page.Ancestors[0] = Ancestor{Id: parentPageID}

	page.Body.Storage.Representation = "storage"
	page.Body.Storage.Value = content

	// Prepare page data for posting
	buff, err := json.Marshal(page)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, pageURL, bytes.NewReader(buff))
	if err != nil {
		return err
	}

	req.SetBasicAuth(Username, Password)
	req.Header.Add("Content-Type", "application/json")

	// Send Confluence API Create
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	// Check status
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		// fmt.Println("Error creating confluence page: " + page.Title)
		fmt.Println(res.Status)
		return errors.New("Error creating confluence page: " + page.Title)
	}

	fmt.Println("Created confluence page: ", page.Title)
	return nil
}

func UpdateConfluencePageWithParentPageID(title string, parentPageID string, content string, spaceKey string) error {

	page, err := GetPageByName(title, spaceKey, "")
	if err != nil {
		return err
	}
	method := "PUT"
	pageURL := BaseURL + "/rest/api/content/" + page.Id

	page.Space.Key = spaceKey
	page.Ancestors = make([]Ancestor, 1)
	page.Ancestors[0] = Ancestor{Id: parentPageID}
	page.Body.Storage.Value = content
	page.Body.Storage.Representation = "storage"
	page.Version.Number = page.Version.Number + 1

	// Prepare page data for posting
	buff, err := json.Marshal(page)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, pageURL, bytes.NewReader(buff))
	if err != nil {
		return err
	}

	req.SetBasicAuth(Username, Password)
	req.Header.Add("Content-Type", "application/json")

	// Send Confluence API Create
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	defer res.Body.Close()
	// Check status
	if res.StatusCode != http.StatusOK {
		// fmt.Println("Error updating confluence page: " + page.Title)
		fmt.Println(res.Status)
		return errors.New("Error updating confluence page: " + page.Title)
	}

	fmt.Println("Updated confluence page: ", page.Title)
	return nil
}

func DeleteConfluencePageWithParentPageID(title string, parentPageID string, spaceKey string) error {

	page, err := GetPageByName(title, spaceKey, "")
	if err != nil {
		return err
	}
	method := "DELETE"
	pageURL := BaseURL + "/rest/api/content/" + page.Id

	req, err := http.NewRequest(method, pageURL, strings.NewReader(""))
	Check(err)

	req.SetBasicAuth(Username, Password)

	// Send Confluence API Create
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	defer res.Body.Close()
	// Check status
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		// fmt.Println("Error deleting confluence page: " + page.Title)
		fmt.Println(res.Status)
		return errors.New("Error deleting confluence page: " + page.Title)
	}

	fmt.Println("Deleted confluence page: ", page.Title)
	return nil
}

func exitIfPageExists(title string, spaceKey string) {
	oldPage, err := GetPageByName(title, spaceKey, "")
	if err != nil {
		log.Fatal(err)
	}
	if len(oldPage.Id) > 0 {
		fmt.Println("Page " + oldPage.Title + " exists")
		os.Exit(1)
	}
}

func GetPageByName(pageName string, spaceKey string, expand string) (ConfluencePage, error) {
	url := BaseURL + "/rest/api/content?spaceKey=" + spaceKey + "&title=" + url.QueryEscape(pageName) + "&expand=space,body.storage,version,container,ancestors" + expand
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ConfluencePage{}, err
	}
	req.SetBasicAuth(Username, Password)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ConfluencePage{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		fmt.Println(res.Status)
		os.Exit(0)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ConfluencePage{}, err
	}
	var results PageResults
	err = json.Unmarshal(body, &results)
	if err != nil {
		return ConfluencePage{}, err
	}
	if len(results.Results) != 1 {
		fmt.Println("Page: " + pageName + " does not exist in space: " + spaceKey)
		return ConfluencePage{}, nil
	}
	return results.Results[0], nil
}

func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type ConfigurationType struct {
	ConfluenceUser     string
	ConfluencePassword string
	ConfluenceHost     string
}

type ConfluencePage struct {
	Id    string `json:"id,omitempty"`
	Type  string `json:"type"`
	Title string `json:"title"`
	Space struct {
		Key string `json:"key"`
	} `json:"space"`
	Ancestors []Ancestor `json:"ancestors"`
	Body      struct {
		Storage struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
	} `json:"body"`
	Version struct {
		Number int `json:"number"`
	} `json:"version"`
}

type Ancestor struct {
	Id string `json:"id"`
}

type PageResults struct {
	Results []ConfluencePage `json:"results"`
	Size    int              `json:"size"`
	Start   int              `json:"start"`
}
