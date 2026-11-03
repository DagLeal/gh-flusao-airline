# Validação da entrega

Ambiente: Windows amd64, Go 1.27.3, Node 22+, Chrome instalado. Builds finais para Windows e Linux amd64.

## Verificações concluídas

| Verificação | Resultado |
|---|---|
| go test ./... -count=1 | 10 testes de integração aprovados |
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
