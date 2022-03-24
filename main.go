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
	"net/http"
	"os/exec"
	"strings"
	"log"
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

// Session of each person
type session struct {
	Username string
	Password string
	Json string
	Id string
}

var sessions []session

// Function to process loading of the index.html file
func index(w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join(templateDirectory, "index.html")
	tmpl, _ := template.ParseFiles(lp)

	tmpl.ExecuteTemplate(w, "layout", nil);
}

// Function to process the loading of the virtual IDE
func ide(w http.ResponseWriter, r *http.Request) {
	var output string = "[None yet]"
	var oldCode = "import robot\n"
	
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
				
				if err != nil {
					log.Println("Python error: ", err)
				}

				output = string(out)

				split := strings.Split(output, "=== JSON Output ===")	
				if len(split) > 1 {
					log.Println(split[1])
					sessions[i].Json = split[1]
				}
			}
		}
	}

	for i := 0; i < len(sessions); i++ {
		if c.Value == sessions[i].Id { 
			var data = ideData{ sessions[i].Username, oldCode, output }
			tmpl.ExecuteTemplate(w, "layout", data);
				var data = ideData{ sessions[i].Username, oldCode, output }
				tmpl.ExecuteTemplate(w, "layout", data);
			return
		}
	}

	http.Redirect(w, r, "/auth/", http.StatusSeeOther)
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
				http.Redirect(w, r, "/ide/", http.StatusSeeOther)
				return
			}
		}
		message = "Invalid login"
	}

	lp := filepath.Join(templateDirectory, "auth.html")
	tmpl, _ := template.ParseFiles(lp)

	tmpl.ExecuteTemplate(w, "layout", message);
	
}

// Function to process the deauthentication of clients
func deauth(w http.ResponseWriter, r *http.Request) {
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
		ses := session {
			Username: sesData[0],
			Password: sesData[1],
			Id: toSha1(sesData[1]),
		}

		sess = append(sess, ses)
	}
	return sess
}


// Main function to do the things
func main() {
	sessions = loadDatabaseKeys()

	f, err := os.OpenFile("main.log", os.O_APPEND | os.O_CREATE | os.O_RDWR, 0666)
	if err != nil {
		log.Printf("Error opening file: %v!", err)
	}

	defer f.Close()

	// log.SetOutput(f)

	mux := http.NewServeMux()
	sfs := http.FileServer(protectedFileSystem{ http.Dir(staticDirectory) })

	mux.Handle("/static/", http.StripPrefix("/static/", sfs))

	mux.HandleFunc("/", index)
	mux.HandleFunc("/ide/", ide)
	mux.HandleFunc("/auth/", auth)
	mux.HandleFunc("/deauth/", deauth)

	log.Fatal(http.ListenAndServe(":8000", mux))
}
