package app

import (
	"context"
	"database/sql"
)

// migrateAdvanced upgrades existing databases without replacing operational rows.
func migrateAdvanced(db *sql.DB) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rs, err := tx.QueryContext(ctx, "SELECT name FROM pragma_table_info('voo')")
	if err != nil {
		return err
	}
	columns := map[string]bool{}
	for rs.Next() {
		var name string
		if err = rs.Scan(&name); err != nil {
			rs.Close()
			return err
		}
		columns[name] = true
	}
	err = rs.Err()
	rs.Close()
	if err != nil {
		return err
	}
	if !columns["valor_centavos"] {
		if _, err = tx.ExecContext(ctx, "ALTER TABLE voo ADD COLUMN valor_centavos INTEGER NOT NULL DEFAULT 0 CHECK(valor_centavos>=0)"); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, "UPDATE voo SET valor_centavos=coalesce((SELECT sum(e.valor_centavos) FROM escala e WHERE e.voo_id=voo.id AND e.situacao<>'Cancelada'),0),versao=versao+1"); err != nil {
			return err
		}
	}
	if !columns["confirmado"] {
		if _, err = tx.ExecContext(ctx, "ALTER TABLE voo ADD COLUMN confirmado INTEGER NOT NULL DEFAULT 0 CHECK(confirmado IN(0,1))"); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `CREATE VIEW IF NOT EXISTS V1 AS
 SELECT v.id AS id_voo,v.empresa_id,v.codigo_voo,emp.nome AS empresa,
 CASE WHEN v.situacao='Agendado' THEN CASE WHEN v.confirmado=1 THEN 'CONFIRMADO' ELSE 'PLANEJADO' END ELSE v.situacao END AS situacao,
 v.valor_centavos/100.0 AS valor,e.id AS id_escala,e.sequencial,
 ao.sigla AS origem,ad.sigla AS destino,e.tempo_escala
 FROM voo v JOIN empresa emp ON emp.id=v.empresa_id
 JOIN escala e ON e.voo_id=v.id
 JOIN aeroporto ao ON ao.id=e.origem_id JOIN aeroporto ad ON ad.id=e.destino_id`)
	if err != nil {
		return err
	}
	return tx.Commit()
}
