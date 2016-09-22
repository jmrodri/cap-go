package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/cdrage/atomicapp-go/nulecule"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func getAnswers(nulecule_path string) map[string]nulecule.Answers {
	fmt.Println("path: " + nulecule_path)
	base := nulecule.New("nulecule-library/"+nulecule_path, "", false)
	err := base.ReadMainFile()
	if err != nil {
		fmt.Println("error reading nulecule", err)
	}

	err = base.LoadAnswers()
	if err != nil {
		fmt.Println("error loading answerse", err)
	}

	j, _ := json.Marshal(base)
	fmt.Println("nulecule: ", string(j))
	return base.AnswersData
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

	json.NewEncoder(w).Encode(getNuleculeList())
}

func NuleculeDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nulecule_id := vars["id"]
	json.NewEncoder(w).Encode(getAnswers(nulecule_id))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/nulecules", Nulecules)
	r.HandleFunc("/nulecules/{id}", NuleculeDetails)
	log.Fatal(http.ListenAndServe(":3001", handlers.CORS()(r)))
}
