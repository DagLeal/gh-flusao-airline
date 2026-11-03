package app

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"erlaine/aeroporto/internal/ui"
)

//go:embed advanced_catalog.json
var advancedCatalogJSON []byte

type AdvancedQuery struct {
	ID                            int
	Title, Group, Kind, SQL, Note string
	Params                        []string
	Defaults                      map[string]string
}

var advancedCatalog = func() []AdvancedQuery {
	var c []AdvancedQuery
	if err := json.Unmarshal(advancedCatalogJSON, &c); err != nil {
		panic(err)
	}
	return c
}()

const advancedCTE = `WITH voos AS (
 SELECT v.id AS id_voo,v.codigo_voo,v.empresa_id,emp.nome AS empresa,
 v.aeronave_id,v.origem_id,v.destino_id,v.partida,v.chegada,
 CASE WHEN v.situacao='Agendado' THEN CASE WHEN v.confirmado=1 THEN 'CONFIRMADO' ELSE 'PLANEJADO' END ELSE v.situacao END AS situacao,
 v.situacao AS estado_operacional,v.valor_centavos,v.valor_centavos/100.0 AS valor,v.confirmado,v.versao
 FROM voo v JOIN empresa emp ON emp.id=v.empresa_id WHERE v.empresa_id=:empresa_id
), escalas AS (
 SELECT e.id AS id_escala,e.voo_id AS id_voo,e.sequencial,e.origem_id AS id_aeroporto_origem,
 e.destino_id AS id_aeroporto_destino,e.tempo_escala,e.valor_centavos,e.situacao,e.versao
 FROM escala e WHERE e.voo_id IS NULL OR e.voo_id IN(SELECT id_voo FROM voos)
), aeroportos AS (
 SELECT a.id AS id_aeroporto,a.nome,a.sigla,c.nome AS cidade,uf.sigla AS uf,
 a.cidade_id,a.situacao,a.versao FROM aeroporto a
 JOIN cidade c ON c.id=a.cidade_id JOIN estado uf ON uf.id=c.estado_id
)
`

type AdvancedTable struct {
	Columns []string
	Rows    [][]string
	Raw     []Row
	HasNext bool
}
type AdvancedData struct {
	Catalog                          []AdvancedQuery
	Query                            AdvancedQuery
	Table                            AdvancedTable
	Filters, Fields                  []Input
	Values                           Row
	Candidates                       []Option
	SQL, CSRF, Company, URL, Success string
	Page                             int
	PreviewID, Version               string
	CanMutate                        bool
}

func advancedQuery(id int) (AdvancedQuery, bool) {
	for _, q := range advancedCatalog {
		if q.ID == id {
			return q, true
		}
	}
	return AdvancedQuery{}, false
}
func advancedID(v url.Values) (int, error) {
	s := v.Get("id")
	if s == "" {
		return 1, nil
	}
	n, e := strconv.Atoi(s)
	if e != nil {
		return 0, fail("Número de consulta inválido")
	}
	if _, ok := advancedQuery(n); !ok {
		return 0, fail("Consulta não encontrada")
	}
	return n, nil
}
func advancedAllowed(w http.ResponseWriter, r *http.Request) bool {
	u := current(r)
	if u.Role != "gerente" || u.Company < 1 {
		http.Error(w, "Área restrita ao gerente.", http.StatusForbidden)
		return false
	}
	return true
}
func moneyCents(s string) (int64, error) {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", ".")
	if !regexp.MustCompile(`^[0-9]{1,8}(\.[0-9]{1,2})?$`).MatchString(s) {
		return 0, fail("Informe um valor válido em reais, com até duas casas decimais")
	}
	parts := strings.Split(s, ".")
	n, _ := strconv.ParseInt(parts[0], 10, 64)
	cents := int64(0)
	if len(parts) > 1 {
		fraction := parts[1]
		if len(fraction) == 1 {
			fraction += "0"
		}
		cents, _ = strconv.ParseInt(fraction, 10, 64)
	}
	n = n*100 + cents
	if n > 1000000000 {
		return 0, fail("Valor acima do limite permitido")
	}
	return n, nil
}
func advancedValues(q AdvancedQuery, v url.Values) (Row, error) {
	out := Row{"empresa": "Azul", "valor": "800.00", "uf": "SP", "codigo": "G31111", "tempo": "45", "cidade": "Recife", "trechos": "3", "chegadas": "1", "sigla": "SDU"}
	for k, val := range q.Defaults {
		out[k] = val
	}
	for _, k := range q.Params {
		if list, ok := v[k]; ok && len(list) > 0 {
			out[k] = strings.TrimSpace(list[0])
		}
		if out[k] == "" || len(out[k]) > 150 {
			return nil, fail("Preencha o filtro " + k + " com até 150 caracteres")
		}
		switch k {
		case "valor":
			if _, err := moneyCents(out[k]); err != nil {
				return nil, err
			}
		case "tempo", "trechos", "chegadas":
			n, e := strconv.Atoi(out[k])
			if e != nil || n < 0 || n > 100000 {
				return nil, fail("Filtro numérico inválido: " + k)
			}
		case "sigla":
			if !regexp.MustCompile("^[A-Za-z0-9]{3}$").MatchString(out[k]) {
				return nil, fail("Sigla deve ter três caracteres alfanuméricos")
			}
			out[k] = strings.ToUpper(out[k])
		case "uf":
			if !regexp.MustCompile("^[A-Za-z]{2}$").MatchString(out[k]) {
				return nil, fail("Informe uma UF com duas letras")
			}
			out[k] = strings.ToUpper(out[k])
		}
	}
	return out, nil
}
func advancedSQL(q AdvancedQuery) string {
	switch q.ID {
	case 35:
		return "SELECT * FROM aeroportos WHERE sigla=:sigla ORDER BY id_aeroporto"
	case 36, 37, 38:
		return "SELECT * FROM voos WHERE codigo_voo=:codigo COLLATE NOCASE"
	case 39:
		return "SELECT e.*,v.codigo_voo FROM escalas e JOIN voos v ON v.id_voo=e.id_voo WHERE v.codigo_voo=:codigo COLLATE NOCASE AND e.sequencial=1"
	default:
		return q.SQL
	}
}
func advancedArgs(sqlText string, u User, v Row) []interface{} {
	values := map[string]interface{}{"empresa_id": u.Company, "empresa": v["empresa"], "codigo": v["codigo"], "uf": v["uf"], "cidade": v["cidade"], "sigla": v["sigla"]}
	for _, k := range []string{"tempo", "trechos", "chegadas"} {
		values[k] = num(v[k])
	}
	n, _ := moneyCents(v["valor"])
	values["valor"] = float64(n) / 100
	args := []interface{}{}
	seen := map[string]bool{}
	for _, name := range regexp.MustCompile(`:[a-z_]+`).FindAllString(sqlText, -1) {
		name = strings.TrimPrefix(name, ":")
		if name == "limite" || name == "deslocamento" {
			continue
		}
		if !seen[name] {
			args = append(args, sql.Named(name, values[name]))
			seen[name] = true
		}
	}
	return args
}
func advancedRead(ctx context.Context, db querier, u User, q AdvancedQuery, v Row, page int) (AdvancedTable, error) {
	sqlText := advancedCTE + advancedSQL(q) + " LIMIT :limite OFFSET :deslocamento"
	args := advancedArgs(sqlText, u, v)
	// Pagination values are always generated by the server.
	args = append(args, sql.Named("limite", 51), sql.Named("deslocamento", (page-1)*50))
	rs, err := db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return AdvancedTable{}, err
	}
	defer rs.Close()
	columns, err := rs.Columns()
	if err != nil {
		return AdvancedTable{}, err
	}
	t := AdvancedTable{Columns: columns, Rows: [][]string{}, Raw: []Row{}}
	for rs.Next() {
		values := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err = rs.Scan(ptrs...); err != nil {
			return t, err
		}
		if len(t.Rows) == 50 {
			t.HasNext = true
			break
		}
		line := []string{}
		raw := Row{}
		for i, c := range columns {
			s := ""
			if values[i] != nil {
				s = fmt.Sprint(values[i])
			}
			raw[c] = s
			display := s
			if c == "valor" || c == "media_valor" {
				if s != "" {
					n, _ := strconv.ParseFloat(s, 64)
					display = "R$ " + strings.ReplaceAll(fmt.Sprintf("%.2f", n), ".", ",")
				} else {
					display = "—"
				}
			}
			line = append(line, display)
		}
		t.Rows = append(t.Rows, line)
		t.Raw = append(t.Raw, raw)
	}
	return t, rs.Err()
}
func (a *App) advancedData(r *http.Request) (AdvancedData, error) {
	id, e := advancedID(r.URL.Query())
	if e != nil {
		return AdvancedData{}, e
	}
	q, _ := advancedQuery(id)
	v, e := advancedValues(q, r.URL.Query())
	if e != nil {
		return AdvancedData{}, e
	}
	page := 1
	if s := r.URL.Query().Get("page"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > 100000 {
			return AdvancedData{}, fail("Página inválida")
		}
		page = n
	}
	u := current(r)
	company, e := one(r.Context(), a.DB, "empresa", u.Company)
	if e != nil {
		return AdvancedData{}, e
	}
	d := AdvancedData{Catalog: advancedCatalog, Query: q, Values: v, CSRF: csrfValue(r), Company: company["nome"], Page: page, CanMutate: company["situacao"] == "Ativa"}
	labels := map[string]string{"empresa": "Nome da empresa", "valor": "Valor de referência (R$)", "uf": "UF", "codigo": "Código do voo", "tempo": "Tempo mínimo (minutos)", "cidade": "Cidade de destino", "trechos": "Quantidade mínima de trechos", "chegadas": "Mais de quantas chegadas", "sigla": "Sigla do aeroporto"}
	for _, key := range q.Params {
		typ := "text"
		if key == "tempo" || key == "trechos" || key == "chegadas" {
			typ = "number"
		}
		d.Filters = append(d.Filters, Input{Field: field(key, labels[key], typ), Value: v[key]})
	}
	vals := url.Values{"id": {strconv.Itoa(id)}}
	for _, k := range q.Params {
		vals.Set(k, v[k])
	}
	d.URL = vals.Encode()
	d.SQL = advancedCTE + q.SQL + ";"
	if q.Kind == "write" {
		d.SQL = q.SQL + ";"
	}
	if id == 40 {
		d.SQL += "\n\n-- Exclusão individual, após verificar todas as referências:\nDELETE FROM aeroporto WHERE id=:registro_id AND versao=:versao;"
	}
	d.Table, e = advancedRead(r.Context(), a.DB, u, q, v, page)
	if e != nil {
		return d, e
	}
	if len(d.Table.Raw) > 0 {
		row := d.Table.Raw[0]
		d.PreviewID = row["id_voo"]
		if id == 39 {
			d.PreviewID = row["id_escala"]
		}
		d.Version = row["versao"]
	}
	switch id {
	case 35:
		opts, _, err := a.options(r.Context(), u, entities["aeroporto"])
		if err != nil {
			return d, err
		}
		d.Fields = []Input{{Field: field("nome", "Nome do aeroporto", "text"), Value: "Aeroporto de Santos Dumont"}, {Field: field("sigla", "Sigla IATA", "text"), Value: v["sigla"]}, {Field: ref("cidade_id", "Cidade / estado", "cidade"), Options: opts["cidade_id"]}}
	case 36:
		opts, _, err := a.options(r.Context(), u, entities["voo"])
		if err != nil {
			return d, err
		}
		d.Fields = []Input{{Field: field("codigo", "Código do voo", "text"), Value: v["codigo"]}, {Field: ref("aeronave_id", "Aeronave", "aeronave"), Options: opts["aeronave_id"]}, {Field: ref("origem_id", "Aeroporto de origem", "aeroporto"), Options: opts["origem_id"]}, {Field: ref("destino_id", "Aeroporto de destino", "aeroporto"), Options: opts["destino_id"]}, {Field: field("partida", "Partida", "datetime-local")}, {Field: field("chegada", "Chegada", "datetime-local")}, {Field: field("valor", "Tarifa do voo (R$)", "text"), Value: "950.00"}}
	case 37:
		d.Fields = []Input{{Field: field("valor", "Novo valor (R$)", "text"), Value: "1050.00"}}
	case 40:
		for _, row := range d.Table.Raw {
			d.Candidates = append(d.Candidates, Option{row["id_aeroporto"] + ":" + row["versao"], row["sigla"] + " · " + row["nome"] + " · " + row["cidade"] + "/" + row["uf"]})
		}
	}
	return d, nil
}
func (a *App) advancedPage(w http.ResponseWriter, r *http.Request) {
	if !advancedAllowed(w, r) {
		return
	}
	if e := a.Sync(r.Context()); e != nil {
		http.Error(w, "Não foi possível atualizar a operação", 500)
		return
	}
	d, e := a.advancedData(r)
	if e != nil {
		http.Error(w, friendly(e), 400)
		return
	}
	u := current(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if e = ui.Layout(u.Name, u.Role, d.CSRF, a.component("advanced", d)).Render(r.Context(), w); e != nil {
		http.Error(w, "Falha ao renderizar consultas", 500)
	}
}
func (a *App) advancedResult(w http.ResponseWriter, r *http.Request) {
	if !advancedAllowed(w, r) {
		return
	}
	if e := a.Sync(r.Context()); e != nil {
		http.Error(w, "Não foi possível atualizar a operação", 500)
		return
	}
	d, e := a.advancedData(r)
	if e != nil {
		http.Error(w, friendly(e), 400)
		return
	}
	w.Header().Set("HX-Push-Url", "/consultas-avancadas?"+d.URL+"&page="+strconv.Itoa(d.Page))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	a.render(w, "advanced", d)
}
func (a *App) advancedCSV(w http.ResponseWriter, r *http.Request) {
	if !advancedAllowed(w, r) {
		return
	}
	d, e := a.advancedData(r)
	if e != nil {
		http.Error(w, friendly(e), 400)
		return
	}
	if d.Query.Kind == "write" {
		http.Error(w, "Esta operação não é um relatório", 400)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=consulta-%02d-pagina-%d.csv", d.Query.ID, d.Page))
	fmt.Fprint(w, "\xef\xbb\xbf")
	writer := csv.NewWriter(w)
	defer writer.Flush()
	writer.Write(d.Table.Columns)
	for _, row := range d.Table.Raw {
		line := []string{}
		for _, c := range d.Table.Columns {
			line = append(line, csvSafe(row[c]))
		}
		writer.Write(line)
	}
}
