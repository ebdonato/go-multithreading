package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type Response struct {
	Origin string
	Body   interface{}
}

const VIA_CEP_URL = "http://viacep.com.br/ws/%s/json/"
const BRASIL_CEP_URL = "https://brasilapi.com.br/api/cep/v1/%s"

func main() {
	resultChannel1 := make(chan *Response)
	resultChannel2 := make(chan *Response)

	var cep string
	var err error

	if len(os.Args) == 1 {
		cep, err = askForCep()
	} else {
		cep = os.Args[1]
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("CEP: %s\n", cep)

	go func() {
		resultChannel1 <- makeRequest(fmt.Sprintf(VIA_CEP_URL, cep))
	}()

	go func() {
		resultChannel2 <- makeRequest(fmt.Sprintf(BRASIL_CEP_URL, cep))
	}()

	select {
	case result1 := <-resultChannel1:
		printResponse(result1)
	case result2 := <-resultChannel2:
		printResponse(result2)
	case <-time.After(1 * time.Second):
		fmt.Println("Timeout: A resposta demorou muito para chegar.")
	}

	fmt.Println("Pronto!")
}

func printResponse(response *Response) {
	if len(response.Origin) == 0 {
		fmt.Println("Nada a exibir.")
	} else {
		fmt.Printf("API: %s\n", response.Origin)
		fmt.Printf("Resposta: %v\n", response.Body)
	}
}

func extractDomain(inputURL string) (string, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", err
	}

	domain := parsedURL.Hostname()
	return domain, nil
}

func makeRequest(url string) *Response {
	origin, err := extractDomain(url)
	if err != nil {
		fmt.Printf("Erro ao processar a url %s: %v\n", url, err)
		return &Response{Origin: "", Body: nil}
	}

	resp, err := resty.New().R().Get(url)
	if err != nil {
		fmt.Printf("Erro ao realizar a requisição na url  %s: %v\n", url, err)
		return &Response{Origin: "", Body: nil}
	}

	return &Response{Origin: origin, Body: resp}
}

func askForCep() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Digite um CEP: ")

	texto, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Erro ao ler a entrada:", err)
		return "", err
	}

	texto = strings.TrimSpace(texto)
	return texto, nil
}
