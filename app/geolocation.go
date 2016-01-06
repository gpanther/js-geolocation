package hello

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	http.HandleFunc("/", redirector)
	http.HandleFunc("/api/ip", ip)
	http.HandleFunc("/api/geolocation", geolocation)
}

func redirector(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "http://example.com", http.StatusFound)
}

func addCorsHeaders(w http.ResponseWriter, r *http.Request) {
	rh := r.Header
	rwh := w.Header()

	if origin := rh.Get("Origin"); "" != origin {
		rwh.Set("Access-Control-Allow-Origin", origin)
		rwh.Set("Access-Control-Allow-Methods", "GET")
		rwh.Set("Vary", "Origin")
	}
}

var validCallbackFuncName = regexp.MustCompile(`^\w+$`)

func serveResult(result interface{}, w http.ResponseWriter, r *http.Request) {
	marshalledResult, err := json.Marshal(result)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// add cache headers
	rwh := w.Header()
	rwh.Set("Cache-Control", "private")
	expires := time.Now().Add(time.Duration(time.Hour)).UTC()
	rwh.Set("Expires", expires.Format("Mon, Jan 02 2006 15:04:05 GMT"))

	format := r.URL.Query().Get("format")
	switch format {
	case "json":
		rwh.Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, string(marshalledResult))
		return
	case "jsonp":
		callback := r.URL.Query().Get("callback")
		if !validCallbackFuncName.MatchString(callback) {
			http.Error(w, "Incorrect request", 400)
			return
		}
		rwh.Set("Content-Type", "application/javascript; charset=utf-8")
		fmt.Fprint(w, callback+"("+string(marshalledResult)+");")
		return
	}

	http.Error(w, "Incorrect request", 400)
}

func ip(w http.ResponseWriter, r *http.Request) {
	addCorsHeaders(w, r)
	if "OPTIONS" == r.Method {
		return
	}

	result := map[string]string{
		"ip": r.RemoteAddr,
	}

	serveResult(result, w, r)
}

func geolocation(w http.ResponseWriter, r *http.Request) {
	addCorsHeaders(w, r)
	if "OPTIONS" == r.Method {
		return
	}

	rh := r.Header
	latlong := strings.Split(rh.Get("X-AppEngine-CityLatLong"), ",")
	lat, _ := strconv.ParseFloat(latlong[0], 64)
	long, _ := strconv.ParseFloat(latlong[1], 64)
	result := map[string]interface{}{
		"country": rh.Get("X-AppEngine-Country"),
		"region":  rh.Get("X-AppEngine-Region"),
		"city":    rh.Get("X-AppEngine-City"),
		"cityLatLong": map[string]float64{
			"lat":  lat,
			"long": long,
		},
	}

	serveResult(result, w, r)
}
