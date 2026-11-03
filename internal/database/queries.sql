-- name: GetUserByEmail :one
SELECT * FROM usuario WHERE email=?;
-- name: GetSessionUser :one
SELECT usuario.* FROM usuario JOIN sessao ON sessao.usuario_id=usuario.id WHERE sessao.token_hash=? AND sessao.expira_em>?;
-- name: CreateSession :exec
INSERT INTO sessao(token_hash,usuario_id,expira_em) VALUES(?,?,?);
-- name: DeleteSession :exec
DELETE FROM sessao WHERE token_hash=?;
-- name: DepartFlights :exec
UPDATE voo SET situacao='Em Voo',versao=versao+1 WHERE situacao='Agendado' AND partida<=?;
-- name: CompleteFlights :exec
UPDATE voo SET situacao=('Conclu'||char(237)||'do'),versao=versao+1 WHERE situacao='Em Voo' AND chegada<=?;
-- name: StartStops :exec
UPDATE escala SET situacao='Em Andamento',versao=versao+1 WHERE situacao='Programada' AND voo_id IN(SELECT id FROM voo WHERE situacao='Em Voo');
-- name: CompleteStops :exec
UPDATE escala SET situacao=('Conclu'||char(237)||'da'),versao=versao+1 WHERE situacao IN('Programada','Em Andamento') AND voo_id IN(SELECT id FROM voo WHERE situacao=('Conclu'||char(237)||'do'));
-- name: CompleteTickets :exec
UPDATE bilhete SET situacao='Usado',versao=versao+1 WHERE situacao IN('Emitido','Check-in Realizado','Embarque Realizado') AND voo_id IN(SELECT id FROM voo WHERE situacao=('Conclu'||char(237)||'do'));
-- name: RecordAudit :exec
INSERT INTO auditoria(usuario_id,entidade,registro_id,acao,motivo) VALUES(?,?,?,?,?);
