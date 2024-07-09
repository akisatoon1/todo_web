package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Task struct {
	ID         int
	Name       string
	Deadline   string
	Created_at string
}

func main() {
	err := connect_db()
	err_log(err, "connect_db")

	http.HandleFunc("/home/{$}", home)
	http.HandleFunc("/add/{$}", add)
	http.HandleFunc("/edit/{$}", edit)
	http.HandleFunc("POST /save_add/{$}", save_add)
	http.HandleFunc("POST /save_edit/{$}", save_edit)
	http.HandleFunc("POST /delete/{$}", delete)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func err_log(err error, title string) {
	if err != nil {
		log.Fatalf("%v: %v\n", title, err.Error())
	}
}

func connect_db() error {
	cfg := mysql.Config{
		User:   os.Getenv("DBUSER"),
		Passwd: os.Getenv("DBPASS"),
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "todo",
	}
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return err
	}
	pingErr := db.Ping()
	if pingErr != nil {
		return pingErr
	}
	return nil
}

func home(w http.ResponseWriter, _ *http.Request) {
	tasks, err := all_tasks()
	err_log(err, "all_tasks")
	tmpl, err := template.ParseFiles("./home.html")
	err_log(err, "template.Parsefiles")
	err = tmpl.Execute(w, tasks)
	err_log(err, "tmpl.Execute")
}

func all_tasks() ([]Task, error) {
	rows, err := db.Query("SELECT id, name, deadline, created_at FROM tasks;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var id int
		var name, deadline, created_at string
		var tmp sql.NullString
		if err := rows.Scan(&id, &name, &tmp, &created_at); err != nil {
			return nil, err
		}
		if tmp.Valid {
			deadline = tmp.String
		} else {
			deadline = "no deadline"
		}

		tasks = append(tasks, Task{ID: id, Name: name, Deadline: deadline, Created_at: created_at})
	}
	return tasks, nil
}

func add(w http.ResponseWriter, _ *http.Request) {
	tmpl, err := template.ParseFiles("./add.html")
	err_log(err, "template.ParseFiles")
	err = tmpl.Execute(w, nil)
	err_log(err, "tmpl.Execute")
}

func edit(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	var Id int
	var name, deadline string
	var tmp sql.NullString
	row := db.QueryRow("SELECT id, name, deadline FROM tasks WHERE id = ?", id)
	err := row.Scan(&Id, &name, &tmp)
	err_log(err, "row.Scan")
	if tmp.Valid {
		deadline = strings.Replace(tmp.String, " ", "T", 1)[:16]
	} else {
		deadline = ""
	}
	task := Task{
		ID:       Id,
		Name:     name,
		Deadline: deadline,
	}
	tmpl, err := template.ParseFiles("./edit.html")
	err_log(err, "template.ParseFiles")
	err = tmpl.Execute(w, task)
	err_log(err, "tmpl.Execute")
}

func save_add(w http.ResponseWriter, r *http.Request) {
	name, deadline := r.FormValue("name"), r.FormValue("deadline")
	var result sql.Result
	var err error
	if deadline != "" {
		deadline = strings.Replace(deadline, "T", " ", 1) + ":00"
		result, err = db.Exec("INSERT INTO tasks (name, deadline) VALUES (?, ?);", name, deadline)
	} else {
		result, err = db.Exec("INSERT INTO tasks (name) VALUES (?);", name)
	}
	err_log(err, "db.Exec")
	num, err := result.RowsAffected()
	err_log(err, "result.RowsAffected")
	fmt.Printf("%d rows is inserted.\n", num)
	http.Redirect(w, r, "../home/", http.StatusFound)
}

func save_edit(w http.ResponseWriter, r *http.Request) {
	id, name, deadline := r.FormValue("id"), r.FormValue("name"), r.FormValue("deadline")
	var result sql.Result
	var err error
	if deadline != "" {
		deadline = strings.Replace(deadline, "T", " ", 1) + ":00"
		result, err = db.Exec("UPDATE tasks SET name = ?, deadline = ? WHERE id = ?;", name, deadline, id)
	} else {
		result, err = db.Exec("UPDATE tasks SET name = ?, deadline = ? WHERE id = ?;", name, nil, id)
	}
	err_log(err, "db.Exec")
	num, err := result.RowsAffected()
	err_log(err, "result.RowsAffected")
	fmt.Printf("%d rows is updated.\n", num)
	http.Redirect(w, r, "../home/", http.StatusFound)
}

func delete(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	result, err := db.Exec("DELETE FROM tasks WHERE id = ?;", id)
	err_log(err, "db.Exec")
	num, err := result.RowsAffected()
	err_log(err, "result.RowsAffected")
	fmt.Printf("%d rows is deleted.\n", num)
	http.Redirect(w, r, "../home/", http.StatusFound)
}
