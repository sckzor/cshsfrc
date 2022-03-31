package main

 /*
 \
 /
 \ WARNING TO ANY AND ALL ON-LOOKERS! THIS PROGRAM HAS TERRIBLE SECURITY IT IS MEANT TO STOP 
 / SMALL CHILD SCRIPT KIDDIES FROM DELETING SOME OF MY FILES, NOT ANYONE ELSE IT SHOULD NOT EVER
 \ BE DEPLOYED TO ANYTHING OTHER THAN THAT DEMOGRAPHIC. !!!!!! YOU HAVE BEEN WARNED !!!!!!!
 /
 \
*/


import (
	"html/template"
	"path/filepath"
	"encoding/hex"
	"encoding/csv"
	"crypto/sha1"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"runtime"
	"log"
	"net"
	"fmt"
	"os"
)

//
// Boiler Plate Code
//

const staticDirectory string = "static/"
const templateDirectory string = "templates/"

type protectedFileSystem struct {
	fs http.FileSystem
}

func (nfs protectedFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		return nil, os.ErrNotExist
	}

	return f, nil
}

// Much better hash function
func toSha1(text string) string {
	algorithm := sha1.New()
	algorithm.Write([]byte(text))
	return hex.EncodeToString(algorithm.Sum(nil))
}

//
// Real Code
//

// Data about the "IDE"
type ideData struct {
	Username string
	Code string
	ExecutionStatus string
}

// Data about the admin page
type adminData struct {
	Users []session
	ExecutionStatus string
	RobotIp string
}

// Data about the authentication page
type authData struct {
	Message string
	Location string
}

// Session of each person
type session struct {
	Username string
	Password string
	Actions string
	Id string
}

var sessions []session

var robotIp string = "10.5.48.2:5800"

const actionFiles string = "actionfiles/"

// Function to process loading of the index.html file
func index(w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join(templateDirectory, "index.html")
	tmpl, _ := template.ParseFiles(lp)

	tmpl.ExecuteTemplate(w, "layout", nil);
}

func saveActions(username string, actions string) {
	err := os.WriteFile(actionFiles + username + ".actionfile", []byte(actions), 0666);
	if err != nil {
		log.Fatal(err)
	}
}

// Function to process the loading of the virtual IDE
func ide(w http.ResponseWriter, r *http.Request) {
	var output string = "[None yet]"
	var oldCode = ""
	
	c, err := r.Cookie("session_token")
	
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			http.Redirect(w, r, "/auth/", http.StatusSeeOther)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	lp := filepath.Join(templateDirectory, "ide.html")
	tmpl, _ := template.ParseFiles(lp)

	if r.Method == "POST" {
		for i := 0; i < len(sessions); i++ {
			if c.Value == sessions[i].Id { 

				oldCode = r.FormValue("python")
				out, err := exec.Command("timeout", "0.2",
						"python3", "python/jail.py", oldCode).CombinedOutput()
				
				output = string(out)

				if strings.Contains(fmt.Sprint(err), "exit status 1") {
					output = "Internal error!"
				}

				if strings.Contains(fmt.Sprint(err), "exit status 124") {
					output = "Command Timed Out!\nEnsure that there are no delays or infinite loops in your code.\nIf the are no erroneous code blocks, try again, this could be an intermittant failure." 
				}

				if output == "" {
					output = "Error! Remember to call robot.run() to execute your code!"
				}

				split := strings.Split(output, "=== Action Output ===\n")	
				if len(split) > 1 {
					log.Println(split[1])
					sessions[i].Actions = split[1]
					saveActions(sessions[i].Username, sessions[i].Actions)
				}
			}
		}
	}

	for i := 0; i < len(sessions); i++ {
		if c.Value == sessions[i].Id { 
			var data = ideData{ sessions[i].Username, oldCode, output }
			tmpl.ExecuteTemplate(w, "layout", data);
			return
		}
	}

	http.Redirect(w, r, "/auth?location=ide", http.StatusSeeOther)
}

func about(w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join(templateDirectory, "about.html")
	tmpl, _ := template.ParseFiles(lp)
	tmpl.ExecuteTemplate(w, "layout", nil);
}

// Helper method to send data to the robot
func sendStringTCP(location string, str string) string {
	conn, err := net.Dial("tcp", location)
	if err != nil {
		return fmt.Sprint(err)
	}

	fmt.Fprintf(conn, str)	
	return "Success"
}

// Function to process the loading of the admin page
func admin(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			http.Redirect(w, r, "/auth/", http.StatusSeeOther)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	if r.Method == "POST" && c.Value == sessions[0].Id {
		robotIp = r.FormValue("robot")
	}


	if c.Value == sessions[0].Id { 
		lp := filepath.Join(templateDirectory, "admin.html")
		tmpl, _ := template.ParseFiles(lp)

		users, ok := r.URL.Query()["user"]

		status := ""
	    
		if ok && len(users) == 1 {
			actions, ok := r.URL.Query()["action"]
			if ok && len(users) == 1 {
				for i := 0; i < len(sessions); i++ {
					if users[0] == sessions[i].Username { 
						if actions[0] == "run" {
							log.Println(sessions[i].Username)
							status = sendStringTCP(robotIp, sessions[i].Actions)
							log.Println("Done sending the stuff")
						} else if actions[0] == "delete" {
							sessions[i].Actions = ""
							os.Remove(actionFiles + sessions[i].Username + ".actionfile")
						}
					}
				}
			}
		}

		data := adminData{ sessions, status, robotIp }

		tmpl.ExecuteTemplate(w, "layout", data);

		return
	}

	http.Redirect(w, r, "/auth?location=admin", http.StatusSeeOther)
}

// Function to process the authentication of clients
func auth(w http.ResponseWriter, r *http.Request) {
	message := ""

	if r.Method == "POST" {
		for i := 0; i < len(sessions); i++ {
			if r.FormValue("username") == sessions[i].Username && 
				r.FormValue("password") == sessions[i].Password {
				cookie := &http.Cookie {
					Name: "session_token",
					Path: "/",
					Value: sessions[i].Id,
				}

				http.SetCookie(w, cookie)	
				http.Redirect(w, r, r.FormValue("location"), http.StatusSeeOther)

				return
			}
		}
		message = "Invalid login"
	}

	lp := filepath.Join(templateDirectory, "auth.html")
	tmpl, _ := template.ParseFiles(lp)

	loc := "/"
	locations, ok := r.URL.Query()["location"]
	if ok && len(locations) == 1 {
		loc += locations[0]
	}
	
	data := authData { message, loc }

	tmpl.ExecuteTemplate(w, "layout", data);
	
}

// Function to process the deauthentication of clients
func deauth(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie {
		Name: "session_token",
		Path: "/",
		Value: "0",
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// CSV processing helper to convert it to a matrix
func csvHelper(fileName string) ([][]string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return [][]string{}, err
	}

	defer f.Close()

	r := csv.NewReader(f)

	// skip first line
	if _, err := r.Read(); err != nil {
		return [][]string{}, err
        }
	data, err := r.ReadAll()

	if err != nil {
		return [][]string{}, err
	}

        return data, nil
}

// Load all of the different accounts that are in the csv
func loadDatabaseKeys() []session {
	sessData, err := csvHelper("logins.csv")

	var sess []session

	if err != nil {
		log.Fatal(err)
	}

	for _, sesData := range sessData {
		actions := ""
		content, err := ioutil.ReadFile(actionFiles + sesData[0] + ".actionfile")

		if err == nil {
			actions = string(content)
		}

		ses := session {
			Username: sesData[0],
			Password: sesData[1],
			Id: toSha1(sesData[1]),
			Actions: actions,
		}

		sess = append(sess, ses)
	}
	return sess
}

// Main function to do the things
func main() {
	sessions = loadDatabaseKeys()

	mux := http.NewServeMux()
	sfs := http.FileServer(protectedFileSystem{ http.Dir(staticDirectory) })

	mux.Handle("/static/", http.StripPrefix("/static/", sfs))

	mux.HandleFunc("/", index)
	mux.HandleFunc("/ide/", ide)
	mux.HandleFunc("/auth/", auth)
	mux.HandleFunc("/about/", about)
	mux.HandleFunc("/admin/", admin)
	mux.HandleFunc("/deauth/", deauth)

	if runtime.GOOS != "linux" {
		log.Fatal("Sadly (or not depending on how you look at it), this script will only run on Linux system due to it's reliance on chroot jails for unprivelaged code execution.")
	}

	log.Fatal(http.ListenAndServe(":80", mux))
}
