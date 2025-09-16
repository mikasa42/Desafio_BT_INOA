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

# --- Função para Instalação do Go ---
install_go() {
    echo "${YELLOW}Go não encontrado. Tentando instalar...${RESET}"

    # Detecta o Sistema Operacional
    OS="$(uname -s)"
    case "${OS}" in
        Linux*)
            echo "Detectado sistema Linux."
            # Detecta a Distribuição
            if [ -f /etc/os-release ]; then
                # shellcheck source=/dev/null
                source /etc/os-release
                DISTRO=$ID
            else
                echo "${YELLOW}Não foi possível determinar a distribuição Linux. Por favor, instale o Go manualmente.${RESET}"
                exit 1
            fi

            case "$DISTRO" in
                ubuntu|debian|mint)
                    echo "${BLUE}Distribuição baseada em Debian/Ubuntu detectada.${RESET}"
                    echo "Será necessário usar 'sudo' para instalar pacotes. Sua senha pode ser solicitada."
                    sudo apt-get update && sudo apt-get install -y golang-go
                    ;;
                fedora|centos|rhel)
                    echo "${BLUE}Distribuição baseada em Fedora/RHEL detectada.${RESET}"
                    echo "Será necessário usar 'sudo' para instalar pacotes. Sua senha pode ser solicitada."
                    sudo dnf install -y golang
                    ;;
                arch)
                    echo "${BLUE}Distribuição Arch Linux detectada.${RESET}"
                    echo "Será necessário usar 'sudo' para instalar pacotes. Sua senha pode ser solicitada."
                    sudo pacman -Syu --noconfirm go
                    ;;
                *)
                    echo "${YELLOW}Distribuição Linux não suportada para instalação automática: ${DISTRO}${RESET}"
                    echo "Por favor, instale o Go manualmente em: https://go.dev/doc/install"
                    exit 1
                    ;;
            esac
            ;;
        Darwin*)
            echo "Detectado sistema macOS."
            if ! command -v brew &> /dev/null; then
                echo "${YELLOW}Homebrew (brew) não encontrado.${RESET}"
                echo "Para instalar o Go automaticamente, por favor, primeiro instale o Homebrew: https://brew.sh/"
                echo "Ou instale o Go manualmente: https://go.dev/doc/install"
                exit 1
            fi
            echo "${BLUE}Usando Homebrew para instalar o Go...${RESET}"
            brew install go
            ;;
        *)
            echo "${YELLOW}Sistema operacional não suportado para instalação automática: ${OS}${RESET}"
            echo "Por favor, instale o Go manualmente em: https://go.dev/doc/install"
            exit 1
            ;;
    esac
    echo "${GREEN}Instalação do Go concluída.${RESET}"
}


# --- Script Principal ---

echo "${BOLD}Iniciando a configuração do ambiente para o projeto '${PROJECT_NAME}'...${RESET}"

# 1. Verificar se o Go está instalado, caso contrário, instalar.
if ! command -v go &> /dev/null; then
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

# O restante do script continua como antes...
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
Usuario = seu-email-de-envio@gmail.com
Senha = sua-senha-de-aplicativo

EOF

# 5. Criar o esqueleto do arquivo main.go
echo "Criando esqueleto do arquivo ${BOLD}main.go${RESET}..."
cat << EOF > main.go
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

func main() {
	log.Println("Arquivo main.go criado. Cole o código completo aqui.")
}
EOF

# 6. Baixar as dependências do Go
echo "Baixando as dependências do Go..."
go mod tidy
go get github.com/mattn/go-sqlite3
go get gonum.org/v1/plot/...

# 7. Baixar a biblioteca da Alpha Vantage
echo "Baixando a biblioteca da Alpha Vantage..."
go get github.com/gocar/alpha-vantage-go

echo ""
echo "${GREEN}${BOLD}Ambiente configurado com sucesso!${RESET}"
echo "----------------------------------------------------"
echo ""
echo "${BOLD}Próximos Passos:${RESET}"
echo "1. ${YELLOW}Edite o arquivo 'config.ini'${RESET} com suas informações."
echo "2. ${YELLOW}Abra 'main.go'${RESET} e cole o código completo do monitor de ações."
echo "3. ${YELLOW}Execute o programa:${RESET}"
echo "   ${BOLD}go run main.go --ativo PETR4 --venda 28.50 --compra 28.00${RESET}"