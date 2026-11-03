package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"erlaine/aeroporto/internal/database"
)

func advancedTestValues(t *testing.T, q AdvancedQuery) Row {
	t.Helper()
	v, e := advancedValues(q, url.Values{})
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func runAdvancedTest(t *testing.T, a *App, id int, v url.Values) AdvancedTable {
	t.Helper()
	q, ok := advancedQuery(id)
	if !ok {
		t.Fatal(id)
	}
	values, e := advancedValues(q, v)
	if e != nil {
		t.Fatal(e)
	}
	result, e := advancedRead(context.Background(), a.DB, User{ID: 1, Role: "gerente", Company: 1}, q, values, 1)
	if e != nil {
		t.Fatalf("query %d: %v", id, e)
	}
	return result
}
func TestAdvancedAllFortyQueriesAndOwnership(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()
	if len(advancedCatalog) != 40 {
		t.Fatal("catalog size", len(advancedCatalog))
	}
	if _, e := a.DB.Exec("INSERT INTO voo(id,codigo_voo,empresa_id,aeronave_id,origem_id,destino_id,partida,chegada,valor_centavos) VALUES(99,'SECRET-OTHER-COMPANY',2,4,1,2,?,?,9999999)", time.Now().AddDate(0, 0, 20).UTC().Format(time.RFC3339), time.Now().AddDate(0, 0, 20).Add(time.Hour).UTC().Format(time.RFC3339)); e != nil {
		t.Fatal(e)
	}
	if _, e := a.DB.Exec("INSERT INTO escala(id,voo_id,sequencial,origem_id,destino_id,tempo_escala,valor_centavos,situacao) VALUES(99,99,1,1,2,180,9999999,'Programada')"); e != nil {
		t.Fatal(e)
	}
	seen := map[int]bool{}
	for _, q := range advancedCatalog {
		if seen[q.ID] {
			t.Fatal("duplicate id", q.ID)
		}
		seen[q.ID] = true
		r := runAdvancedTest(t, a, q.ID, nil)
		if strings.Contains(fmt.Sprint(r.Raw), "SECRET-OTHER-COMPANY") {
			t.Fatal("foreign flight exposed", q.ID)
		}
		if len(r.Columns) == 0 {
			t.Fatal("missing columns", q.ID)
		}
	}
	counts := runAdvancedTest(t, a, 1, nil)
	if len(counts.Raw) != 10 {
		t.Fatal(counts.Raw)
	}
	for _, r := range counts.Raw {
		if r["TABELA"] == "VOO" && r["QUANTIDADE"] != "6" {
			t.Fatal(r)
		}
	}
	average := runAdvancedTest(t, a, 9, nil)
	if average.Raw[0]["media_valor"] != "459.9" {
		t.Fatal(average.Raw)
	}
	values := url.Values{"empresa": {"Azul' OR 1=1 --"}}
	if r := runAdvancedTest(t, a, 4, values); len(r.Raw) != 0 {
		t.Fatal("filter injection")
	}
	if n := count(ctx, a.DB, "SELECT count(*) FROM voo"); n != 7 {
		t.Fatal("read mutated data")
	}
}
func TestAdvancedFiltersCountsAndPagination(t *testing.T) {
	a := testApp(t)
	a.DB.Exec("UPDATE voo SET valor_centavos=CASE id WHEN 1 THEN 50000 WHEN 2 THEN 80000 WHEN 3 THEN 95000 WHEN 4 THEN 95000 WHEN 5 THEN 140000 ELSE 200000 END,confirmado=CASE WHEN id=1 THEN 1 ELSE 0 END")
	if r := runAdvancedTest(t, a, 5, url.Values{"valor": {"800,00"}}); len(r.Raw) != 4 {
		t.Fatal(r.Raw)
	}
	if r := runAdvancedTest(t, a, 21, nil); len(r.Raw) != 1 || r.Raw[0]["codigo_voo"] != "ER 2101" {
		t.Fatal(r.Raw)
	}
	if r := runAdvancedTest(t, a, 23, nil); len(r.Raw) != 2 {
		t.Fatal(r.Raw)
	}
	a.DB.Exec("UPDATE voo SET valor_centavos=50000 WHERE id=2")
	if r := runAdvancedTest(t, a, 23, nil); len(r.Raw) != 3 {
		t.Fatal("ties must be included", r.Raw)
	}
	a.DB.Exec("DELETE FROM escala WHERE voo_id=6")
	if r := runAdvancedTest(t, a, 14, nil); len(r.Raw) != 6 || r.Raw[5]["total_escalas"] != "0" {
		t.Fatal("zero-stop flight", r.Raw)
	}
	for i := 0; i < 55; i++ {
		if _, e := a.DB.Exec("INSERT INTO aeroporto(cidade_id,nome,sigla) VALUES(1,?,?)", fmt.Sprintf("Airport %d", i), fmt.Sprintf("%03d", i)); e != nil {
			t.Fatal(e)
		}
	}
	q, _ := advancedQuery(17)
	v := advancedTestValues(t, q)
	u := User{ID: 1, Role: "gerente", Company: 1}
	first, e := advancedRead(context.Background(), a.DB, u, q, v, 1)
	if e != nil {
		t.Fatal(e)
	}
	second, e := advancedRead(context.Background(), a.DB, u, q, v, 2)
	if e != nil {
		t.Fatal(e)
	}
	if len(first.Raw) != 50 || !first.HasNext || len(second.Raw) != 9 || second.HasNext {
		t.Fatal(len(first.Raw), first.HasNext, len(second.Raw), second.HasNext)
	}
}
func TestAdvancedRoutesRejectOtherProfilesAndSQLInput(t *testing.T) {
	a := testApp(t)
	for _, uid := range []int64{2, 3} {
		s := session(t, a, uid)
		for _, path := range []string{"/consultas-avancadas", "/consultas-avancadas/resultado?id=1", "/consultas-avancadas/csv?id=17"} {
			w := request(a, http.MethodGet, path, s, nil)
			if w.Code != 403 {
				t.Fatal(uid, path, w.Code)
			}
		}
		w := request(a, http.MethodPost, "/consultas-avancadas/operacao", s, url.Values{"id": {"35"}, "nome": {"forbidden"}})
		if w.Code != 403 {
			t.Fatal(w.Code)
		}
		if w = request(a, http.MethodGet, "/", s, nil); strings.Contains(w.Body.String(), ">Consultas avançadas<") {
			t.Fatal("manager menu shown")
		}
	}
	s := session(t, a, 1)
	if w := request(a, http.MethodGet, "/consultas-avancadas?id=1", s, nil); w.Code != 200 || !strings.Contains(w.Body.String(), "Catálogo SQL") {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := request(a, http.MethodGet, "/consultas-avancadas/resultado?id=4&empresa=Azul%27%20OR%201%3D1--", s, nil); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := request(a, http.MethodGet, "/consultas-avancadas?id=999", s, nil); w.Code != 400 {
		t.Fatal(w.Code)
	}
	if w := request(a, http.MethodGet, "/consultas-avancadas?id=5&valor=NaN", s, nil); w.Code != 400 {
		t.Fatal(w.Code)
	}
	if w := request(a, http.MethodGet, "/consultas-avancadas/csv?id=1", s, nil); w.Code != 200 || !strings.Contains(w.Header().Get("Content-Disposition"), "consulta-01") {
		t.Fatal(w.Code)
	}
	if w := request(a, http.MethodPost, "/consultas-avancadas/operacao", s, url.Values{"id": {"1"}, "sql": {"DELETE FROM voo"}}); w.Code != 422 {
		t.Fatal(w.Code)
	}
	r := request(a, http.MethodGet, "/consultas-avancadas/resultado?id=13&codigo=ER+2101", s, nil)
	if r.Header().Get("HX-Push-Url") != "/consultas-avancadas?codigo=ER+2101&id=13&page=1" {
		t.Fatal(r.Header().Get("HX-Push-Url"))
	}
}
func advancedMutateTest(a *App, u User, id int, v Row) (int64, error) {
	tx, e := a.DB.BeginTx(context.Background(), nil)
	if e != nil {
		return 0, e
	}
	defer tx.Rollback()
	record, e := a.mutateAdvanced(context.Background(), tx, u, id, v)
	if e != nil {
		return 0, e
	}
	return record, tx.Commit()
}
func TestAdvancedMutationLifecycleAuditAndGuardedDeletes(t *testing.T) {
	a := testApp(t)
	u := User{ID: 1, Role: "gerente", Company: 1}
	ctx := context.Background()
	airport, e := advancedMutateTest(a, u, 35, Row{"nome": "Aeroporto de Santos Dumont", "sigla": "SDU", "cidade_id": "2"})
	if e != nil {
		t.Fatal(e)
	}
	start := time.Now().In(a.Location).AddDate(0, 0, 10)
	v := Row{"codigo": "G31111", "aeronave_id": "1", "origem_id": "1", "destino_id": "2", "partida": start.Format("2006-01-02T15:04"), "chegada": start.Add(time.Hour).Format("2006-01-02T15:04"), "valor": "950.00", "empresa_id": "2"}
	flight, e := advancedMutateTest(a, u, 36, v)
	if e != nil {
		t.Fatal(e)
	}
	f, _ := one(ctx, a.DB, "voo", flight)
	if f["valor_centavos"] != "95000" || f["empresa_id"] != "1" {
		t.Fatal(f)
	}
	op := Row{"registro_id": fmt.Sprint(flight), "versao": f["versao"], "valor": "1050,00", "confirmed": "yes", "reason": "Ajuste comercial"}
	if _, e = advancedMutateTest(a, u, 37, op); e != nil {
		t.Fatal(e)
	}
	if _, e = advancedMutateTest(a, u, 38, op); e == nil {
		t.Fatal("stale version accepted")
	}
	f, _ = one(ctx, a.DB, "voo", flight)
	op["versao"] = f["versao"]
	if _, e = advancedMutateTest(a, u, 38, op); e != nil {
		t.Fatal(e)
	}
	f, _ = one(ctx, a.DB, "voo", flight)
	if f["confirmado"] != "1" || f["situacao"] != "Agendado" || f["valor_centavos"] != "105000" {
		t.Fatal(f)
	}
	scales, e := rows(ctx, a.DB, "SELECT * FROM escala WHERE voo_id=?", flight)
	if e != nil {
		t.Fatal(e)
	}
	del := Row{"registro_id": scales[0]["id"], "versao": scales[0]["versao"], "confirmed": "yes", "reason": "Revisão da rota"}
	if _, e = advancedMutateTest(a, u, 39, del); e == nil {
		t.Fatal("last stop deleted")
	}
	tx, _ := a.DB.BeginTx(ctx, nil)
	_, e = a.save(ctx, tx, u, "escala", 0, Row{"voo_id": fmt.Sprint(flight), "sequencial": "2", "origem_id": "2", "destino_id": "3", "tempo_escala": "60", "valor_centavos": "10000"})
	if e != nil {
		tx.Rollback()
		t.Fatal(e)
	}
	if e = tx.Commit(); e != nil {
		t.Fatal(e)
	}
	if _, e = advancedMutateTest(a, u, 39, del); e != nil {
		t.Fatal(e)
	}
	f, _ = one(ctx, a.DB, "voo", flight)
	if f["origem_id"] != "2" || f["destino_id"] != "3" {
		t.Fatal("remaining route not synchronized", f)
	}
	if count(ctx, a.DB, "SELECT count(*) FROM escala WHERE id=?", del["registro_id"]) != 0 {
		t.Fatal("delete failed")
	}
	ar, _ := one(ctx, a.DB, "aeroporto", airport)
	if _, e = advancedMutateTest(a, u, 40, Row{"airport": fmt.Sprint(airport) + ":" + ar["versao"], "confirmed": "yes", "reason": "Registro de teste"}); e != nil {
		t.Fatal(e)
	}
	next, e := advancedMutateTest(a, u, 35, Row{"nome": "Outro aeroporto", "sigla": "XXX", "cidade_id": "2"})
	if e != nil {
		t.Fatal(e)
	}
	if next <= airport {
		t.Fatal("audited identity reused", next, airport)
	}
	if count(ctx, a.DB, "SELECT count(*) FROM auditoria WHERE acao LIKE 'Consulta avançada %'") != 7 {
		t.Fatal("audit count")
	}
	if _, e = advancedMutateTest(a, u, 40, Row{"airport": "1:1", "confirmed": "yes", "reason": "Teste de vínculo"}); e == nil {
		t.Fatal("referenced airport deleted")
	}
}
func TestAdvancedForeignCompanyMutationAndGlobalAirportReferences(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()
	u := User{ID: 1, Role: "gerente", Company: 1}
	if _, e := a.DB.Exec("INSERT INTO voo(id,codigo_voo,empresa_id,aeronave_id,origem_id,destino_id,partida,chegada) VALUES(99,'OTHER',2,4,1,2,?,?)", time.Now().AddDate(0, 0, 20).UTC().Format(time.RFC3339), time.Now().AddDate(0, 0, 20).Add(time.Hour).UTC().Format(time.RFC3339)); e != nil {
		t.Fatal(e)
	}
	if _, e := advancedMutateTest(a, u, 37, Row{"registro_id": "99", "versao": "1", "valor": "1", "confirmed": "yes"}); e == nil {
		t.Fatal("foreign update")
	}
	if _, e := a.DB.Exec("INSERT INTO aeroporto(id,cidade_id,nome,sigla) VALUES(50,2,'Aeroporto de outra operação','OUT')"); e != nil {
		t.Fatal(e)
	}
	if _, e := a.DB.Exec("UPDATE voo SET origem_id=50 WHERE id=99"); e != nil {
		t.Fatal(e)
	}
	// No scale references this airport, yet a foreign flight does.
	report := runAdvancedTest(t, a, 40, nil)
	found := false
	for _, r := range report.Raw {
		if r["id_aeroporto"] == "50" {
			found = true
		}
	}
	if !found {
		t.Fatal("report must follow the question: no scales")
	}
	if _, e := advancedMutateTest(a, u, 40, Row{"airport": "50:1", "confirmed": "yes", "reason": "Limpeza"}); e == nil {
		t.Fatal("foreign flight reference ignored")
	}
	if count(ctx, a.DB, "SELECT count(*) FROM aeroporto WHERE id=50") != 1 {
		t.Fatal("airport lost")
	}
}
func TestAdvancedLegacyMigrationPreservesData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, e := sql.Open("sqlite", path)
	if e != nil {
		t.Fatal(e)
	}
	legacy := strings.Replace(database.Schema, ",valor_centavos INTEGER NOT NULL DEFAULT 0 CHECK(valor_centavos>=0),confirmado INTEGER NOT NULL DEFAULT 0 CHECK(confirmado IN(0,1))", "", 1)
	if _, e = db.Exec(legacy); e != nil {
		t.Fatal(e)
	}
	statements := []string{
		"INSERT INTO pais VALUES(1,'Brasil','America','BR','Ativo',1)",
		"INSERT INTO estado VALUES(1,1,'Rio de Janeiro','Sudeste','RJ','Ativo',1)",
		"INSERT INTO cidade VALUES(1,1,'Rio de Janeiro','RIO','Ativa',1)",
		"INSERT INTO empresa VALUES(1,'Legacy','12345678000190','Ativa',1)",
		"INSERT INTO aeroporto VALUES(1,1,'A','AAA','Ativo',1),(2,1,'B','BBB','Ativo',1)",
		"INSERT INTO aeronave VALUES(1,1,'LEG','Model',30,'Ativa',1)",
		"INSERT INTO voo VALUES(1,'LEGACY',1,1,1,2,'2030-01-01T12:00:00Z','2030-01-01T13:00:00Z','Agendado',1)",
		"INSERT INTO escala VALUES(1,1,1,1,2,60,7500,'Programada',1),(2,1,2,2,1,60,99900,'Cancelada',1)",
	}
	for _, s := range statements {
		if _, e = db.Exec(s); e != nil {
			t.Fatal(e)
		}
	}
	db.Close()
	a, e := New(path)
	if e != nil {
		t.Fatal(e)
	}
	r, _ := one(context.Background(), a.DB, "voo", 1)
	if r["valor_centavos"] != "7500" || r["confirmado"] != "0" {
		t.Fatal(r)
	}
	a.DB.Exec("UPDATE voo SET valor_centavos=12345,confirmado=1 WHERE id=1")
	a.DB.Close()
	a, e = New(path)
	if e != nil {
		t.Fatal(e)
	}
	defer a.DB.Close()
	r, _ = one(context.Background(), a.DB, "voo", 1)
	if r["valor_centavos"] != "12345" || r["confirmado"] != "1" {
		t.Fatal("migration repeated", r)
	}
	if n := count(context.Background(), a.DB, "SELECT count(*) FROM V1 WHERE id_voo=1"); n != 2 {
		t.Fatal("view missing", n)
	}
}
