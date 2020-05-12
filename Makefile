
all:
	go build

test1:
	go build
	pongo2cli  --data data.json ex1.tmpl ex1.tmpl

test2:
	go build
	pongo2cli  --debug cache,sleep --data data.json ex1.tmpl ex1.tmpl ex1.tmpl

test3:
	go build
	pongo2cli  --urlpath --debug cache --data data.json 'index.html?page_name=ex1.tmpl'

#var optDbConn = flag.String("conn", "", "Database (PostgreSQL) connection string.")
#var optDbName = flag.String("dbname", "", "Database (PostgreSQL) name.")
#var optQuery = flag.String("sql", "", "Database (PostgreSQL) select to get data.")
#var optUseSubData = flag.Bool("sub-data", false, "use .data as a field for array of data.")
test4:
	go build
	pongo2cli  --urlpath --debug "cache,query"  \
		--conn "user=pschlump dbname=q8s port=5432 host=127.0.0.1 sslmode=disable" \
		--dbname "q8s" \
		--sql "select * from t_ymux_documents" \
		--sub-data \
		'index.html?page_name=ex4.tmpl'
	
