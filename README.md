# API de Gerenciamento de Tarefas (To-Do List)

Este projeto é uma API REST em Go para gerenciamento de tarefas (To-Do List), com operações básicas de CRUD, validação, regras de negócio e testes automatizados.

Principais características:
- API RESTful em Go
- Banco NoSQL (MongoDB) com fallback para armazenamento em memória nos testes
- Validações e regras de negócio configuráveis
- Logger injetável (interface) para facilitar testes
- Testes unitários e de integração

---

**Tecnologias**

- Linguagem: Go
- Banco: MongoDB (pode apontar para uma instância local ou remota)
- Router: Gorilla Mux
- Testes: pacote `testing` e `httptest`
- Documentação da API: `swagger.json` incluído

---

**Requisitos obrigatórios**

- Go 1.18+ instalado
- MongoDB (opcional para execução local; o projeto pode ter fallback em memória para testes)

---

**Como rodar (local)**

1. Ajuste a variável de ambiente `MONGO_URI` caso queira usar um MongoDB diferente do padrão.

Windows (PowerShell):

```powershell
$env:MONGO_URI = "mongodb://localhost:27017"
go run main.go
```

Linux / macOS:

```bash
export MONGO_URI="mongodb://localhost:27017"
go run main.go
```

Por padrão a API roda em `:8080`. Consulte `router/router.go` para ajustar porta/host.

---

**Executando testes**

Para rodar os testes unitários e de integração incluídos no diretório `tests` execute:

```bash
go test ./tests/... -v
```
ou

```bash
gotestsum
```

Os testes cobrem validações, regras de negócio, integrações básicas e a camada de armazenamento em memória.

---

**Variáveis de ambiente importantes**

- `MONGO_URI`: string de conexão com MongoDB (ex: `mongodb://localhost:27017`). Se não configurada, a aplicação pode tentar um fallback em memória para desenvolvimento/testes.

---

**Regras de Negócio (principais)**

- Não é permitido editar uma tarefa cujo `status` seja `completed`. Tentativas de atualização devem retornar `409 Conflict`.
- Campos obrigatórios: `title` (não-vazio).
- Valores válidos para `status`: `pending`, `in_progress`, `completed`.
- `priority` deve ser um inteiro entre `1` e `5` (inclusivo).
- `due_date` (quando fornecida) deve ser uma data futura.

Essas regras são aplicadas na camada de serviço (`models/service.go`) e podem ser configuradas/estendidas.

---

**Validação de dados**

A validação é centralizada usando um conjunto de validadores por campo (Strategy Pattern), tornando fácil adicionar novas regras sem alterar os handlers diretamente.

---

**Endpoints principais**

- `GET /tasks` - lista todas as tarefas
- `GET /tasks/{id}` - obtém tarefa por ID
- `POST /tasks` - cria nova tarefa
- `PUT /tasks/{id}` - atualiza tarefa existente (patch semântica suportada)
- `DELETE /tasks/{id}` - remove tarefa

Formatos e exemplos de payloads podem ser encontrados em `swagger.json`.

---

**Logs e tratamento de erros**

O projeto usa uma interface de `Logger` injetável para seguir o princípio de Inversão de Dependência (DIP). Isso facilita substituir o logger por um `NoOpLogger` durante os testes para manter a saída limpa.

Erros são retornados de forma estruturada pela API (JSON) com códigos HTTP apropriados.

---

**Makefile - Comandos úteis**

O projeto inclui um `Makefile` com comandos para facilitar o desenvolvimento. Compatível com Windows (PowerShell) e Linux/macOS.

Comandos disponíveis:

```bash
make build         # Compila o binário para bin/taskapi (Linux amd64)
make run           # Executa a aplicação localmente (go run main.go)
make test          # Roda todos os testes (go test ./... -v)
make fmt           # Formata o código (gofmt -w .)
make clean         # Remove o diretório bin/

# Docker
make docker-build  # Constrói a imagem Docker
make docker-up     # Sobe todos os serviços (api + mongo) com docker-compose
make docker-run    # Sobe apenas o serviço api com docker-compose
make docker-down   # Para e remove os containers
```

**Nota para Windows**: Se `make` não estiver instalado, use os comandos Docker Compose diretamente (veja seção Docker abaixo) ou instale via Chocolatey:
```powershell
choco install make -y
```

---

**Docker e Docker Compose**

O projeto inclui `Dockerfile` e `docker-compose.yml` prontos para uso.

**1. Usando Docker Compose (recomendado)**

Sobe a API + MongoDB em containers:

```bash
# Com docker-compose (todos os serviços)
docker compose up --build -d

# Ou apenas o serviço API
docker compose -f docker-compose.yml up -d --build api

# Verificar status
docker compose ps

# Parar e remover containers
docker compose down
```

**2. Usando Makefile + Docker**

```bash
# Build da imagem
make docker-build

# Subir todos os serviços (api + mongo)
make docker-up

# Ou apenas o serviço API
make docker-run

# Parar containers
make docker-down
```

**3. Acessar a API**

Após subir os containers:
- API: http://localhost:8080
- MongoDB: localhost:27017

Teste com:
```bash
curl http://localhost:8080/tasks
```

**4. Dockerfile**

O `Dockerfile` usa multi-stage build:
- **Stage 1 (builder)**: Compila a aplicação Go em Alpine Linux
- **Stage 2 (runtime)**: Imagem final mínima (~15MB) baseada em Alpine, executa como usuário não-root

**5. docker-compose.yml**

Define dois serviços:
- **mongo**: MongoDB 6 com volume persistente
- **api**: API compilada, conecta automaticamente ao MongoDB

---

**Swagger / Documentação**

O projeto inclui o arquivo `swagger.json` na raiz. Use-o para gerar documentação interativa (ex: Swagger UI) ou clientes.

---

**Estrutura do projeto (resumo)**

- `main.go` - ponto de entrada
- `router/` - configuração das rotas
- `handlers/` - camadas HTTP/handlers
- `models/` - lógica de domínio e validações
- `store/` - abstração de persistência (MongoDB + in-memory)
- `tests/` - testes unitários e de integração
- `Dockerfile` - imagem Docker multi-stage
- `docker-compose.yml` - orquestração de containers (api + mongo)
- `Makefile` - comandos úteis para build, test, docker
- `swagger.json` - documentação OpenAPI da API

---

**Solução de Problemas (Troubleshooting)**

**1. Erro ao buildar imagem Docker (DNS/Network)**

Se aparecer erro `lookup registry-1.docker.io: no such host`:
- Verifique sua conexão de internet
- Reinicie o Docker Desktop
- Verifique configurações de proxy no Docker Desktop (Settings → Resources → Proxies)
- Teste conectividade: `ping registry-1.docker.io` ou `nslookup registry-1.docker.io`

**2. `make` não reconhecido no Windows**

Instale via Chocolatey:
```powershell
choco install make -y
```

Ou use os comandos Docker Compose diretamente sem o Makefile.

**3. Docker daemon não está rodando**

Windows: Abra o Docker Desktop e aguarde até status "Docker Desktop is running"

Linux/WSL: 
```bash
sudo service docker start
```

**4. WSL não instalado (Windows)**

```powershell
wsl --install
```
Reinicie o computador após a instalação.



