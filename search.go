package main

import (
"fmt"
"time"
"log"
"regexp"
"strings"
"io/ioutil"
"net/http"
"encoding/json"
"html/template"
"github.com/lanior/html"
"github.com/gorilla/mux"
)

type Page struct {
    Title string
    //Body  []byte
}

type Result struct {
	Message []string
	Url string
}

type Word struct {
	word string
	count int
}

func pr(s []string) {
	fmt.Printf("%v\n", s)
}
func p(s string) {
	fmt.Printf("%v\n", s)
}
func e(e error) {
	if(e != nil ){
		fmt.Printf(" ERROR: %v\n", e)			
	}
 
}
func stripTags(s string) string {
	plaintext := ""
	// TODO; use a reflect swtich to handle arg types (string,)
	z := html.NewTokenizer(strings.NewReader(s))
	depth := 0
	Loop:
		for {
			tt:=z.Next()
			switch tt {
			case html.ErrorToken:
				break Loop
			case html.TextToken:
				if depth > 0 {
					plaintext = plaintext+" "+string(z.Text())+" "
				}
			case html.StartTagToken, html.EndTagToken:
				tn, _ := z.TagName()
				if len(tn) == 1 && tn[0] == 'a' {
					if tt == html.StartTagToken {
						depth++
					} else {
						depth--
					}
				}
			}

		}
	return plaintext
}
func getPageChunk(q, url, lwindow, rwindow string) []string{
	res, err := http.Get(url)
	e(err)
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	e(err)

	str_a := string(strings.ToLower(string(body)))
	
	plaintext := stripTags(str_a)
     
    backref := ""
    forwardref := ""

	if lwindow != "0" && lwindow != "1"{
		backref = "(?:(.{1,"+lwindow+"}))"
	}
	if  rwindow != "0" && rwindow != "1"{
		forwardref = "(?:(.{1,"+rwindow+"}))"
	}
	matches := []string{}
	re2,http_err := regexp.Compile(backref+"("+q+")"+forwardref)		
	if(http_err == nil){
		matches = re2.FindAllString(plaintext, -1)		
	}else {
		matches = append(matches,"regex compilation error: "+http_err.Error())
	}

	


	return matches
}


/*
	MultiFetch
	query - search query
	site - url to search
	lwindow, rwindow - the window to the left or right of your query
*/
func MultiFetch(query, site, lwindow, rwindow string) Result {
	c := make(chan Result)
	p(site)
	go func() { 
		c <- Result{ getPageChunk(query, site, lwindow, rwindow), site }
		} ()
	return <-c
}

func GoroSearch(query, lwindow, rwindow string) (results []Result) {
	c:=make(chan Result)

	sites := [...]string{"https://news.ycombinator.com", "https://google.com/news", "http://reddit.com"}

	for _,site := range sites {
		go func (site,lwindow,rwindow string) {
		 	c <-  MultiFetch(query, site, lwindow, rwindow)
		 } (site, lwindow, rwindow)
	}
	

	timeout := time.After(5000 * time.Millisecond)
	for i := 0; i < len(sites); i++ {
		select {
		case result := <- c:
			results = append(results, result)
		case <- timeout:
			//fmt.Println("timed out")
		return
		}
	}
	return
}

func httpSearchHandler(w http.ResponseWriter, r *http.Request) {
	
	req_r := strings.Split(html.EscapeString(r.URL.Path), "/")
	vars := mux.Vars(r)
	lwindow := vars["lwindow"]
	rwindow := vars["rwindow"]

	results := GoroSearch(req_r[2],lwindow,rwindow)

	b,err := json.MarshalIndent(results," " ,"\t")
	e(err)

	fmt.Fprintf(w, "%s\n", b)
}


var index = template.Must(template.ParseFiles(
  "templates/base.html",
  "templates/index.html",
))
func renderTemplate(w http.ResponseWriter, p *Page) {
    index.Execute(w, p)
}
func main() {

    router := mux.NewRouter()
    router.HandleFunc("/search/{query}/{lwindow}/{rwindow}", func(w http.ResponseWriter, r *http.Request) {
		//doing this in preparation to provide preloaded content to the search handler
		// domains, shortcuts
		w.Header().Set("Content-type", "application/json")
		httpSearchHandler(w, r)
	})
	
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, &Page{Title:"Up With It" })
	})

    http.Handle("/", router)

	p("listening on 127.0.0.1:8333")

	log.Fatal(http.ListenAndServe("127.0.0.1:8333", nil))
}
