# Chirpy API

A simple Twitter-like backend service built with **Go**, **PostgreSQL**, and **JWT authentication**.  
It supports user registration, authentication, posting "chirps" (short messages), and updating user details.

---

## 🚀 Features

- User registration with **hashed passwords** (bcrypt)
- User login with **JWT-based authentication**
- Create, view, and fetch individual chirps
- Profanity filtering for chirps
- Update user details (email & password) with JWT validation
- Database migrations using **Goose**
- Database queries managed with **sqlc**
- Admin endpoints for metrics and reset
- Health check endpoint

---

---
## ⚙️ Setup & Installation

### 1. Clone the repository

```bash
git clone https://github.com/<your-username>/<your-repo>.git
cd <your-repo>
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Setup environment variables
Create a .env file in the project root:
``` bash
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
JWT_SECRET=your_secret_key
```

### 4. Run database migrations
```bash 
goose -dir ./migrations postgres "$DB_URL" up
```

### 5. Generate database code with sqlc
```bash 
sqlc generate
```

### 6. Start the server
```bash
go run main.go
```

Server will start at:
```bash
http://localhost:8080
```

### 📡 API Endpoints

#### Users

Register User – POST /api/users
Login – POST /api/login
Update User – PUT /api/users (JWT required)

#### Chirps

Create Chirp – POST /api/chirps
Get All Chirps – GET /api/chirps
Get Chirp by ID – GET /api/chirps/{chirp_id}

#### Admin

Metrics – GET /admin/metrics
Reset DB – POST /admin/reset

### 🛠️ Tech Stack

Go – HTTP server & API
PostgreSQL – Database
Goose – Database migrations
sqlc – Type-safe query generation
JWT – Authentication
bcrypt – Secure password hashing

### 🧪 Example Requests

Register User
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret"}'
```

Create Chirp
```bash
curl -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -d '{"body":"Hello Chirpy!","user_id":"<uuid>"}'

```

Login
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret"}'
```


