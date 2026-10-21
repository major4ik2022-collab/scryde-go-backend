package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Set struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Pdef       float64 `json:"pdef"`
	Mdef       float64 `json:"mdef"`
	Patk       float64 `json:"patk"`
	Matk       float64 `json:"matk"`
	Hp         float64 `json:"hp"`
	Atkspd     float64 `json:"atkspd"`
	Castspd    float64 `json:"castspd"`
	ResistText string  `json:"resist_text"`
}

var db *pgxpool.Pool

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL не задан")
	}

	var err error
	db, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("БД не отвечает: %v", err)
	}
	log.Println("Успешно подключились к Supabase")

	http.HandleFunc("/api/sets", handleGetSets)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleGetSets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.Query(ctx, "SELECT id, name, type, pdef, mdef, patk, matk, hp, atkspd, castspd, resist_text FROM sets")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var sets []Set
	for rows.Next() {
		var s Set
		err := rows.Scan(&s.ID, &s.Name, &s.Type, &s.Pdef, &s.Mdef, &s.Patk, &s.Matk, &s.Hp, &s.Atkspd, &s.Castspd, &s.ResistText)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}
		sets = append(sets, s)
	}

	json.NewEncoder(w).Encode(sets)
}