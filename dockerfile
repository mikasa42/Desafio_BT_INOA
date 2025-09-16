# Estágio 1: o 'builder'
# Usamos uma imagem Go para compilar a aplicação
FROM golang:1.23-alpine AS builder

# Instale as dependências para que o CGO possa funcionar
RUN apk add --no-cache gcc libc-dev

# Define o diretório de trabalho dentro do container
WORKDIR /app

# Copia os arquivos de configuração do módulo Go
COPY go.mod ./
COPY go.sum ./

# Baixa as dependências
RUN go mod tidy

# Copia todo o código-fonte da sua aplicação
COPY . .

# Compila o executável para produção
# - CGO_ENABLED=1: Ativa o CGO para que o go-sqlite3 funcione
# - ldflags="-s -w": Reduz o tamanho do binário final
RUN CGO_ENABLED=1 GOOS=linux go build -o monitor main.go

# Estágio 2: a imagem 'final'
# Usamos uma imagem base mínima para rodar a aplicação
FROM alpine:latest

# Instala as dependências de runtime que o CGO precisa
RUN apk add --no-cache libc6-compat

# Define o diretório de trabalho
WORKDIR /root/

# Copia o executável do estágio 'builder' para a imagem final
COPY --from=builder /app/monitor .

# Copia o arquivo de configuração para a imagem final
COPY config.ini .

# Comando que será executado quando o container iniciar
ENTRYPOINT ["./monitor"]