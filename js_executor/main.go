package main

import (
	"os"
	"io"
	"flag"
	"time"
	"http"
	"log"
	"rand"
	"exec"
	"fmt"
	"strings"
	"./appconfig"
	"github.com/garyburd/go-mongo"
)

const (
	Second = 1000000000
)

/**
 * RemoteJS Execute Server.
 * 
 * # Initialize
 *  - Display setting.
 *  - virtual display start.
 *  - web browser start.
 * # Main Process
 *  - open url.
 *  - javascript execution.
 * # End Process
 *  - kill firefox.
 *  - kill virtual display.
 */
type WorkingBox struct {
	DisplayNo int
	Workings int
	LastExecId mongo.ObjectId
	Firefox *exec.Cmd
	ExecCount int
	Conn mongo.Conn
}
type ExecuteRs struct {
    Id mongo.ObjectId `bson:"_id"`
    Url string `bson:"url"`
    Js string `bson:"js"`
    Json string `bson:"json"`
}
// for Global
var workingBoxes = map[int] *WorkingBox {}
var appConfig appconfig.AppConfig
var sem chan int

/**
  Get enable display number of virtual screen.
 */
func GetDisplay(execId mongo.ObjectId) (int) {
	for i := 0; i < appConfig.MaxVirtualDesktop * 2; i++ {
		display := rand.Intn(appConfig.MaxVirtualDesktop) + 1
		workingBox := workingBoxes[display]
		if workingBox.Workings <= appConfig.ParallelExecCount {
			workingBox.Workings += 1
			workingBox.LastExecId = execId
			return display
		}
		log.Printf("Display is %d, working execute ID.\n", display)
	}

	time.Sleep(Second)
	return GetDisplay(execId)
}

/**
  Get image byte array of given url web site.
 */
func ExecuteJS(url string, js string) []byte {
	log.Print("ExecuteJS begin")
	sem <- 1    // Wait a inactive queue.

	conn, err := GetConnection()

	// Register url and js
	execId, err := RegisterExecuteJS(conn, url, js)
	if err != nil {
		log.Fatal("Can't register execute js.")
	}

	// Setup execute display
	display := GetDisplay(execId)

	// Execute JS at firefox on xfvb.
	environ := os.Environ()
	environ = append(environ, fmt.Sprintf("DISPLAY=:%d.0", display))
	command := appConfig.FirefoxBin
	url = AppendExecIdUrl(url, execId)
	args := []string {command, "-display", fmt.Sprintf(":%d", display), "-remote", fmt.Sprintf("openUrl(%s)", url), "-P", fmt.Sprintf("%s%d", appConfig.ProfileName, display)}
	RunCommand(command, args, environ, nil)
	log.Printf("OpenURL DisplayNo: %d, URL: %s", display, url);

	// Waiting for registered result json.
	json := GetExecutedJS(conn, execId, 1)
	workingBoxes[display].ExecCount += 1 // Increment executed count. 
	log.Printf("DisplayNo: %d ExecCount: %d\n", display, workingBoxes[display].ExecCount)

	workingBoxes[display].Workings -= 1
	conn.Close()

	<-sem // Release
	log.Print("ExecuteJS end")
	return json
}

func AppendExecIdUrl(url string, execId mongo.ObjectId) string {
	sep := "&"
	if strings.Index(url, "?") == -1  {
		url = url + "?"
		sep = ""
	}

	url = fmt.Sprintf("%s%s__eid=%s", url, sep, execId)
	log.Printf("AppendExecIdUrl=%s", url)
	return url
}

func RegisterExecuteJS(conn mongo.Conn, url string, js string) (mongo.ObjectId, os.Error) {
	c := GetExecuteCollection(conn)
	id := mongo.NewObjectId()
	err := c.Insert(&ExecuteRs{Id: id, Url: url, Js: js})
	return id, err
}

func GetExecutedJS(conn mongo.Conn, execId mongo.ObjectId, retry int) []byte {
	if retry > appConfig.MaxRetryCount {
		fmt.Printf("ERROR: failed to get result json(%q)\n", execId)
		return []byte{}
	}

	c := GetExecuteCollection(conn)
	var rs ExecuteRs
	err := c.Find(map[string]interface{}{"_id": execId}).One(&rs);
	//err := c.Find(ExecuteRs{Id: execId}).One(&rs)
	if err != nil || rs.Json == "" {
		fmt.Printf("INFO: ExecID=%q waiting...\n", execId)
		time.Sleep(Second)
		return GetExecutedJS(conn, execId, retry + 1)
	}
	return []byte(rs.Json)
}

func KillFirefox(display int) {
	if (workingBoxes[display].Firefox != nil) {
		err := workingBoxes[display].Firefox.Process.Kill()
		if err != nil {
			log.Fatal(err)
			log.Fatal("failed to kill process.")
		}
	}
}

func RunFirefox(display int, workingBox *WorkingBox) {
	environ := os.Environ()
	environ = append(environ, fmt.Sprintf("DISPLAY=:%d.0", display))
	go func (d int, env []string) {
		command := appConfig.FirefoxBin
		args := []string {command, "-display", fmt.Sprintf(":%d", display), "-width", "1024", "-height", "800", "-P", fmt.Sprintf("%s%d", appConfig.ProfileName, display)}
		RunCommand(command, args, env, workingBox)
	}(display, environ)
	time.Sleep(Second * 3)
}

func InitVirtualScreen() {
	log.Println(">>>>> InitVirtualScreen")
	for i := 0; i < appConfig.MaxVirtualDesktop; i++ {
		display := i + 1
		environ := os.Environ()
		environ = append(environ, fmt.Sprintf("DISPLAY=:%d.0", i + 1))

		workingBox := &WorkingBox{DisplayNo: display, Workings: 0, LastExecId: ""}
		workingBoxes[display] = workingBox

		// Run Xvfb
		go func (d int, env []string) {
			command := "/usr/bin/Xvfb"
			args := []string {command, fmt.Sprintf(":%d", d), "-screen", "0", "1024x768x24"}
			RunCommand(command, args, env, nil)
		}(display, environ)
		time.Sleep(Second * 3)

		// Run Firefox
		RunFirefox(display, workingBox);
	}
	rand.Seed(time.Nanoseconds() % 1e9)
	log.Println("<<<<< InitVirtualScreen")
}

func RunCommand(command string, args []string, environ []string, workingBox *WorkingBox) {
	cmd := exec.Command(command)
	cmd.Env = environ
	cmd.Args = args
	cmd.Dir = "."
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("failed to retrieve pipe. %s", err)
		os.Exit(-1)
	}

	if (workingBox != nil) {
		workingBox.Firefox = cmd
	}
	log.Printf("Run [%s]", command)
	err = cmd.Start()
	if err != nil {
		log.Fatal("failed to execute external command. %s", err)
		os.Exit(-1)
	}

	WriteFileLines(stdout)
}

func WriteFileLines(reader io.Reader) {
	var (
		err os.Error
		n int
	)
	buf := make([]byte, 1024)

	log.Println("WriteFileLines");
	for {
		if n, err = reader.Read(buf); err != nil {
			break
		}
		log.Print(string(buf[0:n]))
	}
	if err == os.EOF {
		log.Println("stdout end");
		err = nil
	} else {
		log.Println("ERROR: " + err.String());
	}
}

// hello world, the web server
func PageExecuteJS(w http.ResponseWriter, req *http.Request) {
	url := req.FormValue("url")
	js := req.FormValue("js")
	header := w.Header()
	if url == "" {
		log.Printf("ERROR: url is required. (%s)\n", url)
		w.WriteHeader(http.StatusInternalServerError)
		header.Set("Content-Type", "text/plian;charset=UTF-8;")
		io.WriteString(w, "Internal Server Error: please input url.\n")
		return
	}
	if strings.Index(url, "http") != 0 {
		log.Printf("ERROR: url is invalid. (%s)\n", url)
		w.WriteHeader(http.StatusInternalServerError)
		header.Set("Content-Type", "text/plian;charset=UTF-8;")
		io.WriteString(w, "Internal Server Error: please input valid url.\n")
		return
	}
	if js == "" {
		log.Printf("ERROR: js is required. (%s)\n", url)
		w.WriteHeader(http.StatusInternalServerError)
		header.Set("Content-Type", "text/plian;charset=UTF-8;")
		io.WriteString(w, "Internal Server Error: please input js.\n")
		return
	}

	json := ExecuteJS(url, js)
	if len(json) == 0 {
		log.Printf("ERROR: failed to execute. (%s)\n", url)
		json = []byte("{}")
	}
	header.Set("Content-Type", "application/json")
	w.Write(json)
}

func PageInternalJs(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	header := w.Header()
	conn, err := GetConnection()
	c := GetExecuteCollection(conn)
	var (
		rs ExecuteRs
		jsonb []byte
	)
	execId, err := mongo.NewObjectIdHex(id)
	log.Printf("Search ExecuteId=%s", execId)
	err = c.Find(map[string]interface{}{"_id": execId}).One(&rs);
	if err != nil {
		jsonb = []byte(`{error: "Not found."}`)
	} else {
		jsonb = []byte(rs.Js)
	}

	conn.Close();

	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write(jsonb)
}

func PageInternalUpdateJson(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	updateJson := req.FormValue("json")
	header := w.Header()
	conn, err := GetConnection()
	c := GetExecuteCollection(conn)
	var (
		rs ExecuteRs
		jsonb []byte
	)
	execId, err := mongo.NewObjectIdHex(id)
	log.Printf("Search ExecuteId=%s", execId)
	err = c.Find(map[string]interface{}{"_id": execId}).One(&rs);
	if err != nil {
		jsonb = []byte(`{error: "Not found."}`)
	} else {
		rs.Json = updateJson;
		err = c.Update(mongo.M{"_id": execId}, rs)
		if err != nil {
			jsonb = []byte(`{result: "failed"}`)
		} else {
			jsonb = []byte(`{result: "success"}`)
		}
	}

	conn.Close()

	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Content-Type", "application/json")
	w.Write(jsonb)
}

func GetConnection() (mongo.Conn, os.Error) {
	conn, err := mongo.Dial(appConfig.DbHost)
	return conn, err
}

func GetExecuteCollection(conn mongo.Conn) (mongo.Collection) {
	return mongo.Collection{conn, fmt.Sprintf("%s.executes", appConfig.DbName), mongo.DefaultLastErrorCmd}
}

func init() {
	var (
		configFilename string
		err os.Error
	)
	flag.StringVar(&configFilename, "f", "./appconfig.conf", "config file name")
	appConfig, err = appconfig.Parse(configFilename)
	if err != nil {
		fmt.Println("ERROR: failed to load config.")
		os.Exit(-1)
	}

	// Parallel executing channel
	sem = make(chan int, appConfig.ParallelExecCount)
 }

func main() {
	var err os.Error

	// Virtual Screen
	InitVirtualScreen()

	portNo := 1975
	http.HandleFunc("/execute_js", PageExecuteJS)
	http.HandleFunc("/internal/js", PageInternalJs)
	http.HandleFunc("/internal/update_json", PageInternalUpdateJson)
	err = http.ListenAndServe(fmt.Sprintf(":%d", portNo), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.String())
	}
	log.Printf("INFO: start server on %d\n", portNo)
}
