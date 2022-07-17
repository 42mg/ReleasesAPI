package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

var tokens []string

func main() {
	tokens = os.Args[1:]
	if len(tokens) == 0 {
		log.Fatalln("401 Unauthorized - Bad credentials")
	}
	gHandler := http.HandlerFunc(gFunc)
	http.Handle("/", gHandler)
	http.HandleFunc("/favicon.ico", noFavicon)
	http.ListenAndServe(":8080", nil)
}

func noFavicon(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(204)
}

func gFunc(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	repo := query.Get("repo")
	include := query.Get("include")
	exclude := query.Get("exclude")
	download := query.Get("download")
	incl := strings.Split(include, " ")
	excl := strings.Split(exclude, " ")
	repo = "https://api.github.com/repos/" + strings.TrimSuffix(strings.TrimPrefix(repo, "https://github.com/"), "/") + "/releases/latest"
	req, err := http.NewRequest("GET", repo, nil)
	if err != nil {
		log.Fatalln(err)
	}
	if len(tokens) > 1 {
		z := tokens[len(tokens)-1]
		copy(tokens[1:], tokens)
		tokens[0] = z
	}
	req.Header.Set("Authorization", "token "+tokens[0])
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	assets := gjson.Get(string(respData), "assets.#.browser_download_url")
	var releases []string
	if (include != "") && (exclude != "") {
		var tmp []string
		for _, release := range assets.Array() {
			x := release.Str
			i := 0
			for _, v := range incl {
				if !strings.Contains(x, v) {
					break
				}
				i++
			}
			if i == len(incl) {
				tmp = append(tmp, x)
			}
		}
		for _, t := range tmp {
			i := 0
			for _, v := range excl {
				if !strings.Contains(t, v) {
					i++
				}
			}
			if i == len(excl) {
				releases = append(releases, t)
			}
		}
	} else {
		for _, release := range assets.Array() {
			x := release.Str
			if (include != "") && (exclude == "") {
				i := 0
				for _, v := range incl {
					if !strings.Contains(x, v) {
						break
					}
					i++
				}
				if i == len(incl) {
					releases = append(releases, x)
				}
			} else if (include == "") && (exclude != "") {
				i := 0
				for _, v := range excl {
					if !strings.Contains(x, v) {
						i++
					}
				}
				if i == len(excl) {
					releases = append(releases, x)
				}
			} else {
				releases = append(releases, x)
			}
		}
	}
	if (download == "true") && (len(releases) == 1) {
		http.Redirect(w, r, releases[0], 301)
	} else {
		responseText := strings.Join(releases[:], "\n")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(responseText))
	}
}
