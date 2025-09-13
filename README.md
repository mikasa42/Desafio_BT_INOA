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

