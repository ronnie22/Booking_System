package main

import (
	"Ticket_Booking_system/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "github.com/lib/pq"
	"github.com/julienschmidt/httprouter"
	"github.com/satori/go.uuid"
	"strconv"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://ronnie:password@localhost/theatre?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("You connected to your database.")
}

func main() {
	r := httprouter.New()
	r.GET("/", index)
	r.POST("/show", ShowTickets)
	r.POST("/insert",insert)
	r.POST("/update",update)
	http.ListenAndServe("localhost:8080", r)
}


func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params)  {
	s := `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Index</title>
</head>
<body>
<h1>Welcome</h1>
</body>
</html>
	`
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s))
}


func ShowTickets(w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	Pt := struct {
		TkId string
	}{}
	json.NewDecoder(r.Body).Decode(&Pt)
	if(Pt.TkId==""){
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
	st := make([]models.TicketsShow,0)

	rows, err:= db.Query("SELECT * FROM customer_data WHERE start_at=$1",Pt.TkId)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	for rows.Next() {
		t := models.TicketsShow{}
		err := rows.Scan(&t.Id,&t.Name,&t.PhnNumber,&t.Number,&t.StartAt,&t.EndAt,&t.Expire) // order matters
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		st = append(st, t)

	}

	tks, er := json.Marshal(st)
	if er != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	fmt.Fprintf(w, "%s\n", tks)
}

func insert(w http.ResponseWriter, r *http.Request, _ httprouter.Params)  {
	u := models.User{}
	json.NewDecoder(r.Body).Decode(&u)
	co,bo:=availabe(u.Start_at,w,r)
	if bo{

		num, err := uuid.NewV4()
		if err != nil {
			log.Fatalln(err)
		}
		s := u.Start_at[:2]
		st, _ := strconv.Atoi(s)
		st += 2
		end := strconv.Itoa(st) + ":30:00"
		number := num.String()

		_, err = db.Exec("INSERT INTO customer_data (name , ph_number, number , start_at  ,end_at ) VALUES ($1, $2, $3, $4,$5)", u.Name, u.Phn_num, number, u.Start_at, end)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		m := make(map[string]string)
		m["Ticket_Id"] = number
		uj, err := json.Marshal(m)
		if err != nil {
			fmt.Println(err)
		}

		co-=1
		_, err = db.Exec("UPDATE timings SET tickets_count=&1 WHERE start_at=$2;",co, u.Start_at)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // 201
		fmt.Fprintf(w, "%s\n", uj)

	}
}

func update(w http.ResponseWriter, r *http.Request,  _ httprouter.Params){
	TkUp:= struct {
		TkId string
		UpTi string
	}{}
	json.NewDecoder(r.Body).Decode(&TkUp)
	if(TkUp.TkId=="" && TkUp.UpTi==""){
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	s := TkUp.UpTi[:2]
	st, _ := strconv.Atoi(s)
	st += 2
	end := strconv.Itoa(st) + ":30:00"
	fmt.Println(TkUp.UpTi,end,TkUp.TkId)

	_, err := db.Exec("UPDATE customer_data SET start_at=&1 , end_at=$2 WHERE number=$3;",TkUp.UpTi,end,TkUp.TkId)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func delete(w http.ResponseWriter, r *http.Request,  _ httprouter.Params){

}




func availabe(start string,w http.ResponseWriter,r *http.Request) (int,bool) {
	row := db.QueryRow("SELECT * FROM timings where start_at=$1;",start)
	t:=models.Timing{}
	err := row.Scan(&t.Id,&t.Count, &t.Start, &t.End)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return 0,false
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return 0,false
	}

	if(t.Count>0){
		return t.Count,true
	}

	return 0,false
}

