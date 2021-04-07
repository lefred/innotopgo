package innotop

import (
	"database/sql"

	"github.com/lefred/innotopgo/db"
)

func GetAHI(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `SELECT ROUND(
            (
              SELECT Variable_value FROM sys.metrics
              WHERE Variable_name = 'adaptive_hash_searches'
            ) /
            (
              (
               SELECT Variable_value FROM sys.metrics
               WHERE Variable_name = 'adaptive_hash_searches_btree'
              )  + (
               SELECT Variable_value FROM sys.metrics
               WHERE Variable_name = 'adaptive_hash_searches'
              )
            ) * 100,2
          ) 'AHIRatio',
		  ROUND(
            (
              SELECT Variable_value FROM sys.metrics
              WHERE Variable_name = 'adaptive_hash_searches'
            ) /
            (
              (
               SELECT Variable_value FROM sys.metrics
               WHERE Variable_name = 'adaptive_hash_searches_btree'
              )  + (
               SELECT Variable_value FROM sys.metrics
               WHERE Variable_name = 'adaptive_hash_searches'
              )
            ) * 100
          ) 'AHIRatioInt',
		  (
					SELECT variable_value
					FROM performance_schema.global_variables
					WHERE variable_name = 'innodb_adaptive_hash_index'
		  ) AHIEnabled,
		  (
					SELECT variable_value
					FROM performance_schema.global_variables
					WHERE variable_name = 'innodb_adaptive_hash_index_parts'
		  ) AHIParts
	`
	rows, err := db.Query(mydb, stmt)
	if err != nil {
		return nil, nil, err
	}
	cols, data, err := db.GetData(rows)
	if err != nil {
		return nil, nil, err
	}
	return cols, data, err
}

func GetBPFill(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `SELECT ROUND(A.num * 100.0 / B.num)  BufferPoolFull, BP_Size, BP_instances
	         FROM (
				    SELECT variable_value num FROM performance_schema.global_status
					WHERE variable_name = 'Innodb_buffer_pool_pages_data') A,
				  (
					SELECT variable_value num FROM performance_schema.global_status
					WHERE variable_name = 'Innodb_buffer_pool_pages_total') B,
				  (
					SELECT format_bytes(variable_value) as BP_Size
					FROM performance_schema.global_variables
					WHERE variable_name = 'innodb_buffer_pool_size') C,
				  (
					SELECT variable_value as BP_instances
					FROM performance_schema.global_variables
					WHERE variable_name = 'innodb_buffer_pool_instances') D
	`
	rows, err := db.Query(mydb, stmt)
	if err != nil {
		return nil, nil, err
	}
	cols, data, err := db.GetData(rows)
	if err != nil {
		return nil, nil, err
	}
	return cols, data, err
}

func GetRedoInfo(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `SELECT CONCAT(
		            (
						SELECT FORMAT_BYTES(
							STORAGE_ENGINES->>'$."InnoDB"."LSN"' - STORAGE_ENGINES->>'$."InnoDB"."LSN_checkpoint"'
							               )
								FROM performance_schema.log_status),
						" / ",
						format_bytes(
							(SELECT VARIABLE_VALUE
								FROM performance_schema.global_variables
								WHERE VARIABLE_NAME = 'innodb_log_file_size'
							)  * (
							 SELECT VARIABLE_VALUE
							 FROM performance_schema.global_variables
							 WHERE VARIABLE_NAME = 'innodb_log_files_in_group'))
					) CheckpointInfo,
					(
						SELECT ROUND(((
							SELECT STORAGE_ENGINES->>'$."InnoDB"."LSN"' - STORAGE_ENGINES->>'$."InnoDB"."LSN_checkpoint"'
							FROM performance_schema.log_status) / ((
								SELECT VARIABLE_VALUE
								FROM performance_schema.global_variables
								WHERE VARIABLE_NAME = 'innodb_log_file_size'
							) * (
							SELECT VARIABLE_VALUE
							FROM performance_schema.global_variables
							WHERE VARIABLE_NAME = 'innodb_log_files_in_group')) * 100),2)
					)  AS CheckpointAge,
					(
						SELECT ROUND(((
							SELECT STORAGE_ENGINES->>'$."InnoDB"."LSN"' - STORAGE_ENGINES->>'$."InnoDB"."LSN_checkpoint"'
							FROM performance_schema.log_status) / ((
								SELECT VARIABLE_VALUE
								FROM performance_schema.global_variables
								WHERE VARIABLE_NAME = 'innodb_log_file_size'
							) * (
							SELECT VARIABLE_VALUE
							FROM performance_schema.global_variables
							WHERE VARIABLE_NAME = 'innodb_log_files_in_group')) * 100))
					)  AS CheckpointAgeInt,
					format_bytes( (
						SELECT VARIABLE_VALUE
						FROM performance_schema.global_variables
						WHERE variable_name = 'innodb_log_file_size')
					) AS InnoDBLogFileSize,
					(
						SELECT VARIABLE_VALUE
						FROM performance_schema.global_variables
						WHERE variable_name = 'innodb_log_files_in_group'
					) AS NbFiles,
					(
						SELECT VARIABLE_VALUE
						FROM performance_schema.global_status
						WHERE VARIABLE_NAME = 'Innodb_redo_log_enabled'
					) AS RedoEnabled,
					(
						SELECT VARIABLE_VALUE
						FROM performance_schema.global_status
						WHERE VARIABLE_NAME = 'Uptime'
					) AS Uptime
	`
	rows, err := db.Query(mydb, stmt)
	if err != nil {
		return nil, nil, err
	}
	cols, data, err := db.GetData(rows)
	if err != nil {
		return nil, nil, err
	}
	return cols, data, err
}
