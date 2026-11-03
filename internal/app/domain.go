package app

import (
	"context"
	"database/sql"
	"erlaine/aeroporto/internal/database"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func fail(s string) error { return errors.New(s) }
func count(ctx context.Context, q querier, s string, args ...interface{}) int {
	var n int
	if e := q.QueryRowContext(ctx, s, args...).Scan(&n); e != nil {
		return -1
	}
	return n
}
func requireState(ctx context.Context, q querier, k string, id int64, states ...string) (Row, error) {
	r, e := one(ctx, q, k, id)
	if e != nil {
		return nil, fail("Referência não encontrada: " + k)
	}
	for _, s := range states {
		if r["situacao"] == s {
			return r, nil
		}
	}
	return nil, fail("Registro indisponível: " + k + " em " + r["situacao"])
}
func (a *App) canTouch(ctx context.Context, tx *sql.Tx, u User, k string, r Row) error {
	if u.Role == "gerente" {
		switch k {
		case "empresa":
			if num(r["id"]) != u.Company {
				return fail("Empresa fora do seu acesso")
			}
		case "aeronave", "voo":
			if num(r["empresa_id"]) != u.Company {
				return fail("Empresa fora do seu acesso")
			}
		case "escala":
			if r["voo_id"] != "" {
				f, e := one(ctx, tx, "voo", num(r["voo_id"]))
				if e != nil || num(f["empresa_id"]) != u.Company {
					return fail("Voo fora do seu acesso")
				}
			}
		case "bilhete":
			f, e := one(ctx, tx, "voo", num(r["voo_id"]))
			if e != nil || num(f["empresa_id"]) != u.Company {
				return fail("Bilhete fora do seu acesso")
			}
		}
	}
	if u.Role == "passageiro" && (k != "bilhete" || num(r["passageiro_id"]) != u.Passenger) {
		return fail("Acesso negado")
	}
	return nil
}
func editable(k string, r Row) bool {
	switch k {
	case "voo":
		return r["situacao"] == "Agendado"
	case "escala":
		return r["situacao"] == "Planejada"
	case "aeronave":
		return r["situacao"] == "Cadastrada"
	case "aeroporto":
		return r["situacao"] == "Cadastrado"
	case "bilhete":
		return false
	case "passageiro":
		return r["situacao"] == "Cadastrado" || r["situacao"] == "Ativo"
	default:
		return !strings.HasPrefix(r["situacao"], "Exclu") && r["situacao"] != "Inativo" && r["situacao"] != "Inativa"
	}
}
func (a *App) save(ctx context.Context, tx *sql.Tx, u User, k string, id int64, v Row) (int64, error) {
	en, ok := entities[k]
	if !ok || !a.canWrite(u, k) {
		return 0, fail("Você não pode alterar este cadastro")
	}
	if u.Role == "gerente" {
		if _, e := requireState(ctx, tx, "empresa", u.Company, "Ativa"); e != nil {
			return 0, fail("Sua empresa não está operando")
		}
	}
	old := Row{}
	if id > 0 {
		var e error
		old, e = one(ctx, tx, k, id)
		if e != nil {
			return 0, e
		}
		if e = a.canTouch(ctx, tx, u, k, old); e != nil {
			return 0, e
		}
		if !editable(k, old) {
			return 0, fail("Este estado permite apenas leitura")
		}
		if v["versao"] != old["versao"] {
			return 0, fail("Este registro foi alterado por outra pessoa. Atualize a grade antes de salvar")
		}
	}
	if k == "bilhete" && v["new_passenger"] == "yes" {
		passengerID, e := a.save(ctx, tx, u, "passageiro", 0, Row{"documento": v["new_documento"], "nome": v["new_nome"], "telefone": v["new_telefone"], "milhagem": "0", "passaporte": ""})
		if e != nil {
			return 0, e
		}
		v["passageiro_id"] = strconv.FormatInt(passengerID, 10)
	}
	data := Row{}
	for _, f := range en.Fields {
		s := strings.TrimSpace(v[f.Key])
		if len(s) > 500 {
			return 0, fail(f.Label + ": valor muito longo")
		}
		if s == "" && !f.Optional {
			return 0, fail(f.Label + ": campo obrigatório")
		}
		if f.Type == "number" && s != "" {
			n, e := strconv.ParseInt(s, 10, 64)
			if e != nil || n < 0 || n > 1000000000 {
				return 0, fail(f.Label + ": número inteiro inválido")
			}
		}
		if f.Ref != "" && s != "" {
			if _, e := one(ctx, tx, f.Ref, num(s)); e != nil {
				return 0, fail(f.Label + ": referência inválida")
			}
		}
		data[f.Key] = s
	}
	if k == "voo" && data["codigo_voo"] == "" {
		data["codigo_voo"] = "ER-" + strings.ToUpper(token()[:8])
	}
	if u.Role == "gerente" && (k == "aeronave" || k == "voo") {
		data["empresa_id"] = strconv.FormatInt(u.Company, 10)
	}
	if e := a.canTouch(ctx, tx, u, k, data); e != nil && k != "empresa" {
		return 0, e
	}
	if k == "empresa" && id == 0 {
		return 0, fail("Novas empresas devem ser provisionadas pelo administrador com a CLI")
	}
	switch k {
	case "pais":
		if len(data["sigla"]) < 2 || len(data["sigla"]) > 3 {
			return 0, fail("Sigla do país deve ter 2 ou 3 caracteres")
		}
		data["sigla"] = strings.ToUpper(data["sigla"])
	case "estado":
		if _, e := requireState(ctx, tx, "pais", num(data["pais_id"]), "Ativo", "Cadastrado"); e != nil {
			return 0, e
		}
	case "cidade":
		if _, e := requireState(ctx, tx, "estado", num(data["estado_id"]), "Ativo", "Cadastrado"); e != nil {
			return 0, e
		}
	case "empresa":
		if !regexp.MustCompile("^[0-9]{14}$").MatchString(data["cnpj"]) {
			return 0, fail("CNPJ deve conter 14 dígitos")
		}
	case "aeroporto":
		data["sigla"] = strings.ToUpper(data["sigla"])
		if !regexp.MustCompile("^[A-Z0-9]{3}$").MatchString(data["sigla"]) {
			return 0, fail("Sigla deve ter 3 caracteres alfanuméricos")
		}
		if _, e := requireState(ctx, tx, "cidade", num(data["cidade_id"]), "Ativa", "Cadastrada"); e != nil {
			return 0, e
		}
	case "aeronave":
		if num(data["capacidade"]) < 1 || num(data["capacidade"]) > 1000 {
			return 0, fail("Capacidade deve ser de 1 a 1000 assentos")
		}
	case "voo":
		for _, key := range []string{"partida", "chegada"} {
			t, e := time.ParseInLocation("2006-01-02T15:04", data[key], a.Location)
			if e != nil {
				return 0, fail("Data/hora inválida")
			}
			data[key] = t.UTC().Format(time.RFC3339)
		}
		if data["chegada"] <= data["partida"] {
			return 0, fail("Chegada deve ser posterior à partida")
		}
		if data["partida"] <= a.now() {
			return 0, fail("Agende a partida para uma data futura")
		}
		craft, e := requireState(ctx, tx, "aeronave", num(data["aeronave_id"]), "Ativa")
		if e != nil {
			return 0, e
		}
		if craft["empresa_id"] != data["empresa_id"] {
			return 0, fail("Aeronave pertence a outra empresa")
		}
		if count(ctx, tx, "SELECT count(*) FROM voo WHERE aeronave_id=? AND id<>? AND situacao IN('Agendado','Em Voo') AND partida<? AND chegada>?", data["aeronave_id"], id, data["chegada"], data["partida"]) != 0 {
			return 0, fail("Aeronave já possui voo nesse intervalo")
		}
		if id > 0 && count(ctx, tx, "SELECT count(*) FROM bilhete WHERE voo_id=? AND situacao<>'Cancelado'", id) > numInt(craft["capacidade"]) {
			return 0, fail("A nova aeronave não comporta os bilhetes emitidos")
		}
		if id > 0 {
			if count(ctx, tx, "SELECT coalesce(max(assento),0) FROM bilhete WHERE voo_id=? AND situacao<>'Cancelado'", id) > numInt(craft["capacidade"]) {
				return 0, fail("Aeronave menor não comporta os números de assento já emitidos")
			}
			routeChanged := data["origem_id"] != old["origem_id"] || data["destino_id"] != old["destino_id"]
			if routeChanged && (count(ctx, tx, "SELECT count(*) FROM bilhete WHERE voo_id=? AND situacao<>'Cancelado'", id) > 0 || count(ctx, tx, "SELECT count(*) FROM escala WHERE voo_id=?", id) > 1) {
				return 0, fail("Cancele os bilhetes ou revise as escalas antes de alterar a rota")
			}
			if routeChanged {
				if _, e := tx.ExecContext(ctx, "UPDATE escala SET origem_id=?,destino_id=?,versao=versao+1 WHERE voo_id=? AND origem_id=? AND destino_id=?", data["origem_id"], data["destino_id"], id, old["origem_id"], old["destino_id"]); e != nil {
					return 0, e
				}
			}
		}
		if data["origem_id"] == data["destino_id"] {
			return 0, fail("Origem e destino devem ser diferentes")
		}
		for _, key := range []string{"origem_id", "destino_id"} {
			if _, e := requireState(ctx, tx, "aeroporto", num(data[key]), "Ativo"); e != nil {
				return 0, e
			}
		}
	case "escala":
		if num(data["sequencial"]) < 1 || num(data["tempo_escala"]) < 1 {
			return 0, fail("Sequência e tempo devem ser positivos")
		}
		if data["origem_id"] == data["destino_id"] {
			return 0, fail("Origem e destino devem ser diferentes")
		}
		for _, key := range []string{"origem_id", "destino_id"} {
			if _, e := requireState(ctx, tx, "aeroporto", num(data[key]), "Ativo"); e != nil {
				return 0, e
			}
		}
		if data["voo_id"] != "" {
			if _, e := requireState(ctx, tx, "voo", num(data["voo_id"]), "Agendado"); e != nil {
				return 0, e
			}
			data["situacao"] = "Programada"
		}
	case "passageiro":
		if id > 0 && data["documento"] != old["documento"] {
			return 0, fail("Documento é imutável após criação")
		}
		if id > 0 && old["situacao"] == "Ativo" && data["nome"] != old["nome"] {
			return 0, fail("Nome de passageiro ativo não pode ser alterado")
		}
		if !regexp.MustCompile("^(?:[0-9]{11}|[A-Za-z]{1,3}[0-9]{5,9})$").MatchString(data["documento"]) {
			return 0, fail("Informe CPF com 11 dígitos ou passaporte com letras e números")
		}
	case "bilhete":
		if _, e := requireState(ctx, tx, "passageiro", num(data["passageiro_id"]), "Cadastrado", "Ativo"); e != nil {
			return 0, e
		}
		f, e := requireState(ctx, tx, "voo", num(data["voo_id"]), "Agendado", "Em Voo")
		if e != nil {
			return 0, e
		}
		if count(ctx, tx, "SELECT count(*) FROM escala WHERE voo_id=? AND situacao IN('Programada','Em Andamento','Concluída')", f["id"]) < 1 {
			return 0, fail("Voo precisa de ao menos uma escala programada antes de emitir bilhetes")
		}
		craft, e := one(ctx, tx, "aeronave", num(f["aeronave_id"]))
		if e != nil {
			return 0, e
		}
		if num(data["assento"]) < 1 || num(data["assento"]) > num(craft["capacidade"]) {
			return 0, fail("Assento fora da capacidade da aeronave")
		}
		if count(ctx, tx, "SELECT count(*) FROM bilhete WHERE voo_id=? AND situacao<>'Cancelado'", f["id"]) >= numInt(craft["capacidade"]) {
			return 0, fail("Todos os assentos estão ocupados")
		}
		data["codigo_bilhete"] = "BLT-" + strings.ToUpper(token()[:12])
		data["origem_id"] = f["origem_id"]
		data["destino_id"] = f["destino_id"]
		data["hora_prev_partida"] = f["partida"]
		data["hora_prev_chegada"] = f["chegada"]
	}
	if id == 0 {
		if data["situacao"] == "" {
			data["situacao"] = en.Initial
		}
	}
	keys, vals, params := []string{}, []interface{}{}, []string{}
	// Column names originate only from the immutable server-side metadata and the explicit derived fields above.
	for key, value := range data {
		keys = append(keys, key)
		if value == "" && key == "voo_id" {
			vals = append(vals, nil)
		} else {
			vals = append(vals, value)
		}
		params = append(params, "?")
	}
	if id == 0 {
		res, e := tx.ExecContext(ctx, "INSERT INTO "+k+"("+strings.Join(keys, ",")+") VALUES("+strings.Join(params, ",")+")", vals...)
		if e != nil {
			return 0, e
		}
		id, e = res.LastInsertId()
		if e != nil {
			return 0, e
		}
	} else {
		sets := []string{}
		for _, key := range keys {
			sets = append(sets, key+"=?")
		}
		vals = append(vals, id, old["versao"])
		res, e := tx.ExecContext(ctx, "UPDATE "+k+" SET "+strings.Join(sets, ",")+",versao=versao+1 WHERE id=? AND versao=?", vals...)
		if e != nil {
			return 0, e
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return 0, fail("Conflito de edição; atualize a grade")
		}
	}
	for _, rel := range []struct{ child, parent, key, initial, active string }{{"estado", "pais", "pais_id", "Cadastrado", "Ativo"}, {"cidade", "estado", "estado_id", "Cadastrado", "Ativo"}, {"aeroporto", "cidade", "cidade_id", "Cadastrada", "Ativa"}} {
		if k == rel.child {
			if _, e := tx.ExecContext(ctx, "UPDATE "+rel.parent+" SET situacao=?,versao=versao+1 WHERE id=? AND situacao=?", rel.active, data[rel.key], rel.initial); e != nil {
				return 0, e
			}
		}
	}
	if k == "voo" && old["id"] == "" {
		start, _ := time.Parse(time.RFC3339, data["partida"])
		end, _ := time.Parse(time.RFC3339, data["chegada"])
		if _, e := tx.ExecContext(ctx, "INSERT INTO escala(voo_id,sequencial,origem_id,destino_id,tempo_escala,valor_centavos,situacao) VALUES(?,1,?,?,?,0,'Programada')", id, data["origem_id"], data["destino_id"], int(end.Sub(start).Minutes())); e != nil {
			return 0, e
		}
	}
	if k == "bilhete" {
		if _, e := tx.ExecContext(ctx, "UPDATE passageiro SET situacao='Ativo',versao=versao+1 WHERE id=?", data["passageiro_id"]); e != nil {
			return 0, e
		}
	}
	e := a.Q.WithTx(tx).RecordAudit(ctx, database.RecordAuditParams{UsuarioID: sql.NullInt64{Int64: u.ID, Valid: u.ID > 0}, Entidade: k, RegistroID: id, Acao: "Salvar", Motivo: ""})
	return id, e
}
func numInt(s string) int { return int(num(s)) }
func (a *App) transition(ctx context.Context, tx *sql.Tx, u User, k string, id int64, target, reason, gate, version string, confirmed bool) error {
	if _, ok := entities[k]; !ok {
		return fail("Cadastro inválido")
	}
	r, e := one(ctx, tx, k, id)
	if e != nil {
		return e
	}
	if e = a.canTouch(ctx, tx, u, k, r); e != nil {
		return e
	}
	if !a.canWrite(u, k) && !(u.Role == "passageiro" && k == "bilhete") {
		return fail("Você não tem permissão para essa ação")
	}
	if u.Role == "gerente" {
		if _, e = requireState(ctx, tx, "empresa", u.Company, "Ativa"); e != nil && k != "empresa" {
			return fail("Sua empresa não está operando")
		}
	}
	if version != r["versao"] {
		return fail("Registro alterado; atualize os detalhes")
	}
	allowed := false
	for _, s := range transitions[k][r["situacao"]] {
		if s == target {
			allowed = true
		}
	}
	if !allowed {
		return fail("Transição de estado não permitida")
	}
	if !confirmed {
		return fail("Confirme a operação")
	}
	critical := target == "Cancelado" || target == "Cancelada" || target == "Em Manutenção" || target == "Inativo" || target == "Inativa" || target == "Suspensa"
	if critical && strings.TrimSpace(reason) == "" {
		return fail("Informe a justificativa")
	}
	if len(reason) > 2000 || len(gate) > 30 {
		return fail("Texto muito longo")
	}
	switch k {
	case "voo":
		if r["situacao"] == "Em Voo" && target == "Cancelado" && len(strings.TrimSpace(reason)) < 10 {
			return fail("Cancelamento em voo exige justificativa com pelo menos 10 caracteres")
		}
		if _, e = tx.ExecContext(ctx, "UPDATE escala SET situacao='Cancelada',versao=versao+1 WHERE voo_id=? AND situacao NOT IN('Concluída','Cancelada')", id); e != nil {
			return e
		}
		if _, e = tx.ExecContext(ctx, "UPDATE bilhete SET situacao='Cancelado',motivo=?,cancelado_em=?,versao=versao+1 WHERE voo_id=? AND situacao<>'Usado'", reason, a.now(), id); e != nil {
			return e
		}
	case "escala":
		if r["voo_id"] != "" {
			if _, e = requireState(ctx, tx, "voo", num(r["voo_id"]), "Agendado"); e != nil {
				return fail("Escala só pode ser cancelada antes da partida")
			}
			if _, e = tx.ExecContext(ctx, "UPDATE bilhete SET situacao='Cancelado',motivo=?,cancelado_em=?,versao=versao+1 WHERE voo_id=? AND situacao IN('Emitido','Check-in Realizado')", "Escala cancelada: "+reason, a.now(), r["voo_id"]); e != nil {
				return e
			}
		}
	case "bilhete":
		f, e := requireState(ctx, tx, "voo", num(r["voo_id"]), "Agendado", "Em Voo")
		if e != nil {
			return e
		}
		if (target == "Check-in Realizado" || u.Role == "passageiro") && f["partida"] <= a.now() {
			return fail("Ação indisponível após a partida")
		}
		if u.Role == "passageiro" && target == "Embarque Realizado" {
			return fail("Somente o atendente pode confirmar o embarque")
		}
		if target == "Check-in Realizado" && strings.TrimSpace(gate) == "" {
			return fail("Portão de embarque é obrigatório")
		}
		if target == "Embarque Realizado" {
			t, _ := time.Parse(time.RFC3339, f["partida"])
			if time.Until(t) > 2*time.Hour {
				return fail("Embarque disponível a partir de 2 horas antes da partida")
			}
		}
		col := ""
		switch target {
		case "Check-in Realizado":
			col = "checkin_em"
		case "Embarque Realizado":
			col = "embarque_em"
		case "Cancelado":
			col = "cancelado_em"
		}
		if col != "" {
			if _, e = tx.ExecContext(ctx, "UPDATE bilhete SET "+col+"=?,portao_embarque=CASE WHEN ?<>'' THEN ? ELSE portao_embarque END,motivo=? WHERE id=?", a.now(), gate, gate, reason, id); e != nil {
				return e
			}
		}
	case "aeronave":
		if target == "Em Manutenção" || target == "Inativa" || target == "Excluída" {
			if count(ctx, tx, "SELECT count(*) FROM voo WHERE aeronave_id=? AND situacao IN('Agendado','Em Voo')", id) != 0 {
				return fail("Aeronave possui voos agendados ou em progresso")
			}
		}
	case "aeroporto":
		if target == "Em Manutenção" || target == "Inativo" || target == "Excluído" {
			if count(ctx, tx, "SELECT count(*) FROM escala WHERE (origem_id=? OR destino_id=?) AND situacao IN('Programada','Em Andamento')", id, id) != 0 || count(ctx, tx, "SELECT count(*) FROM voo WHERE (origem_id=? OR destino_id=?) AND situacao IN('Agendado','Em Voo')", id, id) != 0 {
				return fail("Aeroporto possui voos ou escalas ativas")
			}
		}
	case "passageiro":
		if target == "Inativo" && count(ctx, tx, "SELECT count(*) FROM bilhete WHERE passageiro_id=? AND situacao IN('Emitido','Check-in Realizado','Embarque Realizado')", id) != 0 {
			return fail("Passageiro ainda tem bilhetes ativos")
		}
		if target == "Excluído" && count(ctx, tx, "SELECT count(*) FROM bilhete WHERE passageiro_id=?", id) != 0 {
			return fail("Passageiro tem histórico de bilhetes")
		}
	case "pais":
		if target == "Inativo" && count(ctx, tx, "SELECT count(*) FROM estado WHERE pais_id=? AND situacao='Ativo'", id) != 0 {
			return fail("País possui estados ativos")
		}
	case "estado":
		if target == "Inativo" && count(ctx, tx, "SELECT count(*) FROM cidade WHERE estado_id=? AND situacao IN('Ativa','Cadastrada')", id) != 0 {
			return fail("Estado possui cidades ativas")
		}
	case "cidade":
		if target == "Inativa" && count(ctx, tx, "SELECT count(*) FROM aeroporto WHERE cidade_id=? AND situacao NOT IN('Inativo','Excluído')", id) != 0 {
			return fail("Cidade possui aeroportos ativos")
		}
	}
	if strings.HasPrefix(target, "Exclu") {
		rels := map[string][]struct{ table, key string }{"pais": {{"estado", "pais_id"}}, "estado": {{"cidade", "estado_id"}}, "cidade": {{"aeroporto", "cidade_id"}}, "empresa": {{"aeronave", "empresa_id"}, {"voo", "empresa_id"}}}
		for _, rel := range rels[k] {
			if count(ctx, tx, "SELECT count(*) FROM "+rel.table+" WHERE "+rel.key+"=?", id) != 0 {
				return fail("Registro possui referências; preserve o histórico")
			}
		}
	}
	if target == "Em Manutenção" {
		if _, e = tx.ExecContext(ctx, "INSERT INTO manutencao(entidade,registro_id,motivo,inicio) VALUES(?,?,?,?)", k, id, reason, a.now()); e != nil {
			return e
		}
	}
	if r["situacao"] == "Em Manutenção" {
		if _, e = tx.ExecContext(ctx, "UPDATE manutencao SET fim=? WHERE entidade=? AND registro_id=? AND fim IS NULL", a.now(), k, id); e != nil {
			return e
		}
	}
	res, e := tx.ExecContext(ctx, "UPDATE "+k+" SET situacao=?,versao=versao+1 WHERE id=? AND versao=?", target, id, version)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fail("Conflito de edição")
	}
	if k == "bilhete" || k == "voo" || k == "escala" {
		if _, e = tx.ExecContext(ctx, "UPDATE passageiro SET situacao='Inativo',versao=versao+1 WHERE situacao='Ativo' AND EXISTS(SELECT 1 FROM bilhete WHERE passageiro_id=passageiro.id AND situacao='Cancelado') AND NOT EXISTS(SELECT 1 FROM bilhete WHERE passageiro_id=passageiro.id AND situacao IN('Emitido','Check-in Realizado','Embarque Realizado'))"); e != nil {
			return e
		}
	}
	return a.Q.WithTx(tx).RecordAudit(ctx, database.RecordAuditParams{UsuarioID: sql.NullInt64{Int64: u.ID, Valid: true}, Entidade: k, RegistroID: id, Acao: fmt.Sprintf("%s → %s", r["situacao"], target), Motivo: reason})
}
