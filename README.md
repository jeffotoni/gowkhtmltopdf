# gowkhtmltopdf
Uma image docker com o programa wkhtmltopdf e um server em Go para converter html em pdf

A API irá receber um HTML e retornar um PDF, o html terá que ser passado como json para API.

# Rodando API com Docker

```sh

$ docker run --restart=always -d -p 5010:5010 --name gowkhtmltopdf jeffotoni/gowkhtmltopdf:latest

$ docker logs -f <id-container>

$ curl -X POST localhost:5010/v1/api/topdf -H "Content-Type: application/json" --data @table.html.json --output /tmp/meuteste.pdf

```