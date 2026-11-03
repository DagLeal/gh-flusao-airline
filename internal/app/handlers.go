package app

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"erlaine/aeroporto/internal/ui"
	"github.com/go-chi/chi/v5"
)

type Option struct{ ID, Label string }
type Grid struct {
	Entity                                                                 Entity
	Rows                                                                   []Row
	Fields                                                                 []Field
	Options                                                                map[string][]Option
	Labels                                                                 map[string]map[string]string
	CanWrite                                                               bool
	Page, Total                                                            int
	Q, Status, Sort, Dir, Parent, ParentID, Date, Origin, Destination, URL string
}
type Input struct {
	Field    Field
	Value    string
	Options  []Option
	Readonly bool
}
type Form struct {
	Inline Row
	Seats  []Seat

	Entity                                     Entity
	ID, Version, CSRF, Parent, ParentID, Error string
	Inputs                                     []Input
	IsNew                                      bool
}
type Detail struct {
	Entity             Entity
	Row                Row
	Fields             []Input
	Actions            []string
	CSRF               string
	Children           []Grid
	Audit, Maintenance []Row
	Seats              []Seat
	Print              bool
}
type Seat struct {
	Number   int
	Occupied bool
}

func (a *App) options(ctx context.Context, u User, en Entity) (map[string][]Option, map[string]map[string]string, error) {
	opts := map[string][]Option{}
	labels := map[string]map[string]string{}
	for _, f := range en.Fields {
		if f.Ref == "" {
			continue
		}
		key := f.Ref
		clause := ""
		args := []interface{}{}
		if u.Role == "gerente" {
			switch key {
			case "empresa":
				clause = " AND id=?"
				args = append(args, u.Company)
			case "voo", "aeronave":
				clause = " AND empresa_id=?"
				args = append(args, u.Company)
			}
		}
		rs, e := rows(ctx, a.DB, "SELECT * FROM "+key+" WHERE 1=1"+clause+" ORDER BY id DESC LIMIT 2000", args...)
		if e != nil {
			return nil, nil, e
		}
		labels[f.Key] = map[string]string{}
		for _, r := range rs {
			l := r["nome"]
			switch key {
			case "voo":
				l = r["codigo_voo"] + " · " + a.localTime(r["partida"])
			case "aeronave":
				l = r["codigo_aeronave"] + " · " + r["modelo"]
			case "aeroporto":
				l = r["sigla"] + " · " + r["nome"]
			case "passageiro":
				l = r["nome"] + " · " + r["documento"]
			}
			labels[f.Key][r["id"]] = l
			if !strings.HasPrefix(r["situacao"], "Exclu") {
				opts[f.Key] = append(opts[f.Key], Option{r["id"], l + " [" + r["situacao"] + "]"})
			}
		}
	}
	return opts, labels, nil
}
func (a *App) localTime(s string) string {
	t, e := time.Parse(time.RFC3339, s)
	if e != nil {
		return s
	}
	return t.In(a.Location).Format("02/01/2006 15:04")
}
func (a *App) inputTime(s string) string {
	t, e := time.Parse(time.RFC3339, s)
	if e != nil {
		return s
	}
	return t.In(a.Location).Format("2006-01-02T15:04")
}
func parentField(parent, k string) string {
	pairs := map[string]map[string]string{"voo": {"escala": "voo_id", "bilhete": "voo_id"}, "passageiro": {"bilhete": "passageiro_id"}, "empresa": {"aeronave": "empresa_id", "voo": "empresa_id"}, "aeronave": {"voo": "aeronave_id"}, "pais": {"estado": "pais_id"}, "estado": {"cidade": "estado_id"}, "cidade": {"aeroporto": "cidade_id"}}
	return pairs[parent][k]
}
func (a *App) makeGrid(r *http.Request, k string) (Grid, error) {
	en, ok := entities[k]
	u := current(r)
	if !ok || !a.canRead(u, k) {
		return Grid{}, fail("Acesso negado")
	}
	v := r.URL.Query()
	g := Grid{Entity: en, Fields: en.Fields, CanWrite: a.canWrite(u, k), Page: 1, Q: v.Get("q"), Status: v.Get("status"), Sort: v.Get("sort"), Dir: v.Get("dir"), Parent: v.Get("parent"), ParentID: v.Get("parent_id"), Date: v.Get("date"), Origin: v.Get("origin"), Destination: v.Get("destination")}
	if p := num(v.Get("page")); p > 0 && p < 100000 {
		g.Page = int(p)
	}
	if k == "bilhete" {
		g.Fields = append([]Field{field("codigo_bilhete", "Localizador", "text")}, en.Fields...)
	}
	where, args := a.scope(u, k)
	where = " WHERE 1=1" + where
	if g.Parent != "" {
		f := parentField(g.Parent, k)
		if f == "" {
			return g, fail("Relacionamento inválido")
		}
		if _, e := a.accessible(r.Context(), u, g.Parent, num(g.ParentID)); e != nil {
			return g, e
		}
		where += " AND " + f + "=?"
		args = append(args, g.ParentID)
	}
	if g.Q != "" {
		clauses := []string{"CAST(id AS TEXT) LIKE ?"}
		args = append(args, "%"+g.Q+"%")
		for _, f := range en.Fields {
			if f.Type == "text" || f.Type == "tel" {
				clauses = append(clauses, f.Key+" LIKE ?")
				args = append(args, "%"+g.Q+"%")
			}
		}
		if k == "bilhete" {
			clauses = append(clauses, "codigo_bilhete LIKE ?")
			args = append(args, "%"+g.Q+"%")
		}
		where += " AND (" + strings.Join(clauses, " OR ") + ")"
	}
	if g.Status == "" {
		where += " AND situacao NOT IN('Excluído','Excluída')"
	}
	if g.Status != "" {
		where += " AND situacao=?"
		args = append(args, g.Status)
	}
	if k == "voo" {
		if g.Date != "" {
			t, e := time.ParseInLocation("2006-01-02", g.Date, a.Location)
			if e != nil {
				return g, fail("Data inválida")
			}
			where += " AND partida>=? AND partida<?"
			args = append(args, t.UTC().Format(time.RFC3339), t.AddDate(0, 0, 1).UTC().Format(time.RFC3339))
		}
		for _, x := range []struct{ k, v string }{{"origem_id", g.Origin}, {"destino_id", g.Destination}} {
			if x.v != "" {
				where += " AND " + x.k + "=?"
				args = append(args, x.v)
			}
		}
		if g.Origin != "" && g.Origin == g.Destination {
			return g, fail("Origem e destino devem ser diferentes")
		}
	}
	sort := "id"
	for _, f := range en.Fields {
		if f.Key == g.Sort {
			sort = f.Key
		}
	}
	if g.Sort == "situacao" {
		sort = "situacao"
	}
	g.Sort = sort
	dir := "DESC"
	if g.Dir == "asc" {
		dir = "ASC"
	}
	g.Dir = strings.ToLower(dir)
	if e := a.DB.QueryRowContext(r.Context(), "SELECT count(*) FROM "+k+where, args...).Scan(&g.Total); e != nil {
		return g, e
	}
	rs, e := rows(r.Context(), a.DB, "SELECT * FROM "+k+where+" ORDER BY "+sort+" "+dir+" LIMIT 20 OFFSET ?", append(args, (g.Page-1)*20)...)
	if e != nil {
		return g, e
	}
	g.Rows = rs
	for _, row := range g.Rows {
		row["_editable"] = ""
		if g.CanWrite && editable(k, row) {
			row["_editable"] = "yes"
		}
		for _, f := range en.Fields {
			if f.Type == "datetime-local" {
				row[f.Key] = a.inputTime(row[f.Key])
			}
		}
	}
	g.Options, g.Labels, e = a.options(r.Context(), u, en)
	vals := url.Values{}
	for key, list := range v {
		if key != "page" && key != "entity" {
			vals[key] = list
		}
	}
	g.URL = "/grid/" + k + "?" + vals.Encode()
	return g, e
}
func (a *App) home(w http.ResponseWriter, r *http.Request) {
	if e := a.Sync(r.Context()); e != nil {
		http.Error(w, "Falha ao atualizar operações", 500)
		return
	}
	k := r.URL.Query().Get("entity")
	if k == "" {
		k = "voo"
	}
	g, e := a.makeGrid(r, k)
	if e != nil {
		http.Error(w, friendly(e), 403)
		return
	}
	u := current(r)
	stats := []struct{ Label, Value, Note string }{}
	for _, s := range []struct{ Label, Entity, Where string }{{"Voos programados", "voo", "situacao='Agendado'"}, {"Em operação", "voo", "situacao='Em Voo'"}, {"Bilhetes emitidos", "bilhete", "situacao IN('Emitido','Check-in Realizado','Embarque Realizado')"}, {"Escalas concluídas", "escala", "situacao='Concluída'"}} {
		scope, args := a.scope(u, s.Entity)
		n := count(r.Context(), a.DB, "SELECT count(*) FROM "+s.Entity+" WHERE "+s.Where+scope, args...)
		stats = append(stats, struct{ Label, Value, Note string }{s.Label, strconv.Itoa(n), "Visão da sua operação"})
	}
	data := map[string]interface{}{"Grid": g, "Stats": stats, "User": u, "Date": time.Now().In(a.Location).Format("02 Jan 2006"), "Demo": osDemo()}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if e = ui.Layout(u.Name, u.Role, csrfValue(r), a.component("home", data)).Render(r.Context(), w); e != nil {
		http.Error(w, "Erro ao renderizar", 500)
	}
}
func osDemo() bool { return env("DEMO_MODE", "false") == "true" }
func (a *App) gridHandler(w http.ResponseWriter, r *http.Request) {
	if e := a.Sync(r.Context()); e != nil {
		http.Error(w, "Falha de atualização", 500)
		return
	}
	g, e := a.makeGrid(r, chi.URLParam(r, "entity"))
	if e != nil {
		http.Error(w, friendly(e), 400)
		return
	}
	a.render(w, "grid", g)
}
func (a *App) formData(r *http.Request, k string, v Row) (Form, error) {
	en, ok := entities[k]
	u := current(r)
	if !ok || !a.canWrite(u, k) {
		return Form{}, fail("Acesso negado")
	}
	f := Form{Entity: en, ID: r.URL.Query().Get("id"), CSRF: csrfValue(r), Parent: r.URL.Query().Get("parent"), ParentID: r.URL.Query().Get("parent_id"), IsNew: true}
	if v == nil {
		v = Row{}
		if f.ID != "" {
			var e error
			v, e = a.accessible(r.Context(), u, k, num(f.ID))
			if e != nil {
				return f, e
			}
			if !editable(k, v) {
				return f, fail("Registro permite apenas leitura")
			}
			f.IsNew = false
		}
	} else {
		f.ID = v["id"]
		f.IsNew = f.ID == ""
	}
	f.Inline = v
	f.Version = v["versao"]
	opts, _, e := a.options(r.Context(), u, en)
	if e != nil {
		return f, e
	}
	parentKey := parentField(f.Parent, k)
	if f.Parent != "" {
		if parentKey == "" {
			return f, fail("Relacionamento inválido")
		}
		if _, e = a.accessible(r.Context(), u, f.Parent, num(f.ParentID)); e != nil {
			return f, e
		}
		if f.IsNew {
			v[parentKey] = f.ParentID
		}
	}
	for _, field := range en.Fields {
		value := v[field.Key]
		if field.Type == "datetime-local" {
			value = a.inputTime(value)
		}
		if value == "" && field.Type == "number" {
			if field.Key == "milhagem" || field.Key == "valor_centavos" {
				value = "0"
			} else {
				value = "1"
			}
		}
		if u.Role == "gerente" && field.Key == "empresa_id" {
			value = strconv.FormatInt(u.Company, 10)
		}
		readonly := !f.IsNew && ((k == "passageiro" && field.Key == "documento") || (k == "aeronave" && field.Key == "codigo_aeronave"))
		f.Inputs = append(f.Inputs, Input{field, value, opts[field.Key], readonly})
	}
	if k == "bilhete" && v["voo_id"] != "" {
		f.Seats, _ = a.seats(r.Context(), u, num(v["voo_id"]))
	}
	return f, nil
}
func (a *App) formHandler(w http.ResponseWriter, r *http.Request) {
	f, e := a.formData(r, chi.URLParam(r, "entity"), nil)
	if e != nil {
		http.Error(w, friendly(e), 403)
		return
	}
	a.render(w, "form", f)
}
func posted(r *http.Request) Row {
	v := Row{}
	for k := range r.PostForm {
		v[k] = r.PostForm.Get(k)
	}
	return v
}
func (a *App) saveHandler(w http.ResponseWriter, r *http.Request) {
	k := chi.URLParam(r, "entity")
	if e := a.Sync(r.Context()); e != nil {
		http.Error(w, "Falha de atualização", 500)
		return
	}
	v := posted(r)
	tx, e := a.DB.BeginTx(r.Context(), nil)
	if e != nil {
		http.Error(w, "Banco indisponível", 500)
		return
	}
	defer tx.Rollback()
	_, e = a.save(r.Context(), tx, current(r), k, num(v["id"]), v)
	if e == nil {
		e = tx.Commit()
	}
	if e != nil {
		tx.Rollback()
		r.URL.RawQuery = url.Values{"id": {v["id"]}, "parent": {v["parent"]}, "parent_id": {v["parent_id"]}}.Encode()
		f, err := a.formData(r, k, v)
		if err != nil {
			http.Error(w, friendly(err), 403)
			return
		}
		f.Error = friendly(e)
		a.render(w, "form", f)
		return
	}
	w.Header().Set("HX-Trigger", "saved")
	ioHTML(w, "<div data-close-modal></div>")
}
func ioHTML(w http.ResponseWriter, s string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, s)
}
func (a *App) batchHandler(w http.ResponseWriter, r *http.Request) {
	k := chi.URLParam(r, "entity")
	if !a.canWrite(current(r), k) {
		http.Error(w, "Acesso negado", 403)
		return
	}
	if e := a.Sync(r.Context()); e != nil {
		http.Error(w, "Falha de atualização", 500)
		return
	}
	ids := r.PostForm["changed"]
	if len(ids) == 0 {
		w.Header().Set("HX-Trigger", "unchanged")
		w.WriteHeader(204)
		return
	}
	if len(ids) > 20 {
		http.Error(w, "Limite de 20 linhas por salvamento", 400)
		return
	}
	tx, e := a.DB.BeginTx(r.Context(), nil)
	if e != nil {
		http.Error(w, "Banco indisponível", 500)
		return
	}
	defer tx.Rollback()
	for _, id := range ids {
		v := Row{}
		prefix := "cell_" + id + "_"
		for key := range r.PostForm {
			if strings.HasPrefix(key, prefix) {
				v[strings.TrimPrefix(key, prefix)] = r.PostForm.Get(key)
			}
		}
		if _, e = a.save(r.Context(), tx, current(r), k, num(id), v); e != nil {
			tx.Rollback()
			http.Error(w, "Linha "+id+": "+friendly(e), 422)
			return
		}
	}
	if e = tx.Commit(); e != nil {
		http.Error(w, "Falha ao salvar grade", 500)
		return
	}
	w.Header().Set("HX-Trigger", "gridSaved")
	w.WriteHeader(204)
}
func (a *App) detailData(r *http.Request, k string, id int64) (Detail, error) {
	u := current(r)
	row, e := a.accessible(r.Context(), u, k, id)
	if e != nil {
		return Detail{}, e
	}
	en := entities[k]
	d := Detail{Entity: en, Row: row, CSRF: csrfValue(r)}
	_, labels, e := a.options(r.Context(), u, en)
	if e != nil {
		return d, e
	}
	for _, f := range en.Fields {
		value := row[f.Key]
		if f.Ref != "" {
			value = labels[f.Key][value]
		}
		if f.Type == "datetime-local" {
			value = a.localTime(value)
		}
		d.Fields = append(d.Fields, Input{Field: f, Value: value})
	}
	if k == "bilhete" {
		for _, key := range []string{"origem_id", "destino_id"} {
			airport, err := one(r.Context(), a.DB, "aeroporto", num(row[key]))
			if err != nil {
				return d, err
			}
			label := "Origem"
			if key == "destino_id" {
				label = "Destino"
			}
			d.Fields = append(d.Fields, Input{Field: field(key, label, "text"), Value: airport["sigla"] + " · " + airport["nome"]})
		}
		flight, err := one(r.Context(), a.DB, "voo", num(row["voo_id"]))
		if err != nil {
			return d, err
		}
		d.Fields = append(d.Fields, Input{Field: field("partida", "Partida do voo", "text"), Value: a.localTime(flight["partida"])}, Input{Field: field("chegada", "Chegada do voo", "text"), Value: a.localTime(flight["chegada"])})
	}
	if a.canWrite(u, k) || (u.Role == "passageiro" && k == "bilhete") {
		d.Actions = transitions[k][row["situacao"]]
	}
	if u.Role == "passageiro" {
		filtered := []string{}
		for _, act := range d.Actions {
			if act != "Embarque Realizado" {
				filtered = append(filtered, act)
			}
		}
		d.Actions = filtered
	}
	children := map[string][]string{"voo": {"escala", "bilhete"}, "passageiro": {"bilhete"}, "empresa": {"aeronave", "voo"}, "aeronave": {"voo"}, "pais": {"estado"}, "estado": {"cidade"}, "cidade": {"aeroporto"}}
	for _, child := range children[k] {
		if !a.canRead(u, child) {
			continue
		}
		copyR := r.Clone(r.Context())
		copyURL := *r.URL
		copyR.URL = &copyURL
		copyR.URL.RawQuery = url.Values{"parent": {k}, "parent_id": {strconv.FormatInt(id, 10)}}.Encode()
		g, e := a.makeGrid(copyR, child)
		if e != nil {
			return d, e
		}
		d.Children = append(d.Children, g)
	}
	if u.Role != "passageiro" {
		d.Audit, e = rows(r.Context(), a.DB, "SELECT acao,motivo,criado_em FROM auditoria WHERE entidade=? AND registro_id=? ORDER BY id DESC LIMIT 15", k, id)
		if e != nil {
			return d, e
		}
		d.Maintenance, e = rows(r.Context(), a.DB, "SELECT motivo,inicio,coalesce(fim,'Em andamento') AS fim FROM manutencao WHERE entidade=? AND registro_id=? ORDER BY id DESC LIMIT 15", k, id)
		if e != nil {
			return d, e
		}
	}
	if k == "voo" {
		craft, e := one(r.Context(), a.DB, "aeronave", num(row["aeronave_id"]))
		if e != nil {
			return d, e
		}
		occupied, e := rows(r.Context(), a.DB, "SELECT assento FROM bilhete WHERE voo_id=? AND situacao<>'Cancelado'", id)
		if e != nil {
			return d, e
		}
		taken := map[int64]bool{}
		for _, v := range occupied {
			taken[num(v["assento"])] = true
		}
		for i := 1; i <= numInt(craft["capacidade"]); i++ {
			d.Seats = append(d.Seats, Seat{i, taken[int64(i)]})
		}
	}
	return d, nil
}
func (a *App) detailHandler(w http.ResponseWriter, r *http.Request) {
	if e := a.Sync(r.Context()); e != nil {
		http.Error(w, "Falha de atualização", 500)
		return
	}
	d, e := a.detailData(r, chi.URLParam(r, "entity"), num(chi.URLParam(r, "id")))
	if e != nil {
		http.Error(w, friendly(e), 403)
		return
	}
	a.render(w, "detail", d)
}
func (a *App) transitionHandler(w http.ResponseWriter, r *http.Request) {
	if e := a.Sync(r.Context()); e != nil {
		http.Error(w, "Falha de atualização", 500)
		return
	}
	k := chi.URLParam(r, "entity")
	id := num(chi.URLParam(r, "id"))
	tx, e := a.DB.BeginTx(r.Context(), nil)
	if e != nil {
		http.Error(w, "Banco indisponível", 500)
		return
	}
	defer tx.Rollback()
	e = a.transition(r.Context(), tx, current(r), k, id, r.FormValue("target"), r.FormValue("reason"), r.FormValue("gate"), r.FormValue("versao"), r.FormValue("confirmed") == "yes")
	if e == nil {
		e = tx.Commit()
	}
	if e != nil {
		http.Error(w, friendly(e), 422)
		return
	}
	w.Header().Set("HX-Trigger", "stateChanged")
	d, e := a.detailData(r, k, id)
	if e != nil {
		http.Error(w, friendly(e), 400)
		return
	}
	a.render(w, "detail", d)
}
func (a *App) exportHandler(w http.ResponseWriter, r *http.Request) {
	k := chi.URLParam(r, "entity")
	g, e := a.makeGrid(r, k)
	if e != nil {
		http.Error(w, friendly(e), 403)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+k+"-pagina.csv")
	fmt.Fprint(w, "\xef\xbb\xbf")
	c := csv.NewWriter(w)
	defer c.Flush()
	header := []string{"ID"}
	for _, f := range g.Fields {
		header = append(header, f.Label)
	}
	header = append(header, "Situação")
	c.Write(header)
	for _, row := range g.Rows {
		v := []string{row["id"]}
		for _, f := range g.Fields {
			s := row[f.Key]
			if f.Ref != "" {
				s = g.Labels[f.Key][s]
			}
			v = append(v, csvSafe(s))
		}
		v = append(v, row["situacao"])
		c.Write(v)
	}
}
func csvSafe(s string) string {
	if len(s) > 0 && strings.ContainsAny(s[:1], "=+-@\t\r\n") {
		return "'" + s
	}
	return s
}
func (a *App) printHandler(w http.ResponseWriter, r *http.Request) {
	d, e := a.detailData(r, "bilhete", num(chi.URLParam(r, "id")))
	if e != nil {
		http.Error(w, "Acesso negado", 403)
		return
	}
	d.Print = true
	a.render(w, "print", d)
}
