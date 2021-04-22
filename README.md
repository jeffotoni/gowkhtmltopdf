# gowkhtmltopdf
Uma image docker com o programa wkhtmltopdf e um server em Go para converter html em pdf

A API irá receber um HTML e retornar um PDF, o html terá que ser passado como json para API.

# Rodando API com Docker

```sh

$ docker run --restart=always -d -p 5010:5010 --name gowkhtmltopdf \
	jeffotoni/gowkhtmltopdf:latest

$ docker logs -f <id-container>

$ curl -X POST localhost:5010/v1/api/topdf -H "Content-Type: application/json" \
--data @table.html.json --output /tmp/meuteste.pdf

```

# Build em sua máquina local

```bash

$ docker build --no-cache -f DockerfileAlpine --build-arg PORT=5010 \
	-t xxxxxxxxxxxxx/gowkhtmltopdf:latest .

$ docker run --restart=always -d -p 5010:5010 --name gowkhtmltopdf \
	jeffotoni/gowkhtmltopdf:latest

// -- or

$ docker run -p 5010:5010 --name gowkhtmltopdf -e X_KEY=xxxxxx \
	jeffotoni/gohtmltopdf

$ docker logs -f <id-container>

$ curl -X POST localhost:5010/v1/api/topdf -H "Content-Type: application/json" \
	 -H "Authorization:Basic xxxxxx" \
	--data @table.html.json --output /tmp/meuteste.pdf

```

# Gerando Json do seu HTML antes de enviar para API

A API recebe um JSON, o nome do arquivo e o html que deseja converter e retorna o PDF.
Para gerar seu HTML em JSON só rodar o programinha abaixo.

Campos permitidos
```json
{
	"Nome":"meu_pdf",
	"Html":"<base 64 do HTML aqui>",
	"grayscale":false,
	"nocollate":false,
	"image_dpi":600,
	"image_quality":94,
	"page_size":"A4",
	"orientation":"Portrait",
	"dpi":600,
	"margin_bottom":2,
	"margin_top":2,
	"margin_left":2,
	"margin_right":2
}
```

Este programa irá converter seu HTML em JSON e os campos necessários para geração do PDF
```sh

$ cd gerahtmljson
$ go run main.go --file table.html

```

# Rodando o server sem usar docker

```sh

$ go run gowkhtmltopdf.go

$ curl -X POST localhost:5010/v1/api/topdf -H "Content-Type: application/json" \
	--data @table.html.json --output /tmp/meuteste.pdf

```

