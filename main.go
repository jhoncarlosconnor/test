package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/go-playground/validator/v10"
)

type ActivityModel struct {
	ID             int    `json:"id" validate:"required"`
	Action         string `json:"action" validate:"required"`
	RegisterOffline int   `json:"registerOffline" validate:"required"`
	RegisterOnline  int   `json:"registerOnline" validate:"required"`
	Check          bool   `json:"check" validate:"required"`
	View           bool   `json:"view"`
	Date           string `json:"date" validate:"required"`
	Deploy         bool   `json:"deploy"`
}

var (
	rdb       *redis.Client
	validate  *validator.Validate
	ctx       = context.Background()
)

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Cambia si es necesario
	})

	validate = validator.New()
	r := chi.NewRouter()

	r.Post("/activity", handleActivity)

	http.ListenAndServe(":8080", r)
}

func handleActivity(w http.ResponseWriter, r *http.Request) {
	var activity ActivityModel
	if err := json.NewDecoder(r.Body).Decode(&activity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validar el modelo
	if err := validate.Struct(activity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verificar si el ID ya existe
	key := fmt.Sprintf("activity:%d", activity.ID)
	exists, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Serializar a JSON
	activityJSON, err := json.Marshal(activity)
	if err != nil {
		http.Error(w, "Error serializando la actividad", http.StatusInternalServerError)
		return
	}

	if exists > 0 {
		// Actualizar
		_, err = rdb.Set(ctx, key, activityJSON, 0).Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode("Activity updated")
	} else {
		// Almacenar nuevo
		_, err = rdb.Set(ctx, key, activityJSON, 0).Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode("Activity created")
	}
}

