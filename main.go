/* api endpoint:
curl --location --request POST 'http://127.0.0.1:8000/bid' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer test1234' \
--data-raw '{
 "id": "id_123",
 "width": 600,
 "height": 328,
 "banner": {
 "type": 1
 }
}'
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

// BidRequest represents the incoming bid request payload
type BidRequest struct {
	ID     string `json:"id"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Banner struct {
		Type int `json:"type"`
	} `json:"banner"`
}

// HandleBid handles the POST /bid endpoint
func HandleBid(w http.ResponseWriter, r *http.Request) {
	// check authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader != "Bearer test1234" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// decode the JSON body
	var req BidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// fetch data from redis
	redisResp := getDataFromRedis(req)
	if len(redisResp) == 0 || redisResp == "" {
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
	}

	// set the content type and write the response
	if req.Banner.Type == 1 {
		w.Header().Set("Content-Type", "text/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(redisResp))
	} else if req.Banner.Type == 2 {
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(redisResp))
	}

}

// getDataFromRedis to get data from redis
func getDataFromRedis(req BidRequest) (resp string) {
	// redis client initialisation
	redisClient := redis.NewClient(
		&redis.Options{
			Addr:        "",
			Password:    "",
			IdleTimeout: 180 * time.Second,
			MaxConnAge:  10 * time.Second,
			DialTimeout: 1 * time.Second,
			MaxRetries:  5,
		},
	)

	// get redis hash key
	redisHGet := redisClient.HGetAll(context.Background(), req.ID)
	data := redisHGet.Val()
	if len(data) == 0 {
		return
	}
	if req.Banner.Type == 1 {
		// get redis js key
		dataJS := redisClient.Get(context.Background(), req.ID+"_js").Val()
		// decode js data
		resp = strings.ReplaceAll(dataJS, "{click}", data["click"])
		resp = strings.ReplaceAll(resp, "{impression}", data["impression"])

	} else if req.Banner.Type == 2 {
		// get redis xml key
		dataXML := redisClient.Get(context.Background(), req.ID+"_xml").Val()
		// decode xml key
		resp = strings.ReplaceAll(dataXML, "{click}", data["click"])
		resp = strings.ReplaceAll(resp, "{impression}", data["impression"])
		resp = strings.ReplaceAll(resp, "{video_url}", data["video_url"])
		resp = strings.ReplaceAll(resp, "{video_start}", data["video_start"])
		resp = strings.ReplaceAll(resp, "{video_end}", data["video_end"])
	}
	return
}

func main() {
	// init the mux router
	router := mux.NewRouter()
	router.HandleFunc("/bid", HandleBid).Methods("POST")
	// serve the app at port 8000
	fmt.Println("Server at 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
