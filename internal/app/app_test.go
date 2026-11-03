package app

import (
	"context"
	"erlaine/aeroporto/internal/database"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

func testApp(t *testing.T) *App {
	t.Helper()
	a, e := New(t.TempDir() + "/test.db")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { a.DB.Close() })
	if e = a.SeedDemo(context.Background()); e != nil {
		t.Fatal(e)
	}
	return a
}
func session(t *testing.T, a *App, id int64) string {
	t.Helper()
	v := token()
	e := a.Q.CreateSession(context.Background(), database.CreateSessionParams{TokenHash: digest(v), UsuarioID: id, ExpiraEm: time.Now().Add(time.Hour).UTC().Format(time.RFC3339)})
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func request(a *App, method, path, session string, data url.Values) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if session != "" {
		r.AddCookie(&http.Cookie{Name: "session", Value: session})
		r.Header.Set("X-CSRF-Token", digest("csrf:"+session))
	}
	w := httptest.NewRecorder()
	a.Routes().ServeHTTP(w, r)
	return w
}
func issue(a *App, passenger, seat string) (int64, error) {
	tx, e := a.DB.BeginTx(context.Background(), nil)
	if e != nil {
		return 0, e
	}
	defer tx.Rollback()
	id, e := a.save(context.Background(), tx, User{ID: 2, Role: "atendente"}, "bilhete", 0, Row{"passageiro_id": passenger, "voo_id": "1", "assento": seat})
	if e != nil {
		return 0, e
	}
	return id, tx.Commit()
}
func TestConcurrentSeatReservation(t *testing.T) {
	a := testApp(t)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, e := issue(a, "2", "33"); results <- e }()
	}
	wg.Wait()
	close(results)
	success := 0
	for e := range results {
		if e == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("expected exactly one seat reservation, got %d", success)
	}
	if n := count(context.Background(), a.DB, "SELECT count(*) FROM bilhete WHERE voo_id=1 AND assento=33 AND situacao<>'Cancelado'"); n != 1 {
		t.Fatal(n)
	}
}
func TestTicketLifecycleAndSeatRelease(t *testing.T) {
	a := testApp(t)
	id, e := issue(a, "2", "24")
	if e != nil {
		t.Fatal(e)
	}
	act := func(target, reason, gate string) error {
		tx, e := a.DB.BeginTx(context.Background(), nil)
		if e != nil {
			return e
		}
		defer tx.Rollback()
		r, _ := one(context.Background(), tx, "bilhete", id)
		e = a.transition(context.Background(), tx, User{ID: 2, Role: "atendente"}, "bilhete", id, target, reason, gate, r["versao"], true)
		if e != nil {
			return e
		}
		return tx.Commit()
	}
	if e = act("Embarque Realizado", "", ""); e == nil {
		t.Fatal("boarding before check-in must fail")
	}
	if e = act("Check-in Realizado", "", ""); e == nil {
		t.Fatal("check-in without gate must fail")
	}
	if e = act("Check-in Realizado", "", "B12"); e != nil {
		t.Fatal(e)
	}
	if e = act("Cancelado", "Pedido do passageiro", ""); e != nil {
		t.Fatal(e)
	}
	if _, e = issue(a, "3", "24"); e != nil {
		t.Fatal("cancelled seat must be reusable", e)
	}
	r, _ := one(context.Background(), a.DB, "passageiro", 2)
	if r["situacao"] != "Inativo" {
		t.Fatal(r)
	}
}
func TestPassengerIsolationAndCSRF(t *testing.T) {
	a := testApp(t)
	id, e := issue(a, "2", "25")
	if e != nil {
		t.Fatal(e)
	}
	s := session(t, a, 3)
	if w := request(a, "GET", fmt.Sprintf("/detail/bilhete/%d", id), s, nil); w.Code != 403 {
		t.Fatalf("other passenger ticket: %d", w.Code)
	}
	if w := request(a, "GET", "/export/bilhete", s, nil); w.Code != 200 || strings.Contains(w.Body.String(), "Lucas Santos") {
		t.Fatal("CSV exposed another passenger", w.Body.String())
	}
	if w := request(a, "POST", "/save/voo", s, url.Values{}); w.Code != 200 && w.Code != 403 {
		t.Fatal(w.Code)
	}
	r := httptest.NewRequest("POST", "/logout", strings.NewReader("csrf=invalid"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{Name: "session", Value: s})
	w := httptest.NewRecorder()
	a.Routes().ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("CSRF should be rejected", w.Code)
	}
}
func TestBatchRollbackAndOptimisticConcurrency(t *testing.T) {
	a := testApp(t)
	s := session(t, a, 2)
	data := url.Values{"changed": {"2", "3"}}
	for _, id := range []string{"2", "3"} {
		r, _ := one(context.Background(), a.DB, "passageiro", num(id))
		for _, f := range entities["passageiro"].Fields {
			data.Set("cell_"+id+"_"+f.Key, r[f.Key])
		}
		data.Set("cell_"+id+"_versao", r["versao"])
	}
	data.Set("cell_2_nome", "Nome que deve reverter")
	data.Set("cell_3_milhagem", "-1")
	w := request(a, "POST", "/batch/passageiro", s, data)
	if w.Code != 422 {
		t.Fatal(w.Code, w.Body.String())
	}
	r, _ := one(context.Background(), a.DB, "passageiro", 2)
	if r["nome"] != "Lucas Santos" {
		t.Fatal("batch partially committed", r)
	}
	data.Del("changed")
	data.Add("changed", "2")
	data.Set("cell_2_versao", "999")
	if w = request(a, "POST", "/batch/passageiro", s, data); w.Code != 422 {
		t.Fatal("stale version must fail", w.Code)
	}
	data.Set("cell_2_versao", "1")
	data.Set("cell_2_nome", "Lucas Atualizado")
	if w = request(a, "POST", "/batch/passageiro", s, data); w.Code != 204 {
		t.Fatal(w.Code, w.Body.String())
	}
}
func TestFlightCancellationCascadesAndClock(t *testing.T) {
	a := testApp(t)
	tx, e := a.DB.BeginTx(context.Background(), nil)
	if e != nil {
		t.Fatal(e)
	}
	e = a.transition(context.Background(), tx, User{ID: 1, Role: "gerente", Company: 1}, "voo", 1, "Cancelado", "Cancelamento operacional", "", "1", true)
	if e != nil {
		tx.Rollback()
		t.Fatal(e)
	}
	if e = tx.Commit(); e != nil {
		t.Fatal(e)
	}
	for _, k := range []string{"voo", "escala", "bilhete"} {
		r, _ := one(context.Background(), a.DB, k, 1)
		if !strings.HasPrefix(r["situacao"], "Cancelad") {
			t.Fatal(k, r)
		}
	}
	start := time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339)
	end := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	if _, e = a.DB.Exec("UPDATE voo SET partida=?,chegada=? WHERE id=2", start, end); e != nil {
		t.Fatal(e)
	}
	if e = a.Sync(context.Background()); e != nil {
		t.Fatal(e)
	}
	r, _ := one(context.Background(), a.DB, "voo", 2)
	if r["situacao"] != "Concluído" {
		t.Fatal(r)
	}
	r, _ = one(context.Background(), a.DB, "escala", 2)
	if r["situacao"] != "Concluída" {
		t.Fatal(r)
	}
}
func TestFlightValidationAndMandatoryInitialStop(t *testing.T) {
	a := testApp(t)
	u := User{ID: 1, Role: "gerente", Company: 1}
	v := Row{"codigo_voo": "", "empresa_id": "1", "aeronave_id": "1", "origem_id": "1", "destino_id": "2", "partida": time.Now().In(a.Location).AddDate(0, 0, 10).Format("2006-01-02T15:04"), "chegada": time.Now().In(a.Location).AddDate(0, 0, 10).Add(time.Hour).Format("2006-01-02T15:04")}
	tx, e := a.DB.BeginTx(context.Background(), nil)
	if e != nil {
		t.Fatal(e)
	}
	id, e := a.save(context.Background(), tx, u, "voo", 0, v)
	if e != nil {
		tx.Rollback()
		t.Fatal(e)
	}
	if e = tx.Commit(); e != nil {
		t.Fatal(e)
	}
	if n := count(context.Background(), a.DB, "SELECT count(*) FROM escala WHERE voo_id=? AND situacao='Programada'", id); n != 1 {
		t.Fatal("flight must start with a stop", n)
	}
	v["aeronave_id"] = "4"
	tx, _ = a.DB.BeginTx(context.Background(), nil)
	_, e = a.save(context.Background(), tx, u, "voo", 0, v)
	tx.Rollback()
	if e == nil {
		t.Fatal("foreign company aircraft accepted")
	}
}
func TestAllPagesRenderAndLogin(t *testing.T) {
	a := testApp(t)
	for _, uid := range []int64{1, 2, 3} {
		s := session(t, a, uid)
		for _, path := range []string{"/", "/?entity=bilhete", "/detail/voo/1", "/print/1"} {
			w := request(a, "GET", path, s, nil)
			if w.Code != 200 {
				t.Fatalf("user %d %s: %d %s", uid, path, w.Code, w.Body.String())
			}
		}
	}
	w := request(a, "GET", "/login", "", nil)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Acesse seu espaço") {
		t.Fatal(w.Code, w.Body.String())
	}
	c := w.Result().Cookies()[0]
	d := url.Values{"csrf": {c.Value}, "email": {"gerente@demo.local"}, "password": {"Demo@2026!"}}
	r := httptest.NewRequest("POST", "/login", strings.NewReader(d.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(c)
	w = httptest.NewRecorder()
	a.Routes().ServeHTTP(w, r)
	if w.Code != 303 || w.Header().Get("Location") != "/" {
		b, _ := io.ReadAll(w.Result().Body)
		t.Fatal(w.Code, string(b))
	}
}
