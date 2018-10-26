package main

import (
	"time"

	. "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	//"io"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	PORT_SERVER string
)

const (
	HEDER_X_KEY     = "xxxxxxxxxxxxxxxxxxxx"
	LIMIT_BYTE_BODY = 5000000 // 5MB
	maxClients      = 1000    // simultaneos
	NewLimiter      = 1000    // 1k requests per second

	_      = iota
	KB int = 1 << (2 * iota)
	MB int = 100 << (3 * iota)
	GB int = 1000 << (3 * iota)

	MaxHeaderByte = MB
)

type jsonHtml struct {
	Html string `json:"html"`
	Nome string `json:"nome"`
}

func maxClientsFunc(h http.Handler, n int) http.HandlerFunc {

	sema := make(chan struct{}, n)

	return func(w http.ResponseWriter, r *http.Request) {

		sema <- struct{}{}

		defer func() { <-sema }()

		h.ServeHTTP(w, r)
	}
}

// init..
func init() {

	PORT_SERVER = os.Getenv("PORT")

	if PORT_SERVER == "" {
		PORT_SERVER = "5010"
	}

	http.HandleFunc("/ping", Ping)
	{
		println("run server port:"+PORT_SERVER, " Max Header: ", MB)
	}
}

func main() {

	// /v1/api/topdf
	handlerApiHtmltoPdf := http.HandlerFunc(headerHtmltoPdf)

	// fazendo o controle de conexoes
	http.Handle("/v1/api/topdf", maxClientsFunc(handlerApiHtmltoPdf, maxClients))

	confServer := &http.Server{

		Addr: ":" + PORT_SERVER,

		MaxHeaderBytes: MaxHeaderByte, // Size accepted by package

		ReadTimeout: 5 * time.Second,

		WriteTimeout: 10 * time.Second,
	}

	log.Fatal(confServer.ListenAndServe())
}

func headerHtmltoPdf(w http.ResponseWriter, r *http.Request) {

	jsonHtmlObj := jsonHtml{}

	r.Body = http.MaxBytesReader(w, r.Body, LIMIT_BYTE_BODY)

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)

	//err := json.Unmarshal(jsonObj, &jsonHtmlObj)
	err := decoder.Decode(&jsonHtmlObj)
	if err != nil {
		log.Println("Erro unmarshal: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`Erro na autenticacao do servico`))
		return
	}

	byteFile := gerarHtmltoPdf(jsonHtmlObj.Html)

	file := jsonHtmlObj.Nome
	mime := http.DetectContentType(byteFile)
	fileSize := len(string(byteFile))

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Disposition", "attachment; filename="+file+"")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Content-Length", strconv.Itoa(fileSize))
	w.Header().Set("Content-Control", "private, no-transform, no-store, must-revalidate")

	log.Println("Pdf gerado com sucesso para o arquivo: ", file, " tamanho: ", fileSize, " Ip: ", r.RemoteAddr)

	w.WriteHeader(http.StatusOK)

	//stream the body to the client without fully loading it into memory
	// io.Copy(w, r.body)
	w.Write(byteFile)

}

func Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Pong!"))
}

func gerarHtmltoPdf(htmlStr string) []byte {

	// Create new PDF generator
	pdfg, err := NewPDFGenerator()
	if err != nil {
		log.Println("err: ", err)
	}

	if htmlStr == "" {

		htmlStr = `<html><body><h1 style="color:red;">Erro ao gerar seu PDF, o conteúdo veio vazio...<h1></body></html>`
	}

	pdfg.AddPage(NewPageReader(strings.NewReader(htmlStr)))

	// set dpi of the content
	pdfg.Dpi.Set(350)

	// set margins to zero at all direction
	pdfg.MarginBottom.Set(0)
	pdfg.MarginTop.Set(0)
	pdfg.MarginLeft.Set(0)
	pdfg.MarginRight.Set(0)

	// Create PDF document in internal buffer
	err = pdfg.Create()
	if err != nil {
		log.Fatal(err)
	}

	byteFile := pdfg.Bytes()

	return byteFile
}
