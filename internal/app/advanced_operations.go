package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"erlaine/aeroporto/internal/database"
)

func nextAuditedID(ctx context.Context, tx *sql.Tx, k string) (int64, error) {
	if k != "escala" && k != "aeroporto" {
		return 0, fail("Entidade inválida")
	}
	var id int64
	err := tx.QueryRowContext(ctx, "SELECT max(n)+1 FROM (SELECT coalesce(max(id),0) AS n FROM "+k+" UNION ALL SELECT coalesce(max(registro_id),0) FROM auditoria WHERE entidade=?)", k).Scan(&id)
	return id, err
}
func (a *App) advancedAudit(ctx context.Context, tx *sql.Tx, u User, k string, id int64, action, reason string, snapshot Row) error {
	if snapshot != nil {
		data, e := json.Marshal(snapshot)
		if e != nil {
			return e
		}
		reason += " | Registro anterior: " + string(data)
	}
	return a.Q.WithTx(tx).RecordAudit(ctx, database.RecordAuditParams{UsuarioID: sql.NullInt64{Int64: u.ID, Valid: true}, Entidade: k, RegistroID: id, Acao: action, Motivo: reason})
}
func (a *App) mutateAdvanced(ctx context.Context, tx *sql.Tx, u User, id int, values Row) (int64, error) {
	if u.Role != "gerente" || u.Company < 1 {
		return 0, fail("Área restrita ao gerente")
	}
	if _, err := requireState(ctx, tx, "empresa", u.Company, "Ativa"); err != nil {
		return 0, fail("Sua empresa não está operando")
	}
	if id < 35 || id > 40 {
		return 0, fail("Esta consulta não permite alterações")
	}
	if id >= 37 && values["confirmed"] != "yes" {
		return 0, fail("Confirme a operação")
	}
	reason := strings.TrimSpace(values["reason"])
	if len(reason) > 2000 {
		return 0, fail("Justificativa muito longa")
	}
	if id >= 39 && len(reason) < 3 {
		return 0, fail("Informe a justificativa para a exclusão")
	}
	var recordID int64
	entity := "voo"
	snapshot := Row(nil)
	switch id {
	case 35:
		entity = "aeroporto"
		result, err := a.save(ctx, tx, u, entity, 0, Row{"nome": values["nome"], "sigla": values["sigla"], "cidade_id": values["cidade_id"]})
		if err != nil {
			return 0, err
		}
		recordID = result
	case 36:
		price, err := moneyCents(values["valor"])
		if err != nil {
			return 0, err
		}
		result, err := a.save(ctx, tx, u, "voo", 0, Row{"codigo_voo": values["codigo"], "empresa_id": strconv.FormatInt(u.Company, 10), "aeronave_id": values["aeronave_id"], "origem_id": values["origem_id"], "destino_id": values["destino_id"], "partida": values["partida"], "chegada": values["chegada"], "valor_centavos": strconv.FormatInt(price, 10)})
		if err != nil {
			return 0, err
		}
		recordID = result
	case 37, 38:
		recordID = num(values["registro_id"])
		row, err := one(ctx, tx, "voo", recordID)
		if err != nil {
			return 0, fail("Voo não encontrado")
		}
		if err = a.canTouch(ctx, tx, u, "voo", row); err != nil {
			return 0, err
		}
		if row["versao"] != values["versao"] {
			return 0, fail("O voo foi alterado. Atualize a prévia antes de executar")
		}
		if row["situacao"] != "Agendado" || row["partida"] <= a.now() {
			return 0, fail("A operação exige um voo agendado com partida futura")
		}
		snapshot = row
		if id == 37 {
			price, err := moneyCents(values["valor"])
			if err != nil {
				return 0, err
			}
			data := Row{}
			for k, v := range row {
				data[k] = v
			}
			data["partida"] = a.inputTime(row["partida"])
			data["chegada"] = a.inputTime(row["chegada"])
			data["valor_centavos"] = strconv.FormatInt(price, 10)
			if _, err = a.save(ctx, tx, u, "voo", recordID, data); err != nil {
				return 0, err
			}
		} else {
			if row["confirmado"] == "1" {
				return 0, fail("O voo já está confirmado")
			}
			res, err := tx.ExecContext(ctx, "UPDATE voo SET confirmado=1,versao=versao+1 WHERE id=? AND empresa_id=? AND versao=?", recordID, u.Company, values["versao"])
			if err != nil {
				return 0, err
			}
			n, _ := res.RowsAffected()
			if n != 1 {
				return 0, fail("Conflito de edição")
			}
		}
	case 39:
		entity = "escala"
		recordID = num(values["registro_id"])
		row, err := one(ctx, tx, "escala", recordID)
		if err != nil {
			return 0, fail("Escala não encontrada")
		}
		if err = a.canTouch(ctx, tx, u, entity, row); err != nil {
			return 0, err
		}
		if row["versao"] != values["versao"] {
			return 0, fail("A escala foi alterada. Atualize a prévia")
		}
		if row["sequencial"] != "1" || row["voo_id"] == "" {
			return 0, fail("Selecione a escala de sequencial 1 de um voo")
		}
		flight, err := requireState(ctx, tx, "voo", num(row["voo_id"]), "Agendado")
		if err != nil {
			return 0, err
		}
		if flight["partida"] <= a.now() {
			return 0, fail("Não é possível excluir uma escala após a partida")
		}
		if row["situacao"] != "Programada" && row["situacao"] != "Planejada" {
			return 0, fail("Estado da escala não permite exclusão")
		}
		if count(ctx, tx, "SELECT count(*) FROM bilhete WHERE voo_id=? AND situacao<>'Cancelado'", row["voo_id"]) != 0 {
			return 0, fail("O voo possui bilhetes. Preserve a rota e o histórico")
		}
		n := count(ctx, tx, "SELECT count(*) FROM escala WHERE voo_id=? AND id<>? AND situacao NOT IN('Cancelada','Concluída')", row["voo_id"], recordID)
		if n < 1 {
			return 0, fail("Não é possível excluir o único trecho ativo do voo")
		}
		res, err := tx.ExecContext(ctx, "DELETE FROM escala WHERE id=? AND versao=?", recordID, values["versao"])
		if err != nil {
			return 0, err
		}
		n64, _ := res.RowsAffected()
		if n64 != 1 {
			return 0, fail("Conflito de edição")
		}
		// Removing the first segment also changes the operational route.
		remaining, err := rows(ctx, tx, "SELECT origem_id,destino_id FROM escala WHERE voo_id=? AND situacao NOT IN('Cancelada','Concluída') ORDER BY sequencial", row["voo_id"])
		if err != nil {
			return 0, err
		}
		if len(remaining) == 0 || remaining[0]["origem_id"] == remaining[len(remaining)-1]["destino_id"] {
			return 0, fail("A rota restante precisa ter origem e destino diferentes")
		}
		if _, err = tx.ExecContext(ctx, "UPDATE voo SET origem_id=?,destino_id=?,versao=versao+1 WHERE id=?", remaining[0]["origem_id"], remaining[len(remaining)-1]["destino_id"], row["voo_id"]); err != nil {
			return 0, err
		}
		if err = a.advancedAudit(ctx, tx, u, "voo", num(row["voo_id"]), "Rota atualizada após exclusão de escala", reason, flight); err != nil {
			return 0, err
		}
		snapshot = row
	case 40:
		entity = "aeroporto"
		parts := strings.Split(values["airport"], ":")
		if len(parts) != 2 {
			return 0, fail("Selecione um aeroporto da prévia")
		}
		recordID = num(parts[0])
		row, err := one(ctx, tx, entity, recordID)
		if err != nil {
			return 0, fail("Aeroporto não encontrado")
		}
		if row["versao"] != parts[1] {
			return 0, fail("O aeroporto foi alterado. Atualize a prévia")
		}
		for _, table := range []string{"escala", "voo", "bilhete"} {
			if count(ctx, tx, "SELECT count(*) FROM "+table+" WHERE origem_id=? OR destino_id=?", recordID, recordID) != 0 {
				return 0, fail("O aeroporto possui referências em " + table + " e não pode ser excluído")
			}
		}
		if count(ctx, tx, "SELECT count(*) FROM manutencao WHERE entidade='aeroporto' AND registro_id=?", recordID) != 0 {
			return 0, fail("O aeroporto possui histórico de manutenção")
		}
		res, err := tx.ExecContext(ctx, "DELETE FROM aeroporto WHERE id=? AND versao=?", recordID, parts[1])
		if err != nil {
			return 0, err
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return 0, fail("Conflito de edição")
		}
		snapshot = row
	}
	err := a.advancedAudit(ctx, tx, u, entity, recordID, fmt.Sprintf("Consulta avançada %02d", id), reason, snapshot)
	return recordID, err
}
func (a *App) advancedOperation(w http.ResponseWriter, r *http.Request) {
	if !advancedAllowed(w, r) {
		return
	}
	id, e := strconv.Atoi(r.PostForm.Get("id"))
	if e != nil {
		http.Error(w, "Operação inválida", 400)
		return
	}
	if e = a.Sync(r.Context()); e != nil {
		http.Error(w, "Não foi possível atualizar os estados", 500)
		return
	}
	tx, e := a.DB.BeginTx(r.Context(), nil)
	if e != nil {
		http.Error(w, "Banco indisponível", 500)
		return
	}
	defer tx.Rollback()
	_, e = a.mutateAdvanced(r.Context(), tx, current(r), id, posted(r))
	if e == nil {
		e = tx.Commit()
	}
	if e != nil {
		http.Error(w, friendly(e), 422)
		return
	}
	params := url.Values{"id": {strconv.Itoa(id)}}
	if code := r.PostForm.Get("codigo"); code != "" {
		params.Set("codigo", code)
	}
	if sigla := r.PostForm.Get("sigla"); sigla != "" {
		params.Set("sigla", sigla)
	}
	r.URL.RawQuery = params.Encode()
	if r.Header.Get("HX-Request") == "" {
		http.Redirect(w, r, "/consultas-avancadas?"+params.Encode(), 303)
		return
	}
	d, e := a.advancedData(r)
	if e != nil {
		http.Error(w, friendly(e), 400)
		return
	}
	d.Success = "Operação concluída. Os dados e a auditoria foram atualizados."
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	a.render(w, "advanced", d)
}
