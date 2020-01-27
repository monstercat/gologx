package server

import (
	"github.com/jmoiron/sqlx"

	"github.com/monstercat/logx"
)

func InsertHostMessage(db sqlx.Ext, msg logx.HostMessage, service string ) error {
	query := `
INSERT INTO log(service_id, log_type, log_time, message, context)
VALUES($1,$2,$3,$4,$5,$6);
`
	_, err := db.Exec(query, service, msg.Type, msg.Time, msg.Message, msg.Context)
	return err
}
