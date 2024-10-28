package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/go-playground/validator/v10"
    "github.com/joho/godotenv"
    "os"
    "sort"
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

    _ = godotenv.Load(".env")
	rdb = redis.NewClient(&redis.Options{
		Addr: os.Getenv("URL"),
        Password: os.Getenv("PASSWORD"),
		DB:       0,
	})

	validate = validator.New()
	r := chi.NewRouter()

	r.Post("/activity", handleActivity)
    r.Delete("/activity/{id}", deleteActivity)

    fmt.Println("http://localhost:80")
	http.ListenAndServe(":80", r)
}

func printAllActivities() {
	keys, err := rdb.Keys(ctx, "activity:*").Result()
	if err != nil {
		fmt.Println("Error getting keys:", err)
		return
	}

	var activities []ActivityModel
	for _, key := range keys {
		data, err := rdb.Get(ctx, key).Result()
		if err != nil {
			fmt.Println("Error getting activity:", err)
			continue
		}

		var activity ActivityModel
		if err := json.Unmarshal([]byte(data), &activity); err != nil {
			fmt.Println("Error unmarshalling activity:", err)
			continue
		}
		activities = append(activities, activity)
	}

	// Ordenar las actividades por ID (puedes cambiar la lógica de ordenación según necesites)
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].ID < activities[j].ID
	})

	// Imprimir las actividades
	fmt.Println("Current Activities:")
	for _, activity := range activities {
		fmt.Printf("ID: %d, Action: %s, Date: %s\n", activity.ID, activity.Action, activity.Date)
	}
}

func deleteActivity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	key := fmt.Sprintf("activity:%s", id)

	// Eliminar la actividad de Redis
	result, err := rdb.Del(ctx, key).Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result == 0 {
		http.Error(w, `{"code": 500,"data":"Activity not found"}`, http.StatusNotFound)
		return
	}

	fmt.Printf("Activity with ID %s deleted\n", id)
	printAllActivities() 
    _, _ = w.Write([]byte(`{"code": 200,"data":"Activity remove"}`))
}

func handleActivity(w http.ResponseWriter, r *http.Request) {
	var activity ActivityModel
	if err := json.NewDecoder(r.Body).Decode(&activity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validate.Struct(activity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	key := fmt.Sprintf("activity:%d", activity.ID)
	exists, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	activityJSON, err := json.Marshal(activity)
	if err != nil {
		http.Error(w, `{"code": 500,"data":"Activity error"}`, http.StatusInternalServerError)
		return
	}

	if exists > 0 {
		_, err = rdb.Set(ctx, key, activityJSON, 0).Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
        printAllActivities()
		w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"code": 200,"data":"Activity updated"}`))
	} else {
		_, err = rdb.Set(ctx, key, activityJSON, 0).Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
        printAllActivities()
		w.WriteHeader(http.StatusCreated)
        _, _ = w.Write([]byte(`{"code": 201,"data":"Activity created"}`))
	}
}

