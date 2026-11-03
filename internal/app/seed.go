package app

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strings"
	"time"
)

func (a *App) CreateUser(ctx context.Context, email, name, password, role string, company, passenger int64) error {
	if len(password) < 12 && os.Getenv("DEMO_MODE") != "true" {
		return fail("Use senha de pelo menos 12 caracteres")
	}
	if role != "gerente" && role != "atendente" && role != "passageiro" {
		return fail("Papel inválido")
	}
	if !strings.Contains(email, "@") || name == "" {
		return fail("Nome e e-mail são obrigatórios")
	}
	if role == "gerente" && company < 1 {
		return fail("Gerente precisa de empresa")
	}
	if role == "passageiro" && passenger < 1 {
		return fail("Passageiro precisa de cadastro")
	}
	hash, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if e != nil {
		return e
	}
	var c, p interface{}
	if company > 0 {
		c = company
	}
	if passenger > 0 {
		p = passenger
	}
	_, e = a.DB.ExecContext(ctx, "INSERT INTO usuario(email,nome,senha_hash,papel,empresa_id,passageiro_id) VALUES(?,?,?,?,?,?)", strings.ToLower(email), name, string(hash), role, c, p)
	return e
}
func (a *App) Bootstrap(ctx context.Context) error {
	n := count(ctx, a.DB, "SELECT count(*) FROM usuario")
	if n < 0 {
		return fail("Falha ao consultar usuários")
	}
	if n > 0 {
		return nil
	}
	if os.Getenv("DEMO_MODE") == "true" {
		return a.SeedDemo(ctx)
	}
	password := os.Getenv("ADMIN_PASSWORD")
	email := os.Getenv("ADMIN_EMAIL")
	if len(password) < 12 || !strings.Contains(email, "@") {
		return fail("Primeiro acesso: configure ADMIN_EMAIL e ADMIN_PASSWORD (mínimo 12 caracteres), ou DEMO_MODE=true somente para demonstração")
	}
	tx, e := a.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	res, e := tx.ExecContext(ctx, "INSERT INTO empresa(nome,cnpj) VALUES(?,?)", env("COMPANY_NAME", "Minha companhia aérea"), env("COMPANY_CNPJ", "00000000000000"))
	if e != nil {
		return e
	}
	id, e := res.LastInsertId()
	if e != nil {
		return e
	}
	h, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if e != nil {
		return e
	}
	if _, e = tx.ExecContext(ctx, "INSERT INTO usuario(email,nome,senha_hash,papel,empresa_id) VALUES(?,?,?,'gerente',?)", strings.ToLower(email), "Gerente", string(h), id); e != nil {
		return e
	}
	return tx.Commit()
}
func (a *App) SeedDemo(ctx context.Context) error {
	tx, e := a.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	statements := []string{
		"INSERT INTO pais(id,nome,continente,sigla) VALUES(1,'Brasil','América do Sul','BR')",
		"INSERT INTO estado(id,pais_id,nome,regiao,sigla) VALUES(1,1,'São Paulo','Sudeste','SP'),(2,1,'Rio de Janeiro','Sudeste','RJ'),(3,1,'Distrito Federal','Centro-Oeste','DF'),(4,1,'Pernambuco','Nordeste','PE')",
		"INSERT INTO cidade(id,estado_id,nome,sigla) VALUES(1,1,'Guarulhos','GRU'),(2,2,'Rio de Janeiro','RIO'),(3,3,'Brasília','BSB'),(4,4,'Recife','REC')",
		"UPDATE cidade SET situacao='Ativa'",
		"INSERT INTO empresa(id,nome,cnpj) VALUES(1,'Erlaine Linhas Aéreas','12345678000190'),(2,'Companhia Exemplo','98765432000100')",
		"INSERT INTO aeroporto(id,cidade_id,nome,sigla,situacao) VALUES(1,1,'São Paulo / Guarulhos','GRU','Ativo'),(2,2,'Rio de Janeiro / Galeão','GIG','Ativo'),(3,3,'Brasília Internacional','BSB','Ativo'),(4,4,'Recife / Guararapes','REC','Ativo')",
		"INSERT INTO aeronave(id,empresa_id,codigo_aeronave,modelo,capacidade,situacao) VALUES(1,1,'PR-ERL','Airbus A320neo',180,'Ativa'),(2,1,'PR-ERA','Embraer E195-E2',132,'Ativa'),(3,1,'PR-ERE','Boeing 737-800',186,'Em Manutenção'),(4,2,'PR-OUT','ATR 72',70,'Ativa')",
		"INSERT INTO passageiro(id,documento,nome,milhagem,telefone) VALUES(1,'12345678901','Marina Oliveira',12500,'11999990000'),(2,'98765432100','Lucas Santos',4200,'21988880000'),(3,'AB1234567','Ana Costa',0,'61977770000')",
	}
	for _, s := range statements {
		if _, e = tx.ExecContext(ctx, s); e != nil {
			return e
		}
	}
	base := time.Now().In(a.Location).AddDate(0, 0, 1)
	for i := 0; i < 6; i++ {
		start := time.Date(base.Year(), base.Month(), base.Day()+i/3, 8+(i%3)*4, 0, 0, 0, a.Location).UTC()
		dest := 2 + i%3
		craft := 1 + i%2
		res, e := tx.ExecContext(ctx, "INSERT INTO voo(codigo_voo,empresa_id,aeronave_id,origem_id,destino_id,partida,chegada) VALUES(?,1,?,1,?,?,?)", fmt.Sprintf("ER %04d", 2101+i), craft, dest, start.Format(time.RFC3339), start.Add(2*time.Hour).Format(time.RFC3339))
		if e != nil {
			return e
		}
		id, _ := res.LastInsertId()
		if _, e = tx.ExecContext(ctx, "INSERT INTO escala(voo_id,sequencial,origem_id,destino_id,tempo_escala,valor_centavos,situacao) VALUES(?,1,1,?,120,45990,'Programada')", id, dest); e != nil {
			return e
		}
	}
	if _, e = tx.ExecContext(ctx, "INSERT INTO bilhete(codigo_bilhete,passageiro_id,voo_id,origem_id,destino_id,assento,hora_prev_partida,hora_prev_chegada) SELECT 'BLT-DEMO-001',1,id,origem_id,destino_id,12,partida,chegada FROM voo WHERE id=1"); e != nil {
		return e
	}
	if _, e = tx.ExecContext(ctx, "UPDATE passageiro SET situacao='Ativo' WHERE id=1"); e != nil {
		return e
	}
	h, e := bcrypt.GenerateFromPassword([]byte("Demo@2026!"), bcrypt.DefaultCost)
	if e != nil {
		return e
	}
	for _, u := range []struct {
		email, name, role string
		c, p              interface{}
	}{{"gerente@demo.local", "Erlaine Gerente", "gerente", 1, nil}, {"atendente@demo.local", "Equipe de atendimento", "atendente", nil, nil}, {"passageiro@demo.local", "Marina Oliveira", "passageiro", nil, 1}} {
		if _, e = tx.ExecContext(ctx, "INSERT INTO usuario(email,nome,senha_hash,papel,empresa_id,passageiro_id) VALUES(?,?,?,?,?,?)", u.email, u.name, string(h), u.role, u.c, u.p); e != nil {
			return e
		}
	}
	return tx.Commit()
}
