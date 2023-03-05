package main

import (
	"bufio"
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

const (
	progressFilename = ".progress.txt"
	progressLineFmt  = "%s %s %s"
)

var (
	connMysql = flag.String("mysql", "", "Mysql connection string")
	dictFile  = flag.String("dict", "", "Passwords dictionary file")
	workers   = flag.Int("w", 20, "number of workers")
	outFile   = flag.String("out", "out.txt", "File to output results")
)

func main() {
	flag.Parse()

	app := New()
	if err := app.Conn(*connMysql); err != nil {
		panic(err)
	}
	defer func() {
		if app != nil && app.db != nil {
			app.db.Close()
		}
	}()
	fmt.Println("Connected to DB")

	if err := app.LoadUsers(); err != nil {
		panic(err)
	}
	fmt.Printf("Got %d users\n", len(app.users))

	b, err := os.ReadFile(*dictFile)
	if err != nil {
		fmt.Println("cannot open dictionary")
		return
	}
	dictFileMd5 := md5.Sum(b)

	in, out := make(chan userRec), make(chan userRec)
	done, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(*workers)
	for i := 0; i < *workers; i++ {
		go func() {
			defer wg.Done()
			decryptPass(done, in, out, *dictFile)
		}()
	}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch
		fmt.Println("Shutting down...")
		cancel()
	}()

	go func() {
		defer close(out)
		wg.Wait()
	}()

	go func() {
		defer close(in)
		for _, u := range app.users {
			if !alreadyDone(u, *dictFile, dictFileMd5) {
				select {
				case in <- u:
				case <-done.Done():
					return
				}
			} else {
				fmt.Printf("Skip already checked: %s[%s]\n", u.Username, u.Email)
			}
		}
	}()

	f, err := os.OpenFile(*outFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for u := range out {
		if u.ClearPasswd != "" {
			fmt.Fprintf(f, "[+] Authentication success: %s[%s] %s\n", u.Username, u.Email, u.ClearPasswd)
		} else {
			markProgress(u, *dictFile, dictFileMd5)
			//fmt.Fprintf(f, "[-] Authentication fail: %s[%s] %s\n", u.Username, u.Email, u.Password)
		}
	}

}

func markProgress(u userRec, dictFile string, dictMd5 [md5.Size]byte) {
	f, err := os.OpenFile(progressFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("no progress", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "%s %s %x %v\n", u.Password, dictFile, dictMd5, time.Now())
}

func alreadyDone(u userRec, dictFile string, dictMD5 [md5.Size]byte) bool {
	f, err := os.Open(progressFilename)
	if err != nil {
		fmt.Println("cannot read progress", err)
		return false
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.Split(sc.Text(), " ")
		if line[0] == u.Password && line[1] == dictFile && line[2] == fmt.Sprintf("%x", dictMD5) {
			return true
		}
	}

	return false
}

type userRec struct {
	Username, Email, Password, PasswordHint string
	ClearPasswd                             string
}

type application struct {
	db    *sql.DB
	users []userRec
}

func New() *application {
	return &application{users: make([]userRec, 0)}
}

func (a *application) Conn(connStr string) error {
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	a.db = db

	return nil
}

func (a *application) LoadUsers() error {
	if a.db == nil {
		return errors.New("not connected to db")
	}

	qry := "select username, email, password, COALESCE(passwordHint, '') from app_user order by id desc"
	rows, err := a.db.QueryContext(context.TODO(), qry)
	if err != nil {
		return err
	}
	defer rows.Close()

	var usr, email, passwd, passwdHint string
	for rows.Next() {
		if err := rows.Scan(&usr, &email, &passwd, &passwdHint); err != nil {
			fmt.Println(err)
		} else {
			a.users = append(a.users, userRec{
				Username:     usr,
				Email:        email,
				Password:     passwd,
				PasswordHint: passwdHint,
			})
		}
	}

	return nil
}

func decryptPass(done context.Context, in <-chan userRec, out chan<- userRec, dict string) {

	f, err := os.Open(dict)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	fmt.Println("start worker")
main:
	for u := range in {
		fmt.Printf("Checking %s[%s] - %s\n", u.Username, u.Email, u.Password)
		found := make(chan struct{})
		var wg sync.WaitGroup
		limiter := make(chan struct{}, 100)
		f.Seek(0, io.SeekStart)
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			select {
			case <-found:
				continue main
			case <-done.Done():
				fmt.Println(done.Err())
				return
			default:
				passwd := sc.Text()
				hash := strings.TrimPrefix(u.Password, "{bcrypt}")
				wg.Add(1)
				go func(u userRec, passwd string, hash string) {
					defer func() {
						<-limiter
						wg.Done()
					}()
					limiter <- struct{}{}
					if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(passwd)); err == nil {
						u.ClearPasswd = passwd
						select {
						case out <- u:
						case <-done.Done():
							return
						}
						close(found)
					}
				}(u, passwd, hash)
			}
		}
		wg.Wait()
		close(limiter)
		select {
		case out <- u:
		case <-done.Done():
			return
		}
	}
}
