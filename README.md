# Birth Reminder

Aplicacao Go de execucao unica que consulta aniversarios em uma planilha privada
do Google Sheets e envia mensagens por um webhook de WhatsApp. Foi pensada para
rodar diariamente pelo `cron` ou por um CronJob do Kubernetes, sem servidor web,
banco de dados ou interface administrativa.

## Como funciona

1. Le as colunas `nome`, `aniversario` e, opcionalmente, `ativo`.
2. Compara somente dia e mes no fuso configurado.
3. Envia uma mensagem para aniversarios do dia e um aviso para os do dia seguinte.
4. Nao chama o webhook quando nao ha aniversariantes.

## Planilha

Crie uma aba chamada `Funcionarios`:

| nome | aniversario | ativo |
| --- | --- | --- |
| Maria Souza | 17/09/1995 | SIM |

`aniversario` aceita `DD/MM`, `DD/MM/AAAA` ou `AAAA-MM-DD`. Quando a coluna
`ativo` existir, os valores aceitos sao `SIM`, `ATIVO`, `TRUE`, `1`, `S` e `YES`.

## Google Sheets

1. Habilite a Google Sheets API em um projeto do Google Cloud.
2. Crie uma conta de servico e baixe a credencial JSON.
3. Compartilhe a planilha como leitora com o `client_email` da conta de servico.
4. Mantenha o JSON fora deste repositorio e restrinja suas permissoes no sistema.

## Configuracao

Copie o exemplo e ajuste os valores:

```sh
cp .env.example .env
```

| Variavel | Obrigatoria | Padrao |
| --- | --- | --- |
| `GOOGLE_SHEET_ID` | sim | - |
| `GOOGLE_APPLICATION_CREDENTIALS` | sim | - |
| `GOOGLE_SHEET_RANGE` | nao | `Funcionarios!A:C` |
| `TIMEZONE` | nao | `America/Sao_Paulo` |
| `LOG_FORMAT` | nao | `human` |
| `WHATSAPP_WEBHOOK_URL` | sim | - |
| `WHATSAPP_RECIPIENT` | sim | - |
| `WHATSAPP_SENT_BY` | nao | `birth-reminder` |

O webhook deve aceitar um `POST` JSON neste formato:

```json
{
  "phone": "recipient-id",
  "message": "message text",
  "sent_by": "birth-reminder",
  "linkPreview": true
}
```

Em producao, `WHATSAPP_WEBHOOK_URL` deve usar HTTPS. Enderecos HTTP sao aceitos
somente em interfaces de loopback (`localhost`, `127.0.0.1` e `::1`) para testes.

## Executar

Requer Go 1.26 ou superior:

```sh
./run-local.sh
```

O script interpreta `.env` apenas como pares simples `CHAVE=VALOR`; comandos e
expressoes shell nao sao executados.

Ou compile um binario:

```sh
go build -o birth-reminder ./cmd/birthday-reminder
```

O arquivo `cron.example` mostra uma execucao diaria as 08:45 no fuso de Sao
Paulo. Ajuste caminhos, horario e fuso antes de instalar no crontab.

## Docker

```sh
docker build -t birth-reminder .
```

Monte a credencial como arquivo somente leitura e forneca as configuracoes por
variaveis de ambiente. Nunca copie a credencial para dentro da imagem.

## Seguranca

- Nao versione `.env`, credenciais JSON, kubeconfigs ou URLs autenticadas.
- Trate URLs de webhook como segredos quando elas autorizarem envios.
- Compartilhe a planilha apenas com a conta de servico necessaria.
- Use um gerenciador de segredos em ambientes de producao.
- Se um segredo entrar no Git, rotacione-o; apagar em outro commit nao remove o historico.

Consulte [SECURITY.md](SECURITY.md) para reportar vulnerabilidades.

## Licenca

Distribuido sob a licenca MIT. Consulte [LICENSE](LICENSE).
