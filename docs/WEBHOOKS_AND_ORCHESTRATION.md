# GeevheitNeural — Webhooks de Saída e Orquestração do Grafo de Agentes

Documentação de integração externa e arquitetura determinística de orquestração (Fase 15).

---

## 1. Webhooks de Saída

O GeevheitNeural emite webhooks HTTP em tempo quase-real para conectar a operação a automações externas (ex: n8n, Zapier, Make, sistemas ERP ou microserviços proprietários).

### Headers HTTP Enviados

A cada disparo de evento, a API envia os seguintes headers de segurança:

| Header | Descrição | Exemplo |
|---|---|---|
| `Content-Type` | Formato do payload | `application/json` |
| `User-Agent` | Identificador do despachante | `GeevheitNeural-Webhook/1.0` |
| `X-Geevheit-Event` | Tipo do evento de negócio | `invoice.paid` |
| `X-Geevheit-Timestamp` | Unix timestamp (segundos) do disparo | `1774328400` |
| `X-Geevheit-Signature` | Assinatura HMAC-SHA256 do corpo | `sha256=a1b2c3d4e5f6...` |

### Cálculo da Assinatura HMAC-SHA256

A assinatura é calculada sobre a concatenação do timestamp com o corpo JSON bruto:
$$\text{payload\_to\_sign} = \text{timestamp} + "." + \text{raw\_body}$$
$$\text{signature} = \text{sha256} = \text{hex}(\text{HMAC-SHA256}(\text{secret}, \text{payload\_to\_sign}))$$

### Mitigações de Segurança Integradas
1. **Proteção Anti-SSRF:** O despachante rejeita URLs direcionadas para IPs privados (RFC 1918: `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), loopbacks (`127.0.0.1`, `localhost`) e metadados de nuvem (`169.254.169.254`).
2. **Prevenção a Replay Attacks:** O receptor deve descartar requisições cujo `X-Geevheit-Timestamp` tenha diferença superior a 5 minutos (300s) em relação ao horário local.
3. **Imutabilidade e Auditoria:** Cada tentativa é gravada em `geev.webhook_deliveries` com tempo de resposta em milissegundos e código de status retornado.

---

### Exemplo de Verificação em Node.js / Express

```javascript
const crypto = require('crypto');
const express = require('express');

const app = express();

// IMPORTANTE: Use o buffer bruto do corpo da requisição para verificar o HMAC
app.use(express.json({
  verify: (req, res, buf) => {
    req.rawBody = buf;
  }
}));

function verifyGeevheitSignature(req, secret) {
  const signatureHeader = req.headers['x-geevheit-signature'];
  const timestamp = req.headers['x-geevheit-timestamp'];

  if (!signatureHeader || !timestamp) {
    throw new Error('Headers de assinatura Geevheit ausentes');
  }

  // Prevenção a replay attacks (tolerância: 5 minutos)
  const currentTime = Math.floor(Date.now() / 1000);
  if (Math.abs(currentTime - parseInt(timestamp, 10)) > 300) {
    throw new Error('Timestamp do webhook expirado ou fora da janela de tolerância');
  }

  const payloadToSign = `${timestamp}.${req.rawBody.toString('utf8')}`;
  const expectedHash = crypto
    .createHmac('sha256', secret)
    .update(payloadToSign)
    .digest('hex');

  const expectedSignature = `sha256=${expectedHash}`;

  // Comparação em tempo constante para evitar timing attacks
  const isValid = crypto.timingSafeEqual(
    Buffer.from(signatureHeader),
    Buffer.from(expectedSignature)
  );

  return isValid;
}

app.post('/webhook/geevheit', (req, res) => {
  const WEBHOOK_SECRET = process.env.GEEVHEIT_WEBHOOK_SECRET;

  try {
    if (!verifyGeevheitSignature(req, WEBHOOK_SECRET)) {
      return res.status(401).send('Assinatura inválida');
    }

    const event = req.headers['x-geevheit-event'];
    const data = req.body;

    console.log(`Evento recebido com sucesso: ${event}`, data);

    // Processamento do evento (ex: atualizar CRM, enviar mensagem)
    res.status(200).json({ received: true });
  } catch (err) {
    console.error('Falha na validação do webhook:', err.message);
    res.status(400).send(`Erro: ${err.message}`);
  }
});

app.listen(3000, () => console.log('Servidor de Webhook rodando na porta 3000'));
```

---

### Exemplo de Configuração no n8n (Code Node)

No n8n, utilize um nó **Webhook** configurado como `POST` e em seguida adicione um nó **Code (JavaScript)**:

```javascript
const crypto = require('crypto');

const secret = $env.GEEVHEIT_WEBHOOK_SECRET;
const headers = $input.first().json.headers;
const bodyString = JSON.stringify($input.first().json.body);

const signature = headers['x-geevheit-signature'];
const timestamp = headers['x-geevheit-timestamp'];

if (!signature || !timestamp) {
  throw new Error('Headers de assinatura Geevheit ausentes');
}

const payload = `${timestamp}.${bodyString}`;
const computedSignature = 'sha256=' + crypto.createHmac('sha256', secret).update(payload).digest('hex');

if (signature !== computedSignature) {
  throw new Error('Assinatura HMAC inválida. Rejeitando requisição.');
}

return [{
  json: {
    event: headers['x-geevheit-event'],
    timestamp: timestamp,
    data: $input.first().json.body
  }
}];
```

---

## 2. Grafo de Agentes: Da Especificação ao Aceite do Cliente (F15-03)

O item **F15-03** estabelece a necessidade de estruturar o fluxo de desenvolvimento autônomo e colaborativo de forma **determinística, auditável e segura**.

### A Arquitetura do Grafo

```
                     ┌───────────────────────────────────────────────┐
                     │   Novo Pedido (/intake ou Comando de Voz)     │
                     └───────────────────────┬───────────────────────┘
                                             │
                                             ▼
                                  [1. Agente Requisitos]
                           (Estrutura escopo, orçamento e card)
                                             │
                                             ▼
                                 [PORTÃO 1: Proposta] ◄─── DECISÃO EXCLUSIVA DO DONO (2FA)
                                             │
                                             ▼
                                   [2. Agente Arquiteto]
                             (Gera schema, OpenAPI e plano)
                                             │
                                             ▼
                                   [PORTÃO 2: Prompt]   ◄─── DECISÃO EXCLUSIVA DO DONO (2FA)
                                             │
                                             ▼
                                     [3. Agente Dev]
                            (Implementa código em branch sandbox)
                                             │
                                             ▼
                                      [4. Agente QA]
                              (Testes, coverage, SAST e linters)
                                             │
                                             ▼
                                    [5. Agente Deploy]
                             (Gera ambiente de preview isolado)
                                             │
                                             ▼
                                  [PORTÃO 3: Entrega]   ◄─── DECISÃO EXCLUSIVA DO DONO (2FA)
                                             │
                                             ▼
                                 [6. Aceite do Cliente]
                               (Portal do Cliente / HMAC Token)
```

---

### Detalhamento dos Nós do Grafo

#### 1. Nó de Requisitos (`AgentRequirements`)
- **Entrada:** Formulário de Intake público, mensagem de áudio transcrita ou comando direto do dono.
- **Função:** Categoriza o tipo de demanda (Landing page, API, SaaS, Integração), infere escopo preliminar e estima custo/prazo.
- **Saída:** Cria o `Card` em estágio `proposta` com proposta comercial detalhada e valores em centavos (`int64`).
- **Segurança:** O agente **não avança** o estágio. O card entra em espera pelo **Portão 1**.

#### Portão 1: Aprovação da Proposta (Dono)
- Decisão manual do Dono autenticado com JWT + TOTP (AAL2).
- Gravação imutável em `geev.approvals`.
- Se recusado: card arquivado. Se aprovado: avança para estágio `prompt`.

#### 2. Nó de Arquiteto (`AgentArchitect`)
- **Entrada:** Escopo aprovado.
- **Função:** Desenha a arquitetura técnica, diagrama de dados, dependências e especificações OpenAPI. Gera o prompt instrucional estruturado e o plano de execução (`task list`).
- **Saída:** Anexa os artefatos de engenharia ao Card.
- **Segurança:** O agente submete o plano para revisão e para no **Portão 2**.

#### Portão 2: Aprovação do Prompt/Plano (Dono)
- O Dono valida as instruções de engenharia, os limites de computação e orçamento em centavos (`spend_limit_cents`).
- Gravação em `geev.approvals`.
- Ao aprovar: o pipeline autônomo é liberado.

#### 3. Nó de Desenvolvimento (`AgentDeveloper`)
- **Entrada:** Instruções aprovadas e repositório de trabalho.
- **Execução:**
  - Opera estritamente dentro de container temporário isolado (`sandbox`).
  - Cria branch de feature `feature/<card-id>-<slug>`.
  - Escreve código seguindo SOLID, DDD e convenções do repositório.
  - Commits atômicos no padrão Conventional Commits.
- **Saída:** Pull Request / Commit hash.

#### 4. Nó de QA (`AgentQA`)
- **Entrada:** Código gerado pelo `AgentDeveloper`.
- **Validações Executadas:**
  1. Testes unitários e de integração (`go test -race`, Jest/Vitest).
  2. Verificação de linters estritos (`golangci-lint`, ESLint sem warnings).
  3. Varredura estática de segurança (SAST anti-vulnerabilidades e verificação de segredos expostos via `gitleaks`/`trivy`).
  4. Validação de cobertura mínima (ex: 80% das camadas de domínio e aplicação).
- **Decisão Determinística:**
  - Se houver falha: feedback com logs estruturados retorna ao `AgentDeveloper` (limite máximo de 3 ciclos de auto-reparo).
  - Se aprovado: prossegue para o nó de Deploy.

#### 5. Nó de Deploy (`AgentDeployer`)
- **Entrada:** Build aprovado pelo QA.
- **Função:** Realiza deploy automatizado em ambiente de preview efêmero (ex: Vercel Preview, staging no Render/Fly.io ou banco isolado no Neon).
- **Saída:** URL de homologação gerada e anexada ao card.

#### Portão 3: Aprovação de Entrega (Dono)
- O Dono inspeciona a URL de preview e a evidência de testes.
- Aprovação explícita via celular/dashboard com 2FA.
- O card avança para `entrega` e o link do Portal do Cliente é liberado.

#### 6. Aceite Formal do Cliente (`ClientAcceptancePortal`)
- O cliente acessa `https://hub.suaempresa.com/portal?token=<magic_token>`.
- O cliente revisa a entrega e clica em **"Aceitar Entrega Formalmente"**.
- O backend registra o aceite com IP criptografado (SHA-256), User-Agent, carimbo de tempo auditável e notifica o Dono via Webhook e Notificação Operacional.
- Fatura de liquidação gerada e lançada no Ledger.

---

### Invariantes Invioláveis do Grafo

1. **`ExecutionGuard` Mandatório:** Toda chamada executada por qualquer nó de agente passa pelo `ExecutionGuard`. Se o sistema estiver pausado (`system.is_paused == true`) ou o conector do agente estiver `ativo == false`, a execução é abortada imediatamente.
2. **Nenhum Agente Aprova Portão:** O avanço entre estágios críticos requer `domain.OwnerPrincipal`. Agentes utilizam `domain.AgentPrincipal` e recebem `403 Forbidden` ao tentar interagir com `/v1/cards/{id}/gates/*`.
3. **Imutabilidade Total:** Todas as transições, decisões de portão, saídas de QA e entregas de webhooks são gravadas em tabelas `APPEND-ONLY` com triggers de bloqueio para `UPDATE` e `DELETE`.
4. **Resiliência a Falhas:** Falhas de rede em webhooks externos ou canais de notificação nunca causam rollback de transações de negócio principais.
