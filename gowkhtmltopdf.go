package main

import (
	"io/ioutil"
	"time"

	. "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	//"io"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	PORT_SERVER string
	X_KEY       = os.Getenv("X_KEY")
)

const (
	LIMIT_BYTE_BODY     = 31457280 // 30MB
	maxClients          = 1000     // simultaneos
	NewLimiter          = 1000     // 1k requests per second
	_                   = iota
	KB              int = 1 << (2 * iota)
	MB              int = 100 << (3 * iota)
	GB              int = 1000 << (3 * iota)
	MaxHeaderByte       = MB
)

type jsonHtml struct {
	Html string `json:"html"`
	Nome string `json:"nome"`
}

// Structure of our server configurations
type JsonMsg struct {
	Status  string `json:"status"`
	Message string `json:"message"`
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
		Addr:           ":" + PORT_SERVER,
		MaxHeaderBytes: MaxHeaderByte, // Size accepted by package
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
	}

	log.Fatal(confServer.ListenAndServe())
}

func headerHtmltoPdf(w http.ResponseWriter, r *http.Request) {

	ok, jsonerr, _ := CheckBasic(w, r)
	if !ok {
		msgerr := jsonerr
		log.Println(msgerr)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msgerr))
		return
	}

	content_type := strings.ToLower(r.Header.Get("Content-Type"))
	if content_type != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		jsonstr := `{"status":"error","message":"Content-type é obrigatório!"}`
		w.Write([]byte(jsonstr))
		return
	}

	jsonHtmlObj := jsonHtml{}
	// limit bytes request
	r.Body = http.MaxBytesReader(w, r.Body, LIMIT_BYTE_BODY)
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msgerr := err.Error()
		log.Println(msgerr)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msgerr))
		return
	}

	err = json.Unmarshal(body, &jsonHtmlObj)
	if err != nil {
		log.Println("Erro unmarshal: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	htmlpure := Decode64String(jsonHtmlObj.Html)
	byteFile := gerarHtmltoPdf(htmlpure)

	file := jsonHtmlObj.Nome
	mime := http.DetectContentType(byteFile)
	fileSize := len(string(byteFile))

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Disposition", "attachment; filename="+file+"")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Content-Length", strconv.Itoa(fileSize))
	w.Header().Set("Content-Control", "private, no-transform, no-store, must-revalidate")

	log.Println("Pdf gerado com sucesso para o arquivo: ", file, " size: ", fileSize, " Ip: ", r.RemoteAddr)

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

func Encode64String(content string) string {
	if len(content) > 0 {
		return base64.StdEncoding.EncodeToString([]byte(content))
	}
	return ""
}

func Encode64Byte(content []byte) string {
	if len(string(content)) > 0 {
		return base64.StdEncoding.EncodeToString(content)
	}
	return ""
}

func Decode64String(encoded string) string {
	if len(encoded) > 0 {

		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			log.Println("decode error:", err)
			return ""
		}
		return (string(decoded))
	}
	return ""
}

// validates and generates jwt token
func CheckBasic(w http.ResponseWriter, r *http.Request) (ok bool, msgjson, tokenUserDecodeS string) {

	ok = false
	msgjson = `{"status":"error","message":"tentando autenticar usuário!"}`

	// Authorization Basic base64 Encode
	auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(auth) != 2 || auth[0] != "Basic" {
		msgjson = GetJson(w, "error", http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	tokenBase64 := strings.Trim(auth[1], " ")
	tokenBase64 = strings.TrimSpace(tokenBase64)

	tokenUserEnc := tokenBase64
	// User, Login byte
	tokenUserDecode := Decode64String(tokenUserEnc)
	// User, Login string
	tokenUserDecodeS = strings.TrimSpace(strings.Trim(string(tokenUserDecode), " "))
	UserR := Decode64String(X_KEY)

	// Validate user and password in the database
	if tokenUserDecodeS == string(UserR) {
		ok = true
		return ok, `{"status":"ok"}`, tokenUserDecodeS
	} else {
		stringErr := "Usuário e chaves inválidas"
		//+ auth[0] + " - " + auth[1]
		msgjson = GetJson(w, "error", stringErr, http.StatusUnauthorized)
		return ok, msgjson, tokenUserDecodeS
	}

	defer r.Body.Close()
	return ok, msgjson, tokenUserDecodeS
}

// Returns json by typing on http
func GetJson(w http.ResponseWriter, Status string, Msg string, httpStatus int) string {
	msgJsonStruct := &JsonMsg{Status, Msg}
	msgJson, errj := json.Marshal(msgJsonStruct)
	if errj != nil {
		msg := `{"status":"error","message":"Não conseguimos gerar seu json!"}`
		return msg
	}
	return string(msgJson)
}
