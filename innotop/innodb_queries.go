package innotop

import (
	"database/sql"
	"fmt"

	"github.com/lefred/innotopgo/db"
)

func GetInnoDBStatus(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `select variable_name, variable_value 
	         from performance_schema.global_status 
			 where variable_name like 'innodb_%' or variable_name = 'Uptime'`
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
	stmt := `SELECT ROUND(A.num * 100.0 / B.num)  BufferPoolFull, BP_Size, BP_instances,
					FORMAT(F.num * 100.0 / E.num,2) DiskReadRatio, 
					ROUND(F.num*100/E.num) DiskReadRatioInt
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
					WHERE variable_name = 'innodb_buffer_pool_instances') D,
				  (     
    				SELECT variable_value num 
					FROM performance_schema.global_status       
    				WHERE variable_name = 'Innodb_buffer_pool_read_requests') E,       
    			  (
    				SELECT variable_value num 
					FROM performance_schema.global_status        
    				WHERE variable_name = 'Innodb_buffer_pool_reads'
				  ) F
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

func GetRedoCapacity(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `SELECT 
                                       format_bytes( (
                                                SELECT VARIABLE_VALUE
                                                FROM performance_schema.global_variables
                                                WHERE variable_name = 'innodb_redo_log_capacity2')
                                        ) AS InnoDBLogCapacity,
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
                                                FROM performance_schema.global_variables
                                                WHERE variable_name = 'innodb_redo_log_capacity'
				        ) AS InnoDBLogCapacityRaw,
                                        (
						SELECT VARIABLE_VALUE
                                                FROM performance_schema.global_variables
                                                WHERE variable_name = 'innodb_log_file_size'
                                        ) AS InnoDBLogFileSizeRaw;
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

func GetRedoInfo(mydb *sql.DB, innodb_redo_log_capacity int) ([]string, [][]string, error) {
	stmt := fmt.Sprintf(`SELECT CONCAT(
		            (
						SELECT FORMAT_BYTES(
							STORAGE_ENGINES->>'$."InnoDB"."LSN"' - STORAGE_ENGINES->>'$."InnoDB"."LSN_checkpoint"'
							               )
								FROM performance_schema.log_status),
						" / ",
						format_bytes(%d)
					) CheckpointInfo,
					(
						SELECT ROUND(((
							SELECT STORAGE_ENGINES->>'$."InnoDB"."LSN"' - STORAGE_ENGINES->>'$."InnoDB"."LSN_checkpoint"'
							FROM performance_schema.log_status)) / ((
								SELECT VARIABLE_VALUE
								FROM performance_schema.global_variables
								WHERE VARIABLE_NAME = 'innodb_log_file_size'
							) * %d * 100),2)
					)  AS CheckpointAge,
					(
						SELECT ROUND(((
							SELECT STORAGE_ENGINES->>'$."InnoDB"."LSN"' - STORAGE_ENGINES->>'$."InnoDB"."LSN_checkpoint"'
							FROM performance_schema.log_status) / (
							%d) * 100))
					)  AS CheckpointAgeInt,
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
	`, innodb_redo_log_capacity, innodb_redo_log_capacity, innodb_redo_log_capacity)

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
