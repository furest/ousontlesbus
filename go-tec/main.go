package main

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/valyala/fastjson"
)

var apiURL = "https://tec-api.tapptic.com/api/trips/updates"
var generatedIDs = make([]string, 0)
var colors = map[string]string{
	"L": "red",
	"B": "blue",
	"H": "green",
	"C": "orange",
	"X": "purple",
	"N": "gray",
}

//Position represents a geographical position
type Position struct {
	latitude  float64
	longitude float64
}

//Trip represents a live bus
type Trip struct {
	line      string
	vehicleID string
	speed     float64
	position  Position
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	randStr := hex.EncodeToString(bytes)
	for _, generatedID := range generatedIDs {
		if generatedID == randStr {
			return randomHex(n)
		}
	}
	return randStr, nil
}
func cleanLine(line string) string {
	cleanedLine := strings.Split(line, "-")[0]
	for cleanedLine[1] == '0' {
		cleanedLine = string(cleanedLine[0]) + cleanedLine[2:]
	}
	return cleanedLine
}

func createMarker(tripLine string, vehicleID string, position Position, mapID string) (string, error) {
	var err error
	t, err := template.New("marker.template").ParseFiles("marker.template")
	if err != nil {
		return "", err
	}
	var marker bytes.Buffer
	markerID, _ := randomHex(16)
	iconID, _ := randomHex(16)
	popupID, _ := randomHex(16)
	htmlID, _ := randomHex(16)
	lineNum := cleanLine(tripLine)
	err = t.Execute(&marker, map[string]interface{}{
		"MARKER_ID":    markerID,
		"MARKER_LONG":  position.longitude,
		"MARKER_LAT":   position.latitude,
		"ICON_ID":      iconID,
		"POPUP_ID":     popupID,
		"HTML_ID":      htmlID,
		"BUS_ID":       vehicleID,
		"LINE_NUM":     lineNum,
		"LINE_TRIP":    tripLine,
		"MARKER_COLOR": colors[string(tripLine[0])],
		"MAP_ID":       mapID,
	})
	if err != nil {
		return "", err
	}
	return marker.String(), nil
}

func createMap(mapID string, markers string, center Position) (string, error) {
	var err error
	t, err := template.New("map.template").ParseFiles("map.template")
	if err != nil {
		return "", err
	}
	var busMap bytes.Buffer
	err = t.Execute(&busMap, map[string]interface{}{
		"MAP_ID":      mapID,
		"CENTER_LONG": center.longitude,
		"CENTER_LAT":  center.latitude,
		"MARKERS":     markers,
	})
	if err != nil {
		return "", err
	}
	return busMap.String(), nil
}

func queryTrips(tripIDs []string, c chan []Trip) {
	t := time.Now()
	request := map[string]interface{}{
		"complete": true,
		"datetime": t.Format("2006-01-02 15:04:05"),
		"trip_ids": tripIDs,
	}
	strReq, _ := json.Marshal(request)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(strReq))
	req.Header.Set("User-Agent", "okhttp/3.8.0")
	req.Header.Set("Tec-token", "")
	req.Header.Set("Accept", "application/x.tec.v2+json")
	req.Header.Set("Content-type", "application/json; charset=UTF-8")
	req.SetBasicAuth("public", "alamakota")
	fmt.Println("Begin request")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Request done")
	bodyResp, err := ioutil.ReadAll(resp.Body)
	var p fastjson.Parser
	b, err := p.Parse(string(bodyResp))
	if err != nil {
		log.Fatal(err)
	}
	results := b.GetArray("result")
	trips := []Trip{}
	for _, res := range results {
		tripUpdate := res.Get("tripUpdate")
		hasRealTime := tripUpdate.Get("hasRealtime")
		if hasRealTime.GetBool() {
			trip := Trip{}
			vehicle := tripUpdate.Get("vehiclePosition").Get("vehicle")
			trip.vehicleID = string(vehicle.Get("vehicle").GetStringBytes("id"))
			trip.position.latitude = vehicle.Get("position").GetFloat64("latitude")
			trip.position.longitude = vehicle.Get("position").GetFloat64("longitude")
			if vehicle.Get("position").Exists("speed") {
				trip.speed = vehicle.Get("position").GetFloat64("speed")
			}
			trip.line = string(tripUpdate.Get("trip").GetStringBytes("routeId"))
			trips = append(trips, trip)
		}
	}
	c <- trips
}

func main() {
	db, err := sql.Open("mysql", DB_USER+":"+DB_PASS+"@tcp("+DB_ADDR+")/"+DB_DATABASE)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Query current busses")
	rows, err := db.Query(request)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Current busses acquired")
	defer rows.Close()
	tripIDs := make([]string, 0)
	for rows.Next() {
		var tripID string
		rows.Scan(&tripID)
		tripIDs = append(tripIDs, tripID)
	}

	c := make(chan []Trip)
	nbGoroutines := 0
	for i := 0; i < len(tripIDs); i += 200 {
		if i+200 > len(tripIDs) {
			go queryTrips(tripIDs[i:], c)
		} else {
			go queryTrips(tripIDs[i:i+200], c)
		}
		nbGoroutines++
	}
	trips := []Trip{}
	for i := 0; i < nbGoroutines; i++ {
		tripsSlice := <-c
		trips = append(trips, tripsSlice...)
	}
	fmt.Println("All goroutines are done")
	mapID, _ := randomHex(16)
	var markers string
	for _, bus := range trips {
		marker, _ := createMarker(bus.line, bus.vehicleID, bus.position, mapID)
		markers += marker + "\n"
	}
	busMap, _ := createMap(mapID, markers, Position{50.2296, 5.3586})
	err = ioutil.WriteFile("map.html", []byte(busMap), 0775)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Map created")
}
