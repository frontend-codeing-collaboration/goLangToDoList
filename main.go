package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"

    _ "github.com/mattn/go-sqlite3"
)

type Todo struct {
    ID        int    `json:"id"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
}

var db *sql.DB

func main() {
    var err error
    db, err = sql.Open("sqlite3", "./todos.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create todo table if it doesn't exist
    createTable()

    http.HandleFunc("/todos", listTodosHandler)
    http.HandleFunc("/todos/add", addTodoHandler)
    http.HandleFunc("/todos/update", updateTodoHandler)
    http.HandleFunc("/todos/delete", deleteTodoHandler)

    fmt.Println("Server listening on port 8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func createTable() {
    sqlStmt := `
    CREATE TABLE IF NOT EXISTS todos (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT,
        completed BOOLEAN
    );
    `
    _, err := db.Exec(sqlStmt)
    if err != nil {
        log.Fatalf("Error creating table: %v\n", err)
    }
}

func listTodosHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, title, completed FROM todos")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    todos := []Todo{}
    for rows.Next() {
        var todo Todo
        if err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed); err != nil {
            log.Println(err)
            continue
        }
        todos = append(todos, todo)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(todos)
}

func addTodoHandler(w http.ResponseWriter, r *http.Request) {
    var todo Todo
    if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    stmt, err := db.Prepare("INSERT INTO todos(title, completed) VALUES(?, ?)")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer stmt.Close()

    _, err = stmt.Exec(todo.Title, todo.Completed)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}

func updateTodoHandler(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.URL.Query().Get("id"))
    if err != nil {
        http.Error(w, "Invalid todo ID", http.StatusBadRequest)
        return
    }

    var todo Todo
    if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    stmt, err := db.Prepare("UPDATE todos SET title=?, completed=? WHERE id=?")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer stmt.Close()

    _, err = stmt.Exec(todo.Title, todo.Completed, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func deleteTodoHandler(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.URL.Query().Get("id"))
    if err != nil {
        http.Error(w, "Invalid todo ID", http.StatusBadRequest)
        return
    }

    _, err = db.Exec("DELETE FROM todos WHERE id=?", id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
