package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"erlaine/aeroporto/internal/database"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

//go:embed templates/*.html static/*
var assets embed.FS

type Row map[string]string
type User struct {
	ID                 int64
	Name, Role         string
	Company, Passenger int64
}
type App struct {
	DB       *sql.DB
	Q        *database.Queries
	T        *template.Template
	Location *time.Location
	Secure   bool
}
type key int

const userKey key = 1

func New(path string) (*App, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	for _, q := range []string{"PRAGMA foreign_keys=ON", "PRAGMA journal_mode=WAL", "PRAGMA busy_timeout=5000", database.Schema} {
		if _, err = db.Exec(q); err != nil {
			db.Close()
			return nil, err
		}
	}
	loc, err := time.LoadLocation(env("APP_TIMEZONE", "America/Sao_Paulo"))
	if err != nil {
		return nil, err
	}
	a := &App{DB: db, Q: database.New(db), Location: loc, Secure: os.Getenv("COOKIE_SECURE") == "true"}
	a.T, err = template.New("app").Funcs(template.FuncMap{
		"get": func(r Row, k string) string { return r[k] },
		"datetime": func(s string) string {
			t, e := time.Parse("2006-01-02T15:04", s)
			if e != nil {
				return s
			}
			return t.Format("02/01/2006 15:04")
		},
		"eqs": func(a, b interface{}) bool { return fmt.Sprint(a) == fmt.Sprint(b) },
		"inc": func(n int) int { return n + 1 },
		"prev": func(n int) int {
			if n > 1 {
				return n - 1
			}
			return 1
		},
		"url": url.QueryEscape,
		"badge": func(s string) string {
			switch s {
			case "Ativo", "Ativa", "Agendado", "Programada", "Emitido":
				return "green"
			case "Em Voo", "Em Andamento", "Check-in Realizado":
				return "blue"
			case "Cancelado", "Cancelada", "Inativo", "Inativa", "Excluído", "Excluída", "Suspensa":
				return "red"
			default:
				return "neutral"
			}
		},
	}).ParseFS(assets, "templates/*.html")
	if err != nil {
		db.Close()
		return nil, err
	}
	return a, nil
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func token() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func digest(v string) string       { h := sha256.Sum256([]byte(v)); return hex.EncodeToString(h[:]) }
func current(r *http.Request) User { u, _ := r.Context().Value(userKey).(User); return u }
func num(s string) int64           { n, _ := strconv.ParseInt(s, 10, 64); return n }
func (a *App) now() string         { return time.Now().UTC().Format(time.RFC3339) }
func (a *App) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, loginLimiter)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Referrer-Policy", "same-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; object-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'")
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if a.DB.PingContext(r.Context()) != nil {
			http.Error(w, "database unavailable", 503)
			return
		}
		io.WriteString(w, "ok")
	})
	r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		p := chi.URLParam(r, "*")
		b, e := assets.ReadFile("static/" + p)
		if e != nil {
			http.NotFound(w, r)
			return
		}
		switch filepath.Ext(p) {
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		}
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Write(b)
	})
	r.Get("/login", a.loginPage)
	r.Post("/login", a.login)
	r.Group(func(r chi.Router) {
		r.Use(a.authenticate)
		r.Use(a.csrf)
		r.Get("/", a.home)
		r.Post("/logout", a.logout)
		r.Get("/grid/{entity}", a.gridHandler)
		r.Get("/seats", a.seatsHandler)
		r.Get("/form/{entity}", a.formHandler)
		r.Post("/save/{entity}", a.saveHandler)
		r.Post("/batch/{entity}", a.batchHandler)
		r.Get("/detail/{entity}/{id}", a.detailHandler)
		r.Post("/transition/{entity}/{id}", a.transitionHandler)
		r.Get("/export/{entity}", a.exportHandler)
		r.Get("/print/{id}", a.printHandler)
	})
	return r
}
func (a *App) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, e := r.Cookie("session")
		if e != nil {
			http.Redirect(w, r, "/login", 303)
			return
		}
		u, e := a.Q.GetSessionUser(r.Context(), database.GetSessionUserParams{TokenHash: digest(c.Value), ExpiraEm: a.now()})
		if e != nil {
			w.Header().Set("HX-Redirect", "/login")
			http.Redirect(w, r, "/login", 303)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, User{u.ID, u.Nome, u.Papel, u.EmpresaID.Int64, u.PassageiroID.Int64})))
	})
}
func (a *App) csrf(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Requisição inválida", 400)
				return
			}
			c, e := r.Cookie("session")
			v := r.Header.Get("X-CSRF-Token")
			if v == "" {
				v = r.FormValue("csrf")
			}
			if e != nil || v != digest("csrf:"+c.Value) {
				http.Error(w, "Token de segurança inválido", 403)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
func csrfValue(r *http.Request) string {
	c, e := r.Cookie("session")
	if e != nil {
		return ""
	}
	return digest("csrf:" + c.Value)
}
func (a *App) loginPage(w http.ResponseWriter, r *http.Request) {
	c, e := r.Cookie("login_csrf")
	v := ""
	if e == nil {
		v = c.Value
	}
	if v == "" {
		v = token()
		http.SetCookie(w, &http.Cookie{Name: "login_csrf", Value: v, Path: "/login", HttpOnly: true, Secure: a.Secure, SameSite: http.SameSiteStrictMode, MaxAge: 900})
	}
	a.render(w, "login", map[string]interface{}{"CSRF": v, "Demo": os.Getenv("DEMO_MODE") == "true", "Error": r.URL.Query().Get("error")})
}
func (a *App) login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16384)
	if r.ParseForm() != nil {
		http.Error(w, "Dados inválidos", 400)
		return
	}
	c, e := r.Cookie("login_csrf")
	if e != nil || c.Value != r.FormValue("csrf") {
		http.Error(w, "Atualize a página de login", 403)
		return
	}
	u, e := a.Q.GetUserByEmail(r.Context(), strings.ToLower(strings.TrimSpace(r.FormValue("email"))))
	if e != nil || bcrypt.CompareHashAndPassword([]byte(u.SenhaHash), []byte(r.FormValue("password"))) != nil {
		time.Sleep(250 * time.Millisecond)
		http.Redirect(w, r, "/login?error=Credenciais+inválidas", 303)
		return
	}
	v := token()
	if e = a.Q.CreateSession(r.Context(), database.CreateSessionParams{TokenHash: digest(v), UsuarioID: u.ID, ExpiraEm: time.Now().UTC().Add(8 * time.Hour).Format(time.RFC3339)}); e != nil {
		http.Error(w, "Não foi possível iniciar sessão", 500)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "session", Value: v, Path: "/", HttpOnly: true, Secure: a.Secure, SameSite: http.SameSiteStrictMode, MaxAge: 28800})
	http.Redirect(w, r, "/", 303)
}
func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	c, _ := r.Cookie("session")
	if c != nil {
		a.Q.DeleteSession(r.Context(), digest(c.Value))
	}
	http.SetCookie(w, &http.Cookie{Name: "session", Path: "/", MaxAge: -1, HttpOnly: true, Secure: a.Secure})
	w.Header().Set("HX-Redirect", "/login")
	http.Redirect(w, r, "/login", 303)
}
func (a *App) render(w io.Writer, name string, data interface{}) {
	if e := a.T.ExecuteTemplate(w, name, data); e != nil {
		log.Printf("template %s: %v", name, e)
	}
}
func (a *App) component(name string, data interface{}) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return a.T.ExecuteTemplate(w, name, data) })
}
func (a *App) Sync(ctx context.Context) error {
	tx, e := a.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	q := a.Q.WithTx(tx)
	for _, fn := range []func() error{func() error { return q.DepartFlights(ctx, a.now()) }, func() error { return q.CompleteFlights(ctx, a.now()) }, func() error { return q.StartStops(ctx) }, func() error { return q.CompleteStops(ctx) }, func() error { return q.CompleteTickets(ctx) }} {
		if e = fn(); e != nil {
			return e
		}
	}
	return tx.Commit()
}
func (a *App) RunClock(ctx context.Context) {
	t := time.NewTicker(15 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if e := a.Sync(ctx); e != nil {
				log.Printf("clock: %v", e)
			}
		}
	}
}
func (a *App) canRead(u User, k string) bool {
	if _, ok := entities[k]; !ok {
		return false
	}
	if u.Role == "gerente" {
		return true
	}
	if u.Role == "atendente" {
		return k == "voo" || k == "escala" || k == "bilhete" || k == "passageiro"
	}
	return k == "voo" || k == "escala" || k == "bilhete"
}
func (a *App) canWrite(u User, k string) bool {
	if u.Role == "gerente" {
		return k != "bilhete" && k != "passageiro"
	}
	return u.Role == "atendente" && (k == "bilhete" || k == "passageiro")
}
func (a *App) scope(u User, k string) (string, []interface{}) {
	if u.Role == "passageiro" {
		if k == "bilhete" {
			return " AND passageiro_id=?", []interface{}{u.Passenger}
		}
		if k == "voo" {
			return " AND situacao IN('Agendado','Em Voo')", nil
		}
		if k == "escala" {
			return " AND situacao IN('Programada','Em Andamento','Concluída')", nil
		}
	}
	if u.Role == "gerente" {
		switch k {
		case "empresa":
			return " AND id=?", []interface{}{u.Company}
		case "voo", "aeronave":
			return " AND empresa_id=?", []interface{}{u.Company}
		case "escala":
			return " AND (voo_id IS NULL OR voo_id IN(SELECT id FROM voo WHERE empresa_id=?))", []interface{}{u.Company}
		case "bilhete":
			return " AND voo_id IN(SELECT id FROM voo WHERE empresa_id=?)", []interface{}{u.Company}
		}
	}
	return "", nil
}

type querier interface {
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func rows(ctx context.Context, q querier, query string, args ...interface{}) ([]Row, error) {
	rs, e := q.QueryContext(ctx, query, args...)
	if e != nil {
		return nil, e
	}
	defer rs.Close()
	cols, e := rs.Columns()
	if e != nil {
		return nil, e
	}
	out := []Row{}
	for rs.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if e = rs.Scan(ptrs...); e != nil {
			return nil, e
		}
		row := Row{}
		for i, c := range cols {
			if vals[i] != nil {
				row[c] = fmt.Sprint(vals[i])
			} else {
				row[c] = ""
			}
		}
		out = append(out, row)
	}
	return out, rs.Err()
}
func one(ctx context.Context, q querier, k string, id int64) (Row, error) {
	if _, ok := entities[k]; !ok {
		return nil, errors.New("Cadastro inválido")
	}
	rs, e := rows(ctx, q, "SELECT * FROM "+k+" WHERE id=?", id)
	if e != nil {
		return nil, e
	}
	if len(rs) == 0 {
		return nil, sql.ErrNoRows
	}
	return rs[0], nil
}
func (a *App) accessible(ctx context.Context, u User, k string, id int64) (Row, error) {
	if !a.canRead(u, k) {
		return nil, errors.New("Acesso negado")
	}
	s, args := a.scope(u, k)
	rs, e := rows(ctx, a.DB, "SELECT * FROM "+k+" WHERE id=?"+s, append([]interface{}{id}, args...)...)
	if e != nil {
		return nil, e
	}
	if len(rs) == 0 {
		return nil, errors.New("Registro não encontrado ou acesso negado")
	}
	return rs[0], nil
}
func friendly(e error) string {
	s := e.Error()
	switch {
	case strings.Contains(s, "UNIQUE"):
		return "Já existe um registro com esse código, documento, sequência ou assento. Atualize a grade."
	case strings.Contains(s, "FOREIGN KEY"):
		return "A referência selecionada não existe ou possui registros relacionados."
	case strings.Contains(s, "CHECK"):
		return "Verifique os valores: quantidades positivas, horários e origem diferente do destino."
	default:
		return s
	}
}
