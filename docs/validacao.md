# Validação da entrega

Ambiente: Windows amd64, Go 1.27.3, Node 22+, Chrome instalado. Builds finais para Windows e Linux amd64.

## Verificações concluídas

| Verificação | Resultado |
|---|---|
| go test ./... -count=1 | 16 testes de integração aprovados |
| go vet ./... | Sem problemas |
| Build Windows com -trimpath -ldflags="-s -w" | Aprovado; executável de aproximadamente 15 MB |
| Build Linux amd64, CGO_ENABLED=0 | Aprovado; executável de aproximadamente 15 MB |
| Templ generate | Aprovado, versão 0.3.1100 |
| sqlc generate | Aprovado, versão 1.31.1; queries ASCII |
| Tailwind build | Aprovado, versão 4.1.18 |
| npm audit --audit-level=high | Zero vulnerabilidades reportadas |
| Navegador Chrome real via Playwright | Fluxos desktop e mobile aprovados |
| Backup VACUUM INTO com servidor em uso | Snapshot criado |
| PRAGMA integrity_check no snapshot | ok |
| PRAGMA foreign_key_check no snapshot | Nenhuma violação |
| Reinício usando banco existente | Dados preservados |
| Primeiro início em produção | Empresa e gerente criados sem carga de demonstração |

## Testes Go

1. Duas emissões concorrentes disputando o mesmo assento: apenas uma é confirmada.
2. Check-in exige portão; embarque sem check-in é recusado; cancelamento libera o assento.
3. Passageiro não acessa bilhete alheio nem pelo CSV; CSRF inválido é recusado.
4. Grade com duas linhas e uma inválida reverte todas; versão obsoleta não sobrescreve alterações.
5. Cancelamento de voo atualiza escalas e bilhetes; relógio conclui voo e escalas.
6. Novo voo cria escala inicial; aeronave de outra empresa é recusada.
7. Páginas, detalhes, impressão e login renderizam para os perfis previstos.
8. País/estado/cidade seguem ativação por dependência e bloqueiam inativação incompatível.
9. Passageiro inline é revertido se a emissão falhar; capacidade de assentos é respeitada.
10. Manutenção é bloqueada com voos pendentes; documento de passageiro é imutável; check-in após a partida é recusado.

## Navegador

O smoke executou login de gerente, seleção mestre/detalhe, carregamento de escalas e bilhetes, criação de voo, edição de células e salvamento da grade, login de atendente, emissão com passageiro inline e seleção no mapa de assentos, check-in com portão, login de passageiro e isolamento de bilhetes.

Viewport mobile: 390 × 844, sem transbordamento horizontal da página. As tabelas têm rolagem interna. A execução final não apresentou erros de JavaScript ou console.

Capturas: [login](screenshots/login.png), [operações](screenshots/operacoes.png), [mestre/detalhe](screenshots/master-detail.png), [bilhetes](screenshots/bilhetes.png), [mobile](screenshots/mobile.png).

## Limites de verificação

O Docker CLI está instalado, mas o daemon Docker Desktop não estava em execução. Portanto, a imagem Docker não foi construída nem executada nesta máquina. O servidor foi compilado para Linux sem CGO, e os scripts de entrypoint foram normalizados para LF.

Não houve publicação em conta Render, acesso a domínio externo nem teste de carga com múltiplos usuários reais. O blueprint, Dockerfile, disco persistente e variáveis estão preparados para o deploy pelo proprietário.


## Atualização: consultas avançadas

Seis testes de integração adicionais aprovados:

11. Execução das 40 prévias/consultas, unicidade dos IDs e isolamento por companhia.
12. Filtros monetários, confirmação, empates de valores, voos sem escalas e paginação.
13. Bloqueio de outros perfis, parâmetros inválidos, tentativa de SQL livre, CSV e URL de navegação.
14. Cadastro de aeroporto/voo, mudança de tarifa, confirmação, exclusão protegida, sincronização da rota, auditoria e não reutilização de IDs.
15. Bloqueio de alteração de voo alheio e verificação global de referências de aeroporto.
16. Migração de banco antigo, tarifa inicial, visão V1, preservação dos dados e idempotência.

O smoke específico de consultas avançadas passou em Chrome: catálogo 1–40, filtros, navegação HTMX com recarga da página, CSV, cadastros, alteração de tarifa, confirmação, bloqueio de exclusão do último trecho, exclusão após inclusão de outro trecho, remoção de aeroporto sem referências, SQL visível, acesso exclusivo ao gerente e mobile sem transbordamento horizontal.

O smoke original também foi reexecutado em outro banco novo e passou, preservando os fluxos de mestre/detalhe, grade, emissão inline, check-in e isolamento de passageiros.

Capturas: [consultas](screenshots/consultas-avancadas.png), [visão V1](screenshots/consultas-avancadas-v1.png), [mobile](screenshots/consultas-avancadas-mobile.png).
