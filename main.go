package main

// todo add text cleanup (start with cleaning up paragraphs to one line)
import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type TReceive struct {
	Error   string
	Meta    map[string]string
	Html    string
	Text    string
	Cleaned string
	Files   []string
}

type Configuration struct {
	AppDir  string
	TmpDir  string
	TikaApp string
	Java    string
}

var configFile = ""
var conf = Configuration{}
var templates = &template.Template{}

const tikaMeta = "--metadata"
const tikaHTML = "--html"
const tikaText = "--text"
const tikaExtract = "--extract"

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file") // the FormFile function takes in the POST input id file
	defer file.Close()
	m := new(TReceive)

	if err != nil {
		m.Error = m.Error + err.Error()
	}

	_, err = os.Stat(conf.TmpDir)
	if os.IsNotExist(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	out, err := ioutil.TempFile(conf.TmpDir, "sally")
	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		m.Error = m.Error + err.Error()
	}

	t, err := ioutil.TempDir(conf.TmpDir, filepath.Base(out.Name())+"-extract")
	if err != nil {
		m.Error = m.Error + err.Error()
	}

	cmdMeta := exec.Command(conf.Java, "-jar", conf.TikaApp, tikaMeta, out.Name())
	cmdHTML := exec.Command(conf.Java, "-jar", conf.TikaApp, tikaHTML, out.Name())
	cmdText := exec.Command(conf.Java, "-jar", conf.TikaApp, tikaText, out.Name())
	cmdExtract := exec.Command(conf.Java, "-jar", conf.TikaApp, tikaExtract, "--extract-dir="+t, out.Name())

	text, err := cmdMeta.Output()
	re := regexp.MustCompile(`(.*): (.*)\n`)
	meta := re.FindAllStringSubmatch(fmt.Sprintf("%s", text), -1)
	if meta != nil {
		m.Meta = make(map[string]string)
		for _, v := range meta {
			m.Meta[v[1]] = v[2]
		}
	}

	text, err = cmdHTML.Output()
	if err == nil {
		m.Html = fmt.Sprintf("%s", text)
	} else {
		m.Html = err.Error()
	}

	text, err = cmdText.Output()
	if err == nil {
		m.Text = fmt.Sprintf("%s", text)
		re := regexp.MustCompile(`\n([^\s])`)
		// Replace newlines followed by non-whitespace with just a space
		m.Cleaned = re.ReplaceAllString(m.Text, " $1")
	} else {
		m.Text = err.Error()
	}

	text, err = cmdExtract.Output()
	if err == nil {
		re := regexp.MustCompile(`Extracting '[^']+' \([^)]+\) to ([^\n]*)\n`)

		// TODO add go routine to watch & cleanup directory
		b := re.FindAllStringSubmatch(fmt.Sprintf("%s", text), -1)

		if b != nil {
			m.Files = make([]string, len(b))
			for i, v := range b {
				m.Files[i] = "/tmp" + strings.Replace(v[1], conf.TmpDir, "", -1)
			}
		}

	}
	err = templates.ExecuteTemplate(w, "receive.html", m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func serveSingle(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	})
}

func main() {
	conf = config()
	sanityCheck()

	http.HandleFunc("/receive", uploadHandler)

	http.Handle("/static/", http.FileServer(http.Dir(conf.AppDir)))
	http.Handle("/tmp/", http.StripPrefix("/tmp/", http.FileServer(http.Dir(conf.TmpDir))))

	serveSingle("/", conf.AppDir+"/index.html")

	http.ListenAndServe(":8080", nil)
}

func sanityCheck() {
	// TODO check error value
	templates, _ = template.ParseFiles(conf.AppDir+"/index.html", conf.AppDir+"/receive.html")
}

func config() Configuration {
	flag.StringVar(&configFile, "config", "", "Path to configuration file to use")
	flag.Parse()

	file, err := os.Open(configFile)
	if err != nil {
		log.Fatal(err)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}

	// Set defaults for empty values
	if conf.AppDir == "" {
		conf.AppDir = os.Getenv("GOPATH") + "/src/github.com/FunnyMonkey/sally-tika"
	}
	if conf.TikaApp == "" {
		conf.TikaApp = os.Getenv("GOPATH") + "/src/github.com/FunnyMonkey/sally-tika/tika-app-1.5.jar"
	}
	if conf.TmpDir == "" {
		conf.TmpDir = os.TempDir()
	}
	if conf.Java == "" {
		conf.Java = "/usr/bin/java"
	}

	_, err = os.Stat(conf.AppDir)
	if err != nil {
		log.Fatal(err)
	}

	_, err = os.Stat(conf.TikaApp)
	if err != nil {
		log.Fatal(err)
	}

	_, err = os.Stat(conf.TmpDir)
	if err != nil {
		log.Fatal(err)
	}

	_, err = os.Stat(conf.Java)
	if err != nil {
		log.Fatal(err)
	}

	return conf
}
