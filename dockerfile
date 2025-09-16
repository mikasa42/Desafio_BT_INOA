# Estágio 1: o 'builder'
# Usamos uma imagem Go para compilar a aplicação
FROM golang:1.20-alpine AS builder

# Define o diretório de trabalho dentro do container
WORKDIR /app

# Copia os arquivos de configuração do módulo Go
COPY go.mod ./
COPY go.sum ./

# Baixa as dependências.
# Apenas baixamos o que é necessário para a compilação
RUN go mod tidy

# Copia todo o código-fonte da sua aplicação
COPY . .

# Compila o executável para produção
# -o monitor: define o nome do arquivo de saída como 'monitor'
# -ldflags="-s -w": reduz o tamanho do binário final
RUN CGO_ENABLED=0 GOOS=linux go build -o monitor main.go

# Estágio 2: a imagem 'final'
# Usamos uma imagem base mínima para rodar a aplicação
FROM alpine:latest

# Define o diretório de trabalho
WORKDIR /root/

# Copia o executável do estágio 'builder' para a imagem final
COPY --from=builder /app/monitor .

# Copia o arquivo de configuração para a imagem final
# Note que o config.ini deve estar na mesma pasta do Dockerfile
COPY config.ini .

# Comando que será executado quando o container iniciar
ENTRYPOINT ["./monitor"]