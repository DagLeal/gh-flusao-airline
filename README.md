# GH Flusão Airline

**Sistema de Gestão de Aeroporto — documentação do projeto no modelo Ágil Scrum.**

Projeto da disciplina **Metodologia e Construção de Projetos de TI 2**, desenvolvido a partir da [documentação UML fornecida](docs/original/Documentacao-UML-Sistema-Aeroporto.pdf). A interface utiliza a identidade **Erlaine Airport Operations**.

| Informação | Referência |
|---|---|
| Repositório | [ozelfls/gh-flusao-airline](https://github.com/ozelfls/gh-flusao-airline) |
| Referência funcional | PDF com 48 páginas, dez entidades e casos de uso UC01–UC08 |
| Situação do produto | Incremento funcional disponível para demonstração local; aceite dos stakeholders pendente |
| Publicação em produção | Pendente; configuração para Render incluída |
| Atualização desta documentação | 03/11/2026 |

Esta documentação organiza o trabalho com os artefatos e compromissos do Scrum. As funcionalidades existentes são apresentadas como evidência técnica. O planejamento abaixo é uma proposta para continuidade: não constitui um histórico de Sprints ou reuniões já realizadas.

![Painel de operações](docs/screenshots/operacoes.png)

## Sumário

- [Visão e Meta do Produto](#visão-e-meta-do-produto)
- [Stakeholders e Scrum Team](#stakeholders-e-scrum-team)
- [Escopo e requisitos](#escopo-e-requisitos)
- [Product Backlog](#product-backlog)
- [Histórias de usuário e critérios de aceite](#histórias-de-usuário-e-critérios-de-aceite)
- [Planejamento das Sprints](#planejamento-das-sprints)
- [Sprint Backlog proposto](#sprint-backlog-proposto)
- [Eventos e acompanhamento](#eventos-e-acompanhamento)
- [Definição de Pronto](#definição-de-pronto)
- [Incremento e evidências](#incremento-e-evidências)
- [Riscos e decisões](#riscos-e-decisões)
- [Arquitetura e tecnologias](#arquitetura-e-tecnologias)
- [Executar e validar](#executar-e-validar)
- [Documentação complementar](#documentação-complementar)

## Visão e Meta do Produto

O sistema centraliza o planejamento de voos, a disponibilidade de aeronaves e aeroportos e o atendimento de passageiros. O objetivo é permitir que a equipe acompanhe cada operação e seus registros relacionados sem depender de controles separados.

**Meta do Produto (Product Goal):** disponibilizar uma aplicação web leve que permita cadastrar uma operação aérea, emitir e acompanhar seus bilhetes e consultar os dados por perfil de acesso, preservando as relações e regras descritas no modelo UML.

A meta será considerada atingida quando uma demonstração validada pelos stakeholders comprovar o fluxo completo — cadastro operacional → voo e escalas → passageiro e bilhete → check-in → embarque → conclusão ou cancelamento —, com persistência dos dados e documentação para execução.

## Stakeholders e Scrum Team

Os perfis abaixo representam usuários do produto, e não responsabilidades Scrum.

| Stakeholder / perfil | Necessidade |
|---|---|
| Gerente da empresa aérea | Administrar frota, aeroportos, voos e escalas da operação |
| Atendente de aeroporto | Cadastrar passageiros, emitir bilhetes e executar atendimento de embarque |
| Passageiro | Consultar voos e acessar somente seus próprios bilhetes |
| Docente / avaliador | Conferir aderência ao enunciado, UML, execução e documentação |
| Responsável pelo projeto | Validar prioridades e acompanhar a evolução da entrega |

### Equipe e frentes de trabalho

| Frente | Responsáveis | Atuação |
|---|---|---|
| Infraestrutura | Dagner Costa | Configuração dos ambientes, Docker e Render, persistência do SQLite, backup e deploy |
| Desenvolvimento | Daniel Oliveira, Pedro Artur e Dagner Costa | Implementação do backend e frontend, regras de negócio, integração e testes |
| Modelagens UML em geral | Dagner Costa e Daniel Oliveira | Elaboração e manutenção dos modelos e diagramas UML, incluindo classes, casos de uso, estados e sequência |

### Responsabilidades Scrum

| Responsabilidade | Aplicação no projeto | Integrante |
|---|---|---|
| Product Owner | Ordenar o backlog, esclarecer o objetivo do produto e alinhar necessidades com stakeholders | A definir pela equipe |
| Scrum Master | Apoiar o uso de Scrum, facilitar os eventos e ajudar a remover impedimentos | A definir pela equipe |
| Developers | Planejar e executar o trabalho, testar, documentar e entregar incrementos utilizáveis | Daniel Oliveira, Pedro Artur e Dagner Costa |

A equipe técnica está identificada nas frentes acima. As designações de Product Owner e Scrum Master permanecem a definir pela equipe.

## Escopo e requisitos

### Escopo funcional

- Cadastro de País, Estado, Cidade, Empresa, Aeroporto, Aeronave, Voo, Escala, Passageiro e Bilhete.
- Gestão dos estados operacionais e manutenção do histórico.
- Autenticação e autorização para gerente, atendente e passageiro.
- Emissão de bilhetes com capacidade e assentos controlados.
- Check-in, embarque, cancelamentos e atualização automática dos estados por horário.
- Telas mestre/detalhe e grade interativa com edição de várias linhas.
- Busca, filtros, ordenação, paginação, exportação CSV e impressão de bilhete.
- Banco SQLite local, backup consistente e configuração de deploy.

### Requisitos de qualidade

| ID | Requisito verificável |
|---|---|
| RNF01 | Executar em um único serviço Go, sem Node ou servidor de banco separado em produção |
| RNF02 | Preservar referências do modelo e impedir assentos ativos duplicados no mesmo voo |
| RNF03 | Validar permissões no servidor e isolar os bilhetes de cada passageiro |
| RNF04 | Salvar a edição de várias linhas de forma atômica e detectar versões obsoletas |
| RNF05 | Permitir uso em desktop e celular, com rolagem interna nas grades |
| RNF06 | Manter os dados após reinício e permitir backup verificável |
| RNF07 | Oferecer instruções reproduzíveis de execução, testes e publicação |

## Product Backlog

A ordem abaixo é uma proposta de priorização a ser revisada pelo Product Owner. Não foram atribuídos pontos ou velocidade retrospectivamente; as estimativas serão discutidas pelo time durante o refinamento.

**Legenda:** “Implementado” indica disponibilidade técnica e não equivale a aceite do stakeholder nem a encerramento de uma Sprint.

| Ordem | ID | Item / valor para o usuário | Referência | Situação técnica |
|---|---|---|---|---|
| 1 | US01 | Acessar o sistema conforme o perfil | Transversal | Implementado; aceite pendente |
| 2 | US02 | Organizar a localização dos aeroportos | País, Estado, Cidade | Implementado; aceite pendente |
| 3 | US03 | Administrar frota e aeroportos disponíveis | UC03, UC04 | Implementado; aceite pendente |
| 4 | US04 | Planejar e acompanhar voos da empresa | UC01 | Implementado; aceite pendente |
| 5 | US05 | Organizar escalas e associá-las a voos | UC02 | Implementado; aceite pendente |
| 6 | US06 | Cadastrar e manter passageiros | UC06 | Implementado; aceite pendente |
| 7 | US07 | Emitir e acompanhar bilhetes até o embarque | UC05 | Implementado; aceite pendente |
| 8 | US08 | Consultar voos com os filtros do PDF | UC07 | Parcial: consulta de ida disponível; retorno e duração pendentes |
| 9 | US09 | Consultar bilhetes e histórico individual | UC08 | Parcial: consulta e ações disponíveis; período De/Até pendente |
| 10 | US10 | Trabalhar com grades e registros relacionados | Transversal | Implementado; aceite pendente |
| 11 | US11 | Preservar dados e produzir backups | RNF02, RNF06 | Implementado e verificado tecnicamente |
| 12 | US12 | Executar verificações reproduzíveis | RNF03–RNF07 | Testes e evidências disponíveis |
| 13 | US13 | Acessar a aplicação em hospedagem persistente | RNF01, RNF06, RNF07 | Docker/Render configurados; publicação pendente |
| 14 | US14 | Consultar voos em calendário visual | Alternativa UC01 | Pendente |
| 15 | US15 | Formalizar revisão e aceite do incremento | Meta do Produto | Pendente |

Os detalhes das divergências do PDF, das políticas adotadas e das funcionalidades complementares ainda ausentes estão em [rastreabilidade e decisões](docs/uml-e-decisoes.md). Não fazem parte deste escopo integrações com companhias reais, pagamentos, SMS ou e-mail.

## Histórias de usuário e critérios de aceite

Os critérios descrevem o comportamento esperado e servem de roteiro para a revisão do incremento.

| ID | História de usuário | Critérios de aceite |
|---|---|---|
| US01 | Como usuário, quero entrar com meu perfil para acessar as operações autorizadas. | Credenciais válidas iniciam a sessão; acesso direto sem autorização é recusado; passageiro não obtém bilhetes de terceiros. |
| US02 | Como gerente, quero organizar países, estados e cidades para localizar os aeroportos. | Referências geográficas devem existir; adicionar dependente ativa o pai conforme o diagrama; inativação incompatível e exclusão com referências são recusadas. |
| US03 | Como gerente, quero controlar frota e aeroportos para selecionar recursos disponíveis. | Capacidade positiva; códigos únicos; aeroportos vinculados a cidades; ativação e manutenção respeitam estados e operações pendentes. |
| US04 | Como gerente, quero cadastrar voos para planejar a operação da minha empresa. | Chegada posterior à partida; recursos ativos e aeronave da empresa; conflito de horário recusado; edição restrita a agendados; cancelamento preserva histórico e atualiza dependentes. |
| US05 | Como gerente, quero organizar escalas para representar os trechos de um voo. | Sequencial positivo e único por voo; tempo positivo e tarifa não negativa; origem diferente do destino; associação somente a voo agendado. |
| US06 | Como atendente, quero cadastrar passageiros para emitir bilhetes com dados consistentes. | Documento válido no formato suportado e único; documento imutável; milhagem não negativa; histórico impede exclusão incompatível. |
| US07 | Como atendente, quero emitir bilhetes e realizar check-in e embarque para atender o passageiro. | Assento livre e dentro da capacidade; cadastro inline e emissão na mesma transação; check-in exige portão; embarque exige check-in; cancelamento permitido libera o assento. |
| US08 | Como passageiro, quero buscar voos para escolher uma viagem. | Filtrar data, origem e destino; rejeitar origem igual ao destino; consultar horários, escalas e assentos; completar retorno e duração máxima previstos no PDF. |
| US09 | Como passageiro, quero consultar meus bilhetes para acompanhar a viagem. | Listagem, detalhe, CSV e impressão restritos ao titular; ações respeitam estado e partida; completar filtro de período com datas válidas. |
| US10 | Como operador, quero editar linhas e consultar detalhes sem recarregar a página inteira. | Seleção carrega dependentes; Tab permite navegar entre células; salvar confirma todo o conjunto ou reverte tudo; conflito de versão gera aviso. |
| US11 | Como responsável pela operação, quero preservar e recuperar os dados. | Reinício mantém registros; backup do banco em uso produz snapshot íntegro; destino existente não é sobrescrito; restauração possui instruções. |
| US12 | Como equipe, quero verificar os fluxos críticos antes de entregar alterações. | Testes reprodutíveis de concorrência, autorização e transações; verificações de navegador documentadas; falhas e limites informados. |
| US13 | Como usuário, quero acessar uma instância publicada com dados duráveis. | Serviço responde ao health check; autenticação funciona em HTTPS; reinício/deploy preserva o SQLite; backup e publicação são documentados. |
| US14 | Como gerente, quero consultar um calendário para visualizar os voos por data. | Navegar por datas; selecionar voo e abrir detalhes; restringir alterações pelo estado e pela empresa. |
| US15 | Como stakeholder, quero revisar uma demonstração para confirmar o atendimento das necessidades. | Executar o roteiro dos critérios; registrar resultados, feedback e pendências; atualizar o backlog com as decisões. |

## Planejamento das Sprints

**Cadência proposta:** Sprints de duas semanas, com início e término a combinar pela equipe. A capacidade e as estimativas ainda precisam ser definidas; os recortes seguintes não são compromissos já assumidos.

| Sprint proposta | Meta da Sprint | Itens candidatos | Incremento esperado |
|---|---|---|---|
| Próxima Sprint | Completar e revisar a consulta de viagens e bilhetes | US08, US09, recorte de US15 | Consulta de retorno/duração e histórico por período demonstráveis, com feedback registrado |
| Sprint seguinte | Facilitar a visão de planejamento da operação | US14, ajustes priorizados da Review | Calendário integrado à seleção de voos e detalhes |
| Sprint de publicação | Disponibilizar o incremento validado com persistência | US13, revisão de US11 e US12, US15 | Instância hospedada verificada, backup e roteiro de entrega |

Os incrementos já disponíveis serão usados como base. O planejamento deve ser revisto a cada Sprint Planning, considerando o aprendizado da Review e a capacidade real do time.

## Sprint Backlog proposto

**Meta:** completar e revisar a consulta de viagens e bilhetes.

| Tarefa | Item | Condição de conclusão | Dependência | Situação |
|---|---|---|---|---|
| Confirmar comportamento esperado dos filtros com o stakeholder | US08, US09 | Exemplos de ida/volta, duração e período registrados | Disponibilidade do stakeholder | A fazer |
| Implementar consulta de retorno e duração máxima | US08 | Retorno respeita rota inversa e datas; limite de duração filtra resultados | Regras confirmadas | A fazer |
| Implementar período De/Até para bilhetes | US09 | Datas válidas; intervalo coerente; escopo do titular mantido | Regras confirmadas | A fazer |
| Verificar rotas com escalas intermediárias | US08 | Busca comparada aos exemplos e às associações do PDF | Casos de teste definidos | A fazer |
| Cobrir novos filtros com testes | US08, US09, US12 | Casos válidos, inválidos e isolamento aprovados | Implementação dos filtros | A fazer |
| Preparar demonstração e atualizar documentação | US15 | Evidências e resultados da Review registrados | Incremento verificável | A fazer |

Responsáveis e estimativas serão definidos pelos Developers no planejamento. Nenhuma Sprint foi formalmente iniciada nesta documentação.

## Eventos e acompanhamento

| Evento / atividade | Aplicação proposta ao projeto | Registro |
|---|---|---|
| Sprint Planning | Definir a meta e selecionar trabalho compatível com a capacidade | Meta, itens e plano da Sprint |
| Daily Scrum | Inspecionar diariamente o progresso e adaptar o plano; até 15 minutos | Impedimentos e mudanças relevantes |
| Sprint Review | Demonstrar o incremento e discutir o próximo trabalho com stakeholders | Resultados, feedback e atualização do backlog |
| Sprint Retrospective | Inspecionar a forma de trabalhar e escolher melhorias | Ação de melhoria, responsável e acompanhamento |
| Refinamento contínuo | Esclarecer histórias, critérios, dependências e estimativas | Product Backlog atualizado |

O refinamento é uma atividade contínua, não um evento formal adicional do Scrum. O [modelo de registros](docs/scrum/registros.md) deve ser preenchido quando os eventos ocorrerem.

### Quadro de acompanhamento

Fluxo de trabalho sugerido: **A fazer → Em andamento → Em verificação → Concluído**.

Cada item deve ter ID do backlog, responsável, critérios, evidência e impedimentos. O GitHub Projects pode hospedar o quadro; sua criação e o cadastro de issues ainda não foram realizados.

Não há dados suficientes para informar velocidade, burndown ou quantidade de Sprints concluídas. As métricas deverão partir de registros reais da equipe.

## Definição de Pronto

**Definition of Done proposta**, a ser acordada pelo Scrum Team:

- [ ] Comportamento do item implementado e demonstrável conforme seus critérios.
- [ ] Relações e regras do domínio preservadas no backend.
- [ ] Permissões e isolamento de dados verificados, quando aplicáveis.
- [ ] Testes pertinentes aprovados e evidência vinculada ao item.
- [ ] Interface e mensagens verificadas nas resoluções pertinentes.
- [ ] Código revisado e verificações técnicas sem falhas conhecidas que impeçam o uso.
- [ ] Documentação e limitações atualizadas.
- [ ] Incremento integrado e executável a partir do repositório.

Esta lista é um acordo proposto, não um certificado retroativo de conclusão. A Review inspeciona o incremento e pode gerar mudanças no backlog. Publicação em produção é um item próprio, sem confundir “implementado”, “verificado”, “aceito” e “publicado”.

## Incremento e evidências

A versão atual permite uma demonstração local do fluxo principal, com banco persistente e usuários fictícios. Os resultados disponíveis são:

| Evidência técnica | Referência |
|---|---|
| Dez testes Go de integração | [Testes de aplicação](internal/app/app_test.go) e [testes de domínio](internal/app/domain_test.go) |
| Fluxos de desktop e celular em navegador real | [Smoke de navegador](scripts/browser-smoke.cjs) |
| Telas do incremento | [Operações](docs/screenshots/operacoes.png), [mestre/detalhe](docs/screenshots/master-detail.png), [bilhetes](docs/screenshots/bilhetes.png), [mobile](docs/screenshots/mobile.png) |
| Build Windows/Linux, persistência, backup e verificações | [Relatório de validação](docs/validacao.md) |
| Modelo e regras de UC01–UC08 | [Rastreabilidade UML](docs/uml-e-decisoes.md) |
| Configuração de hospedagem | [Blueprint Render](render.yaml) e [Dockerfile](Dockerfile) |

As verificações técnicas estão documentadas na entrega original. A documentação Scrum não substitui o aceite da equipe avaliadora, nem indica que houve publicação ou eventos Scrum já executados.

## Riscos e decisões

| Risco / dependência | Situação e tratamento |
|---|---|
| Divergências entre diagramas e descrições do PDF | Decisões registradas em [UML e decisões](docs/uml-e-decisoes.md); confirmar com stakeholders |
| Responsabilidades Scrum e agenda | Equipe técnica definida; Product Owner, Scrum Master e datas das Sprints permanecem a definir |
| SQLite em ambiente efêmero | Usar Render com disco persistente; Vercel exigiria mudança de persistência |
| Validação do contêiner | Dockerfile preparado, mas a imagem ainda não foi executada nesta entrega |
| Hospedagem e custo do disco | Publicação depende da configuração da conta e do serviço persistente |
| Dados e segredos locais no Git | Banco, dependências, executáveis e arquivos de ambiente privados excluídos pelo .gitignore |
| Funcionalidades complementares pendentes | Retorno, duração, calendário e período constam no backlog |

## Arquitetura e tecnologias

| Camada | Tecnologia |
|---|---|
| Backend | Go |
| Framework HTTP | Chi |
| Frontend | Go Templates + Templ |
| Estilização | Tailwind CSS |
| Banco | SQLite |
| Driver | modernc.org/sqlite |
| Queries geradas | sqlc |
| Interatividade | HTMX |
| Hospedagem preparada | Render com disco persistente |

O executável incorpora templates e arquivos estáticos. O código sqlc atende autenticação, sessões, atualização de estados e auditoria; o CRUD genérico das grades usa metadados internos e parâmetros SQL. Consulte a [organização das queries](docs/uml-e-decisoes.md#organização-das-queries).

## Executar e validar

Requisito para iniciar: **Go 1.26 ou superior**. O código gerado está incluído; Node é necessário somente para recompilar CSS e executar o smoke de navegador.

No PowerShell:

```powershell
git clone https://github.com/ozelfls/gh-flusao-airline.git
cd gh-flusao-airline
$env:DEMO_MODE = "true"
$env:COOKIE_SECURE = "false"
go run ./cmd/server
```

Abra **http://localhost:8080**. Os dados são criados na primeira execução e persistem em `data/aeroporto.db`.

| Perfil da demonstração | E-mail | Senha |
|---|---|---|
| Gerente | gerente@demo.local | Demo@2026! |
| Atendente | atendente@demo.local | Demo@2026! |
| Passageiro | passageiro@demo.local | Demo@2026! |

As credenciais são exclusivas do ambiente fictício. Para produção, use `DEMO_MODE=false` e configure o primeiro acesso conforme o guia.

Verificações Go:

```sh
go test ./... -count=1
go vet ./...
go build ./cmd/server
```

O [guia técnico](docs/guia-tecnico.md) contém Docker Compose, criação de contas, backup, geração de Templ/sqlc/Tailwind, teste de navegador, variáveis de ambiente e deploy no Render.

## Documentação complementar

- [Guia técnico de execução e deploy](docs/guia-tecnico.md).
- [Rastreabilidade UML, decisões e limites](docs/uml-e-decisoes.md).
- [Resultados e limites da validação técnica](docs/validacao.md).
- [Modelo para registros reais das Sprints](docs/scrum/registros.md).
- [Documento UML original](docs/original/Documentacao-UML-Sistema-Aeroporto.pdf).

Referência metodológica: Ken Schwaber e Jeff Sutherland, [Guia do Scrum — edição de novembro de 2020, português brasileiro](https://scrumguides.org/docs/scrumguide/v2020/2020-Scrum-Guide-PortugueseBR-3.0.pdf). User stories, o quadro sugerido e o período de duas semanas são escolhas de organização propostas para este projeto.
