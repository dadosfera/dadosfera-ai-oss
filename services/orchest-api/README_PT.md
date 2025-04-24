# API Dadosfera AI

A API Dadosfera AI serve como back-end para a plataforma Dadosfera AI. Ela é responsável por executar
e gerenciar pipelines (e armazenar seus estados). Além disso, gerencia o ambiente Jupyter
que permite ao usuário editar seus notebooks dentro da Dadosfera AI.

## orchest-api

Container executando a API Flask.

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

Container executando o Celery worker que recebe tarefas em segundo plano da `orchest-api`. Um
container adicional executando RabbitMQ é usado como broker entre a API e o Celery.

## Estrutura da API

A API é organizada em namespaces, cada um responsável por um recurso específico:

- **jobs**: Gerenciamento de tarefas programadas
- **runs**: Execução de pipelines
- **pipelines**: Gerenciamento de pipelines
- **projects**: Gerenciamento de projetos
- **environments**: Gerenciamento de ambientes de execução
- **sessions**: Gerenciamento de sessões interativas
- **services**: Serviços da plataforma
- **notifications**: Sistema de notificações
- **info**: Informações sobre a plataforma

## Como usar o Swagger UI

1. Acesse a URL do Swagger UI
2. Explore os endpoints disponíveis, organizados por namespaces
3. Clique em um endpoint para ver detalhes sobre parâmetros e respostas
4. Use o botão "Experimente" para testar a API diretamente pelo navegador
5. Preencha os parâmetros necessários e execute a requisição
6. Veja a resposta da API com o corpo da resposta e códigos HTTP

## Configuração

A documentação Swagger pode ser personalizada através do arquivo `swagger-config.json`, que permite ajustar:

- Título e descrição da API
- Informações de contato
- Termos de serviço
- Traduções de tags e mensagens de erro
- Configurações da interface do Swagger UI 