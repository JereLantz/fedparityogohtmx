package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

    _ "github.com/mattn/go-sqlite3"
)

type formData struct{
    Id int
    Name string
    Email string
    Location string
}

func newFormData(id int, name, email, location string) formData{
    return formData{
        Id: id,
        Name: name,
        Email: email,
        Location: location,
    }
}

type pageFormData = []formData

func newPageFormData () pageFormData{
    return pageFormData{}
}

func delDataOfID (data pageFormData, id int) pageFormData{
    newData := newPageFormData()

    for _, item := range data {
        if item.Id != id{
            newData = append(newData, item)
        }
    }
    return newData
}

func main(){
    router := http.NewServeMux()
    server := http.Server{
        Addr: ":42069",
        Handler: router,
    }

    conn, err := connectDB()
    if err != nil {
        log.Println("error connecting to db: ", err)
    }

    err = createTable(conn)
    if err != nil {
        log.Println("Error creating the db table", err)
    }

    //TODO: hae databasesta tiedot pagecontenttiin
    defer conn.Close()

    pageTemplates, err := template.ParseFiles("templates/index.html")
    if err != nil {
        log.Printf("Error parsing templates: %s", err)
    }

    pageContent, err := getAllData(conn)
    if err != nil {
        log.Printf("Error fetching data: %s\n", err)
    }

    router.HandleFunc("GET /page/", func(w http.ResponseWriter, r *http.Request) {
        pageContent, err = getAllData(conn)
        if err != nil {
            log.Printf("Error fetching data: %s\n", err)
        }
        w.WriteHeader(200)
        pageTemplates.Execute(w, pageContent)
    })

    router.HandleFunc("GET /page/css/style.css", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "templates/css/style.css")
    })

    router.HandleFunc("POST /api/addContact", func(w http.ResponseWriter, r *http.Request) {
        newData, err := addUserToDB(conn,r.FormValue("name"), r.FormValue("email"), r.FormValue("location"))
        if err != nil {
            w.WriteHeader(500)
            return
        }

        pageContent = append(pageContent, newData)
        w.WriteHeader(200)
        pageTemplates.ExecuteTemplate(w, "tableRow", newData)
    })

    router.HandleFunc("DELETE /api/delContact/", func(w http.ResponseWriter, r *http.Request) {
        path := r.URL.Path
        parts := strings.Split(path, "/")
        id, err:= strconv.Atoi(parts[len(parts)-1])
        if err != nil{
            log.Printf("Error converting id from string to an int. %s", err)
            w.WriteHeader(500)
            return
        }

        err = deleteUserFromDB(conn, id)
        if err != nil {
            log.Printf("Error removing user with the id: %d from the database. %s\n", id, err)
            w.WriteHeader(500)
            return
        }
        pageContent = delDataOfID(pageContent, id)
        w.WriteHeader(200)
    })

    log.Printf("Http service started on port %s", server.Addr)
    log.Fatal(server.ListenAndServe())
}

func connectDB() (*sql.DB, error){
    db, err := sql.Open("sqlite3", "data.db")
    if err != nil {
        return nil, err
    }

    err = db.Ping()
    if err != nil {
        return nil, err
    }

    return db, nil
}

func createTable(db *sql.DB) error{
    query := `CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT not null,
        email text not null,
        location text not null
    );`

    _, err := db.Exec(query)
    if err != nil {
        return err
    }
    return nil
}

func getAllData(db *sql.DB) (pageFormData, error){
    queData := pageFormData{}
    query := "select * from users;"

    data, err := db.Query(query)
    if err != nil {
        return nil, err
    }

    defer data.Close()

    for data.Next(){
        newData := formData{}
        data.Scan(&newData.Id, &newData.Name, &newData.Email, &newData.Location)

        queData = append(queData, newData)
    }

    return queData, nil
}

func addUserToDB(db *sql.DB, name, email, location string) (formData, error){
    returnData := formData{
        Name: name,
        Email: email,
        Location: location,
    }
    query := "INSERT INTO users (name, email, location) VALUES (?,?,?);"

    result, err := db.Exec(query, name, email, location)
    if err != nil {
        return formData{}, err
    }

    lastID, err := result.LastInsertId()
    if err != nil {
        return formData{}, err
    }
    returnData.Id = int(lastID)
    return returnData, nil
}

func deleteUserFromDB(db *sql.DB, id int) error{
    //TODO:
    query := "DELETE FROM users WHERE id=?;"

    _, err := db.Exec(query, id)
    if err != nil {
        return err
    }

    return nil
}
