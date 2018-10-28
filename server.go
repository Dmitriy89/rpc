package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type HttpConn struct {
	in  io.Reader
	out io.Writer
}

func (c *HttpConn) Read(p []byte) (n int, err error)  { return c.in.Read(p) }
func (c *HttpConn) Write(d []byte) (n int, err error) { return c.out.Write(d) }
func (c *HttpConn) Close() error                      { return nil }

type Auth struct {
	Uuid    string
	Login   string
	DateReg time.Time
	Update  []struct {
		NewLogin string
	}
}

type rawTime []byte

func (t rawTime) Time() (time.Time, error) {
	return time.Parse("2006-01-02", string(t))
}

type Data struct{}

func connectDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:12345@/test")
	if err != nil {
		return nil, fmt.Errorf("Error connect database: %s", err)
	}
	return db, nil
}

func (mt *Data) Add(a *Auth, msg *string) error {

	if a.Uuid == "" && a.Login == "" {
		return fmt.Errorf("params are empty or incorrectly specified")
	}

	db, err := connectDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, errEx := db.Exec("insert into test(uuid ,login, dateReg) values(?,?,?);", a.Uuid, a.Login, time.Now())
	if errEx != nil {
		return errEx
	}

	*msg = fmt.Sprintf("Data added to database: uuid - %s / login - %s", a.Uuid, a.Login)
	return nil
}

func (mt *Data) Get(a *Auth, msg *string) error {

	var (
		uuid    string
		dateReg rawTime
	)

	if a.Login == "" {
		return fmt.Errorf("Enter login")
	}

	db, err := connectDB()
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRow(`select uuid, dateReg from test where login = ?;`, a.Login)

	if err := row.Scan(&uuid, &dateReg); err != nil {
		return err
	}

	t, err := dateReg.Time()
	if err != nil {
		return err
	}

	*msg = fmt.Sprintf("Report: %s / %s", uuid, t.Format("2006.02.01"))

	return nil
}

func (mt *Data) Set(a *Auth, msg *string) error {

	var tx *sql.Tx

	if a.Uuid == "" {
		return fmt.Errorf("Enter uuid")
	}

	db, err := connectDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err = db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	stmt, errPp := tx.Prepare("update test set login = ? where uuid = ?;")
	if errPp != nil {
		return errPp
	}
	defer stmt.Close()

	for _, val := range a.Update {
		_, errEx := stmt.Exec(val.NewLogin, a.Uuid)
		if errEx != nil {
			return errEx
		}
	}
	errCom := tx.Commit()
	if errCom != nil {
		return errCom
	}

	*msg = fmt.Sprintf("Data update")

	return nil
}

func main() {

	fmt.Printf("TestHTTPServer\n")

	server := rpc.NewServer()
	server.Register(&Data{})

	listener, errLn := net.Listen("tcp", ":8080")
	if errLn != nil {
		log.Fatal("listen error:", errLn)
	}
	defer listener.Close()

	http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rpc" {
			serverCodec := jsonrpc.NewServerCodec(&HttpConn{in: r.Body, out: w})
			w.Header().Set("Content-type", "application/json")
			err := server.ServeRequest(serverCodec)
			if err != nil {
				log.Printf("Error while serving JSON request: %v", err)
				return
			}
		}

	}))

}
