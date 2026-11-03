package main

import (
	"context"
	"erlaine/aeroporto/internal/app"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	_ "time/tzdata"
)

func main() {
	path := os.Getenv("DATABASE_PATH")
	if path == "" {
		path = "data/aeroporto.db"
	}
	a, e := app.New(path)
	if e != nil {
		log.Fatal(e)
	}
	defer a.DB.Close()
	if len(os.Args) > 1 && os.Args[1] == "backup" {
		f := flag.NewFlagSet("backup", flag.ExitOnError)
		output := f.String("output", "", "Arquivo de destino (novo)")
		f.Parse(os.Args[2:])
		if *output == "" {
			log.Fatal("Informe --output backups/aeroporto.db")
		}
		if _, err := os.Stat(*output); err == nil {
			log.Fatal("Destino já existe; use um novo arquivo")
		}
		if err := os.MkdirAll(filepath.Dir(*output), 0700); err != nil {
			log.Fatal(err)
		}
		if _, err := a.DB.ExecContext(context.Background(), "VACUUM INTO ?", *output); err != nil {
			log.Fatal(err)
		}
		log.Printf("Backup consistente criado: %s", *output)
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "create-user" {
		f := flag.NewFlagSet("create-user", flag.ExitOnError)
		email := f.String("email", "", "E-mail")
		name := f.String("name", "", "Nome")
		role := f.String("role", "atendente", "gerente, atendente ou passageiro")
		company := f.Int64("company", 0, "ID da empresa")
		passenger := f.Int64("passenger", 0, "ID do passageiro")
		f.Parse(os.Args[2:])
		if e = a.CreateUser(context.Background(), *email, *name, os.Getenv("USER_PASSWORD"), *role, *company, *passenger); e != nil {
			log.Fatal(e)
		}
		log.Print("Usuário criado")
		return
	}
	if e = a.Bootstrap(context.Background()); e != nil {
		log.Fatal(e)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if e = a.Sync(ctx); e != nil {
		log.Fatal(e)
	}
	go a.RunClock(ctx)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: ":" + port, Handler: a.Routes(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	go func() {
		log.Printf("Erlaine disponível em http://localhost:%s", port)
		if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			log.Fatal(e)
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if e = srv.Shutdown(shutdown); e != nil {
		log.Print(e)
	}
}
