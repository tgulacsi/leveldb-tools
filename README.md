# Purpose
`leveldb-tools` provides a simple way to dump your LevelDB (esp.
[goleveldb](https://github.com/syndtr/goleveldb)
) database in a general, easily parsable
[cdbmake](http://cr.yp.to/cdb/cdbmake.html)
format.

It also provides a way to load from such formatted input.

# Usage
## Dump LevelDB database

	leveldb-tools dump my.leveldb >my.cdbmake

## Load from cdbmake-like source

	leveldb-tools load my.leveldb <my.cdbmake

For example, cznic's
[kv](https://github.com/cznic/kv)
database can be dumped in cdbmake format:

	kvaudit -d my.kvdb | leveldb-tools load my.leveldb
