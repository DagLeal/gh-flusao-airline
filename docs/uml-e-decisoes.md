# Rastreabilidade e decisões

Fonte: Documentacao-UML-Sistema-Aeroporto.pdf, 48 páginas, lido como texto e imagens dos diagramas. Nenhum arquivo da pasta original foi alterado.

## Modelo conceitual

| Entidade do PDF | Tabela | Relações preservadas |
|---|---|---|
| País | pais | 1:N estado |
| Estado | estado | N:1 pais, 1:N cidade |
| Cidade | cidade | N:1 estado, 1:N aeroporto |
| Aeroporto | aeroporto | N:1 cidade; origem/destino de escalas, voos e bilhetes |
| Empresa | empresa | 1:N aeronave e voo |
| Aeronave | aeronave | N:1 empresa, 1:N voo |
| Voo | voo | N:1 empresa e aeronave; 1:N escala e bilhete |
| Escala | escala | N:1 voo quando associada; dois aeroportos |
| Passageiro | passageiro | 1:N bilhete |
| Bilhete | bilhete | N:1 passageiro e voo; dois aeroportos |

Todos os atributos do diagrama conceitual têm representação. Os nomes `partida` e `chegada` combinam data e hora em timestamps UTC; o formulário converte o fuso da operação. Valores monetários usam centavos inteiros e tempo de escala usa minutos inteiros. Os campos de situação, versão, classe, eventos de bilhete, usuários, sessões, auditoria e manutenção complementam o modelo conceitual para executar as regras narradas.

## Casos de uso

| Caso | Implementação principal |
|---|---|
| UC01 – Cadastrar voo | Criar/editar agendado, aeronave e aeroportos ativos, empresa da sessão, horários coerentes, bloqueio de conflitos da aeronave, cancelamento e cascatas |
| UC02 – Cadastrar escala | Planejada sem voo, associação a agendado, sequencial único, tempo positivo, tarifa não negativa, cancelamento e impacto em bilhetes |
| UC03 – Cadastrar aeronave | Registro, ativação, manutenção, retorno, inativação, exclusão lógica e histórico |
| UC04 – Cadastrar aeroporto | Cadastro com cidade, sigla válida e única, ativação, manutenção, inativação e exclusão lógica |
| UC05 – Cadastrar bilhete | Emissão transacional, passageiro inline, mapa de assentos, capacidade, assento único, check-in com portão, embarque e cancelamento |
| UC06 – Cadastrar passageiro | Documento único e imutável, nome, telefone, passaporte, milhagem, histórico de bilhetes e restrições de inativação/exclusão |
| UC07 – Consultar voo | Consulta autenticada por data, origem, destino, código e situação; detalhes, escalas e mapa de assentos |
| UC08 – Consultar bilhete | Isolamento por passageiro em listagem, detalhe, CSV, impressão e ações; check-in e cancelamento antes da partida |

O relógio executa a cada 15 segundos e antes de consultas/mutações: Agendado → Em Voo → Concluído; Programada → Em Andamento → Concluída; bilhetes não cancelados do voo concluído → Usado. Os updates incrementam versões para invalidar edições obsoletas.

## Decisões para divergências do PDF

1. **Aeronaves e aeroportos com histórico:** os fluxos alternativos permitem excluir mantendo histórico, enquanto algumas exceções proíbem. A implementação preserva as linhas e referências por exclusão lógica, bloqueando voos/escalas ativos.
2. **Voo com pelo menos uma escala versus escala sem voo:** cada voo novo recebe um trecho inicial programado, na mesma transação. Escalas independentes podem ser planejadas e depois associadas. A tarifa inicial é zero e o tempo inicial é a duração estimada; para uma rota comercial, cadastre os trechos tarifados conforme a operação.
3. **Cancelamento de escala e bilhetes:** o modelo associa bilhete ao voo, sem bilhete_escala. Por isso, cancelar uma escala programada cancela os bilhetes ainda não embarcados do voo, com motivo registrado; não é possível identificar um subconjunto por trecho.
4. **Exclusão:** o CRUD principal preserva o histórico por exclusão lógica. País, estado, cidade e empresa só podem alcançar exclusão sem referências dependentes; passageiro exige ausência de qualquer bilhete histórico.
5. **Aeronave em manutenção:** o PDF cita “N dias” sem definir configuração. A política adotada bloqueia manutenção ou retirada enquanto houver qualquer voo agendado ou em progresso, independentemente de quantos dias faltam.
6. **Check-in e embarque:** check-in termina na partida; embarque por atendente inicia duas horas antes, podendo ser confirmado com voo em andamento conforme a narrativa. Passageiro não confirma embarque.
7. **Geografia:** país e estado começam em Cadastrado, cidade em Cadastrada; a criação do primeiro dependente ativa o pai, conforme os diagramas. Não há reativação após a inativação terminal.
8. **Empresa inicial:** o provisionamento cria a empresa já ativa para permitir o primeiro gerente operar. Novas contas e vínculos são administrativos, sem seleção livre de perfil no login.
9. **Milhagem:** o modelo não define distância por voo nem fator de acúmulo; a atualização é manual pelo atendente. Não foi inventada uma conversão de tempo em distância.
10. **Documento:** valida formato de CPF de 11 dígitos ou passaporte com letras/números e unicidade. Não faz validação de dígitos verificadores do CPF nem consulta documental externa.
11. **Assento:** utiliza numeração inteira de 1 até a capacidade, compatível com a opção número/código do UC05. Não presume a configuração de letras e fileiras de cada fabricante.
12. **Alteração de rota:** preserva trechos e bilhetes. A mudança de origem/destino é bloqueada se existirem bilhetes ativos ou múltiplas escalas. Em um trecho direto sem bilhetes, sincroniza a escala inicial.
13. **Permissões:** gerente administra sua companhia/frota/voos e os cadastros compartilhados; atendente opera passageiros/bilhetes; passageiro acessa apenas seus bilhetes. A validação ocorre no backend, inclusive por URL direta.

## Fronteiras desta versão

- Não integra pagamentos, companhias reais, SMS, e-mail ou scanners de portão.
- A consulta de voo oferece filtros de ida; não monta itinerário ida/volta nem filtro de duração máxima.
- Não há calendário visual; a operação usa grade e filtros de data.
- O histórico de bilhetes é consultável pela grade e situação; não há filtro específico “De/Até”.
- Não há reset de senha por e-mail, auto cadastro ou painel de administração de contas.
- A impressão oferece comprovante e código de rastreamento, sem QR code.
- Exportação CSV é da página exibida (20 linhas), com escaping contra fórmulas de planilha.
- Seleções de referências carregam até 2.000 registros por cadastro. Para bases maiores, substitua por busca paginada.
- A grade permite editar múltiplas linhas existentes; inclusão usa modal e remoção usa transição de estado.
- SQLite é para um serviço com disco local persistente. Não foi projetado para réplicas distribuídas.

## Organização das queries

sqlc gera autenticação, sessões, relógio dos estados e auditoria a partir de `internal/database/queries.sql`. O CRUD compartilhado e os filtros/ordenação das grades usam SQL montado a partir de metadados internos imutáveis, com parâmetros para todos os valores. Nomes de tabela/coluna e direção de ordenação nunca vêm livremente do cliente.

O arquivo de queries usa apenas ASCII: os rótulos “Concluído/Concluída” são compostos por `char(237)`. Isso contorna uma falha observada na geração de offsets Unicode pelo parser SQLite do sqlc 1.31.1, mantendo o texto em português no banco e sem editar código gerado manualmente.

## Segurança e operação

Senhas bcrypt; cookies HttpOnly e SameSite Strict, Secure em HTTPS; sessões de oito horas armazenadas pelo hash do token; proteção CSRF em POST, inclusive no login; CSP sem JavaScript inline; limite de tamanho de requisição; limite de tentativas de login por IP; isolamento de dados e controle de versão de registros. O endpoint `/healthz` verifica a conexão do banco sem expor os dados.

Os dados compartilhados de geografia e aeroportos são globais, enquanto aeronaves e voos têm vínculo de empresa. Escalas planejadas sem voo são compartilhadas entre gerentes; ganham escopo de empresa após associação. A aplicação pressupõe uma equipe confiável para administrar os cadastros compartilhados.


## Ampliação: consultas avançadas

O catálogo de 40 itens está documentado em [Consultas avançadas](consultas-avancadas.md). As consultas de voos e seus indicadores respeitam a empresa da sessão; geografia, aeroportos e passageiros seguem o escopo compartilhado já existente.

O voo passa a possuir tarifa própria em centavos e um indicador de confirmação comercial. Na migração de um banco anterior, a tarifa é inicializada uma única vez pela soma das escalas não canceladas, com incremento de versão para invalidar formulários antigos. Novos voos têm tarifa própria, independente de ajustes posteriores nos valores dos trechos.

PLANEJADO representa Agendado sem confirmação; CONFIRMADO representa Agendado com confirmação. Quando o estado operacional muda, a consulta mostra o novo estado. Confirmar não antecipa decolagem nem conclusão.

As operações SQL 39 e 40 permitem exclusão física excepcional, em vez da exclusão lógica do CRUD. A primeira escala só pode ser removida antes da partida, sem bilhetes não cancelados e mantendo outro trecho ativo; a origem/destino do voo são sincronizados com a rota restante. Aeroportos exigem ausência de referências em todas as companhias e ausência de histórico de manutenção. O registro anterior é mantido na auditoria. IDs de aeroportos e escalas removidos não são reutilizados pelas operações da aplicação.

O catálogo é fixo e parametrizado. Nenhum texto SQL recebido do cliente é executado. O gerente não pode alterar a empresa usada para autorizar uma consulta ou mutação.
