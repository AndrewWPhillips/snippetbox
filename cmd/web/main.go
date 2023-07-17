package main

import (
	"crypto/tls"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/andrewwphillips/snippetbox/pkg/models"
	"github.com/andrewwphillips/snippetbox/pkg/models/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
)

// application holds application-wide dependencies
type application struct {
	infoLog, errorLog *log.Logger // INFO (stdout) and ERROR (stderr) loggers
	session           *sessions.Session
	snippets          interface {
		Insert(string, string, string) (int, error)
		Get(int) (*models.Snippet, error)
		Latest() ([]*models.Snippet, error)
		Close()
	}
	templateCache map[string]*template.Template
	users         interface {
		Insert(string, string, string) (int, error)
		Authenticate(string, string) (int, string, error)
		Get(int) (*models.User, error)
		Close()
	}
}

// main is the program entry point
func main() {
	// Get command line options and initialise the app struct
	root := flag.String("static-dir", "./ui/static/", "Path to static assets")
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:pass@/snippetbox", "MySQL data source name")
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")
	flag.Parse()

	app := application{
		templateCache: newTemplateCache("./ui/html/"),
		infoLog:       log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		errorLog:      log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		snippets:      mysql.NewSnippetModel(*dsn),
		users:         mysql.NewUserModel(*dsn),
		session:       sessions.New([]byte(*secret)),
	}
	defer app.snippets.Close()
	defer app.users.Close()
	app.session.Lifetime = 12 * time.Hour // sessions expire after 12 hours
	app.session.Secure = true

	// Start the HTTPS server
	app.infoLog.Println("Starting server on", *addr)

	// This code was replaced when we moved to HTTPS
	//server := &http.Server{
	//	Addr:     *addr,
	//	ErrorLog: app.errorLog,
	//	Handler:  app.routes(*root),
	//}
	//app.errorLog.Fatal(server.ListenAndServe())

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	server := &http.Server{
		Addr:      *addr,
		ErrorLog:  app.errorLog,
		Handler:   app.routes(*root),
		TLSConfig: tlsConfig,
		// The following 4 timeouts may cause problems while debugging
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// NOTE: ListenAndServeTLS params (TLS private key and certificate files) were generated using this command line:
	// > go run /c/progra~1/go1.19/src/crypto/tls/generate_cert.go --ecdsa-curve P256 --host=localhost

	app.errorLog.Fatal(server.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem"))
}
