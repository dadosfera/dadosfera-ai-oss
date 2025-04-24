# Dadosfera AI API

The Dadosfera AI API serves as the back-end for the Dadosfera AI platform. It is responsible for executing
and managing pipelines (and storing their states). In addition, it manages the Jupyter environment
that allows the user to edit their notebooks within Dadosfera AI.

## orchest-api

Container running the Flask API.

### Documentação da API (Swagger UI)

A API possui documentação interativa através do Swagger UI disponível em:

```
http://<servidor>/api/swagger
```

Ou acessando diretamente a raiz da aplicação, que redirecionará automaticamente para a documentação:

```
http://<servidor>/
```

A documentação fornece:
- Listagem de todos os endpoints disponíveis
- Métodos HTTP suportados (GET, POST, PUT, DELETE)
- Parâmetros necessários para cada endpoint
- Exemplos de requisição e resposta
- Interface para testar as chamadas diretamente pelo navegador

## celery-worker

Container running the Celery worker that receives background tasks from the `orchest-api`. An
additional container running RabbitMQ is run to serve as a broker between the API and Celery.
