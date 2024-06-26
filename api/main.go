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
var zone01URL = "https://www.zone01.be/hercules/resultaten"

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
		log.Print(err)
		return
	}
	bodyResp, err := ioutil.ReadAll(resp.Body)
	var p fastjson.Parser
	b, err := p.Parse(string(bodyResp))
	if err != nil {
		log.Print(err)
		return
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
	sqlargs := []interface{}{0, 0}
	line := c.Query("line")
	_SQL_REQUEST := SQL_REQUEST
	if line != "" {
		_SQL_REQUEST += " WHERE r.route_id LIKE ?"
		sqlargs = append(sqlargs, line+"%")
	}
	rows, err := db.Query(_SQL_REQUEST, sqlargs...)
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(500)
		return
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

//provides list of points to draw the shape of the trip
func getTripShape(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	tripID := c.Param("trip_id")
	db := c.MustGet("DB").(*sql.DB)
	rows, err := db.Query("SELECT shape_pt_lat,shape_pt_long FROM shapes WHERE shape_id = (SELECT shape_id from trips WHERE trip_id = ?) ORDER BY shape_pt_sequence;", tripID)
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(500)
		return
	}
	defer rows.Close()
	pt_couple_arr := make([][]string, 0)
	for rows.Next() {
		var lat string
		var long string
		rows.Scan(&lat, &long)
		point := make([]string, 2)
		point[0] = lat
		point[1] = long
		pt_couple_arr = append(pt_couple_arr, point)
	}
	c.IndentedJSON(200, pt_couple_arr)
}

//Provides the list of all currently geolocated trips
func getLiveTrips(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	db := c.MustGet("DB").(*sql.DB)
	sqlargs := []interface{}{0, 0}
	line := c.Query("line")
	_SQL_REQUEST := SQL_REQUEST
	if line != "" {
		_SQL_REQUEST += " WHERE r.route_id LIKE ?"
		sqlargs = append(sqlargs, line+"%")
	}
	rows, err := db.Query(_SQL_REQUEST, sqlargs...)
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(500)
		return
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

//Provides the list of all currently geolocated trips
func getUpdateTrip(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	tripID := c.Param("trip_id")
	ch := make(chan []LiveTrip)
	tripsMap := make(map[string]Trip)
	tripsMap[tripID] = Trip{ID: tripID, BeginTime: "", EndTime: "", RouteID: "", RouteShortName: "", RouteLongName: ""}
	go queryTrips(tripsMap, ch)
	livetrip := <-ch
	if len(livetrip) == 0 {
		c.AbortWithStatus(404)
		return
	}
	c.IndentedJSON(http.StatusOK, livetrip[0])
}

//returns infos about the vehicule ID requested. Uses zone01.be.
func getFindplate(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	busID := c.Param("id")
	form := url.Values{}
	form.Add("q", busID)
	req, err := http.NewRequest("GET", zone01URL, nil)
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(500)
		return
	}
	query := req.URL.Query()
	if len(busID) > 4 {
		query.Add("q", busID)
	} else {
		query.Add("q", "OTW "+busID)
	}

	req.URL.RawQuery = query.Encode()

	resp, _ := http.DefaultClient.Do(req)

	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(500)
		return
	}

	table := doc.Find("#herculesdetails")
	if table.Length() == 0 {
		c.AbortWithStatus(404)
		return
	}

	//Loop on the lines containing busses and each of their affectation
	table.EachWithBreak(func(i int, s *goquery.Selection) bool {
		vehicle := s
		affectations := vehicle.Find("tr")
		found := false
		affectations.EachWithBreak(func(i2 int, s2 *goquery.Selection) bool {
			owner := s2.Find(".owners").Text()
			busNr := s2.Find(".busnr").Text()
			licensePlate := s2.Find(".kenteken").Text()
			link, _ := s2.Find("a").Attr("href")
			status, _ := s2.Find("img").Attr("src")
			if !strings.Contains(status, "/A.png") {
				return true //Affectation is not in service anymore
			}
			if busNr != busID {
				return true //Bus ID is not correct
			}
			if strings.Contains(strings.ToLower(owner), "otw") {
				found = true
				c.IndentedJSON(200, map[string]interface{}{
					"license_plate": licensePlate,
					"link":          "https://www.zone01.be/hercules/" + link,
				})
				return false
			}
			for j := 0; j < len(BAD_OWNERS); j++ {
				if strings.Contains(strings.ToLower(owner), strings.ToLower(BAD_OWNERS[j])) {
					return true // Bus belongs to a company which we know does not contract for OTW.
				}
			}
			found = true
			c.IndentedJSON(200, map[string]interface{}{
				"license_plate": licensePlate,
				"link":          "https://www.zone01.be/hercules/" + link,
			})
			return false
		})
		return !found
	})
}

func main() {
	gin.SetMode(GIN_MODE)
	router := gin.Default()
	router.SetTrustedProxies(TRUSTED_PROXIES)
	router.Use(Database())
	router.GET("/trips", getTrips)
	router.GET("/livetrips", getLiveTrips)
	router.GET("/updatetrip/:trip_id", getUpdateTrip)
	router.GET("/findplate/:id", getFindplate)
	router.GET("/tripshape/:trip_id", getTripShape)

	router.Run(LISTEN_ADDR + ":" + LISTEN_PORT)
}
