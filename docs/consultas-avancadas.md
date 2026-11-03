# Consultas avançadas — catálogo SQL

A área está em **Consultas avançadas**, disponível somente ao gerente. Os 40 itens correspondem à lista de exercícios fornecida. Os relatórios têm filtros editáveis, limite de 50 linhas por página e CSV da página. Ações de escrita usam formulários com validação do domínio; não existe editor para executar SQL livre.

## Adaptação ao modelo normalizado

| Referência do exercício | Representação na aplicação |
|---|---|
| VOO.EMPRESA | empresa.nome, via voo.empresa_id |
| VOO.VALOR | voo.valor_centavos / 100.0, em reais |
| AEROPORTO.CIDADE e UF | cidade.nome e estado.sigla, por chaves estrangeiras |
| ID_VOO, ID_ESCALA e ID_AEROPORTO | IDs canônicos e aliases nas CTEs abaixo |
| PLANEJADO | Estado operacional Agendado sem confirmação comercial |
| CONFIRMADO | Estado operacional Agendado com confirmado = 1 |

A consulta de uma empresa diferente da sessão retorna vazio, mesmo quando o nome é alterado no filtro. A empresa de um voo novo é a do gerente; cadastrar um voo “da GOL” exige operar como gerente da companhia correspondente. Os exemplos não renomeiam nem criam empresas automaticamente.

A tarifa é armazenada em centavos para evitar arredondamento na escrita. Filtros e resultados monetários da área são apresentados em reais. A migração inicializa tarifas existentes pela soma das escalas não canceladas e não repete o cálculo. Tarifas do voo e de seus trechos passam a ser independentes.

O enunciado 36 menciona G3111, mas o SQL e os itens seguintes usam G31111. O catálogo adota **G31111** como sugestão e permite informar outro código.

## Base comum dos relatórios

As consultas de leitura abaixo são precedidas por esta CTE. O parâmetro :empresa_id é definido pelo backend a partir da sessão; não pode ser escolhido pelo navegador. Os demais parâmetros correspondem aos filtros de cada item.

```sql
WITH voos AS (
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
```

Uma consulta copiada isoladamente deve incluir a base comum e receber seus parâmetros. O [arquivo SQL de leitura](sql/consultas-leitura.sql) reúne somente os relatórios, cada um com a sua CTE completa.

## Visão V1

SQLite utiliza CREATE VIEW IF NOT EXISTS em vez do CREATE OR REPLACE VIEW do enunciado. A criação ocorre na migração do servidor, e a consulta 34 aplica explicitamente o escopo da companhia.

```sql
CREATE VIEW IF NOT EXISTS V1 AS
 SELECT v.id AS id_voo,v.empresa_id,v.codigo_voo,emp.nome AS empresa,
 CASE WHEN v.situacao='Agendado' THEN CASE WHEN v.confirmado=1 THEN 'CONFIRMADO' ELSE 'PLANEJADO' END ELSE v.situacao END AS situacao,
 v.valor_centavos/100.0 AS valor,e.id AS id_escala,e.sequencial,
 ao.sigla AS origem,ad.sigla AS destino,e.tempo_escala
 FROM voo v JOIN empresa emp ON emp.id=v.empresa_id
 JOIN escala e ON e.voo_id=v.id
 JOIN aeroporto ao ON ao.id=e.origem_id JOIN aeroporto ad ON ad.id=e.destino_id;
```

## Catálogo

### 01. Quantidade de registros por tabela

Tabelas de domínio. Voos, frota, escalas vinculadas e bilhetes respeitam a empresa do gerente; cadastros geográficos, aeroportos e passageiros são compartilhados.

```sql
SELECT 'PAIS' AS TABELA, COUNT(*) AS QUANTIDADE FROM pais
UNION ALL SELECT 'ESTADO', COUNT(*) FROM estado
UNION ALL SELECT 'CIDADE', COUNT(*) FROM cidade
UNION ALL SELECT 'EMPRESA', COUNT(*) FROM empresa WHERE id=:empresa_id
UNION ALL SELECT 'AEROPORTO', COUNT(*) FROM aeroporto
UNION ALL SELECT 'AERONAVE', COUNT(*) FROM aeronave WHERE empresa_id=:empresa_id
UNION ALL SELECT 'VOO', COUNT(*) FROM voos
UNION ALL SELECT 'ESCALA', COUNT(*) FROM escalas
UNION ALL SELECT 'PASSAGEIRO', COUNT(*) FROM passageiro
UNION ALL SELECT 'BILHETE', COUNT(*) FROM bilhete WHERE voo_id IN(SELECT id_voo FROM voos);
```

### 02. Todos os atributos dos voos

```sql
SELECT * FROM voos ORDER BY id_voo;
```

### 03. Códigos, empresas e valores dos voos

```sql
SELECT codigo_voo,empresa,valor FROM voos ORDER BY codigo_voo;
```

### 04. Voos da empresa Azul

Azul é o valor inicial do filtro. Uma empresa diferente da sua sessão não retorna voos.

Filtros iniciais: empresa = Azul.

```sql
SELECT codigo_voo,situacao,valor FROM voos WHERE empresa=:empresa COLLATE NOCASE ORDER BY codigo_voo;
```

### 05. Voos acima de R$ 800,00

Filtros iniciais: valor = 800.00.

```sql
SELECT codigo_voo,empresa,valor FROM voos WHERE valor>:valor ORDER BY valor,codigo_voo;
```

### 06. Voos por valor crescente

```sql
SELECT codigo_voo,empresa,valor FROM voos ORDER BY valor ASC,codigo_voo;
```

### 07. Aeroportos do estado de São Paulo

Filtros iniciais: uf = SP.

```sql
SELECT nome,sigla,cidade,uf FROM aeroportos WHERE uf=:uf COLLATE NOCASE ORDER BY nome;
```

### 08. Empresas cadastradas

```sql
SELECT DISTINCT empresa FROM voos ORDER BY empresa;
```

### 09. Valor médio dos voos

```sql
SELECT ROUND(AVG(valor),2) AS media_valor FROM voos;
```

### 10. Valor médio por empresa

```sql
SELECT empresa,ROUND(AVG(valor),2) AS media_valor FROM voos GROUP BY empresa ORDER BY empresa;
```

### 11. Empresas com média acima de R$ 900,00

Filtros iniciais: valor = 900.00.

```sql
SELECT empresa,ROUND(AVG(valor),2) AS media_valor FROM voos GROUP BY empresa HAVING AVG(valor)>:valor ORDER BY empresa;
```

### 12. Voos e aeroportos de cada escala

```sql
SELECT v.codigo_voo,v.empresa,e.sequencial,ao.sigla AS origem,ad.sigla AS destino FROM voos v JOIN escalas e ON e.id_voo=v.id_voo JOIN aeroportos ao ON ao.id_aeroporto=e.id_aeroporto_origem JOIN aeroportos ad ON ad.id_aeroporto=e.id_aeroporto_destino ORDER BY v.codigo_voo,e.sequencial;
```

### 13. Escalas do voo AD1001

Filtros iniciais: codigo = AD1001.

```sql
SELECT v.codigo_voo,e.sequencial,e.tempo_escala FROM voos v JOIN escalas e ON e.id_voo=v.id_voo WHERE v.codigo_voo=:codigo COLLATE NOCASE ORDER BY e.sequencial;
```

### 14. Quantidade de escalas por voo

Inclui voos sem escala, com quantidade zero.

```sql
SELECT v.codigo_voo,COUNT(e.id_escala) AS total_escalas FROM voos v LEFT JOIN escalas e ON e.id_voo=v.id_voo GROUP BY v.id_voo,v.codigo_voo ORDER BY v.codigo_voo;
```

### 15. Escalas acima de 45 minutos

Filtros iniciais: tempo = 45.

```sql
SELECT v.codigo_voo,v.empresa,e.sequencial,e.tempo_escala FROM voos v JOIN escalas e ON e.id_voo=v.id_voo WHERE e.tempo_escala>:tempo ORDER BY v.codigo_voo,e.sequencial;
```

### 16. Aeroportos de destino das escalas

```sql
SELECT e.id_escala,e.sequencial,a.sigla,a.nome,a.cidade FROM escalas e JOIN aeroportos a ON a.id_aeroporto=e.id_aeroporto_destino ORDER BY e.id_escala;
```

### 17. Todos os atributos dos aeroportos

```sql
SELECT * FROM aeroportos ORDER BY id_aeroporto;
```

### 18. Código, empresa e valor de todos os voos

```sql
SELECT codigo_voo,empresa,valor FROM voos ORDER BY codigo_voo;
```

### 19. Nome, sigla e cidade dos aeroportos de SP

Filtros iniciais: uf = SP.

```sql
SELECT nome,sigla,cidade FROM aeroportos WHERE uf=:uf COLLATE NOCASE ORDER BY nome;
```

### 20. Escalas acima de 120 minutos

Filtros iniciais: tempo = 120.

```sql
SELECT id_voo,sequencial,tempo_escala FROM escalas WHERE tempo_escala>:tempo ORDER BY id_voo,sequencial;
```

### 21. Códigos dos voos confirmados

```sql
SELECT codigo_voo FROM voos WHERE situacao='CONFIRMADO' ORDER BY codigo_voo;
```

### 22. Voos do mais caro para o mais barato

```sql
SELECT codigo_voo,valor FROM voos ORDER BY valor DESC,codigo_voo;
```

### 23. Voos de menor e maior valor

Inclui todos os empates nos valores mínimo e máximo.

```sql
SELECT empresa,codigo_voo,valor FROM voos WHERE valor=(SELECT MIN(valor) FROM voos) OR valor=(SELECT MAX(valor) FROM voos) ORDER BY valor,codigo_voo;
```

### 24. Escalas que partem do Rio de Janeiro

Filtros iniciais: uf = RJ.

```sql
SELECT e.sequencial,a.sigla AS sigla_origem,a.cidade,a.uf FROM escalas e JOIN aeroportos a ON a.id_aeroporto=e.id_aeroporto_origem WHERE a.uf=:uf COLLATE NOCASE ORDER BY e.id_voo,e.sequencial;
```

### 25. Escalas com destino a Recife

Filtros iniciais: cidade = Recife.

```sql
SELECT e.sequencial,a.sigla AS sigla_destino,a.nome,a.cidade FROM escalas e JOIN aeroportos a ON a.id_aeroporto=e.id_aeroporto_destino WHERE a.cidade=:cidade COLLATE NOCASE ORDER BY e.id_voo,e.sequencial;
```

### 26. Origem e destino de cada escala

```sql
SELECT e.id_escala,e.sequencial,ao.sigla AS origem,ao.cidade AS cidade_origem,ad.sigla AS destino,ad.cidade AS cidade_destino FROM escalas e JOIN aeroportos ao ON ao.id_aeroporto=e.id_aeroporto_origem JOIN aeroportos ad ON ad.id_aeroporto=e.id_aeroporto_destino ORDER BY e.id_escala;
```

### 27. Código, empresa e aeroportos de cada trecho

```sql
SELECT v.codigo_voo,v.empresa,e.sequencial,ao.sigla AS origem,ad.sigla AS destino FROM voos v JOIN escalas e ON e.id_voo=v.id_voo JOIN aeroportos ao ON ao.id_aeroporto=e.id_aeroporto_origem JOIN aeroportos ad ON ad.id_aeroporto=e.id_aeroporto_destino ORDER BY v.codigo_voo,e.sequencial;
```

### 28. Voos que passam por aeroportos de São Paulo

Cada voo aparece uma vez, mesmo quando vários trechos passam por SP.

Filtros iniciais: uf = SP.

```sql
SELECT DISTINCT v.codigo_voo,v.empresa,v.valor FROM voos v JOIN escalas e ON e.id_voo=v.id_voo JOIN aeroportos ao ON ao.id_aeroporto=e.id_aeroporto_origem JOIN aeroportos ad ON ad.id_aeroporto=e.id_aeroporto_destino WHERE ao.uf=:uf COLLATE NOCASE OR ad.uf=:uf COLLATE NOCASE ORDER BY v.codigo_voo;
```

### 29. Quantidade de trechos por voo e empresa

```sql
SELECT v.codigo_voo,v.empresa,COUNT(e.id_escala) AS total_trechos FROM voos v LEFT JOIN escalas e ON e.id_voo=v.id_voo GROUP BY v.id_voo,v.codigo_voo,v.empresa ORDER BY v.codigo_voo;
```

### 30. Voos com três ou mais trechos

Filtros iniciais: trechos = 3.

```sql
SELECT v.codigo_voo,v.empresa,COUNT(e.id_escala) AS total_trechos FROM voos v JOIN escalas e ON e.id_voo=v.id_voo GROUP BY v.id_voo,v.codigo_voo,v.empresa HAVING COUNT(e.id_escala)>=:trechos ORDER BY v.codigo_voo;
```

### 31. Tempo médio e quantidade de escalas por voo

```sql
SELECT v.codigo_voo,v.empresa,ROUND(AVG(e.tempo_escala),2) AS tempo_medio,COUNT(e.id_escala) AS total_trechos FROM voos v LEFT JOIN escalas e ON e.id_voo=v.id_voo GROUP BY v.id_voo,v.codigo_voo,v.empresa ORDER BY v.codigo_voo;
```

### 32. Voos acima da média de valor

```sql
SELECT codigo_voo,empresa,valor FROM voos WHERE valor>(SELECT AVG(valor) FROM voos) ORDER BY valor,codigo_voo;
```

### 33. Aeroportos com mais de uma chegada

Filtros iniciais: chegadas = 1.

```sql
SELECT a.sigla,a.nome,a.cidade,COUNT(e.id_escala) AS total_chegadas FROM aeroportos a JOIN escalas e ON e.id_aeroporto_destino=a.id_aeroporto GROUP BY a.id_aeroporto,a.sigla,a.nome,a.cidade HAVING COUNT(e.id_escala)>:chegadas ORDER BY total_chegadas DESC,a.sigla;
```

### 34. Visão V1: voos e seus trechos

A visão V1 é criada pela migração de banco. A consulta aplica a empresa do gerente.

```sql
SELECT codigo_voo,empresa,situacao,valor,sequencial,origem,destino,tempo_escala FROM V1 WHERE empresa_id=:empresa_id ORDER BY codigo_voo,sequencial;
```

### 35. Inserir aeroporto Santos Dumont

A cidade é selecionada no cadastro geográfico; a UF é obtida pelo estado. O aeroporto inicia como Cadastrado.

Filtros iniciais: sigla = SDU.

```sql
INSERT INTO aeroporto(nome,sigla,cidade_id) VALUES(:nome,:sigla,:cidade_id);
```

### 36. Inserir voo G31111

O enunciado alterna G3111 e G31111. Usamos G31111 como sugestão, conforme as operações seguintes. A empresa é a da sua sessão; aeronave, rota e horários são obrigatórios.

Filtros iniciais: codigo = G31111.

```sql
INSERT INTO voo(codigo_voo,empresa_id,aeronave_id,origem_id,destino_id,partida,chegada,valor_centavos) VALUES(:codigo,:empresa_id,:aeronave_id,:origem_id,:destino_id,:partida,:chegada,:valor_centavos);
```

### 37. Alterar o valor do voo G31111

Alteração somente em voo agendado, com controle de versão.

Filtros iniciais: codigo = G31111.

```sql
UPDATE voo SET valor_centavos=:valor_centavos,versao=versao+1 WHERE id=:registro_id AND empresa_id=:empresa_id AND situacao='Agendado' AND versao=:versao;
```

### 38. Confirmar o voo G31111

Confirmação comercial independente do estado operacional: o voo continua Agendado até a partida.

Filtros iniciais: codigo = G31111.

```sql
UPDATE voo SET confirmado=1,versao=versao+1 WHERE id=:registro_id AND empresa_id=:empresa_id AND situacao='Agendado' AND confirmado=0 AND versao=:versao;
```

### 39. Excluir a primeira escala do voo G31111

Exclusão física da escala de sequencial 1, antes da partida, sem bilhetes ativos e mantendo ao menos outro trecho ativo. A operação é auditada.

Filtros iniciais: codigo = G31111.

```sql
DELETE FROM escala WHERE id=:registro_id AND voo_id IN(SELECT id FROM voo WHERE empresa_id=:empresa_id AND situacao='Agendado') AND sequencial=1 AND versao=:versao;
```

O backend confere estado, titularidade da companhia, versão, ausência de bilhetes e existência de outro trecho. A rota do voo é ajustada à sequência restante na mesma transação.

### 40. Aeroportos sem escalas

A lista considera todas as escalas, inclusive de outras empresas. A exclusão é individual, com confirmação, e exige ausência de referências em escalas, voos e bilhetes.

```sql
SELECT a.* FROM aeroportos a WHERE NOT EXISTS(SELECT 1 FROM escala e WHERE e.origem_id=a.id_aeroporto OR e.destino_id=a.id_aeroporto) ORDER BY a.id_aeroporto;
```

A área apresenta os aeroportos sem escala e permite selecionar **um registro por operação**. O backend também impede exclusão com voos, bilhetes ou manutenção referenciados, mesmo de outras empresas.

```sql
DELETE FROM aeroporto WHERE id = :registro_id AND versao = :versao;
```

## Regras das operações de escrita

- 35 e 36 reutilizam as validações de cadastro, com gravação e auditoria transacionais.
- 37 exige voo agendado com partida futura e controle de versão.
- 38 registra confirmação comercial; mantém o estado operacional e não permite confirmar duas vezes.
- 39 e 40 exigem confirmação e justificativa. Uma falha reverte toda a operação, inclusive seus efeitos relacionados.
- Os comandos exibidos representam o núcleo SQL. O endpoint executa também as validações, atualizações relacionadas e auditoria descritas acima.
- Exclusões são físicas nesta área excepcional, preservando uma cópia do registro anterior na auditoria; o CRUD principal continua usando exclusão lógica.
- “Empresas cadastradas”, médias e indicadores de voos consideram a companhia do gerente. Cadastros geográficos e aeroportos são compartilhados.

Foram corrigidos os DELETE sem FROM, as entidades HTML dos operadores e aspas, a contagem para voos sem escalas e a duplicação de voos que passam por mais de um aeroporto de SP. O item 23 usa uma única consulta e mantém os empates de mínimo/máximo.

Referências técnicas: [CREATE VIEW no SQLite](https://www.sqlite.org/lang_createview.html) e [ALTER TABLE no SQLite](https://www.sqlite.org/lang_altertable.html).
