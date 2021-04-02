package innotop

import (
	"database/sql"

	"github.com/lefred/innotopgo/db"
)

func Processlist(mydb *sql.DB, displaytype string) error {
	stmt := `select pps.PROCESSLIST_COMMAND AS command,
                                  pps.THREAD_ID AS thd_id, pps.PROCESSLIST_ID AS conn_id,
                                  conattr_pid.ATTR_VALUE AS pid, pps.PROCESSLIST_STATE AS state,
                                  if((pps.NAME in ('thread/sql/one_connection','thread/thread_pool/tp_one_connection')),
                                   concat(pps.PROCESSLIST_USER,'@',pps.PROCESSLIST_HOST),
                                   replace(pps.NAME,'thread/','')) AS user,
                                  pps.PROCESSLIST_DB AS db, sys.format_statement(pps.PROCESSLIST_INFO) AS current_statement,
                                  if(isnull(esc.END_EVENT_ID), format_pico_time(esc.TIMER_WAIT),NULL) AS statement_latency,
                                  format_pico_time(esc.LOCK_TIME) AS lock_latency,
                                  if(isnull(esc.END_EVENT_ID),esc.TIMER_WAIT,0) AS sort_time
                            from (((((((performance_schema.threads pps
                            left join performance_schema.events_waits_current ewc
                                on((pps.THREAD_ID = ewc.THREAD_ID)))
                            left join performance_schema.events_stages_current estc
                                on((pps.THREAD_ID = estc.THREAD_ID)))
                            left join performance_schema.events_statements_current esc
                                on((pps.THREAD_ID = esc.THREAD_ID)))
                            left join performance_schema.events_transactions_current etc
                                on((pps.THREAD_ID = etc.THREAD_ID)))
                            left join sys.x$memory_by_thread_by_current_bytes mem
                                on((pps.THREAD_ID = mem.thread_id)))
                            left join performance_schema.session_connect_attrs conattr_pid
                                on(((conattr_pid.PROCESSLIST_ID = pps.PROCESSLIST_ID) and (conattr_pid.ATTR_NAME = '_pid'))))
                            left join performance_schema.session_connect_attrs conattr_progname
                                on(((conattr_progname.PROCESSLIST_ID = pps.PROCESSLIST_ID)
                                and (conattr_progname.ATTR_NAME = 'program_name'))))
                            where pps.PROCESSLIST_ID is not null
                              and pps.PROCESSLIST_COMMAND <> 'Daemon'
                              and user <> 'sql/event_scheduler'
                            order by sort_time desc
                        `
	rows := db.Query(mydb, stmt)
	cols, data, err := db.GetData(rows)
	if err != nil {
		panic(err)
	}
	if displaytype == "simple" {
		DisplaySimple(cols, data)
	} else {
		DisplayProcesslist(cols, data)
	}

	return err
}
