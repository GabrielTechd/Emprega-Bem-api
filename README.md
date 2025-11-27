# 🚀 EmpregaBem API

<div align="center">

![Go](https://img.shields.io/badge/Go-1.25.4-00ADD8?style=flat-square&logo=go)
![MongoDB](https://img.shields.io/badge/MongoDB-v2-47A248?style=flat-square&logo=mongodb)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

**API REST para plataforma de recrutamento**

[📚 Documentação](./DOCUMENTACAO.md) • [🔗 Rotas](./ROTAS.md)

</div>

---

## 📖 Sobre

Plataforma de recrutamento que conecta empresas e candidatos. API desenvolvida em **Go** com **MongoDB**.

**Recursos:**
- 🔐 Autenticação JWT + bcrypt
- 🔍 Busca com filtros
- ⭐ Sistema de favoritos
- 📝 Gestão de candidaturas
- 📊 Métricas de vagas

**Stack:** Go 1.25.4 • MongoDB • JWT • bcrypt

## 🚀 Quick Start

**Pré-requisitos:** Go 1.25.4+ e MongoDB

```bash
# 1. Clonar e instalar
git clone <url-do-repositorio>
cd api-empregabem
go mod download

# 2. Configurar .env
PORT=8080
DATABASE_URL=mongodb+srv://usuario:senha@cluster.mongodb.net/
JWT_SECRET=sua_chave_minimo_32_caracteres
CORS_ORIGINS=http://localhost:5173

# 3. Executar
air                        # dev (hot reload)
go run cmd/api/main.go     # produção

# 4. Testar
curl http://localhost:8080/api
```

## 📚 API

**22 endpoints** divididos em:
- Públicas (5) - Health, registro, login, listagem
- Empresas (7) - CRUD vagas, gerenciar candidatos
- Candidatos (6) - Candidaturas, favoritos
- Manutenção (1)

📖 **[Ver todas as rotas →](./ROTAS.md)**


## 🏗 Arquitetura

```
cmd/api/          → Entry point
internal/http/    → Handlers, middleware, router
companies/        → Domínio empresas
candidates/       → Domínio candidatos
jobs/             → Domínio vagas
applications/     → Domínio candidaturas
database/         → MongoDB
```

**Padrões:** Repository • Middleware • REST • DDD

## 🗄️ Banco

**Collections:** `companies` • `candidates` • `jobs` • `applications` • `saved_jobs`

📖 **[Ver detalhes →](./DOCUMENTACAO.md)**

---

<div align="center">

**Desenvolvido em Go**
**api para estudo**

[Documentação](./DOCUMENTACAO.md) • [Rotas](./ROTAS.md)

</div>
