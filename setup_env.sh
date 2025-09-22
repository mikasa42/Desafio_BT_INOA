#!/bin/bash

# Encerra o script imediatamente se um comando falhar.
set -e

# --- Cores e Estilos ---
BOLD=$(tput bold)
GREEN=$(tput setaf 2)
YELLOW=$(tput setaf 3)
BLUE=$(tput setaf 4)
RESET=$(tput sgr0)

# --- Nome do Projeto ---
PROJECT_NAME="stock-alert-go"
MODULE_NAME="stock-alert-go"

# --- Função para Instalação e Atualização do Go ---
install_go() {
    echo "${YELLOW}Go não encontrado ou versão desatualizada. Tentando instalar a versão mais recente...${RESET}"

    # Baixa a versão mais recente do Go
    GO_VERSION="1.23.0"
    GO_ARCHIVE="go${GO_VERSION}.linux-amd64.tar.gz"
    GO_DOWNLOAD_URL="https://go.dev/dl/${GO_ARCHIVE}"

    if ! command -v wget &> /dev/null; then
        echo "O comando 'wget' não foi encontrado. Por favor, instale-o (ex: sudo apt install wget) ou baixe o Go manualmente."
        exit 1
    fi
    wget "$GO_DOWNLOAD_URL"

    # Remove qualquer instalação antiga e extrai a nova
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "$GO_ARCHIVE"
    rm "$GO_ARCHIVE"

    # Adiciona o Go ao PATH do usuário
    if ! grep -q "/usr/local/go/bin" "$HOME/.profile"; then
        echo "export PATH=\$PATH:/usr/local/go/bin" >> "$HOME/.profile"
        echo "${GREEN}Adicionado /usr/local/go/bin ao seu PATH. Por favor, reinicie seu terminal ou execute 'source ~/.profile'.${RESET}"
    fi

    echo "${GREEN}Instalação do Go concluída.${RESET}"
}


# --- Script Principal ---

echo "${BOLD}Iniciando a configuração do ambiente para o projeto '${PROJECT_NAME}'...${RESET}"

# 1. Verificar se o Go está instalado e é uma versão recente
GO_VERSION_REQUIRED=1.21
if ! command -v go &> /dev/null || [ "$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')" -lt "$GO_VERSION_REQUIRED" ]; then
    install_go
fi

# Confirmação final da instalação
echo -n "Verificando a instalação do Go... "
if command -v go &> /dev/null; then
    echo "${GREEN}Go encontrado: $(go version)${RESET}"
else
    echo "${YELLOW}A instalação do Go falhou. Por favor, verifique os erros acima e tente instalar manualmente.${RESET}"
    exit 1
fi

# 2. Criar a estrutura do projeto
if [ -d "$PROJECT_NAME" ]; then
    echo "${YELLOW}O diretório '${PROJECT_NAME}' já existe. Pulando a criação.${RESET}"
else
    echo "Criando diretório do projeto: ${BOLD}$PROJECT_NAME${RESET}"
    mkdir "$PROJECT_NAME"
fi
cd "$PROJECT_NAME"

# 3. Inicializar o módulo Go
echo "Inicializando o módulo Go: ${BOLD}${MODULE_NAME}${RESET}"
go mod init "$MODULE_NAME"

# 4. Criar o arquivo de configuração config.ini
echo "Criando arquivo de configuração ${BOLD}config.ini${RESET} com valores de exemplo..."
cat << EOF > config.ini
[Email]
Destinatario = seu-email-de-destino@exemplo.com

[SMTP]
Servidor = smtp.gmail.com
Porta = 587
Usuario = aulasfisica4@gmail.com
Senha = jbjb dyju kpyr iogj

[AlphaVantage]
ChaveAPI = N723FNULRCE2TO84
EOF

# 5. Bibliotecas necessarias
echo "Importando bibliotecas necessarias ..."
cat << 'EOF' > main.go
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
	"bufio"
  "strings"
)

EOF

# 6. Baixar as dependências do Go
echo "Baixando as dependências do Go..."
go mod tidy

echo ""
echo "${GREEN}${BOLD}Ambiente configurado com sucesso!${RESET}"
echo "----------------------------------------------------"
echo ""
echo "${BOLD}Próximos Passos:${RESET}"
echo " ${YELLOW}Execute o programa:${RESET}"
echo "   ${BOLD}go run main.go --ativo {ativo} --venda {valor} --compra {valor}${RESET}"