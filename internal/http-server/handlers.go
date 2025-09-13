package http_server

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sandbox-dev-test-task/internal/models"
	"sandbox-dev-test-task/internal/pool"
	"time"
)

const (
	ErrServerOverload = "Internal http-server error. Try again later"
	ErrJSONEncode     = "Unexpected error"
)

func enqueuePostHandler(p pool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var equeReq models.EnqueueReq

		dec := json.NewDecoder(r.Body)

		if err := dec.Decode(&equeReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		superImportantTask := func() (interface{}, error) {
			// Sleep 100-500 ms
			time.Sleep(time.Millisecond * time.Duration(1+rand.Intn(5)) * 100)

			// Error simulating
			if rand.Intn(5) == 0 {
				return nil, fmt.Errorf("fatal error")
			}
			return equeReq.Payload, nil
		}

		id, err := p.Submit(superImportantTask, equeReq.MaxRetries)
		if err != nil {
			// Queue is full
			http.Error(w, ErrServerOverload, http.StatusInternalServerError)
			return
		}

		enqueRes := models.EnqueueRes{
			TaskId: id,
		}

		resData, err := json.Marshal(enqueRes)
		if err != nil {
			http.Error(w, ErrJSONEncode, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resData)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	return
}
