package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Item is the single repository data structure consisting of **only the data i need**
type Item struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

// JSONData contains the GitHub API response
type JSONData struct {
	Count    int `json:"total_count"`
	Username string
	Items    []Item
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/post", fetch)
	http.ListenAndServe(":8080", nil)
}

// fetch() fetches the data from GitHub P.S:The maximum it can fetch is 100 records due to GitHub's Pagination
func fetch(w http.ResponseWriter, r *http.Request) {
	url := "https://api.github.com/search/repositories?q=user:@&per_page=100"

	user := r.FormValue("firstname")

	url = strings.Replace(url, "@", user, 1)

	res, getErr := http.Get(url)
	if getErr != nil {
		log.Fatal(getErr)
	}
	data := JSONData{}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		log.Fatal(err)
	}

	if data.Count > 100 {
		num := (data.Count / 100)
		if data.Count%100 != 0 {
			num++
		}
		fmt.Println("num = ", num)
		for i := 2; i <= num; i++ {
			page := url + "&page=@"

			page = strings.Replace(page, "@", strconv.Itoa(i), 1)
			rest, err := http.Get(page)
			fmt.Println("Gotten page", i, "url=", page)
			if err != nil {
				log.Fatal(err)
			}
			tmp := JSONData{}
			if err := json.NewDecoder(rest.Body).Decode(&tmp); err == nil {
				ref := &data
				for ix := range tmp.Items {
					ref.Items = append(ref.Items, tmp.Items[ix])
				}
			} else {
				log.Fatal(err)
			}
		}
	}

	for i := range data.Items {
		r := &data.Items[i]
		if format, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
			ref := strconv.Itoa(format.Year()) + " " + format.Month().String() + " " +
				strconv.Itoa(format.Hour()) + ":" + strconv.Itoa(format.Minute()) + ":" + strconv.Itoa(format.Second())
			r.CreatedAt = ref
		} else {
			log.Fatal(err)
		}
	}
	data.Username = user
	sendData(&data, user, &w)
}

// sendData() sends the data to the client using a buffer
func sendData(data *JSONData, user string, w *http.ResponseWriter) {

	t := template.New("Response Data")
	t, _ = t.Parse(`
    <p>
	  Github User {{ .Username }} has {{ .Count }} repositories as follows:
    </p>
    <table>
        <tr style="display: table-row">
          <th>Repository Name</th>
		  <th>Description</th>
		  <th>Created At</th>
		</tr>
		{{ with .Items }}

			{{ range . }}
        		<tr style="display: table-row">
          			<td id="fn">{{ .FullName }}</td>
		  			<td id="dsc">{{ .Description }}</td>
		  			<td id="cat">{{ .CreatedAt }}</td>
				</tr>
			{{ end }}

		{{ end }}
      </table>
	`)

	buf := new(bytes.Buffer)
	t.Execute(buf, *data)
	tmp := *w
	tmp.Write(buf.Bytes())
	tmp.(http.Flusher).Flush()
}
