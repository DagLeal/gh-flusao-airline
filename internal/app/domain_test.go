package app

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestGeographicActivationAndFinalDeletion(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()
	u := User{ID: 1, Role: "gerente", Company: 1}
	save := func(k string, v Row) int64 {
		t.Helper()
		tx, e := a.DB.BeginTx(ctx, nil)
		if e != nil {
			t.Fatal(e)
		}
		id, e := a.save(ctx, tx, u, k, 0, v)
		if e != nil {
			tx.Rollback()
			t.Fatal(e)
		}
		if e = tx.Commit(); e != nil {
			t.Fatal(e)
		}
		return id
	}
	country := save("pais", Row{"nome": "Portugal", "continente": "Europa", "sigla": "PT"})
	row, _ := one(ctx, a.DB, "pais", country)
	if row["situacao"] != "Cadastrado" {
		t.Fatal(row)
	}
	state := save("estado", Row{"pais_id": fmt.Sprint(country), "nome": "Lisboa", "regiao": "Centro", "sigla": "LI"})
	row, _ = one(ctx, a.DB, "pais", country)
	if row["situacao"] != "Ativo" {
		t.Fatal(row)
	}
	city := save("cidade", Row{"estado_id": fmt.Sprint(state), "nome": "Lisboa", "sigla": "LIS"})
	row, _ = one(ctx, a.DB, "cidade", city)
	if row["situacao"] != "Cadastrada" {
		t.Fatal(row)
	}
	save("aeroporto", Row{"cidade_id": fmt.Sprint(city), "nome": "Humberto Delgado", "sigla": "LIS"})
	row, _ = one(ctx, a.DB, "cidade", city)
	if row["situacao"] != "Ativa" {
		t.Fatal(row)
	}
	row, _ = one(ctx, a.DB, "pais", country)
	tx, _ := a.DB.BeginTx(ctx, nil)
	e := a.transition(ctx, tx, u, "pais", country, "Inativo", "Encerramento", "", row["versao"], true)
	tx.Rollback()
	if e == nil {
		t.Fatal("country with active state must not be disabled")
	}
}
func TestInlinePassengerRollbackAndCapacity(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()
	u := User{ID: 2, Role: "atendente"}
	tx, _ := a.DB.BeginTx(ctx, nil)
	_, e := a.save(ctx, tx, u, "bilhete", 0, Row{"voo_id": "1", "assento": "12", "new_passenger": "yes", "new_nome": "Novo passageiro", "new_documento": "66554433221"})
	tx.Rollback()
	if e == nil {
		t.Fatal("occupied seat must fail")
	}
	if n := count(ctx, a.DB, "SELECT count(*) FROM passageiro WHERE documento='66554433221'"); n != 0 {
		t.Fatal("inline passenger must roll back on ticket failure")
	}
	tx, _ = a.DB.BeginTx(ctx, nil)
	id, e := a.save(ctx, tx, u, "bilhete", 0, Row{"voo_id": "1", "assento": "40", "new_passenger": "yes", "new_nome": "Novo passageiro", "new_documento": "66554433221"})
	if e != nil {
		tx.Rollback()
		t.Fatal(e)
	}
	if e = tx.Commit(); e != nil {
		t.Fatal(e)
	}
	r, _ := one(ctx, a.DB, "bilhete", id)
	p, _ := one(ctx, a.DB, "passageiro", num(r["passageiro_id"]))
	if p["situacao"] != "Ativo" {
		t.Fatal(p)
	}
	if _, e = issue(a, "2", "181"); e == nil {
		t.Fatal("seat beyond aircraft capacity must fail")
	}
}
func TestMaintenanceAndImmutableDocument(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()
	tx, _ := a.DB.BeginTx(ctx, nil)
	e := a.transition(ctx, tx, User{ID: 1, Role: "gerente", Company: 1}, "aeronave", 1, "Em Manutenção", "Revisão programada", "", "1", true)
	tx.Rollback()
	if e == nil {
		t.Fatal("aircraft with scheduled flights must not enter maintenance")
	}
	tx, _ = a.DB.BeginTx(ctx, nil)
	r, _ := one(ctx, tx, "passageiro", 2)
	r["documento"] = "99887766554"
	_, e = a.save(ctx, tx, User{ID: 2, Role: "atendente"}, "passageiro", 2, r)
	tx.Rollback()
	if e == nil {
		t.Fatal("document mutation must fail")
	}
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	if _, e = a.DB.Exec("UPDATE voo SET partida=? WHERE id=1", past); e != nil {
		t.Fatal(e)
	}
	tx, _ = a.DB.BeginTx(ctx, nil)
	e = a.transition(ctx, tx, User{ID: 3, Role: "passageiro", Passenger: 1}, "bilhete", 1, "Check-in Realizado", "", "B12", "1", true)
	tx.Rollback()
	if e == nil {
		t.Fatal("check-in after departure must fail")
	}
}
