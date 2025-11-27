# 📚 Documentação das Rotas da API - EmpregaBem

## 🌐 Base URL
```
http://localhost:8080
```

---

## 🔓 ROTAS PÚBLICAS

### 1. Health Check
```http
GET /api
```
Verifica se a API está funcionando.

**Resposta:**
```json
{
  "mensagem": "API funcionando!"
}
```

---

### 2. Registrar Empresa
```http
POST /company/register
```

**Body:**
```json
{
  "cnpj": "12345678901234",
  "name": "Tech Solutions LTDA",
  "email": "contato@techsolutions.com",
  "password": "senha123",
  "location": "São Paulo, SP",
  "website": "https://techsolutions.com",
  "about": "Empresa de tecnologia focada em soluções inovadoras"
}
```

**Resposta (201):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "company": {
    "id": "674612fa3b2c1a4d8e9f0123",
    "cnpj": "12345678901234",
    "name": "Tech Solutions LTDA",
    "email": "contato@techsolutions.com",
    "location": "São Paulo, SP",
    "website": "https://techsolutions.com",
    "about": "Empresa de tecnologia focada em soluções inovadoras",
    "created_at": "2024-11-26T10:00:00Z"
  }
}
```

**Validações:**
- CNPJ: 14 dígitos numéricos, único
- Email: formato válido, único
- Password: mínimo 6 caracteres
- Name, location: obrigatórios

---

### 3. Login Empresa
```http
POST /company/login
```

**Body:**
```json
{
  "email": "contato@techsolutions.com",
  "password": "senha123"
}
```

**Resposta (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "company": {
    "id": "674612fa3b2c1a4d8e9f0123",
    "name": "Tech Solutions LTDA",
    "email": "contato@techsolutions.com"
  }
}
```

---

### 4. Registrar Candidato
```http
POST /candidate/register
```

**Body:**
```json
{
  "name": "João Silva",
  "email": "joao.silva@email.com",
  "password": "senha123",
  "phone": "11999999999",
  "location": "São Paulo, SP",
  "linkedin": "https://linkedin.com/in/joaosilva",
  "portfolio": "https://joaosilva.dev",
  "bio": "Desenvolvedor Full Stack com 3 anos de experiência",
  "skills": ["JavaScript", "React", "Node.js", "MongoDB"],
  "experience": "5 anos de experiência em desenvolvimento web"
}
```

**Resposta (201):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "candidate": {
    "id": "674612fa3b2c1a4d8e9f0124",
    "name": "João Silva",
    "email": "joao.silva@email.com",
    "phone": "11999999999",
    "location": "São Paulo, SP",
    "linkedin": "https://linkedin.com/in/joaosilva",
    "portfolio": "https://joaosilva.dev",
    "bio": "Desenvolvedor Full Stack com 3 anos de experiência",
    "skills": ["JavaScript", "React", "Node.js", "MongoDB"],
    "experience": "5 anos de experiência em desenvolvimento web",
    "created_at": "2024-11-26T10:00:00Z"
  }
}
```

---

### 5. Login Candidato
```http
POST /candidate/login
```

**Body:**
```json
{
  "email": "joao.silva@email.com",
  "password": "senha123"
}
```

**Resposta (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "candidate": {
    "id": "674612fa3b2c1a4d8e9f0124",
    "name": "João Silva",
    "email": "joao.silva@email.com"
  }
}
```

---

### 6. Listar Todas as Vagas
```http
GET /jobs
```

Lista todas as vagas ativas. Suporta filtros via query parameters.

**Query Parameters (opcionais):**
- `location` - Busca parcial (ex: "São Paulo")
- `jobType` - Tipo: "remoto", "presencial", "híbrido"
- `level` - Nível: "junior", "pleno", "senior"
- `minSalary` - Salário mínimo (ex: 3000)

**Exemplos:**
```http
GET /jobs
GET /jobs?level=senior
GET /jobs?level=senior&minSalary=5000
GET /jobs?location=São Paulo&jobType=remoto
GET /jobs?location=remoto&level=pleno&minSalary=4000
```

**Resposta (200):**
```json
{
  "vagas": [
    {
      "id": "674612fa3b2c1a4d8e9f0125",
      "company_id": "674612fa3b2c1a4d8e9f0123",
      "title": "Desenvolvedor Full Stack",
      "description": "Desenvolvimento de aplicações web modernas",
      "company": "Tech Solutions LTDA",
      "location": "São Paulo, SP",
      "salary": 8000,
      "job_type": "híbrido",
      "level": "pleno",
      "requirements": ["JavaScript", "React", "Node.js", "MongoDB"],
      "benefits": ["Vale-refeição", "Vale-transporte", "Plano de saúde"],
      "is_active": true,
      "views": 156,
      "applicants": 23,
      "priority": 0,
      "created_at": "2024-11-26T10:00:00Z",
      "updated_at": "2024-11-26T10:00:00Z"
    }
  ]
}
```

---

### 7. Ver Detalhes de uma Vaga
```http
GET /jobs/{id}
```

Retorna os detalhes de uma vaga específica. **Não incrementa** visualizações automaticamente.

**Resposta (200):**
```json
{
  "id": "674612fa3b2c1a4d8e9f0125",
  "company_id": "674612fa3b2c1a4d8e9f0123",
  "title": "Desenvolvedor Full Stack",
  "description": "Desenvolvimento de aplicações web modernas",
  "company": "Tech Solutions LTDA",
  "location": "São Paulo, SP",
  "salary": 8000,
  "job_type": "híbrido",
  "level": "pleno",
  "requirements": ["JavaScript", "React", "Node.js", "MongoDB"],
  "benefits": ["Vale-refeição", "Vale-transporte", "Plano de saúde"],
  "is_active": true,
  "views": 156,
  "applicants": 23,
  "priority": 0,
  "created_at": "2024-11-26T10:00:00Z",
  "updated_at": "2024-11-26T10:00:00Z"
}
```

---

### 8. Registrar Visualização
```http
POST /jobs/{id}/view
```

Incrementa o contador de visualizações da vaga. Use este endpoint quando o usuário realmente visualizar a vaga (ex: abrir a página de detalhes).

**Resposta (200):**
```json
{
  "mensagem": "Visualização registrada com sucesso"
}
```

---

## 🔐 ROTAS PROTEGIDAS - EMPRESAS

Todas as rotas abaixo requerem:
- Header: `Authorization: Bearer TOKEN`
- Token de uma conta **empresa**

---

### 9. Criar Vaga
```http
POST /company/jobs
```

**Body:**
```json
{
  "title": "Desenvolvedor Full Stack",
  "description": "Desenvolvimento de aplicações web modernas usando React e Node.js",
  "company": "Tech Solutions LTDA",
  "location": "São Paulo, SP",
  "salary": 8000,
  "job_type": "híbrido",
  "level": "pleno",
  "requirements": ["JavaScript", "React", "Node.js", "MongoDB"],
  "benefits": ["Vale-refeição", "Vale-transporte", "Plano de saúde"],
  "priority": 0
}
```

**Validações:**
- `title`, `description`, `company`, `location`: obrigatórios
- `job_type`: "remoto", "presencial" ou "híbrido"
- `level`: "junior", "pleno" ou "senior"
- `priority`: 0 (normal) ou 1 (destaque)
- Vaga criada como ativa (`is_active: true`) por padrão
- Contadores inicializados em 0 (`views: 0`, `applicants: 0`)

**Resposta (201):**
```json
{
  "mensagem": "Vaga criada com sucesso",
  "vaga": {
    "id": "674612fa3b2c1a4d8e9f0125",
    "company_id": "674612fa3b2c1a4d8e9f0123",
    "title": "Desenvolvedor Full Stack",
    "description": "Desenvolvimento de aplicações web modernas",
    "company": "Tech Solutions LTDA",
    "location": "São Paulo, SP",
    "salary": 8000,
    "job_type": "híbrido",
    "level": "pleno",
    "requirements": ["JavaScript", "React", "Node.js", "MongoDB"],
    "benefits": ["Vale-refeição", "Vale-transporte", "Plano de saúde"],
    "is_active": true,
    "views": 0,
    "applicants": 0,
    "priority": 0,
    "created_at": "2024-11-26T10:00:00Z",
    "updated_at": "2024-11-26T10:00:00Z"
  }
}
```

---

### 10. Listar Vagas da Empresa
```http
GET /company/jobs
```

Lista todas as vagas criadas pela empresa autenticada (ativas e inativas).

**Resposta (200):**
```json
{
  "vagas": [
    {
      "id": "674612fa3b2c1a4d8e9f0125",
      "title": "Desenvolvedor Full Stack",
      "company": "Tech Solutions LTDA",
      "location": "São Paulo, SP",
      "salary": 8000,
      "job_type": "híbrido",
      "level": "pleno",
      "is_active": true,
      "views": 156,
      "applicants": 23,
      "created_at": "2024-11-26T10:00:00Z"
    }
  ]
}
```

---

### 11. Atualizar Vaga
```http
PUT /company/jobs/{id}
```

Atualiza os dados de uma vaga. Apenas a empresa que criou a vaga pode editá-la.

**Body:**
```json
{
  "title": "Desenvolvedor Full Stack Sênior",
  "description": "Desenvolvimento de aplicações web complexas",
  "location": "São Paulo, SP - Híbrido",
  "salary": 12000,
  "job_type": "híbrido",
  "level": "senior",
  "requirements": ["JavaScript", "React", "Node.js", "MongoDB", "Docker"],
  "benefits": ["Vale-refeição", "Vale-transporte", "Plano de saúde", "Gympass"],
  "is_active": true,
  "priority": 1
}
```

**Resposta (200):**
```json
{
  "mensagem": "Vaga atualizada com sucesso"
}
```

---

### 12. Ativar/Desativar Vaga
```http
PATCH /company/jobs/{id}/status
```

Alterna o status de ativo/inativo da vaga.

**Resposta (200):**
```json
{
  "mensagem": "Status da vaga atualizado com sucesso",
  "is_active": false
}
```

---

### 13. Excluir Vaga
```http
DELETE /company/jobs/{id}
```

Remove permanentemente uma vaga. Apenas a empresa que criou pode excluir.

**Resposta (200):**
```json
{
  "mensagem": "Vaga excluída com sucesso"
}
```

---

### 14. Listar Candidatos de uma Vaga
```http
GET /company/jobs/{id}/applicants
```

Lista todos os candidatos que se candidataram a uma vaga específica da empresa.

**Resposta (200):**
```json
{
  "candidaturas": [
    {
      "id": "674612fa3b2c1a4d8e9f0126",
      "job_id": "674612fa3b2c1a4d8e9f0125",
      "candidate_id": "674612fa3b2c1a4d8e9f0124",
      "candidate_name": "João Silva",
      "candidate_email": "joao@email.com",
      "candidate_phone": "11999999999",
      "candidate_linkedin": "https://linkedin.com/in/joaosilva",
      "candidate_portfolio": "https://joaosilva.dev",
      "candidate_skills": ["JavaScript", "React", "Node.js"],
      "status": "pending",
      "applied_at": "2024-11-26T11:00:00Z",
      "viewed_at": null,
      "updated_at": "2024-11-26T11:00:00Z"
    }
  ]
}
```

**Status possíveis:**
- `pending` - Candidatura enviada, aguardando análise
- `viewed` - Empresa visualizou o perfil
- `in_review` - Em análise
- `shortlisted` - Pré-selecionado
- `interview` - Agendado para entrevista
- `rejected` - Rejeitado
- `accepted` - Aprovado/Contratado

---

### 15. Atualizar Status de Candidatura
```http
PATCH /company/applications/{id}/status
```

Atualiza o status de uma candidatura específica.

**Body:**
```json
{
  "status": "interview"
}
```

**Status válidos:**
- `pending`, `viewed`, `in_review`, `shortlisted`, `interview`, `rejected`, `accepted`

**Resposta (200):**
```json
{
  "mensagem": "Status atualizado com sucesso"
}
```

---

## 🔐 ROTAS PROTEGIDAS - CANDIDATOS

Todas as rotas abaixo requerem:
- Header: `Authorization: Bearer TOKEN`
- Token de uma conta **candidato**

---

### 16. Candidatar-se a uma Vaga
```http
POST /candidate/applications
```

Envia candidatura para uma vaga. Não permite candidaturas duplicadas.

**Body:**
```json
{
  "job_id": "674612fa3b2c1a4d8e9f0125"
}
```

**Resposta (201):**
```json
{
  "mensagem": "Candidatura enviada com sucesso",
  "application_id": "674612fa3b2c1a4d8e9f0126"
}
```

**Erros:**
- 400: Candidatura duplicada
- 404: Vaga não encontrada ou inativa

---

### 17. Listar Minhas Candidaturas
```http
GET /candidate/applications
```

Lista todas as candidaturas do candidato autenticado.

**Resposta (200):**
```json
{
  "candidaturas": [
    {
      "id": "674612fa3b2c1a4d8e9f0126",
      "job_id": "674612fa3b2c1a4d8e9f0125",
      "job_title": "Desenvolvedor Full Stack",
      "company_name": "Tech Solutions LTDA",
      "job_location": "São Paulo, SP",
      "job_salary": 8000,
      "status": "in_review",
      "applied_at": "2024-11-26T11:00:00Z",
      "viewed_at": "2024-11-26T14:00:00Z",
      "updated_at": "2024-11-26T15:00:00Z"
    }
  ]
}
```

---

### 18. Cancelar Candidatura
```http
DELETE /candidate/applications/{id}
```

Cancela uma candidatura. **Só é permitido cancelar se o status ainda for "pending"**.

**Resposta (200):**
```json
{
  "mensagem": "Candidatura cancelada com sucesso"
}
```

**Erros:**
- 400: Não é possível cancelar candidatura com status alterado pela empresa
- 403: Você não pode cancelar esta candidatura
- 404: Candidatura não encontrada

---

### 19. Salvar Vaga como Favorita
```http
POST /candidate/saved-jobs
```

Adiciona uma vaga aos favoritos do candidato.

**Body:**
```json
{
  "job_id": "674612fa3b2c1a4d8e9f0125"
}
```

**Resposta (201):**
```json
{
  "mensagem": "Vaga salva com sucesso"
}
```

**Erros:**
- 400: Vaga já está nos favoritos
- 404: Vaga não encontrada

---

### 20. Listar Vagas Favoritas
```http
GET /candidate/saved-jobs
```

Lista todas as vagas salvas pelo candidato.

**Resposta (200):**
```json
{
  "saved_jobs": [
    {
      "id": "674612fa3b2c1a4d8e9f0127",
      "candidate_id": "674612fa3b2c1a4d8e9f0124",
      "job_id": "674612fa3b2c1a4d8e9f0125",
      "job": {
        "id": "674612fa3b2c1a4d8e9f0125",
        "title": "Desenvolvedor Full Stack",
        "company": "Tech Solutions LTDA",
        "location": "São Paulo, SP",
        "salary": 8000,
        "job_type": "híbrido",
        "level": "pleno",
        "description": "Desenvolvimento de aplicações web modernas",
        "requirements": ["JavaScript", "React", "Node.js"],
        "benefits": ["Vale-refeição", "Vale-transporte"],
        "is_active": true,
        "views": 156,
        "applicants": 23,
        "created_at": "2024-11-26T10:00:00Z"
      },
      "saved_at": "2024-11-26T16:00:00Z"
    }
  ]
}
```

---

### 21. Remover Vaga dos Favoritos
```http
DELETE /candidate/saved-jobs/{job_id}
```

Remove uma vaga dos favoritos.

**Resposta (200):**
```json
{
  "mensagem": "Vaga removida dos favoritos"
}
```

**Erros:**
- 404: Vaga não está nos favoritos

---

## 🛠 ROTA DE MANUTENÇÃO

### 22. Corrigir Contadores de Vagas
```http
POST /maintenance/fix-counters
```

Inicializa os campos `views` e `applicants` em vagas que não os possuem. Útil para corrigir vagas criadas antes da implementação dos contadores.

**Resposta (200):**
```json
{
  "mensagem": "Contadores corrigidos com sucesso",
  "vagas_atualizadas": 15
}
```

---

## 🔐 Autenticação

Todas as rotas protegidas requerem um token JWT no header:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

O token é retornado nas rotas de registro e login e tem validade de **24 horas**.

---

## ⚠️ Códigos de Erro

| Código | Descrição |
|--------|-----------|
| 200 | OK - Requisição bem-sucedida |
| 201 | Created - Recurso criado com sucesso |
| 400 | Bad Request - Dados inválidos ou requisição malformada |
| 401 | Unauthorized - Token ausente, inválido ou expirado |
| 403 | Forbidden - Sem permissão para acessar este recurso |
| 404 | Not Found - Recurso não encontrado |
| 409 | Conflict - Conflito (ex: email já cadastrado) |
| 500 | Internal Server Error - Erro interno do servidor |

---

## 📝 Notas Importantes

1. **CNPJ** deve ter exatamente 14 dígitos numéricos
2. **Email** deve ser único para empresas e candidatos
3. **Senhas** são criptografadas com bcrypt antes de serem armazenadas
4. **Tokens JWT** expiram em 24 horas
5. **Vagas inativas** não aparecem na listagem pública
6. **Candidaturas duplicadas** não são permitidas (mesmo candidato + mesma vaga)
7. **Candidaturas** só podem ser canceladas se o status for "pending"
8. **Empresas** só podem editar/excluir suas próprias vagas
9. **Contadores** (views, applicants) são atualizados atomicamente no MongoDB
10. **Status de candidaturas** só pode ser alterado pela empresa

---

## 🚀 Exemplos de Uso

### Fluxo Completo - Empresa

```bash
# 1. Registrar empresa
curl -X POST http://localhost:8080/company/register \
  -H "Content-Type: application/json" \
  -d '{
    "cnpj": "12345678901234",
    "name": "Tech Solutions",
    "email": "tech@email.com",
    "password": "senha123",
    "location": "São Paulo, SP"
  }'

# Resposta contém o token
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 2. Criar vaga
curl -X POST http://localhost:8080/company/jobs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Desenvolvedor Go",
    "description": "Vaga para dev Go",
    "company": "Tech Solutions",
    "location": "São Paulo, SP",
    "salary": 8000,
    "job_type": "remoto",
    "level": "pleno"
  }'

# 3. Listar minhas vagas
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/company/jobs

# 4. Ver candidatos de uma vaga
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/company/jobs/{job_id}/applicants
```

### Fluxo Completo - Candidato

```bash
# 1. Registrar candidato
curl -X POST http://localhost:8080/candidate/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "João Silva",
    "email": "joao@email.com",
    "password": "senha123",
    "phone": "11999999999",
    "location": "São Paulo, SP",
    "skills": ["JavaScript", "React", "Node.js"]
  }'

TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 2. Buscar vagas
curl "http://localhost:8080/jobs?level=pleno&minSalary=5000"

# 3. Ver detalhes de uma vaga
curl http://localhost:8080/jobs/{job_id}

# 4. Registrar visualização
curl -X POST http://localhost:8080/jobs/{job_id}/view

# 5. Candidatar-se
curl -X POST http://localhost:8080/candidate/applications \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"job_id": "674612fa3b2c1a4d8e9f0125"}'

# 6. Salvar nos favoritos
curl -X POST http://localhost:8080/candidate/saved-jobs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"job_id": "674612fa3b2c1a4d8e9f0125"}'

# 7. Ver minhas candidaturas
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/candidate/applications

# 8. Ver meus favoritos
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/candidate/saved-jobs
```

---

**Desenvolvido com ❤️ em Go**
