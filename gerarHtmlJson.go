package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

type TaleHtml struct {
	Nome string
	Html string
}

func main() {

	flagfile := flag.String("file", "help\n gerarHtmlJson --file table.html", "html")
	flag.Parse()

	if *flagfile != "help\n gerarHtmlJson --file table.html" {

		nomeFile := *flagfile
		pathfile := "./" + nomeFile

		// abrir o arquivo e pegar o conteudo
		dat, err := ioutil.ReadFile(pathfile)
		check(err)

		// html
		html := string(dat)

		// criando estrutura
		m := TaleHtml{"meu_primeiro_pdf.pdf", html}

		// convertendo json
		b, err := json.Marshal(m)
		check(err)

		pathfile = "./" + nomeFile + ".json"
		// gravara no arquivo nosso json
		err = ioutil.WriteFile(pathfile, b, 0644)
		check(err)
	}
}

func check(e error) {
	if e != nil {
		log.Println(e)
	}
}
