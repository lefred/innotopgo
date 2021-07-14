# innotopgo

Innotop for MySQL 8 written in Go

## 0.3.0-dev 2021-07-14 (not yet released)
- adding Error Log Dashboard <E>

## 0.2.0 2021-04-10
- adding InnoDB Info Dashboard <I>
- adding Memory Info Dashboard <M>
- adding warning in EXPLAIN ANALYZE
- splitting EXPLAIN ANALYZE: <a> with timeout of 5min, <A> no timeout
- better handling of MySQL disconnection
- possibility to make a debug build

## 0.1.1 2021-04-06
- missing <K> option in help
- remove panic in kill.go
- display the error message when a thread_id is not available to be killed
- simplify label prints in details

## 0.1.0 2021-04-06
- Initial Release
