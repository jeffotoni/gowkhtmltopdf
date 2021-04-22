package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

type TaleHtml struct {
	Nome         string
	Html         string
	NoCollate    bool   `json:"nocollate,omitempty"`
	PageSize     string `json:"page_size,omitempty"`
	Orientation  string `json:"orientation,omitempty"`
	Dpi          uint   `json:"dpi,omitempty"`
	MarginBottom uint   `json:"margin_bottom,omitempty"`
	MarginTop    uint   `json:"margin_top,omitempty"`
	MarginLeft   uint   `json:"margin_left,omitempty"`
	MarginRight  uint   `json:"margin_right,omitempty"`
	ImageDpi     uint   `json:"image_dpi,omitempty"`
	ImageQuality uint   `json:"image_quality,omitempty"`
	Grayscale    bool   `json:"grayscale,omitempty"`
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
		m := TaleHtml{
			Grayscale:    false,
			NoCollate:    false,
			ImageDpi:     600,
			ImageQuality: 94,
			PageSize:     "A4",
			Orientation:  "Portrait",
			Dpi:          600,
			MarginBottom: 2,
			MarginTop:    2,
			MarginLeft:   2,
			MarginRight:  2,
			Nome:         "meu_primeiro_pdf.pdf",
			Html:         htmlEnc,
		}

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
