# 📚 Documentação Técnica Completa - EmpregaBem API

## Índice

- [Visão Geral](#visão-geral)
- [Arquitetura Detalhada](#arquitetura-detalhada)
- [Banco de Dados](#banco-de-dados)
- [Autenticação e Segurança](#autenticação-e-segurança)
- [Configuração](#configuração)
- [Deploy](#deploy)
- [Testes](#testes)
- [Performance](#performance)
- [Troubleshooting](#troubleshooting)

---

## Visão Geral

### Stack Completo

```
Frontend vite + react (não incluído neste repo)
          ↓
    [REST API - Go]
          ↓
    [MongoDB Atlas]
```

### Fluxo de Autenticação

```
1. Usuário → POST /company/login ou /candidate/login
2. API valida credenciais (bcrypt)
3. API gera JWT token (válido 24h)
4. Usuário armazena token
5. Usuário envia token em todas as requisições protegidas
6. Middleware valida token e extrai user_id + user_type
7. Handler processa requisição com contexto do usuário
```

### Fluxo de Candidatura

```
1. Candidato busca vagas → GET /jobs (público)
2. Candidato visualiza vaga → POST /jobs/{id}/view (incrementa contador)
3. Candidato se candidata → POST /candidate/applications
   ↓
   - Valida se vaga existe e está ativa
   - Valida se não há candidatura duplicada
   - Cria candidatura com status "pending"
   - Incrementa contador de candidatos na vaga (atômico)
4. Empresa visualiza candidatos → GET /company/jobs/{id}/applicants
   ↓
   - Auto-sincroniza contador de candidatos com contagem real
5. Empresa atualiza status → PATCH /company/applications/{id}/status
6. Candidato pode cancelar apenas se status = "pending"
```

---

## Arquitetura Detalhada

### Estrutura de Pastas

```
empregabemapi/
│
├── cmd/
│   └── api/
│       └── main.go                 # Entry point
│                                   # - Carrega .env
│                                   # - Conecta MongoDB
│                                   # - Inicializa repositories
│                                   # - Configura router
│                                   # - Inicia servidor HTTP
│
├── internal/
│   └── http/
│       ├── handlers/               # Controllers (lógica de negócio)
│       │   ├── applications.go    # CRUD candidaturas
│       │   │   - Apply()          # Criar candidatura
│       │   │   - List()           # Listar do candidato
│       │   │   - Cancel()         # Cancelar (só pending)
│       │   │   - ListJobApplicants() # Empresa ver candidatos
│       │   │   - UpdateApplicationStatus() # Empresa mudar status
│       │   │
│       │   ├── companies.go       # Auth empresas
│       │   │   - Register()       # Criar conta
│       │   │   - Login()          # Autenticar
│       │   │
│       │   ├── candidates.go      # Auth candidatos
│       │   │   - Register()       # Criar conta
│       │   │   - Login()          # Autenticar
│       │   │
│       │   ├── jobs.go            # CRUD vagas
│       │   │   - List()           # Listar (público + filtros)
│       │   │   - GetByID()        # Detalhes vaga
│       │   │   - RegisterView()   # Incrementar views
│       │   │   - Create()         # Empresa criar
│       │   │   - Update()         # Empresa editar
│       │   │   - Delete()         # Empresa deletar
│       │   │   - ToggleStatus()   # Ativar/desativar
│       │   │
│       │   ├── saved_jobs.go      # Favoritos
│       │   │   - Save()           # Adicionar favorito
│       │   │   - List()           # Listar favoritos
│       │   │   - Remove()         # Remover favorito
│       │   │
│       │   └── maintenance.go     # Manutenção
│       │       - FixJobCounters() # Inicializar contadores
│       │
│       ├── middleware/
│       │   └── auth.go            # Autenticação e autorização
│       │       - AuthMiddleware() # Valida JWT, injeta user no context
│       │       - CompanyOnly()    # Permite apenas empresas
│       │       - CandidateOnly()  # Permite apenas candidatos
│       │
│       └── router.go              # Configuração de rotas
│           - SetupRoutes()        # Registra todos os endpoints
│
├── companies/
│   ├── model.go                   # type Company struct
│   └── repository.go              # MongoRepository para companies
│       - Create()                 # Inserir empresa
│       - GetByEmail()             # Buscar por email
│       - GetByCNPJ()              # Buscar por CNPJ
│       - GetByID()                # Buscar por ID
│
├── candidates/
│   ├── model.go                   # type Candidate struct
│   └── repository.go              # MongoRepository para candidates
│       - Create()                 # Inserir candidato
│       - GetByEmail()             # Buscar por email
│       - GetByID()                # Buscar por ID
│
├── jobs/
│   ├── model.go                   # type Job struct
│   └── repository.go              # MongoRepository para jobs
│       - Create()                 # Inserir vaga (views=0, applicants=0)
│       - List()                   # Listar todas
│       - Search()                 # Listar com filtros (location, jobType, level, minSalary)
│       - GetByID()                # Buscar por ID
│       - Update()                 # Atualizar vaga
│       - Delete()                 # Deletar vaga
│       - IncrementViews()         # $inc views (atômico)
│       - IncrementApplicants()    # $inc applicants (atômico)
│       - DecrementApplicants()    # $inc applicants:-1 (atômico)
│       - SetApplicantsCount()     # $set applicants (correção manual)
│
├── applications/
│   ├── model.go                   # type Application struct
│   └── repository.go              # MongoRepository para applications
│       - Create()                 # Criar candidatura
│       - GetByID()                # Buscar por ID
│       - GetByCandidateID()       # Listar do candidato
│       - GetByJobID()             # Listar da vaga
│       - Exists()                 # Verificar duplicata
│       - Delete()                 # Deletar candidatura
│       - UpdateStatus()           # Atualizar status
│       - CountByJobID()           # Contar candidatos (para sync)
│
├── users/
│   └── repository.go              # Busca multi-collection
│       - FindByEmailAndType()     # Busca em companies OU candidates
│
└── database/
    └── connection.go              # Conexão MongoDB
        - Connect()                # Cria client MongoDB
        - GetDatabase()            # Retorna *mongo.Database
```

### Padrões de Design

#### 1. Repository Pattern

Cada domínio tem seu repository que abstrai o MongoDB:

```go
type JobRepository interface {
    Create(ctx context.Context, job *Job) error
    GetByID(ctx context.Context, id string) (*Job, error)
    List(ctx context.Context) ([]*Job, error)
    // ...
}

type MongoRepository struct {
    collection *mongo.Collection
}

func (r *MongoRepository) Create(ctx context.Context, job *Job) error {
    // Lógica MongoDB
}
```

**Benefícios:**
- Desacopla lógica de negócio do banco
- Facilita testes (mock do repository)
- Permite trocar banco de dados facilmente

#### 2. Middleware Pattern

```go
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token := extractToken(r)
        claims, err := validateJWT(token)
        if err != nil {
            http.Error(w, "Unauthorized", 401)
            return
        }
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        next(w, r.WithContext(ctx))
    }
}
```

**Benefícios:**
- Reutilização de lógica de autenticação
- Separação de responsabilidades
- Composição de middlewares

#### 3. Context Pattern

```go
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    // Usa userID para buscar perfil
}
```

**Benefícios:**
- Passa dados entre middlewares e handlers
- Timeout control
- Cancelamento de requisições

---

## Banco de Dados

### Modelo de Dados Completo

#### Collection: `companies`

```javascript
{
  _id: ObjectId("674612fa3b2c1a4d8e9f0123"),
  cnpj: "12345678901234",           // String, unique, indexed
  name: "Tech Solutions LTDA",       // String
  email: "tech@email.com",           // String, unique, indexed
  password: "$2a$10$hashed...",       // String (bcrypt hash)
  location: "São Paulo, SP",         // String
  website: "https://techsolutions.com", // String (opcional)
  about: "Descrição da empresa",     // String (opcional)
  created_at: ISODate("2024-11-26"), // Date
  updated_at: ISODate("2024-11-26")  // Date
}
```

**Índices:**
```javascript
db.companies.createIndex({ "email": 1 }, { unique: true })
db.companies.createIndex({ "cnpj": 1 }, { unique: true })
```

---

#### Collection: `candidates`

```javascript
{
  _id: ObjectId("674612fa3b2c1a4d8e9f0124"),
  name: "João Silva",                // String
  email: "joao@email.com",           // String, unique, indexed
  password: "$2a$10$hashed...",       // String (bcrypt hash)
  phone: "11999999999",              // String
  location: "São Paulo, SP",         // String
  linkedin: "https://linkedin.com/...", // String (opcional)
  portfolio: "https://joao.dev",     // String (opcional)
  bio: "Desenvolvedor Full Stack",   // String (opcional)
  skills: ["JavaScript", "React"],   // Array<String>
  experience: "5 anos",              // String (opcional)
  created_at: ISODate("2024-11-26"), // Date
  updated_at: ISODate("2024-11-26")  // Date
}
```

**Índices:**
```javascript
db.candidates.createIndex({ "email": 1 }, { unique: true })
```

---

#### Collection: `jobs`

```javascript
{
  _id: ObjectId("674612fa3b2c1a4d8e9f0125"),
  company_id: ObjectId("674612fa3b2c1a4d8e9f0123"), // Referência a companies
  title: "Desenvolvedor Full Stack",                 // String
  description: "Desenvolvimento de apps web...",     // String
  company: "Tech Solutions LTDA",                    // String (desnormalizado)
  location: "São Paulo, SP",                         // String
  salary: 8000.0,                                    // Float64
  
  // Categorização
  job_type: "híbrido",               // String: remoto | presencial | híbrido
  level: "pleno",                    // String: junior | pleno | senior
  
  // Detalhes
  requirements: [                    // Array<String>
    "JavaScript",
    "React",
    "Node.js"
  ],
  benefits: [                        // Array<String>
    "Vale-refeição",
    "Plano de saúde"
  ],
  
  // Métricas
  views: 156,                        // Int (incremento atômico)
  applicants: 23,                    // Int (incremento/decremento atômico)
  
  // Status
  is_active: true,                   // Boolean
  priority: 0,                       // Int: 0=normal, 1=destaque
  
  // Timestamps
  created_at: ISODate("2024-11-26"),
  updated_at: ISODate("2024-11-26")
}
```

**Índices:**
```javascript
db.jobs.createIndex({ "company_id": 1 })           // Buscar vagas da empresa
db.jobs.createIndex({ "is_active": 1 })            // Filtrar ativas
db.jobs.createIndex({ "location": 1 })             // Busca por localização
db.jobs.createIndex({ "job_type": 1 })             // Filtro tipo
db.jobs.createIndex({ "level": 1 })                // Filtro nível
db.jobs.createIndex({ "salary": 1 })               // Filtro salário
db.jobs.createIndex({ "created_at": -1 })          // Ordenar por recentes
```

---

#### Collection: `applications`

```javascript
{
  _id: ObjectId("674612fa3b2c1a4d8e9f0126"),
  job_id: ObjectId("674612fa3b2c1a4d8e9f0125"),       // Referência a jobs
  candidate_id: ObjectId("674612fa3b2c1a4d8e9f0124"), // Referência a candidates
  
  // Status workflow
  status: "pending",                 // pending | viewed | in_review | shortlisted | 
                                     // interview | rejected | accepted
  
  // Timestamps
  applied_at: ISODate("2024-11-26"),  // Data da candidatura
  viewed_at: ISODate("2024-11-27"),   // Quando empresa visualizou (null se não viu)
  updated_at: ISODate("2024-11-27")   // Última atualização
}
```

**Índices:**
```javascript
// Evita candidatura duplicada (mesmo candidato + mesma vaga)
db.applications.createIndex(
  { "job_id": 1, "candidate_id": 1 }, 
  { unique: true }
)

db.applications.createIndex({ "candidate_id": 1 })  // Listar candidaturas do candidato
db.applications.createIndex({ "job_id": 1 })        // Listar candidatos da vaga
db.applications.createIndex({ "status": 1 })        // Filtrar por status
```

---

#### Collection: `saved_jobs`

```javascript
{
  _id: ObjectId("674612fa3b2c1a4d8e9f0127"),
  candidate_id: ObjectId("674612fa3b2c1a4d8e9f0124"), // Referência a candidates
  job_id: ObjectId("674612fa3b2c1a4d8e9f0125"),       // Referência a jobs
  saved_at: ISODate("2024-11-26")                     // Data que salvou
}
```

**Índices:**
```javascript
// Evita salvar mesma vaga duas vezes
db.saved_jobs.createIndex(
  { "candidate_id": 1, "job_id": 1 },
  { unique: true }
)

db.saved_jobs.createIndex({ "candidate_id": 1 })  // Listar favoritos do candidato
```

---

### Operações Atômicas

#### Incrementar Views (Thread-Safe)

```go
func (r *MongoRepository) IncrementViews(ctx context.Context, jobID string) error {
    id, _ := bson.ObjectIDFromHex(jobID)
    filter := bson.M{"_id": id}
    update := bson.M{"$inc": bson.M{"views": 1}}
    
    _, err := r.collection.UpdateOne(ctx, filter, update)
    return err
}
```

**Operação MongoDB:**
```javascript
db.jobs.updateOne(
  { _id: ObjectId("...") },
  { $inc: { views: 1 } }
)
```

**Por que é atômico?**
- `$inc` é operação atômica do MongoDB
- Não há race condition mesmo com múltiplas requisições simultâneas
- MongoDB garante consistência

#### Incrementar/Decrementar Applicants

```go
// Quando candidato se candidata
IncrementApplicants(ctx, jobID) // $inc: {applicants: 1}

// Quando candidato cancela
DecrementApplicants(ctx, jobID) // $inc: {applicants: -1}
```

---

### Queries Otimizadas

#### Busca com Filtros

```go
func (r *MongoRepository) Search(ctx context.Context, filters SearchFilters) ([]*Job, error) {
    filter := bson.M{}
    
    // Location (case-insensitive, partial match)
    if filters.Location != "" {
        filter["location"] = bson.M{
            "$regex": filters.Location,
            "$options": "i",  // case-insensitive
        }
    }
    
    // JobType (exact match, case-insensitive)
    if filters.JobType != "" {
        filter["job_type"] = bson.M{
            "$regex": "^" + filters.JobType + "$",
            "$options": "i",
        }
    }
    
    // Level (exact match, case-insensitive)
    if filters.Level != "" {
        filter["level"] = bson.M{
            "$regex": "^" + filters.Level + "$",
            "$options": "i",
        }
    }
    
    // MinSalary (>=)
    if filters.MinSalary > 0 {
        filter["salary"] = bson.M{"$gte": filters.MinSalary}
    }
    
    cursor, err := r.collection.Find(ctx, filter)
    // ...
}
```

**Exemplo de query gerada:**
```javascript
db.jobs.find({
  location: { $regex: "São Paulo", $options: "i" },
  job_type: { $regex: "^remoto$", $options: "i" },
  level: { $regex: "^senior$", $options: "i" },
  salary: { $gte: 5000 }
})
```

---

## Autenticação e Segurança

### JWT (JSON Web Token)

#### Estrutura do Token

```json
{
  "user_id": "674612fa3b2c1a4d8e9f0123",
  "user_type": "company",  // ou "candidate"
  "exp": 1732713600        // Unix timestamp (24h)
}
```

#### Geração do Token

```go
func GenerateToken(userID string, userType string) (string, error) {
    claims := jwt.MapClaims{
        "user_id":   userID,
        "user_type": userType,
        "exp":       time.Now().Add(24 * time.Hour).Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    secret := os.Getenv("JWT_SECRET")
    
    return token.SignedString([]byte(secret))
}
```

#### Validação do Token

```go
func ValidateToken(tokenString string) (*jwt.MapClaims, error) {
    secret := os.Getenv("JWT_SECRET")
    
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("método de assinatura inválido")
        }
        return []byte(secret), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return &claims, nil
    }
    
    return nil, fmt.Errorf("token inválido")
}
```

### bcrypt (Hash de Senhas)

#### Criar Hash

```go
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
    return string(bytes), err
}
```

**Custo 10**: ~100ms para gerar hash (bom balanço segurança/performance)

#### Verificar Hash

```go
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

### Fluxo de Autenticação Completo

```
1. POST /company/register
   ├─> Valida dados (CNPJ, email único)
   ├─> Hash da senha com bcrypt
   ├─> Insere no MongoDB (collection companies)
   ├─> Gera JWT token
   └─> Retorna { token, company }

2. POST /company/login
   ├─> Busca empresa por email
   ├─> Verifica senha com bcrypt
   ├─> Gera JWT token
   └─> Retorna { token, company }

3. GET /company/jobs (rota protegida)
   ├─> Middleware extrai token do header
   ├─> Valida JWT e extrai claims
   ├─> Verifica se user_type == "company"
   ├─> Injeta user_id no context
   └─> Handler usa context.Value("user_id")
```

### Middleware de Autorização

```go
// AuthMiddleware: Valida JWT
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Extrai token do header Authorization
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Token não fornecido", 401)
            return
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // Valida token
        claims, err := ValidateToken(tokenString)
        if err != nil {
            http.Error(w, "Token inválido", 401)
            return
        }
        
        // Injeta dados no context
        ctx := context.WithValue(r.Context(), "user_id", (*claims)["user_id"])
        ctx = context.WithValue(ctx, "user_type", (*claims)["user_type"])
        
        next(w, r.WithContext(ctx))
    }
}

// CompanyOnly: Permite apenas empresas
func CompanyOnly(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userType := r.Context().Value("user_type").(string)
        if userType != "company" {
            http.Error(w, "Acesso negado", 403)
            return
        }
        next(w, r)
    }
}

// CandidateOnly: Permite apenas candidatos
func CandidateOnly(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userType := r.Context().Value("user_type").(string)
        if userType != "candidate" {
            http.Error(w, "Acesso negado", 403)
            return
        }
        next(w, r)
    }
}
```

### Proteção de Rotas

```go
// Rota pública
mux.HandleFunc("/jobs", jobsHandler.List)

// Rota protegida (qualquer usuário autenticado)
mux.HandleFunc("/profile", AuthMiddleware(profileHandler.Get))

// Rota apenas para empresas
mux.HandleFunc("/company/jobs", 
    AuthMiddleware(CompanyOnly(jobsHandler.CreateJob)))

// Rota apenas para candidatos
mux.HandleFunc("/candidate/applications",
    AuthMiddleware(CandidateOnly(applicationsHandler.Apply)))
```

---

## Configuração

### Variáveis de Ambiente (.env)

```env
# Servidor
PORT=8080                           # Porta HTTP
ENVIRONMENT=development             # development | production

# MongoDB
DATABASE_URL=mongodb+srv://user:pass@cluster.mongodb.net/?appName=EmpregaBem

# JWT
JWT_SECRET=sua_chave_super_segura_minimo_32_caracteres

# CORS
CORS_ORIGINS=http://localhost:5173,http://localhost:3000

# Email (opcional - futuro)
EMAIL_ENABLED=false
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=email@gmail.com
SMTP_PASSWORD=senha_app
EMAIL_FROM=noreply@empregabem.com
```

### Carregamento de Variáveis

```go
import "github.com/joho/godotenv"

func main() {
    // Carrega .env
    if err := godotenv.Load(); err != nil {
        log.Fatal("Erro ao carregar .env")
    }
    
    // Acessa variáveis
    port := os.Getenv("PORT")
    dbURL := os.Getenv("DATABASE_URL")
    jwtSecret := os.Getenv("JWT_SECRET")
    
    // Validações
    if jwtSecret == "" || len(jwtSecret) < 32 {
        log.Fatal("JWT_SECRET deve ter no mínimo 32 caracteres")
    }
}
```

---

## Deploy

### Docker

#### Dockerfile Otimizado (Multi-stage)

```dockerfile
# Stage 1: Build
FROM golang:1.25-alpine AS builder

# Instalar dependências de build
RUN apk add --no-cache git

WORKDIR /app

# Copiar go.mod e go.sum primeiro (cache de layers)
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fonte
COPY . .

# Compilar (binary estático)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o empregabem cmd/api/main.go

# Stage 2: Runtime
FROM alpine:latest

# Instalar CA certificates (para HTTPS)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar binary do stage de build
COPY --from=builder /app/empregabem .

# Copiar .env (ou usar env vars do docker)
COPY .env .

# Expor porta
EXPOSE 8080

# Comando de execução
CMD ["./empregabem"]
```

#### docker-compose.yml

```yaml
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DATABASE_URL=${DATABASE_URL}
      - JWT_SECRET=${JWT_SECRET}
      - CORS_ORIGINS=${CORS_ORIGINS}
      - ENVIRONMENT=production
    restart: unless-stopped
    networks:
      - empregabem-network

networks:
  empregabem-network:
    driver: bridge
```

#### Comandos Docker

```bash
# Build
docker build -t empregabem-api .

# Run
docker run -d \
  -p 8080:8080 \
  --env-file .env \
  --name empregabem \
  empregabem-api

# Logs
docker logs -f empregabem

# Stop
docker stop empregabem

# Remove
docker rm empregabem
```

### Heroku

```bash
# 1. Login
heroku login

# 2. Criar app
heroku create empregabem-api

# 3. Configurar variáveis
heroku config:set PORT=8080
heroku config:set DATABASE_URL="mongodb+srv://..."
heroku config:set JWT_SECRET="sua_chave_super_segura"
heroku config:set CORS_ORIGINS="https://seusite.com"
heroku config:set ENVIRONMENT=production

# 4. Deploy
git push heroku main

# 5. Ver logs
heroku logs --tail
```

### AWS EC2

```bash
# 1. Conectar ao EC2
ssh -i key.pem ubuntu@ec2-ip

# 2. Instalar Go
wget https://go.dev/dl/go1.25.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 3. Clonar repositório
git clone <repo-url>
cd api-empregabem

# 4. Configurar .env
nano .env
# (editar variáveis)

# 5. Compilar
go build -o empregabem cmd/api/main.go

# 6. Rodar com systemd (serviço)
sudo nano /etc/systemd/system/empregabem.service
```

**empregabem.service:**
```ini
[Unit]
Description=EmpregaBem API
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/api-empregabem
ExecStart=/home/ubuntu/api-empregabem/empregabem
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

```bash
# 7. Iniciar serviço
sudo systemctl daemon-reload
sudo systemctl enable empregabem
sudo systemctl start empregabem

# 8. Ver status
sudo systemctl status empregabem

# 9. Ver logs
sudo journalctl -u empregabem -f
```

---

## Testes

### Testes Unitários (Exemplo)

```go
package jobs_test

import (
    "context"
    "testing"
    "empregabemapi/jobs"
)

func TestCreateJob(t *testing.T) {
    // Setup
    repo := jobs.NewMongoRepository(testDB)
    ctx := context.Background()
    
    job := &jobs.Job{
        Title:       "Test Job",
        Description: "Test Description",
        Company:     "Test Company",
        Location:    "Test Location",
        Salary:      5000,
        JobType:     "remoto",
        Level:       "pleno",
    }
    
    // Execute
    err := repo.Create(ctx, job)
    
    // Assert
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    
    if job.ID.IsZero() {
        t.Error("Expected job ID to be set")
    }
    
    if job.Views != 0 {
        t.Errorf("Expected views to be 0, got %d", job.Views)
    }
}
```

### Testes de Integração (cURL)

```bash
# Health check
curl http://localhost:8080/api

# Registrar empresa
TOKEN=$(curl -s -X POST http://localhost:8080/company/register \
  -H "Content-Type: application/json" \
  -d '{
    "cnpj": "12345678901234",
    "name": "Test Company",
    "email": "test@test.com",
    "password": "senha123",
    "location": "São Paulo, SP"
  }' | jq -r '.token')

echo "Token: $TOKEN"

# Criar vaga
curl -X POST http://localhost:8080/company/jobs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Desenvolvedor Go",
    "description": "Vaga para dev Go",
    "company": "Test Company",
    "location": "São Paulo, SP",
    "salary": 8000,
    "job_type": "remoto",
    "level": "pleno"
  }'

# Listar vagas
curl http://localhost:8080/jobs

# Buscar com filtros
curl "http://localhost:8080/jobs?level=pleno&minSalary=5000"
```

---

## Performance

### Otimizações Implementadas

1. **MongoDB Indexes**: Queries rápidas
2. **Connection Pooling**: MongoDB driver gerencia automaticamente
3. **Context Timeout**: 5s para todas as operações
4. **Operações Atômicas**: $inc para contadores
5. **Desnormalização**: Nome da empresa na vaga (evita JOIN)

### Métricas Esperadas

- **Health Check**: < 5ms
- **Login**: < 100ms (bcrypt)
- **Listar Vagas**: < 50ms (com índices)
- **Busca com Filtros**: < 100ms
- **Criar Candidatura**: < 50ms

### Melhorias Futuras

- [ ] Redis para cache de listagens
- [ ] Paginação (limit + skip)
- [ ] Rate limiting (golang.org/x/time/rate)
- [ ] Compressão gzip
- [ ] CDN para assets estáticos

---

## Troubleshooting

### Problemas Comuns

#### 1. "Token inválido"

**Causa**: JWT_SECRET incorreto ou token expirado

**Solução**:
```bash
# Verificar JWT_SECRET no .env
cat .env | grep JWT_SECRET

# Fazer login novamente para obter novo token
curl -X POST http://localhost:8080/company/login \
  -H "Content-Type: application/json" \
  -d '{"email": "seu@email.com", "password": "senha"}'
```

#### 2. "Erro ao conectar MongoDB"

**Causa**: DATABASE_URL incorreto ou MongoDB fora do ar

**Solução**:
```bash
# Testar conexão
mongosh "mongodb+srv://user:pass@cluster.mongodb.net/"

# Verificar whitelist de IPs no MongoDB Atlas
# Dashboard > Network Access > Add IP Address
```

#### 3. "Candidatura duplicada"

**Causa**: Mesmo candidato tentando se candidatar duas vezes

**Solução**: Verificar no frontend se já existe candidatura antes de enviar

#### 4. "Contador de candidatos incorreto"

**Causa**: Vagas criadas antes da implementação dos contadores

**Solução**:
```bash
# Rodar endpoint de manutenção
curl -X POST http://localhost:8080/maintenance/fix-counters
```

#### 5. "Views incrementando duas vezes"

**Causa**: React Strict Mode renderiza componentes duas vezes

**Solução**: Usar endpoint POST /jobs/{id}/view apenas quando necessário (não no GET)

---

## Logs e Monitoramento

### Adicionar Logs Estruturados

```go
import "log/slog"

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    
    logger.Info("Servidor iniciado", 
        "port", port,
        "environment", os.Getenv("ENVIRONMENT"))
}
```

### Health Checks

```go
func HealthCheck(w http.ResponseWriter, r *http.Request) {
    // Verificar MongoDB
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    err := db.Client().Ping(ctx, nil)
    if err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "unhealthy",
            "mongodb": "down",
        })
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
        "mongodb": "up",
    })
}
```

---

## Referências

- [Go Documentation](https://golang.org/doc/)
- [MongoDB Go Driver](https://pkg.go.dev/go.mongodb.org/mongo-driver)
- [JWT Go Library](https://github.com/golang-jwt/jwt)
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)

---

**Última atualização**: 27 de novembro de 2025
