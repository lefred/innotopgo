package innotop

import (
	"database/sql"

	"github.com/lefred/innotopgo/db"
)

func GetTempMem(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `SELECT FORMAT(B.num * 100.0 / A.num,2) AS TempTablesDiskRatio, 
       ROUND(B.num * 100/A.num) AS TempTablesDiskRatioInt,
       B.num As TempTablesDisk,  A.num As TempTables, 
       C.total_allocated AS TotalAllocated,
	   E.total_allocated AS TotalAllocatedNum,
       Uptime
      FROM (     
        SELECT variable_value num FROM performance_schema.global_status
         WHERE variable_name = 'Created_tmp_tables') A,
         (SELECT variable_value num FROM performance_schema.global_status
           WHERE variable_name = 'Created_tmp_disk_tables') B,
         (SELECT total_allocated FROM sys.memory_global_total) C,
         (SELECT variable_value Uptime FROM performance_schema.global_status
         WHERE variable_name = 'Uptime') D,
		 (SELECT total_allocated FROM sys.x$memory_global_total) E`

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

func GetTempAlloc(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `SELECT event_name, current_alloc, high_alloc 
	         FROM sys.memory_global_by_current_bytes 
			 WHERE event_name LIKE 'memory/temptable/physical_ram'`

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

func GetUserMemAlloc(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `SELECT user, current_allocated, current_max_alloc  
	        FROM sys.memory_by_user_by_current_bytes 
			WHERE user != "background"`

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

func GetCodeMemAlloc(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `SELECT SUBSTRING_INDEX(event_name,'/',2) AS code_area,  
       				format_bytes(SUM(current_alloc)) AS current_alloc,
					sum(current_alloc) current_alloc_num  
       		 FROM sys.x$memory_global_by_current_bytes  
       		 GROUP BY SUBSTRING_INDEX(event_name,'/',2)  
       		 ORDER BY SUM(current_alloc) DESC`

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
