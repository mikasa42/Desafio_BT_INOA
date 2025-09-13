// main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/piquette/finance-go/quote"
	"gopkg.in/ini.v1"
	"gopkg.in/gomail.v2"
)

// Structs para armazenar as configurações de forma organizada
type SmtpConfig struct {
	Servidor string
	Porta    int
	Usuario  string
	Senha    string
}

type EmailConfig struct {
	Destinatario string
}

type AppConfig struct {
	SMTP  SmtpConfig
	Email EmailConfig
}

// Adicione esta função em main.go
func parseArgs() (string, float64, float64) {
	ativo := flag.String("ativo", "", "O código do ativo a ser monitorado (ex: PETR4). (Obrigatório)")
	precoVenda := flag.Float64("venda", 0.0, "Preço de referência para VENDA. (Obrigatório)")
	precoCompra := flag.Float64("compra", 0.0, "Preço de referência para COMPRA. (Obrigatório)")

	flag.Parse()

	if *ativo == "" || *precoVenda == 0.0 || *precoCompra == 0.0 {
		log.Fatal("Todos os parâmetros (--ativo, --venda, --compra) são obrigatórios.")
	}
	return *ativo, *precoVenda, *precoCompra
}

// Adicione esta função em main.go
func carregarConfiguracoes(path string) (*AppConfig, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler o arquivo de configuração: %v", err)
	}

	porta, err := cfg.Section("SMTP").Key("Porta").Int()
	if err != nil {
		return nil, fmt.Errorf("porta SMTP inválida: %v", err)
	}

	config := &AppConfig{
		SMTP: SmtpConfig{
			Servidor: cfg.Section("SMTP").Key("Servidor").String(),
			Porta:    porta,
			Usuario:  cfg.Section("SMTP").Key("Usuario").String(),
			Senha:    cfg.Section("SMTP").Key("Senha").String(),
		},
		Email: EmailConfig{
			Destinatario: cfg.Section("Email").Key("Destinatario").String(),
		},
	}
	return config, nil
}

// Adicione esta função em main.go
func obterCotacao(ativo string) (float64, error) {
	// Para a B3, adicionamos o sufixo .SA
	ticker := fmt.Sprintf("%s.SA", ativo)
	q, err := quote.Get(ticker)
	if err != nil || q == nil {
		return 0, fmt.Errorf("falha ao obter cotação para %s: %v", ticker, err)
	}
	return q.RegularMarketPrice, nil
}

// Adicione esta função em main.go
func enviarEmail(config *AppConfig, assunto, corpo string) {
	m := gomail.NewMessage()
	m.SetHeader("From", config.SMTP.Usuario)
	m.SetHeader("To", config.Email.Destinatario)
	m.SetHeader("Subject", assunto)
	m.SetBody("text/plain", corpo)

	d := gomail.NewDialer(config.SMTP.Servidor, config.SMTP.Porta, config.SMTP.Usuario, config.SMTP.Senha)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Erro ao enviar e-mail: %v", err)
	} else {
		log.Printf("E-mail de alerta enviado para %s", config.Email.Destinatario)
	}
}

// Esta será a função principal do seu programa
func main() {
	ativo, precoVenda, precoCompra := parseArgs()

	config, err := carregarConfiguracoes("config.ini")
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

 // Envia uma mensagem inicial para o e-mail de teste
    assuntoTeste := "Aviso: Monitoramento de Ações Iniciado"
    corpoTeste := "O programa de monitoramento de ações foi iniciado com sucesso."
    log.Printf("Enviando e-mail de teste para o destinatário: %s", config.Email.Destinatario)
    enviarEmail(config, assuntoTeste, corpoTeste)

	if precoCompra >= precoVenda {
		log.Fatal("O preço de compra deve ser menor que o preço de venda.")
	}

	log.Println("--- Iniciando Monitoramento ---")
	log.Printf("Ativo: %s", ativo)
	log.Printf("Alvo de Venda: > %.2f", precoVenda)
	log.Printf("Alvo de Compra: < %.2f", precoCompra)
	log.Println("---------------------------------")

	// Ticker que dispara a cada 1 minuto
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop() // Garante que o ticker seja limpo ao final

	ultimoAlerta := "" // Controla o estado para evitar spam

	// Loop infinito que aguarda o "tick" do nosso relógio
	for range ticker.C {
		precoAtual, err := obterCotacao(ativo)
		if err != nil {
			log.Println(err)
			continue // Pula para a próxima iteração
		}
		log.Printf("Cotação atual de %s: R$ %.2f", ativo, precoAtual)

		// Lógica de Venda
		if precoAtual > precoVenda && ultimoAlerta != "venda" {
			log.Println("!!! ALVO DE VENDA ATINGIDO !!!")
			assunto := fmt.Sprintf("ALERTA DE VENDA: %s", ativo)
			corpo := fmt.Sprintf("O ativo %s ultrapassou seu alvo de R$ %.2f.\n\nCotação atual: R$ %.2f.", ativo, precoVenda, precoAtual)
			enviarEmail(config, assunto, corpo)
			ultimoAlerta = "venda"
		}

		// Lógica de Compra
		if precoAtual < precoCompra && ultimoAlerta != "compra" {
			log.Println("!!! ALVO DE COMPRA ATINGIDO !!!")
			assunto := fmt.Sprintf("ALERTA DE COMPRA: %s", ativo)
			corpo := fmt.Sprintf("O ativo %s caiu abaixo do seu alvo de R$ %.2f.\n\nCotação atual: R$ %.2f.", ativo, precoCompra, precoAtual)
			enviarEmail(config, assunto, corpo)
			ultimoAlerta = "compra"
		}

		// Resetar o alerta se o preço voltar ao normal
		if precoAtual >= precoCompra && precoAtual <= precoVenda {
			if ultimoAlerta != "" {
				log.Println("Preço voltou à faixa normal. Alertas resetados.")
				ultimoAlerta = ""
			}
		}
	}
}