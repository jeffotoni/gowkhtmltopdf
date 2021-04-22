package main

import (
	"encoding/base64"
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
		pathfile := nomeFile

		// abrir o arquivo e pegar o conteudo
		dat, err := ioutil.ReadFile(pathfile)
		if err != nil {
			log.Println("log:", err)
			return
		}

		// html
		html := string(dat)

		htmlEnc := Encode64String(html)
		// criando estrutura
		m := TaleHtml{"meu_primeiro_pdf.pdf", htmlEnc}

		// convertendo json
		b, err := json.Marshal(m)
		if err != nil {
			log.Println("log:", err)
			return
		}

		pathfile = "./" + nomeFile + ".json"
		// gravara no arquivo nosso json
		err = ioutil.WriteFile(pathfile, b, 0644)
		check(err)
	}
}

func Encode64String(content string) string {
	if len(content) > 0 {
		return base64.StdEncoding.EncodeToString([]byte(content))
	}
	return ""
}

func check(e error) {
	if e != nil {
		log.Println("log:", e)
	}
}
