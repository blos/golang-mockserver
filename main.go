package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"log"
	"net/http"
	"math/rand/v2"
	"time"

	"github.com/gin-gonic/gin"
)

type MockServerConfig struct {
	RoutesFile string
	Host       string
	Port       string
	Verbose    bool
}
type ArbitraryJSON = map[string]interface{}
type HeaderType = map[string]string
type MockEntries []MockEntry
type MockMethodDefinition func(*gin.Engine, MockEntry)
type MockMethodDefinitions map[string]MockMethodDefinition
type MockEntry struct {
	Path               string        `json:"path,required"`
	Method             string        `json:"method,required"`
	StatusCode         int           `json:"status_code,omitempty"`
	Header             HeaderType    `json:"header,omitempty"`
	Body               ArbitraryJSON `json:"body,omitempty"`
	ResponseOffsetMode string        `json:"response_offset_mode,omitempty"`
	ResponseOffset     interface{}   `json:"response_offset,omitempty"`
}
type WaitTimeModeType map[string]func(MockEntry) float64

var WaitTimeModes WaitTimeModeType = WaitTimeModeType{
	"":         ConstantWaitTime, // Default case
	"constant": ConstantWaitTime,
	"normal":   NormalDistWaitTime,
}
var AllowedHTTPMethods MockMethodDefinitions = MockMethodDefinitions{
	"DELETE": NewDELETERoute,
	"GET":    NewGETRoute,
	"PATCH":  NewPATCHRoute,
	"POST":   NewPOSTRoute,
	"PUT":    NewPUTRoute,
}

func main() {
	config := GetCLIParameters()
	log.Println("config: ", config)
	log.Println("Starting server...")

	router := gin.Default()

	ParseConfig(config, router)

	baseURL := BuildBaseURL(config)
	srv := &http.Server{
		Addr:    baseURL,
		Handler: router,
	}

	srv.ListenAndServe()
}

func GetEnv(key string, fallback interface{}) interface{} {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func ReadConfigFile(fileName string) MockEntries {
	jsonFile, err := os.OpenFile(fileName, os.O_RDONLY, 0755)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	defer jsonFile.Close()
	var mockEntries MockEntries
	byteValues, _ := io.ReadAll(jsonFile)
	json.Unmarshal(byteValues, &mockEntries)
	return mockEntries
}

func GetCLIParameters() MockServerConfig {
	config := MockServerConfig{
		RoutesFile: "",
		Host:       "",
		Port:       "",
		Verbose:     false,
	}
	// get values from CLI
	flag.StringVar(&config.RoutesFile, "routes-file", "./example.routes.json", "Path to the route definitions")
	flag.StringVar(&config.Host, "host", "0.0.0.0", "Host address the server listens to")
	flag.StringVar(&config.Port, "port", "1080", "Port the server listens to")
	flag.BoolVar(&config.Verbose, "verbose", false, "Makes the logging of the mockserver verbose")
	flag.Parse()
	// override CLI params with environment vars. if they are set
	config.Host = GetEnv("MOCK_SERVER_HOST", config.Host).(string)
	config.Port = GetEnv("MOCK_SERVER_PORT", config.Port).(string)
	config.Verbose = GetEnv("MOCK_SERVER_VERBOSE", config.Verbose).(bool)
	config.RoutesFile = GetEnv("MOCK_ROUTES_FILE", config.RoutesFile).(string)

	return config
}

func BuildBaseURL(config MockServerConfig) string {
	baseURL := fmt.Sprintf("%s:%s", config.Host, config.Port)
	return baseURL
}

func SetHeader(context *gin.Context, entries HeaderType) {
	for key, value := range entries {
		context.Header(key, value)
	}
}

func HandlerFunction(mockentry MockEntry) func(*gin.Context) {
	handler := func(context *gin.Context) {
		// sets the response header
		SetHeader(context, mockentry.Header)
		// sets the response body
		context.JSON(mockentry.StatusCode, mockentry.Body)
		// emulates wait time for timeouts
		HandleWaitTime(mockentry)
		context.Done()
	}
	return handler
}

func ParseConfig(config MockServerConfig, router *gin.Engine) {
	routesJSONFileContents := ReadConfigFile(config.RoutesFile)
	for _, entry := range routesJSONFileContents {

		if config.Verbose {
			PrintRouteConfig(entry)
		}

		allowedMethodFunction, ok := AllowedHTTPMethods[entry.Method]
		if !ok {
			continue
		}
		allowedMethodFunction(router, entry)
	}
}

func PrintRouteConfig(entry MockEntry) {
	log.Println("-------------------")
	log.Println("Path:", entry.Path)
	log.Println("Method:", entry.Method)
	log.Println("Body:", entry.Body)
	log.Println("StatusCode:", entry.StatusCode)
	log.Println("Header:", entry.Header)
	log.Println("ResponseOffsetMode:", entry.ResponseOffsetMode)
	log.Println("ResponseOffset:", entry.ResponseOffset)
	log.Println("-------------------")
}

func NewGETRoute(router *gin.Engine, mockentry MockEntry) {
	router.GET(mockentry.Path, HandlerFunction(mockentry))
}

func NewPUTRoute(router *gin.Engine, mockentry MockEntry) {
	router.PUT(mockentry.Path, HandlerFunction(mockentry))
}

func NewPOSTRoute(router *gin.Engine, mockentry MockEntry) {
	router.POST(mockentry.Path, HandlerFunction(mockentry))
}

func NewPATCHRoute(router *gin.Engine, mockentry MockEntry) {
	router.PATCH(mockentry.Path, HandlerFunction(mockentry))
}

func NewDELETERoute(router *gin.Engine, mockentry MockEntry) {
	router.DELETE(mockentry.Path, HandlerFunction(mockentry))
}

func ConstantWaitTime(entry MockEntry) float64 {
	return entry.ResponseOffset.(float64)
}

func NormalDistWaitTime(entry MockEntry) float64 {
	s, ok := entry.ResponseOffset.(map[string]interface{})
	if !ok {
		log.Fatal("Expected type map[string]interface{}; got ", entry.ResponseOffset)
		return float64(0)
	}
	mean := s["mean"].(float64)
	std := s["std"].(float64)
	return (rand.Float64() * float64(std)) + float64(mean)
}

func HandleWaitTime(entry MockEntry) {
	waitTimeFunction, ok := WaitTimeModes[entry.ResponseOffsetMode]
	if !ok {
		log.Println("No waitTimeFunction found, check your spelling.")
		return
	}
	var waitTime float64 = waitTimeFunction(entry)
	if waitTime < 0 {
		waitTime = 0
	}
	time.Sleep(time.Duration(waitTime) * time.Millisecond)
}
