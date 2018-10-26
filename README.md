# gowkhtmltopdf
Uma image docker com o programa wkhtmltopdf e um server em Go para converter html em pdf

A API irá receber um HTML e retornar um PDF, o html terá que ser passado como json para API.

# Rodando API com Docker

```sh

$ docker run --restart=always -d -p 5010:5010 --name gowkhtmltopdf jeffotoni/gowkhtmltopdf:latest

$ docker logs -f <id-container>

$ curl -X POST localhost:5010/v1/api/topdf -H "Content-Type: application/json" --data @table.html.json --output /tmp/meuteste.pdf

```

# Build em sua máquina local

```sh

$ docker build --no-cache -f DockerfileAlpine --build-arg PORT=5010 -t xxxxxxxxxxxxx/gowkhtmltopdf:latest .

$ docker run --restart=always -d -p 5010:5010 --name gowkhtmltopdf jeffotoni/gowkhtmltopdf:latest

$ docker logs -f <id-container>

$ curl -X POST localhost:5010/v1/api/topdf -H "Content-Type: application/json" --data @table.html.json --output /tmp/meuteste.pdf

```

# Rodando o server sem usar docker

```sh

$ go run gowkhtmltopdf.go

$ curl -X POST localhost:5010/v1/api/topdf -H "Content-Type: application/json" --data @table.html.json --output /tmp/meuteste.pdf

```

# Gerando Json do seu HTML antes de enviar para API

A API recebe um JSON, o nome do arquivo e o html que deseja converter e retorna o PDF.
Para gerar seu HTML em JSON só rodar o programinha abaixo.

```sh

$ go run gerarHtmlJson.go --file table.html

```
