package echo_server

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var TemplateFiles embed.FS

var (
	echoFormat    string
	echoText      string
	responseDelay time.Duration
	certPath      string
	keyPath       string
	listenPort    int
	delay         <-chan time.Time
)

// echoInfo is used to store dynamic properties on
// the echo template
type echoInfo struct {
	App     string
	Host    string
	Request string
	Headers http.Header

	BackgroundColor string
}

func getRequest(w http.ResponseWriter, r *http.Request) {

	// Add delay if enabled
	if responseDelay > 0 {
		<-delay
	}

	name, err := os.Hostname() // container ID
	if err != nil {
		fmt.Println("err: ", err)
	}

	// outPut text for debugging on console when running: go run
	outputText := fmt.Sprintf("ECHO Request Server: \n--------------------\n")
	outputText += fmt.Sprintf("App: \n    %s\n", echoText)
	outputText += fmt.Sprintf("Host: \n    %s\n", name)

	headers := r.Header
	outputText += fmt.Sprintf("Request: \n    http://%s%s\n", r.Host, r.RequestURI)
	outputText += fmt.Sprintf("Headers: \n    %s\n", headers)

	backgroundColor := "bg-light"

	if val := r.Header.Get("iscanary"); val == "true" {
		backgroundColor = "bg-primary"
	}

	data := &echoInfo{
		App:             echoText,
		Host:            name,
		Request:         fmt.Sprintf("http://%s%s\n", r.Host, r.RequestURI),
		Headers:         r.Header,
		BackgroundColor: backgroundColor,
	}

	if echoFormat == "text" {
		w.Write([]byte(outputText))
	} else {
		serveTemplate("templates", "echo.html", data, w)
	}

	// Log to stdout
	fmt.Println(outputText)
}

func serveTemplate(tmplDir string, tmplFile string, data interface{}, w http.ResponseWriter) {
	webContent, err := fs.Sub(fs.FS(TemplateFiles), tmplDir)
	if err != nil {
		log.Error(err)
	}
	parseFS, err := template.ParseFS(webContent, tmplFile)
	if err != nil {
		log.Error(err)
	}
	parseFS.Execute(w, data)
}

func init() {
	flag.StringVar(&echoFormat, "format", "text", "'html': output as html or 'text': plaintext on stdout")
	flag.StringVar(&echoText, "echotext", "", "enter text to echo back to the user")
	flag.DurationVar(&responseDelay, "response-delay", 0, "")
	flag.StringVar(&certPath, "cert-path", "", "")
	flag.StringVar(&keyPath, "key-path", "", "")
	flag.IntVar(&listenPort, "listen-port", 8080, "The port used to listen on. Defaults to 8080")
}

func Run() {
	flag.Parse()
	delay = time.Tick(responseDelay)

	http.HandleFunc("/", getRequest)

	fmt.Printf("Server started! Listening on port %q. ", fmt.Sprintf(":%d", listenPort))

	certExists := true
	keyExists := true

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		certExists = false
		fmt.Println("---err: ", err)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		keyExists = false
		fmt.Println("---err: ", err)
	}

	if certExists && keyExists {
		fmt.Println("Serving on HTTPS.")
		http.ListenAndServeTLS(fmt.Sprintf(":%d", listenPort), certPath, keyPath, nil)
	} else {
		fmt.Println("Serving on HTTP.")
		http.ListenAndServe(fmt.Sprintf(":%d", listenPort), nil)
	}
}
