-- Relatórios de leitura. Vincule os parâmetros antes de executar.
-- :empresa_id é a companhia autorizada; os demais parâmetros estão descritos no catálogo.

-- 01. Quantidade de registros por tabela
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

-- 02. Todos os atributos dos voos
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
SELECT * FROM voos ORDER BY id_voo;

-- 03. Códigos, empresas e valores dos voos
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
SELECT codigo_voo,empresa,valor FROM voos ORDER BY codigo_voo;

-- 04. Voos da empresa Azul
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
SELECT codigo_voo,situacao,valor FROM voos WHERE empresa=:empresa COLLATE NOCASE ORDER BY codigo_voo;

-- 05. Voos acima de R$ 800,00
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
SELECT codigo_voo,empresa,valor FROM voos WHERE valor>:valor ORDER BY valor,codigo_voo;

-- 06. Voos por valor crescente
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
SELECT codigo_voo,empresa,valor FROM voos ORDER BY valor ASC,codigo_voo;

-- 07. Aeroportos do estado de São Paulo
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
SELECT nome,sigla,cidade,uf FROM aeroportos WHERE uf=:uf COLLATE NOCASE ORDER BY nome;

-- 08. Empresas cadastradas
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
SELECT DISTINCT empresa FROM voos ORDER BY empresa;

-- 09. Valor médio dos voos
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
SELECT ROUND(AVG(valor),2) AS media_valor FROM voos;

-- 10. Valor médio por empresa
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
SELECT empresa,ROUND(AVG(valor),2) AS media_valor FROM voos GROUP BY empresa ORDER BY empresa;

-- 11. Empresas com média acima de R$ 900,00
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
SELECT empresa,ROUND(AVG(valor),2) AS media_valor FROM voos GROUP BY empresa HAVING AVG(valor)>:valor ORDER BY empresa;

-- 12. Voos e aeroportos de cada escala
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
SELECT v.codigo_voo,v.empresa,e.sequencial,ao.sigla AS origem,ad.sigla AS destino FROM voos v JOIN escalas e ON e.id_voo=v.id_voo JOIN aeroportos ao ON ao.id_aeroporto=e.id_aeroporto_origem JOIN aeroportos ad ON ad.id_aeroporto=e.id_aeroporto_destino ORDER BY v.codigo_voo,e.sequencial;

-- 13. Escalas do voo AD1001
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
SELECT v.codigo_voo,e.sequencial,e.tempo_escala FROM voos v JOIN escalas e ON e.id_voo=v.id_voo WHERE v.codigo_voo=:codigo COLLATE NOCASE ORDER BY e.sequencial;

-- 14. Quantidade de escalas por voo
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
SELECT v.codigo_voo,COUNT(e.id_escala) AS total_escalas FROM voos v LEFT JOIN escalas e ON e.id_voo=v.id_voo GROUP BY v.id_voo,v.codigo_voo ORDER BY v.codigo_voo;

-- 15. Escalas acima de 45 minutos
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
SELECT v.codigo_voo,v.empresa,e.sequencial,e.tempo_escala FROM voos v JOIN escalas e ON e.id_voo=v.id_voo WHERE e.tempo_escala>:tempo ORDER BY v.codigo_voo,e.sequencial;

-- 16. Aeroportos de destino das escalas
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
SELECT e.id_escala,e.sequencial,a.sigla,a.nome,a.cidade FROM escalas e JOIN aeroportos a ON a.id_aeroporto=e.id_aeroporto_destino ORDER BY e.id_escala;

-- 17. Todos os atributos dos aeroportos
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
SELECT * FROM aeroportos ORDER BY id_aeroporto;

-- 18. Código, empresa e valor de todos os voos
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
SELECT codigo_voo,empresa,valor FROM voos ORDER BY codigo_voo;

-- 19. Nome, sigla e cidade dos aeroportos de SP
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
SELECT nome,sigla,cidade FROM aeroportos WHERE uf=:uf COLLATE NOCASE ORDER BY nome;

-- 20. Escalas acima de 120 minutos
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
SELECT id_voo,sequencial,tempo_escala FROM escalas WHERE tempo_escala>:tempo ORDER BY id_voo,sequencial;

-- 21. Códigos dos voos confirmados
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
SELECT codigo_voo FROM voos WHERE situacao='CONFIRMADO' ORDER BY codigo_voo;

-- 22. Voos do mais caro para o mais barato
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
SELECT codigo_voo,valor FROM voos ORDER BY valor DESC,codigo_voo;

-- 23. Voos de menor e maior valor
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
SELECT empresa,codigo_voo,valor FROM voos WHERE valor=(SELECT MIN(valor) FROM voos) OR valor=(SELECT MAX(valor) FROM voos) ORDER BY valor,codigo_voo;

-- 24. Escalas que partem do Rio de Janeiro
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
SELECT e.sequencial,a.sigla AS sigla_origem,a.cidade,a.uf FROM escalas e JOIN aeroportos a ON a.id_aeroporto=e.id_aeroporto_origem WHERE a.uf=:uf COLLATE NOCASE ORDER BY e.id_voo,e.sequencial;

-- 25. Escalas com destino a Recife
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
SELECT e.sequencial,a.sigla AS sigla_destino,a.nome,a.cidade FROM escalas e JOIN aeroportos a ON a.id_aeroporto=e.id_aeroporto_destino WHERE a.cidade=:cidade COLLATE NOCASE ORDER BY e.id_voo,e.sequencial;

-- 26. Origem e destino de cada escala
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
SELECT e.id_escala,e.sequencial,ao.sigla AS origem,ao.cidade AS cidade_origem,ad.sigla AS destino,ad.cidade AS cidade_destino FROM escalas e JOIN aeroportos ao ON ao.id_aeroporto=e.id_aeroporto_origem JOIN aeroportos ad ON ad.id_aeroporto=e.id_aeroporto_destino ORDER BY e.id_escala;

-- 27. Código, empresa e aeroportos de cada trecho
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
SELECT v.codigo_voo,v.empresa,e.sequencial,ao.sigla AS origem,ad.sigla AS destino FROM voos v JOIN escalas e ON e.id_voo=v.id_voo JOIN aeroportos ao ON ao.id_aeroporto=e.id_aeroporto_origem JOIN aeroportos ad ON ad.id_aeroporto=e.id_aeroporto_destino ORDER BY v.codigo_voo,e.sequencial;

-- 28. Voos que passam por aeroportos de São Paulo
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
SELECT DISTINCT v.codigo_voo,v.empresa,v.valor FROM voos v JOIN escalas e ON e.id_voo=v.id_voo JOIN aeroportos ao ON ao.id_aeroporto=e.id_aeroporto_origem JOIN aeroportos ad ON ad.id_aeroporto=e.id_aeroporto_destino WHERE ao.uf=:uf COLLATE NOCASE OR ad.uf=:uf COLLATE NOCASE ORDER BY v.codigo_voo;

-- 29. Quantidade de trechos por voo e empresa
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
SELECT v.codigo_voo,v.empresa,COUNT(e.id_escala) AS total_trechos FROM voos v LEFT JOIN escalas e ON e.id_voo=v.id_voo GROUP BY v.id_voo,v.codigo_voo,v.empresa ORDER BY v.codigo_voo;

-- 30. Voos com três ou mais trechos
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
SELECT v.codigo_voo,v.empresa,COUNT(e.id_escala) AS total_trechos FROM voos v JOIN escalas e ON e.id_voo=v.id_voo GROUP BY v.id_voo,v.codigo_voo,v.empresa HAVING COUNT(e.id_escala)>=:trechos ORDER BY v.codigo_voo;

-- 31. Tempo médio e quantidade de escalas por voo
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
SELECT v.codigo_voo,v.empresa,ROUND(AVG(e.tempo_escala),2) AS tempo_medio,COUNT(e.id_escala) AS total_trechos FROM voos v LEFT JOIN escalas e ON e.id_voo=v.id_voo GROUP BY v.id_voo,v.codigo_voo,v.empresa ORDER BY v.codigo_voo;

-- 32. Voos acima da média de valor
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
SELECT codigo_voo,empresa,valor FROM voos WHERE valor>(SELECT AVG(valor) FROM voos) ORDER BY valor,codigo_voo;

-- 33. Aeroportos com mais de uma chegada
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
SELECT a.sigla,a.nome,a.cidade,COUNT(e.id_escala) AS total_chegadas FROM aeroportos a JOIN escalas e ON e.id_aeroporto_destino=a.id_aeroporto GROUP BY a.id_aeroporto,a.sigla,a.nome,a.cidade HAVING COUNT(e.id_escala)>:chegadas ORDER BY total_chegadas DESC,a.sigla;

-- 34. Visão V1: voos e seus trechos
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
SELECT codigo_voo,empresa,situacao,valor,sequencial,origem,destino,tempo_escala FROM V1 WHERE empresa_id=:empresa_id ORDER BY codigo_voo,sequencial;

-- 40. Aeroportos sem escalas
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
SELECT a.* FROM aeroportos a WHERE NOT EXISTS(SELECT 1 FROM escala e WHERE e.origem_id=a.id_aeroporto OR e.destino_id=a.id_aeroporto) ORDER BY a.id_aeroporto;
