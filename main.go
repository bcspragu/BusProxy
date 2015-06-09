package main

import (
	"github.com/bcspragu/AvailParser"

	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"appengine"
	"appengine/urlfetch"
)

type StopResponse struct {
	StopName string      `json:"stopName"`
	Routes   []RouteInfo `json:"routes"`
}

type RouteInfo struct {
	Number        string          `json:"number"`
	Name          string          `json:"name"`
	DepartureTime avail.AvailTime `json:"departureTime"`
	Color         string          `json:"color"`
	TextColor     string          `json:"textColor"`
}

var feed *avail.Feed
var routes = make(map[int]avail.Route)
var stops = make(map[int]avail.Stop)

var t = template.Must(template.New("").Delims(`\\\`, `///`).ParseFiles("index.html"))

func init() {
	feed = avail.NewFeed("http://bustracker.pvta.com")
	avail.SetLocation("America/New_York")

	http.HandleFunc("/", serveInfo)
	http.HandleFunc("/board", serveBoard)
	http.HandleFunc("/_ah/warmup", serveLoad)
}

func serveLoad(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	feed.SetClient(urlfetch.Client(c))

	visRoutes, err := feed.VisibleRoutes()
	if err != nil {
		log.Fatal("Loading routes: ", err)
	}

	for _, route := range visRoutes {
		routes[route.RouteId] = route
	}

	allStops, err := feed.Stops()
	if err != nil {
		log.Fatal("Loading stops: ", err)
	}

	for _, stop := range allStops {
		stops[stop.StopId] = stop
	}
}

func serveInfo(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	feed.SetClient(urlfetch.Client(c))

	w.Header().Add("Access-Control-Allow-Origin", "*")
	// We're sending a JSON response
	w.Header().Set("Content-Type", "application/json")

	r.ParseForm()
	if len(r.Form["stop"]) == 0 {
		w.Write([]byte("[]"))
		return
	}
	stops, err := stopList(r.Form["stop"])

	if err != nil {
		serveError(w, c, err)
		return
	}
	depChan := make(chan avail.StopDeparture, len(stops))

	for _, stop := range stops {
		go func(id int, resChan chan<- avail.StopDeparture) {
			dep, err := feed.StopDeparture(id)
			if err != nil {
				serveError(w, c, err)
				resChan <- avail.StopDeparture{}
				return
			}
			resChan <- dep
		}(stop, depChan)
	}

	resCount := 0
	departures := make([]avail.StopDeparture, len(stops))

	for {
		departures[resCount] = <-depChan
		resCount++
		if resCount == len(stops) {
			break
		}
	}

	res := make([]StopResponse, len(stops))
	for i, dep := range departures {
		res[i] = responseFromDeparture(dep)
	}

	data, err := json.Marshal(res)
	if err != nil {
		serveError(w, c, err)
		return
	}

	w.Write(data)
}

func serveBoard(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	err := t.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		serveError(w, c, err)
	}
}

func stopList(stops []string) ([]int, error) {
	var res []int

	for _, stop := range stops {
		s, err := strconv.Atoi(stop)
		if err != nil {
			return []int{}, errors.New("Ill-formed stop number found")
		}
		res = append(res, s)
	}

	return res, nil
}

func responseFromDeparture(stopDep avail.StopDeparture) StopResponse {
	res := StopResponse{
		StopName: stops[stopDep.StopId].Name,
		Routes:   []RouteInfo{},
	}
	uniqueISDs := make(map[string]bool)

	for _, dir := range stopDep.RouteDirections {
		route := routes[dir.RouteId]
		for _, dep := range dir.Departures {
			// If we haven't seen this one before
			if _, ok := uniqueISDs[dep.Trip.InternetServiceDesc]; !ok {
				if dep.EDT.After(time.Now()) {
					uniqueISDs[dep.Trip.InternetServiceDesc] = true
					routeInfo := RouteInfo{
						Number:        route.ShortName,
						Name:          dep.Trip.InternetServiceDesc,
						Color:         route.Color,
						TextColor:     route.TextColor,
						DepartureTime: dep.EDT,
					}
					res.Routes = append(res.Routes, routeInfo)
				}
			}
		}
	}
	return res
}

func serveError(w http.ResponseWriter, c appengine.Context, err error) {
	c.Errorf("Error handling request: %s", err)
	w.Write([]byte("[]"))
}
