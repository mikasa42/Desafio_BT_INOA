package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3" // Driver SQLite

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"

	"gopkg.in/gomail.v2"
	"gopkg.in/ini.v1"
)

// --- Structs de Configuração ---
type SmtpConfig struct {
	Servidor string
	Porta    int
	Usuario  string
	Senha    string
}

type EmailConfig struct {
	Destinatario string
}

type AlphaVantageConfig struct {
	ChaveAPI string
}

type AppConfig struct {
	SMTP         SmtpConfig
	Email        EmailConfig
	AlphaVantage AlphaVantageConfig
}

// --- Funções Auxiliares ---
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
		AlphaVantage: AlphaVantageConfig{
			ChaveAPI: cfg.Section("AlphaVantage").Key("ChaveAPI").String(),
		},
	}
	return config, nil
}

// --- Estruturas para Decodificar o JSON da API ---
type GlobalQuoteResponse struct {
	GlobalQuote struct {
		Symbol        string `json:"01. symbol"`
		Open          string `json:"02. open"`
		High          string `json:"03. high"`
		Low           string `json:"04. low"`
		Price         string `json:"05. price"` // Usaremos este campo
		Volume        string `json:"06. volume"`
		LatestDay     string `json:"07. latest trading day"`
		PreviousClose string `json:"08. previous close"`
		Change        string `json:"09. change"`
		ChangePercent string `json:"10. change percent"`
	} `json:"Global Quote"`
}

// --- Funções para Obter Cotação e Salvar Dados ---
func obterCotacao(ativo string, chaveAPI string) (float64, error) {
	apiURL := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", ativo, chaveAPI)

	resp, err := http.Get(apiURL)
	if err != nil {
		return 0, fmt.Errorf("falha na requisição HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("erro na API: status %s", resp.Status)
	}

	var quoteResponse GlobalQuoteResponse
	err = json.NewDecoder(resp.Body).Decode(&quoteResponse)
	if err != nil {
		return 0, fmt.Errorf("falha ao decodificar JSON da API: %v", err)
	}

	precoStr := quoteResponse.GlobalQuote.Price
	if precoStr == "" {
		return 0, fmt.Errorf("preço não encontrado para o ativo %s", ativo)
	}

	preco, err := strconv.ParseFloat(precoStr, 64)
	if err != nil {
		return 0, fmt.Errorf("falha ao converter o preço '%s': %v", precoStr, err)
	}

	return preco, nil
}

// --- Funções para Gerar Gráfico e Enviar E-mail ---
func enviarEmail(config *AppConfig, assunto, corpo, anexoPath string) {
	m := gomail.NewMessage()
	m.SetHeader("From", config.SMTP.Usuario)
	m.SetHeader("To", config.Email.Destinatario)
	m.SetHeader("Subject", assunto)
	m.SetBody("text/plain", corpo)

	if anexoPath != "" {
		m.Attach(anexoPath)
	}

	d := gomail.NewDialer(config.SMTP.Servidor, config.SMTP.Porta, config.SMTP.Usuario, config.SMTP.Senha)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Erro ao enviar e-mail: %v", err)
	} else {
		log.Printf("E-mail de alerta enviado para %s", config.Email.Destinatario)
	}
}

func gerarGrafico(db *sql.DB, ativo string, periodo string) (string, error) {
	rows, err := db.Query("SELECT timestamp, preco FROM cotacoes WHERE ativo = ? ORDER BY timestamp DESC LIMIT 200", ativo)
	if err != nil {
		return "", fmt.Errorf("falha ao buscar dados do banco: %v", err)
	}
	defer rows.Close()

	var timestamps []time.Time
	var precos []float64

	for rows.Next() {
		var timestamp time.Time
		var preco float64
		if err := rows.Scan(&timestamp, &preco); err != nil {
			return "", fmt.Errorf("falha ao escanear linha do banco: %v", err)
		}
		timestamps = append(timestamps, timestamp)
		precos = append(precos, preco)
	}

	if len(timestamps) == 0 {
		return "", fmt.Errorf("não há dados suficientes para gerar o gráfico")
	}

	pts := make(plotter.XYs, len(precos))
	for i := range precos {
		// Assegura que o gráfico vai do ponto mais antigo ao mais recente
		pts[len(precos)-1-i].X = float64(timestamps[i].Unix())
		pts[len(precos)-1-i].Y = precos[i]
	}

	p := plot.New()

	p.Title.Text = fmt.Sprintf("Histórico de Cotação (%s)", ativo)
	p.X.Label.Text = "Tempo"
	p.Y.Label.Text = "Preço (R$)"

	err = plotutil.AddLinePoints(p, "Preço", pts)
	if err != nil {
		return "", err
	}

	p.X.Tick.Marker = plot.TimeTicks{Format: "02/01\n15:04"}

	// Salva o gráfico em um arquivo temporário
	graficoPath := fmt.Sprintf("grafico_%s.png", ativo)
	if err := p.Save(8*vg.Inch, 4*vg.Inch, graficoPath); err != nil {
		return "", fmt.Errorf("falha ao salvar gráfico: %v", err)
	}
	return graficoPath, nil
}


// --- Função Principal Atualizada ---
func main() {
	ativo, precoVenda, precoCompra := parseArgs()
	config, err := carregarConfiguracoes("config.ini")
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}
	if config.AlphaVantage.ChaveAPI == "" {
		log.Fatal("Chave da API da Alpha Vantage não encontrada no arquivo de configuração.")
	}

	// 1. Inicializa o banco de dados SQLite
	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		log.Fatalf("Erro ao abrir/criar banco de dados: %v", err)
	}
	defer db.Close()

	sqlStmt := `
		CREATE TABLE IF NOT EXISTS cotacoes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ativo TEXT NOT NULL,
			preco REAL NOT NULL,
			timestamp DATETIME NOT NULL
		);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("Erro ao criar tabela: %v", err)
	}

	log.Println("--- Iniciando Monitoramento ---")
	log.Printf("Ativo: %s", ativo)
	log.Printf("Alvo de Venda: > %.2f", precoVenda)
	log.Printf("Alvo de Compra: < %.2f", precoCompra)
	log.Println("---------------------------------")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	ultimoAlerta := ""

	for range ticker.C {
		precoAtual, err := obterCotacao(ativo, config.AlphaVantage.ChaveAPI)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Cotação atual de %s: R$ %.2f", ativo, precoAtual)

		// 2. Salva a cotação no banco de dados
		stmt, err := db.Prepare("INSERT INTO cotacoes(ativo, preco, timestamp) VALUES(?, ?, ?)")
		if err != nil {
			log.Printf("Erro ao preparar inserção no banco: %v", err)
		} else {
			_, err := stmt.Exec(ativo, precoAtual, time.Now())
			if err != nil {
				log.Printf("Erro ao inserir dados no banco: %v", err)
			}
		}

		// 3. Verifica as condições de alerta e envia e-mail com anexo
		if precoAtual > precoVenda && ultimoAlerta != "venda" {
			log.Println("!!! ALVO DE VENDA ATINGIDO !!!")
			assunto := fmt.Sprintf("ALERTA DE VENDA: %s", ativo)
			corpo := fmt.Sprintf("O ativo %s ultrapassou seu alvo de R$ %.2f.\n\nCotação atual: R$ %.2f.", ativo, precoVenda, precoAtual)

			graficoPath, err := gerarGrafico(db, ativo, "24h")
			if err != nil {
				log.Printf("Não foi possível gerar o gráfico: %v", err)
				enviarEmail(config, assunto, corpo, "") // Envia sem anexo
			} else {
				enviarEmail(config, assunto, corpo, graficoPath)
				os.Remove(graficoPath) // Remove o arquivo após o envio
			}
			ultimoAlerta = "venda"
		} else if precoAtual < precoCompra && ultimoAlerta != "compra" {
			log.Println("!!! ALVO DE COMPRA ATINGIDO !!!")
			assunto := fmt.Sprintf("ALERTA DE COMPRA: %s", ativo)
			corpo := fmt.Sprintf("O ativo %s caiu abaixo do seu alvo de R$ %.2f.\n\nCotação atual: R$ %.2f.", ativo, precoCompra, precoAtual)

			graficoPath, err := gerarGrafico(db, ativo, "24h")
			if err != nil {
				log.Printf("Não foi possível gerar o gráfico: %v", err)
				enviarEmail(config, assunto, corpo, "") // Envia sem anexo
			} else {
				enviarEmail(config, assunto, corpo, graficoPath)
				os.Remove(graficoPath) // Remove o arquivo após o envio
			}
			ultimoAlerta = "compra"
		} else if precoAtual >= precoCompra && precoAtual <= precoVenda {
			if ultimoAlerta != "" {
				log.Println("Preço voltou à faixa normal. Alertas resetados.")
				ultimoAlerta = ""
			}
		}
	}
}