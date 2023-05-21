package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/valyala/fastjson"
	"golang.org/x/exp/maps"
)

var apiURL = "https://tec-api.tapptic.com/api/trips/updates"
var zone01URL = "https://www.zone01.be/hercules/resultaat_uitgebreid"

type Trip struct {
	ID             string `json:"id"`
	BeginTime      string `json:"beginTime"`
	EndTime        string `json:"endTime"`
	RouteID        string `json:"routeID"`
	RouteShortName string `json:"routeShortName"`
	RouteLongName  string `json:"routeLongName"`
}

type LiveTrip struct {
	Trip       Trip    `json:"trip"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	VehiculeID string  `json:"vehiculeID"`
	Speed      float64 `json:"speed"`
}

var liveTrips = []LiveTrip{}

func Database() gin.HandlerFunc {
	db, err := sql.Open("mysql", DB_USER+":"+DB_PASS+"@tcp("+DB_ADDR+")/"+DB_DATABASE)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	return func(c *gin.Context) {
		c.Set("DB", db)
		c.Next()
	}
}

func queryTrips(tripMap map[string]Trip, c chan []LiveTrip) {
	t := time.Now()
	request := map[string]interface{}{
		"complete": true,
		"datetime": t.Format("2006-01-02 15:04:05"),
		"trip_ids": maps.Keys(tripMap),
	}
	strReq, _ := json.Marshal(request)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(strReq))
	req.Header.Set("User-Agent", "okhttp/3.8.0")
	req.Header.Set("Tec-token", "")
	req.Header.Set("Accept", "application/x.tec.v2+json")
	req.Header.Set("Content-type", "application/json; charset=UTF-8")
	req.SetBasicAuth("public", "alamakota")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyResp, err := ioutil.ReadAll(resp.Body)
	var p fastjson.Parser
	b, err := p.Parse(string(bodyResp))
	if err != nil {
		log.Fatal(err)
	}
	results := b.GetArray("result")
	liveTrips := []LiveTrip{}
	for _, res := range results {
		tripUpdate := res.Get("tripUpdate")
		hasRealTime := tripUpdate.Get("hasRealtime")
		if !hasRealTime.GetBool() {
			continue
		}
		livetrip := LiveTrip{}
		vehicle := tripUpdate.Get("vehiclePosition").Get("vehicle")
		livetrip.VehiculeID = string(vehicle.Get("vehicle").GetStringBytes("id"))
		livetrip.Latitude = vehicle.Get("position").GetFloat64("latitude")
		livetrip.Longitude = vehicle.Get("position").GetFloat64("longitude")
		if vehicle.Get("position").Exists("speed") {
			livetrip.Speed = vehicle.Get("position").GetFloat64("speed")
		}
		tripId := string(tripUpdate.Get("trip").GetStringBytes("tripId"))
		livetrip.Trip = tripMap[tripId]
		liveTrips = append(liveTrips, livetrip)
	}
	c <- liveTrips
}

//Provides the list of all trips that should be running
func getTrips(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	db := c.MustGet("DB").(*sql.DB)
	rows, err := db.Query(SQL_REQUEST, 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	trips := make([]Trip, 0)
	for rows.Next() {
		var trip_id string
		var begin_time string
		var end_time string
		var route_id string
		var route_short_name string
		var route_long_name string
		rows.Scan(&trip_id, &begin_time, &end_time, &route_id, &route_short_name, &route_long_name)
		trips = append(trips, Trip{ID: trip_id, BeginTime: begin_time, EndTime: end_time, RouteID: route_id, RouteShortName: route_short_name, RouteLongName: route_long_name})
	}
	c.IndentedJSON(http.StatusOK, trips)
}

//Provides the list of all currently geolocated trips
func getLiveTrips(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	db := c.MustGet("DB").(*sql.DB)
	rows, err := db.Query(SQL_REQUEST, 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	//an array of maps containing Trip
	tripsMapArr := make([]map[string]Trip, 1)
	nTrip := 0
	nSlice := 0
	tripsMapArr[nSlice] = make(map[string]Trip)
	for rows.Next() {
		if nTrip >= 200 {
			nTrip = 0
			nSlice++
			newMap := make(map[string]Trip)
			tripsMapArr = append(tripsMapArr, newMap)
		}
		var trip_id string
		var begin_time string
		var end_time string
		var route_id string
		var route_short_name string
		var route_long_name string
		rows.Scan(&trip_id, &begin_time, &end_time, &route_id, &route_short_name, &route_long_name)
		tripsMapArr[nSlice][trip_id] = Trip{ID: trip_id, BeginTime: begin_time, EndTime: end_time, RouteID: route_id, RouteShortName: route_short_name, RouteLongName: route_long_name}
		nTrip++

	}
	ch := make(chan []LiveTrip)
	nbGoroutines := len(tripsMapArr)
	for i := 0; i < nbGoroutines; i++ {
		go queryTrips(tripsMapArr[i], ch)
	}
	liveTrips := []LiveTrip{}
	for i := 0; i < nbGoroutines; i++ {
		liveTripsSlice := <-ch
		liveTrips = append(liveTrips, liveTripsSlice...)
	}
	c.IndentedJSON(http.StatusOK, liveTrips)
}

func getFindplate(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	busID := c.Param("id")
	form := url.Values{}
	form.Add("owner", "OTW")
	form.Add("busnr", busID)
	form.Add("status", "A")
	form.Add("submit", "Zoeken")
	resp, err := http.PostForm(zone01URL, form)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	table := doc.Find("#herculesdetails")
	if table.Length() == 0 {
		c.AbortWithStatus(404)
		return
	}

	rows := table.Find("tr")
	totalLen := rows.Length()
	rows.EachWithBreak(func(i int, s *goquery.Selection) bool {
		owner := s.Find(".owners").Text()
		busNr := s.Find(".busnr").Text()
		licensePlate := s.Find(".kenteken").Text()
		link, _ := s.Find("a").Attr("href")
		if strings.Contains(owner, "OTW") && busNr == busID {
			c.IndentedJSON(200, map[string]interface{}{
				"license_plate": licensePlate,
				"link":          "https://www.zone01.be/hercules/" + link,
			})
			return false
		}
		//if at end of array then abort with 404
		if totalLen == i+1 {
			c.AbortWithStatus(404)
			return false
		}
		return true
	})
}

func main() {
	gin.SetMode(GIN_MODE)
	router := gin.Default()
	router.SetTrustedProxies(TRUSTED_PROXIES)
	router.Use(Database())
	router.GET("/trips", getTrips)
	router.GET("/livetrips", getLiveTrips)
	router.GET("/findplate/:id", getFindplate)

	router.Run(LISTEN_ADDR + ":" + LISTEN_PORT)
}
