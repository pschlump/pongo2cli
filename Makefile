
all:
	go build

test1:
	go build
	t1-pongo2 --data data.json ex1.tmpl ex1.tmpl

test2:
	go build
	t1-pongo2 --debug cache --data data.json ex1.tmpl ex1.tmpl ex1.tmpl
