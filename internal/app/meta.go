package app

type Field struct {
	Key, Label, Type, Ref string
	Optional              bool
}
type Entity struct {
	Key, Title, Singular, Intro, Initial string
	Fields                               []Field
}

func field(k, l, t string) Field { return Field{Key: k, Label: l, Type: t} }
func ref(k, l, r string) Field   { return Field{Key: k, Label: l, Type: "select", Ref: r} }
func optional(f Field) Field     { f.Optional = true; return f }

var entities = map[string]Entity{
	"pais":       {"pais", "Países", "país", "A base geográfica da sua rede de destinos.", "Cadastrado", []Field{field("nome", "Nome", "text"), field("continente", "Continente", "text"), field("sigla", "Sigla", "text")}},
	"estado":     {"estado", "Estados", "estado", "Organize os estados por país.", "Cadastrado", []Field{ref("pais_id", "País", "pais"), field("nome", "Nome", "text"), field("regiao", "Região", "text"), field("sigla", "Sigla", "text")}},
	"cidade":     {"cidade", "Cidades", "cidade", "Os destinos que conectam a sua operação.", "Cadastrada", []Field{ref("estado_id", "Estado", "estado"), field("nome", "Nome", "text"), field("sigla", "Sigla", "text")}},
	"empresa":    {"empresa", "Empresas", "empresa", "Companhias aéreas e disponibilidade operacional.", "Ativa", []Field{field("nome", "Razão social", "text"), field("cnpj", "CNPJ (14 dígitos)", "text")}},
	"aeroporto":  {"aeroporto", "Aeroportos", "aeroporto", "Sua rede de terminais, pronta para conectar destinos.", "Cadastrado", []Field{field("nome", "Nome", "text"), field("sigla", "Sigla IATA", "text"), ref("cidade_id", "Cidade", "cidade")}},
	"aeronave":   {"aeronave", "Frota", "aeronave", "Disponibilidade, capacidade e manutenção em um só lugar.", "Cadastrada", []Field{field("codigo_aeronave", "Código", "text"), field("modelo", "Modelo", "text"), field("capacidade", "Capacidade", "number"), ref("empresa_id", "Empresa", "empresa")}},
	"voo":        {"voo", "Voos e operações", "voo", "Planeje rotas. Acompanhe cada etapa. Conecte todos os detalhes.", "Agendado", []Field{optional(field("codigo_voo", "Código do voo", "text")), ref("empresa_id", "Empresa", "empresa"), ref("aeronave_id", "Aeronave", "aeronave"), ref("origem_id", "Origem", "aeroporto"), ref("destino_id", "Destino", "aeroporto"), field("partida", "Partida", "datetime-local"), field("chegada", "Chegada", "datetime-local")}},
	"escala":     {"escala", "Escalas e rotas", "escala", "Trechos organizados na ordem da sua operação.", "Planejada", []Field{optional(ref("voo_id", "Voo (associação opcional)", "voo")), field("sequencial", "Sequência", "number"), ref("origem_id", "Origem", "aeroporto"), ref("destino_id", "Destino", "aeroporto"), field("tempo_escala", "Tempo em minutos", "number"), field("valor_centavos", "Tarifa em centavos", "number")}},
	"passageiro": {"passageiro", "Passageiros", "passageiro", "Cadastro e histórico de quem viaja com você.", "Cadastrado", []Field{field("documento", "Documento (CPF ou passaporte)", "text"), field("nome", "Nome completo", "text"), field("milhagem", "Milhagem", "number"), optional(field("telefone", "Telefone", "tel")), optional(field("passaporte", "Passaporte", "text"))}},
	"bilhete":    {"bilhete", "Bilhetes", "bilhete", "Da emissão ao embarque, cada passageiro acompanhado.", "Emitido", []Field{ref("passageiro_id", "Passageiro", "passageiro"), ref("voo_id", "Voo", "voo"), field("assento", "Assento (número de 1 à capacidade)", "number")}},
}
var transitions = map[string]map[string][]string{
	"voo":        {"Agendado": {"Cancelado"}, "Em Voo": {"Cancelado"}},
	"escala":     {"Planejada": {"Cancelada"}, "Programada": {"Cancelada"}},
	"bilhete":    {"Emitido": {"Check-in Realizado", "Cancelado"}, "Check-in Realizado": {"Embarque Realizado", "Cancelado"}},
	"aeronave":   {"Cadastrada": {"Ativa", "Excluída"}, "Ativa": {"Em Manutenção", "Inativa"}, "Em Manutenção": {"Ativa"}, "Inativa": {"Excluída"}},
	"aeroporto":  {"Cadastrado": {"Ativo"}, "Ativo": {"Em Manutenção", "Inativo"}, "Em Manutenção": {"Ativo"}, "Inativo": {"Excluído"}},
	"passageiro": {"Ativo": {"Inativo"}, "Inativo": {"Excluído"}},
	"empresa":    {"Cadastrada": {"Ativa", "Excluída"}, "Ativa": {"Suspensa", "Inativa"}, "Suspensa": {"Ativa", "Inativa"}, "Inativa": {"Excluída"}},
	"pais":       {"Cadastrado": {"Excluído"}, "Ativo": {"Inativo"}, "Inativo": {"Excluído"}},
	"estado":     {"Cadastrado": {"Excluído"}, "Ativo": {"Inativo"}, "Inativo": {"Excluído"}},
	"cidade":     {"Cadastrada": {"Excluída"}, "Ativa": {"Inativa"}, "Inativa": {"Excluída"}},
}
