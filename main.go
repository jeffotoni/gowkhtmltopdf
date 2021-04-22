// Package man generate pdf use wkhtmltopdf Wrapper C
// By @jeffotoni
package main

import (
	"io"

	. "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"github.com/ulule/limiter/v3/drivers/store/memory"

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
	sizeKB = 1 << (10 * 1) // 1 refers to the constants ByteSize KB
	sizeMB = 1 << (10 * 2) // 2 refers to the constants ByteSize MB -- example of declaring 5 MB
	sizeGB = 1 << (10 * 3) // 3 refers to the constants ByteSize GB
)

const (
	LIMIT_BYTE_BODY = 30 * sizeMB // 30MB
	maxClients      = 1000        // simultaneos
	NewLimiter      = 1000        // 1k requests per second
	MaxHeaderByte   = 30 * sizeMB
)

type jsonHtml struct {
	Html         string `json:"html"`
	Nome         string `json:"nome"`
	NoCollate    *bool  `json:"nocollate,omitempty"`
	PageSize     string `json:"page_size,omitempty"`
	Orientation  string `json:"orientation,omitempty"`
	Dpi          uint   `json:"dpi,omitempty"`
	MarginBottom uint   `json:"margin_bottom,omitempty"`
	MarginTop    uint   `json:"margin_top,omitempty"`
	MarginLeft   uint   `json:"margin_left,omitempty"`
	MarginRight  uint   `json:"margin_right,omitempty"`
	ImageDpi     uint   `json:"image_dpi,omitempty"`
	ImageQuality uint   `json:"image_quality,omitempty"`
	Grayscale    *bool  `json:"grayscale,omitempty"`
}

type JsonMsg struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func init() {
	PORT_SERVER = os.Getenv("PORT")
	if PORT_SERVER == "" {
		PORT_SERVER = "5010"
	}
}

func main() {
	rate, err := limiter.NewRateFromFormatted("100-S")
	if err != nil {
		log.Fatal(err)
		return
	}
	store := memory.NewStore()
	middleware := stdlib.NewMiddleware(limiter.New(store, rate, limiter.WithTrustForwardHeader(true)))

	mux := http.NewServeMux()
	mux.Handle("/v1/api/topdf", middleware.Handler(http.HandlerFunc(headerHtmltoPdf)))
	mux.Handle("/ping", middleware.Handler(http.HandlerFunc(Ping)))
	cserver := &http.Server{
		Handler:        mux,
		Addr:           ":" + PORT_SERVER,
		MaxHeaderBytes: MaxHeaderByte,
	}

	log.Println("run server port:", PORT_SERVER, " Max Header: ", MaxHeaderByte/1024/1024, "MB")
	log.Fatal(cserver.ListenAndServe())
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

	body, err := io.ReadAll(r.Body)
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
	byteFile := gerarHtmltoPdf(htmlpure, &jsonHtmlObj)

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
	w.Write(byteFile)
}

func Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Pong(🏓)!"))
}

func gerarHtmltoPdf(htmlStr string, obj *jsonHtml) []byte {

	// Create new PDF generator
	pdfg, err := NewPDFGenerator()
	if err != nil {
		log.Println("err: ", err)
	}
	if htmlStr == "" {
		htmlStr = `<html><body><h1 style="color:red;">Erro ao gerar seu PDF, o conteúdo veio vazio...<h1></body></html>`
	}

	pdfg.AddPage(NewPageReader(strings.NewReader(htmlStr)))

	// Landscape or Portrait
	if len(obj.Orientation) > 0 {
		pdfg.Orientation.Set(obj.Orientation)
	} else {
		pdfg.Orientation.Set("Portrait")
	}

	// true or false
	if obj.NoCollate != nil {
		pdfg.NoCollate.Set(*obj.NoCollate)
	} else {
		pdfg.NoCollate.Set(false)
	}
	if obj.Grayscale != nil {
		pdfg.Grayscale.Set(*obj.Grayscale)
	}

	// B0, A9 , A8, A7, A6, A5, A4 , A3 , A2, A1 , A0
	if len(obj.PageSize) > 0 {
		pdfg.PageSize.Set(obj.PageSize)
	} else {
		pdfg.PageSize.Set("A4")
	}
	if obj.Dpi > 0 {
		pdfg.Dpi.Set(obj.Dpi)
	} else {
		pdfg.Dpi.Set(350)
	}
	if obj.ImageDpi > 0 {
		pdfg.ImageDpi.Set(obj.ImageDpi)
	}
	if obj.ImageQuality > 0 {
		pdfg.ImageQuality.Set(obj.ImageQuality)
	}
	if obj.MarginBottom > 0 {
		pdfg.MarginBottom.Set(obj.MarginBottom)
	}
	if obj.MarginTop > 0 {
		pdfg.MarginTop.Set(obj.MarginTop)
	}
	if obj.MarginLeft > 0 {
		pdfg.MarginLeft.Set(obj.MarginLeft)
	}
	if obj.MarginTop > 0 {
		pdfg.MarginRight.Set(obj.MarginTop)
	}

	err = pdfg.Create()
	if err != nil {
		log.Println("Error pdfg.Create:", err)
		var bb []byte
		return bb
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
	msgJsonStruct := &JsonMsg{Status: Status, Message: Msg}
	msgJson, errj := json.Marshal(msgJsonStruct)
	if errj != nil {
		msg := `{"status":"error","message":"Não conseguimos gerar seu json!"}`
		return msg
	}
	return string(msgJson)
}
