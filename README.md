# Stock Quote Alert

Este projeto foi desenvolvido como parte de um **desafio técnico do processo seletivo da INOA**.  
O objetivo é criar um sistema capaz de monitorar a cotação de ativos da **B3** e enviar alertas por e-mail quando determinados limites de preço forem atingidos.

---

## 📌 Requisitos

- O sistema deve **avisar via e-mail** caso a cotação de um ativo:
  - Caia abaixo de um preço de referência de compra.
  - Suba acima de um preço de referência de venda.
- O programa deve ser uma **aplicação de console** (não há necessidade de interface gráfica).
- Deve ser executado via linha de comando, recebendo **3 parâmetros**:
  1. **Ativo** a ser monitorado  
  2. **Preço de referência para venda**  
  3. **Preço de referência para compra**

### Exemplo de uso:
```bash
stock-quote-alert.exe PETR4 22.67 22.59
```


## Configuração do E-mail (Gmail)

- Para que o programa possa enviar alertas por e-mail, é necessário configurar uma Senha de App na sua conta do Gmail. Uma Senha de App é uma senha de 16 dígitos que permite a aplicativos de terceiros (como este script) se conectarem à sua conta de forma segura, sem usar sua senha principal.
### Importante: Este recurso só está disponível se a Verificação em Duas Etapas estiver ativada na sua conta Google.

Siga os passos abaixo para gerar sua senha:
    1. Acesse as configurações de Segurança da sua conta Google em: myaccount.google.com/security.
    2. Na seção "Como fazer login no Google", clique em Verificação em duas etapas e ative-a, caso ainda não esteja.
    3. Após ativar, volte para a mesma página de Segurança e clique em Senhas de app.
    4. Nos menus suspensos, selecione "Mail" como o aplicativo e "Outro" como o dispositivo. Dê um nome, como "Alerta de Ações Go", e clique em Gerar.
    5. Uma senha de 16 caracteres será exibida. Copie-a e salve-a imediatamente, pois ela só é mostrada uma única vez.
    6. Cole esta senha no campo Senha do seu arquivo config.ini. Seu arquivo de configuração deverá ficar parecido com este exemplo:

### Exemplo de uso:
```ini
[Email]
Destinatario = seu.email.destino@gmail.com

[SMTP]
Servidor = smtp.gmail.com
Porta = 587
Usuario = seu.email@gmail.com
Senha = sua_senha_de_app_de_16_digitos
```
