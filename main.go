package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/cdrage/atomicapp-go/nulecule"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const MAIN_FILE = "Nulecule"

type Answers map[string]map[string]string

/*
func loadFromPath(src string) {
	nuleculePath := filepath.Join(src, MAIN_FILE)

	// []byte
	nuleculeData, err := ioutil.ReadFile(nuleculePath)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(nuleculeData, )
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(nuleculeData)
}
*/

func runCommand(cmd string, args ...string) []byte {
	output, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		fmt.Println("Error running " + cmd)
	}
	return output
}

func getAnswers(nulecule_path string) map[string]nulecule.Answers {
	fmt.Println("path: " + nulecule_path)
	base := nulecule.New("nulecule-library/"+nulecule_path, "", false)
	err := base.ReadMainFile()
	if err != nil {
		fmt.Println("error reading nulecule", err)
	}

	base.AnswersDirectory = "nulecule-library/" + nulecule_path
	err = base.LoadAnswers()
	if err != nil {
		fmt.Println("error loading answerse", err)
	}

	j, _ := json.Marshal(base)
	fmt.Println("nulecule: ", string(j))
	return base.AnswersData
}

// returns a map of maps
func parseBasicINI(data string) map[string]map[string]string {
	/*
		find first [ then find matching ]. Everything between them is the first key. Read until next [ or end of string.
	*/
	var answers = make(map[string]map[string]string)
	values := strings.SplitAfter(data, "\n")
	var key string
	for _, str := range values {
		if strings.HasPrefix(str, "[") {
			key = strings.Trim(str, "[]\n")
			answers[key] = make(map[string]string)
		} else {
			subvalue := strings.Split(str, " = ")
			answers[key][subvalue[0]] = strings.Trim(subvalue[1], "\n")
		}
	}

	fmt.Println(answers)
	return answers
}

func getAnswersFromFile(nulecule_path string) map[string]Answers {
	os.Remove("answers.conf")
	/*
		output, err := exec.Command("atomicapp", "genanswers", "nulecule-library/"+nulecule_path).CombinedOutput()
		if err != nil {
			fmt.Println("Error running atomicapp")
		}
	*/
	output := runCommand("atomicapp", "genanswers", "nulecule-library/"+nulecule_path)
	fmt.Println(string(output))
	answers, err := ioutil.ReadFile("answers.conf")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(answers))
	// add root node
	return map[string]Answers{"nulecule": parseBasicINI(string(answers))}
}

func getNuleculeList() map[string][]string {
	files, _ := ioutil.ReadDir("./nulecule-library")
	nulecules := make([]string, 0)
	for _, f := range files {
		if f.IsDir() {
			nulecules = append(nulecules, f.Name())
		}
	}
	return map[string][]string{"nulecules": nulecules}
}

func Nulecules(w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("Gorilla!\n"))
	fmt.Println("Getting nulecules")

	json.NewEncoder(w).Encode(getNuleculeList())
}

func NuleculeDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nulecule_id := vars["id"]
	json.NewEncoder(w).Encode(getAnswersFromFile(nulecule_id))
}

func NuleculeUpdate(w http.ResponseWriter, r *http.Request) {
	// update the nulecule answers file
	vars := mux.Vars(r)
	nulecule_id := vars["id"]
	json.NewEncoder(w).Encode(getAnswersFromFile(nulecule_id))
}

func NuleculeRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nulecule_id := vars["id"]
	output := runCommand("atomic", "run", "nulecule-library/"+nulecule_id)
	fmt.Println(string(output))
	json.NewEncoder(w).Encode(string(output))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/nulecules", Nulecules)
	r.HandleFunc("/nulecules/{id}", NuleculeDetails).Methods("GET")
	r.HandleFunc("/nulecules/{id}", NuleculeUpdate).Methods("POST")
	r.HandleFunc("/nulecules/{id}/deploy", NuleculeRun).Methods("POST")
	log.Fatal(http.ListenAndServe(":3001", handlers.CORS()(r)))
}
