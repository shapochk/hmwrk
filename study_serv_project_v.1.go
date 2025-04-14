package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Хэндлер для доступа к БД через передачу указателя на объект бд
type goodsHandler struct {
	custom_handler func(*sql.DB, http.ResponseWriter, *http.Request)
	goods_DB       *sql.DB
}

// реализуем интерфейс Хэндлер
func (lh *goodsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lh.custom_handler(lh.goods_DB, w, r)
}

// тут идут 3 подряд функции для обработки запросов хэндлером,
// которые потом станут полями объекта структуры goodsHandler
func list_handlerfunc(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	rows, err := db.Query("SELECT id, name, description, price FROM dbs")
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprintln(w, "id;name;description;Price")
	for rows.Next() {
		var id int
		var name, description string
		var price float64
		if err := rows.Scan(&id, &name, &description, &price); err != nil {
			fmt.Fprint(w, err)
			return
		}
		fmt.Fprintf(w, "%d;%s;%s;%g\n", id, name, description, price)
	}
	if err := rows.Err(); err != nil {
		fmt.Fprint(w, err)
		return
	}
}
func addrow_handlerfunc(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	if r.FormValue("name") == "" || r.FormValue("description") == "" || r.FormValue("price") == "" {
		fmt.Fprint(w, "All field parameters must be filled with nonempty values")
		return
	}
	pr, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	//Нет проверки на допустимость значения var pr ну и хер с ним
	if _, err := db.Exec("INSERT INTO dbs (name, description, price) VALUES ($1, $2, $3)", r.FormValue("name"), r.FormValue("description"), pr); err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprintln(w, "New row was successfully inserted")
}
func deleterow_handlerfunc(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.ParseInt(r.FormValue("id"), 10, 32)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	//Нет проверки на допустимсоть значения пер. id и проверки на
	//существование в БД записи с id указанным для удаления
	if _, err := db.Exec("DELETE FROM dbs WHERE id = $1", id); err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprintf(w, "Row with id=%d was successfully deleted", id)
}

// мидлвэйр для логгирования
func middleware_logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("middleware: Start processing %s request\n", r.Method)
		next.ServeHTTP(w, r)
		duration := time.Since(start).Nanoseconds()
		log.Printf("middleware: Request was processed in %d nanoseconds\n", duration)
	})
}
func main() {
	//пароль не пали...
	dsn := "host=localhost port=5432 dbname=myDB user=postgres password=postgres connect_timeout=10 sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	listHandler := &goodsHandler{list_handlerfunc, db}
	addrowHandler := &goodsHandler{addrow_handlerfunc, db}
	deleterowHandler := &goodsHandler{deleterow_handlerfunc, db}

	mux := http.NewServeMux()
	mux.Handle("/list", middleware_logging(listHandler))
	mux.Handle("/addrow", middleware_logging(addrowHandler))
	mux.Handle("/delrow", middleware_logging(deleterowHandler))

	myServ := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	err = myServ.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
