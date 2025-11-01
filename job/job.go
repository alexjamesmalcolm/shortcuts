package job

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Executer[O any] interface {
	Execute() (O, error)
}

func MakeJob[Input Executer[Output], Output any](path string) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotImplemented)
			fmt.Fprintf(w, "Not Implemented: %v", r.Method)
			return
		}
		var input Input
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
				log.Println(err)
			}
			log.Printf("unexpected error while parsing JSON: %v\n", err)
			return
		}
		result, err := input.Execute()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
				log.Println(err)
			}
			return
		}
		jsonStringResult, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
				log.Println(err)
			}
			return
		}
		_, err = w.Write(jsonStringResult)
		if err != nil {
			log.Println(err)
		}
	})
}
