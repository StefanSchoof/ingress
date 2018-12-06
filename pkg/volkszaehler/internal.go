package volkszaehler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	TypeGroup = "group"
)

type Api struct {
	url    string
	client http.Client
	debug  bool
}

func NewAPI(url string, timeout time.Duration, debug bool) *Api {
	return &Api{
		url: detectApiEndpoint(url),
		client: http.Client{
			Timeout: timeout,
		},
		debug: debug,
	}
}

func detectApiEndpoint(url string) string {
	const probe = "/entity.json"

	url = strings.TrimRight(url, "/")
	log.Println("Validating API endpoint")

	resp, err := http.Get(url + probe)
	if err == nil && resp.StatusCode == 200 {
		log.Println("API endpoint validated")
		return url
	}

	// append middleware.php
	detectedURL := url + "/middleware.php"
	log.Println("API endpoint not responding. Trying " + detectedURL)

	resp, err = http.Get(detectedURL + probe)
	if err == nil && resp.StatusCode == 200 {
		log.Println("API endpoint detected, using " + detectedURL)
		return detectedURL
	}

	log.Println("API endpoint still not responding. Will keep retrying using configured uri")
	return url
}

func (api *Api) validate() {
	resp, err := http.Get(api.url)
	log.Println(err)
	log.Fatal(resp)
}

func (api *Api) Get(endpoint string) (*http.Response, error) {
	url := api.url + endpoint

	start := time.Now()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Accept", "application/json")

	resp, err := api.client.Do(req)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	duration := time.Now().Sub(start)
	log.Printf("GET %s (%dms)", url, duration.Nanoseconds()/1e6)

	if api.debug {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Print(err)
		}
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		log.Print(string(body))
	}

	return resp, nil
}

func (api *Api) GetPublicEntities() []Entity {
	context := "/entity.json"
	r, err := api.Get(context)
	if err != nil {
		return []Entity{}
	}

	er := EntitiesResponse{}
	if err := json.NewDecoder(r.Body).Decode(&er); err != nil {
		log.Printf("json decode failed: %v", err)
		return []Entity{}
	}

	return er.Entities
}

func (api *Api) GetEntity(parent string) Entity {
	context := fmt.Sprintf("/entity/%s.json", parent)
	r, err := api.Get(context)
	if err != nil {
		return Entity{}
	}

	er := EntityResponse{}
	if err := json.NewDecoder(r.Body).Decode(&er); err != nil {
		log.Printf("json decode failed: %v", err)
		return Entity{}
	}

	return er.Entity
}

func getGroup(d int64) string {
	if d > 3600*24*365 {
		return "year"
	} else if d > 3600*24*30 {
		return "month"
	} else if d > 3600*24*7 {
		return "week"
	} else if d > 3600*24 {
		return "day"
	} else if d > 3600 {
		return "hour"
	} else if d > 60 {
		return "minute"
	}
	return ""
}

func (api *Api) GetData(uuid string, from time.Time, to time.Time, group string, options string, tuples int) []Tuple {
	f := from.Unix()
	t := to.Unix()
	url := fmt.Sprintf("/data/%s.json?from=%d&to=%d", uuid, f*1000, t*1000)

	if tuples > 0 {
		url += fmt.Sprintf("&tuples=%d", tuples)

		if group == "" {
			period := (t - f) / int64(tuples)
			group = getGroup(period)
		}
	}

	if group != "" {
		url += "&group=" + group
	}

	if options != "" {
		url += "&options=" + options
	}

	r, err := api.Get(url)
	if err != nil {
		return []Tuple{}
	}

	dr := DataResponse{}
	if err := json.NewDecoder(r.Body).Decode(&dr); err != nil {
		log.Printf("json decode failed: %v", err)
		return []Tuple{}
	}

	return dr.Data.Tuples
}

func (api *Api) GetPrognosis(uuid string, period string) PrognosisStruct {
	url := fmt.Sprintf("/prognosis/%s.json?period=%s", uuid, period)

	r, err := api.Get(url)
	if err != nil {
		return PrognosisStruct{}
	}

	pr := PrognosisResponse{}
	if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
		log.Printf("json decode failed: %v", err)
		return PrognosisStruct{}
	}

	return pr.Prognosis
}

func (api *Api) Post(endpoint string, payload string) (*http.Response, error) {
	url := api.url + endpoint

	start := time.Now()
	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Accept", "application/json")

	resp, err := api.client.Do(req)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	duration := time.Now().Sub(start)
	log.Printf("POST %s (%dms)", url, duration.Nanoseconds()/1e6)

	if api.debug {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Print(err)
		}
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		log.Print(string(body))
	}

	return resp, nil
}
