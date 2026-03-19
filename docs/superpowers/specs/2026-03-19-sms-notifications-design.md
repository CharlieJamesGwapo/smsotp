# SMS Notification System — Design Spec

## Overview

A web-based SMS notification system for managing and sending SMS to clients in the Philippines. Built as a Go monolith serving a React SPA, using SQLite for storage and dual SMS gateways (Android phone + Semaphore API fallback).

**Scale:** <100 contacts, single admin user, PH numbers only.

## Architecture

Single Go binary serves both the REST API and the embedded React frontend.

```
┌─────────────────────────────────────────────┐
│              Go Binary (port 8080)           │
│                                              │
│  ┌──────────────┐    ┌───────────────────┐  │
│  │ React SPA    │    │ Go REST API        │  │
│  │ (embedded)   │    │ /api/*             │  │
│  └──────────────┘    └────────┬──────────┘  │
│                               │              │
│                    ┌──────────▼──────────┐   │
│                    │     SQLite DB       │   │
│                    │     sms.db          │   │
│                    └──────────┬──────────┘   │
│                               │              │
│              ┌────────────────┼────────┐     │
│              ▼                         ▼     │
│  ┌──────────────────┐  ┌─────────────────┐  │
│  │ Android Gateway  │  │ Semaphore API   │  │
│  │ (primary, free)  │  │ (fallback, opt) │  │
│  └──────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────┘
```

### Tech Stack

**Backend:**
- Go 1.22+
- Chi router (lightweight, idiomatic)
- SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- JWT authentication (`golang-jwt/jwt`)
- `embed.FS` to serve React build

**Frontend:**
- React 18 with TypeScript
- React Router for navigation
- Tailwind CSS for styling
- Axios for API calls
- React Hot Toast for notifications

### Key Design Decisions

- **Monolith:** Single binary deployment. Go embeds the React production build via `embed.FS`. During development, React and Go run separately.
- **SQLite:** No external database server needed. Single file `sms.db`. Sufficient for <100 contacts.
- **Dual gateway with optional fallback:** Android phone gateway is primary (free, uses SIM load). Semaphore is optional fallback — configurable from the settings page. System works fine with phone gateway only.
- **Static IP recommendation:** User should set a DHCP reservation for the Android phone to avoid IP changes.

## Database Schema

### admin
| Column        | Type    | Notes               |
|---------------|---------|---------------------|
| id            | INTEGER | PRIMARY KEY         |
| username      | TEXT    | UNIQUE, NOT NULL    |
| password_hash | TEXT    | NOT NULL (bcrypt)   |
| created_at    | DATETIME| DEFAULT CURRENT_TIMESTAMP |

### contacts
| Column     | Type    | Notes               |
|------------|---------|---------------------|
| id         | INTEGER | PRIMARY KEY         |
| name       | TEXT    | NOT NULL            |
| phone      | TEXT    | NOT NULL, UNIQUE    |
| email      | TEXT    | nullable            |
| notes      | TEXT    | nullable            |
| created_at | DATETIME| DEFAULT CURRENT_TIMESTAMP |

### groups
| Column      | Type    | Notes               |
|-------------|---------|---------------------|
| id          | INTEGER | PRIMARY KEY         |
| name        | TEXT    | NOT NULL, UNIQUE    |
| description | TEXT    | nullable            |
| created_at  | DATETIME| DEFAULT CURRENT_TIMESTAMP |

### contact_groups (join table)
| Column     | Type    | Notes                          |
|------------|---------|--------------------------------|
| group_id   | INTEGER | FK → groups.id, ON DELETE CASCADE |
| contact_id | INTEGER | FK → contacts.id, ON DELETE CASCADE |
| PRIMARY KEY | —      | (group_id, contact_id)         |

### templates
| Column     | Type    | Notes               |
|------------|---------|---------------------|
| id         | INTEGER | PRIMARY KEY         |
| name       | TEXT    | NOT NULL            |
| body       | TEXT    | NOT NULL            |
| created_at | DATETIME| DEFAULT CURRENT_TIMESTAMP |

Template variables use `{variable}` syntax in the body, e.g., `"Hi {name}, your OTP is {code}"`.

### messages
| Column        | Type    | Notes                              |
|---------------|---------|------------------------------------|
| id            | INTEGER | PRIMARY KEY                        |
| template_id   | INTEGER | FK → templates.id, nullable        |
| contact_id    | INTEGER | FK → contacts.id                   |
| phone         | TEXT    | NOT NULL (denormalized for history) |
| body          | TEXT    | NOT NULL (final resolved message)  |
| status        | TEXT    | "pending", "sent", "failed"        |
| gateway_used  | TEXT    | "phone", "semaphore", nullable     |
| scheduled_at  | DATETIME| nullable (null = send immediately) |
| sent_at       | DATETIME| nullable                           |
| error_message | TEXT    | nullable                           |
| created_at    | DATETIME| DEFAULT CURRENT_TIMESTAMP          |

## API Endpoints

All endpoints require JWT token in `Authorization: Bearer <token>` header, except `/api/auth/login`.

### Auth
| Method | Endpoint          | Description            |
|--------|-------------------|------------------------|
| POST   | /api/auth/login   | Login, returns JWT     |
| GET    | /api/auth/me      | Get current admin info |

### Contacts
| Method | Endpoint              | Description                  |
|--------|-----------------------|------------------------------|
| GET    | /api/contacts         | List contacts (search/filter)|
| POST   | /api/contacts         | Create contact               |
| PUT    | /api/contacts/:id     | Update contact               |
| DELETE | /api/contacts/:id     | Delete contact               |
| POST   | /api/contacts/import  | Bulk import from CSV         |

### Groups
| Method | Endpoint                    | Description              |
|--------|-----------------------------|--------------------------|
| GET    | /api/groups                 | List all groups          |
| POST   | /api/groups                 | Create group             |
| PUT    | /api/groups/:id             | Update group             |
| DELETE | /api/groups/:id             | Delete group             |
| POST   | /api/groups/:id/members     | Add contacts to group    |
| DELETE | /api/groups/:id/members     | Remove contacts from group|

### Templates
| Method | Endpoint              | Description       |
|--------|-----------------------|-------------------|
| GET    | /api/templates        | List templates    |
| POST   | /api/templates        | Create template   |
| PUT    | /api/templates/:id    | Update template   |
| DELETE | /api/templates/:id    | Delete template   |

### Messages
| Method | Endpoint               | Description                    |
|--------|------------------------|--------------------------------|
| POST   | /api/messages/send     | Send SMS now                   |
| POST   | /api/messages/schedule | Schedule SMS for later         |
| GET    | /api/messages          | Message history with filters   |
| GET    | /api/messages/:id      | Single message detail          |

### Dashboard
| Method | Endpoint              | Description                         |
|--------|-----------------------|-------------------------------------|
| GET    | /api/dashboard/stats  | Contacts count, messages stats, etc.|

### Settings
| Method | Endpoint        | Description                              |
|--------|-----------------|------------------------------------------|
| GET    | /api/settings   | Get gateway config                       |
| PUT    | /api/settings   | Update gateway IP, Semaphore key, etc.   |

## Frontend Pages

| Route              | Page                | Description                                          |
|--------------------|---------------------|------------------------------------------------------|
| /login             | Login               | Admin login form                                     |
| /dashboard         | Dashboard           | Stats: contact count, messages sent, success rate    |
| /contacts          | Contacts            | Table with search, add/edit/delete, CSV import       |
| /groups            | Groups              | Group CRUD, assign/remove contacts                   |
| /templates         | Templates           | Template CRUD with variable preview                  |
| /messages/send     | Send SMS            | Select recipients (contacts/groups), write or pick template, preview, send now or schedule |
| /messages/history  | Message History     | Logs with filters (date, status, gateway)            |
| /settings          | Settings            | Gateway IP, Semaphore API key, fallback toggle       |

**UI:** Sidebar navigation, desktop-focused, Tailwind CSS, responsive but optimized for desktop.

## SMS Sending Flow

1. User selects recipients (individual contacts and/or groups) and composes message (custom or from template)
2. Backend resolves all recipients — expands groups, deduplicates phone numbers
3. If template used, replaces variables (`{name}`, etc.) per contact
4. For each recipient:
   - Try Android phone gateway (`POST http://<phone-ip>:8080/send-sms`)
   - On success → log as "sent" with `gateway_used = "phone"`
   - On failure + Semaphore enabled → try Semaphore API
   - On Semaphore success → log as "sent" with `gateway_used = "semaphore"`
   - On all fail → log as "failed" with error message
5. 1-second delay between each SMS (configurable)

### Scheduled Messages

- Background goroutine runs every 60 seconds
- Queries messages where `scheduled_at <= now AND status = 'pending'`
- Sends through the same flow above

## Project Structure

```
smsotp/
├── main.go                  # Entry point, server setup
├── go.mod
├── go.sum
├── sms.db                   # SQLite database (created at runtime)
├── internal/
│   ├── auth/                # JWT middleware, login handler
│   ├── contacts/            # Contact CRUD handlers
│   ├── groups/              # Group CRUD handlers
│   ├── templates/           # Template CRUD handlers
│   ├── messages/            # Send, schedule, history handlers
│   ├── dashboard/           # Stats handler
│   ├── settings/            # Gateway config handler
│   ├── gateway/             # SMS gateway interface
│   │   ├── gateway.go       # Interface definition
│   │   ├── phone.go         # Android phone gateway
│   │   └── semaphore.go     # Semaphore API gateway
│   ├── scheduler/           # Background scheduled message sender
│   └── database/            # SQLite setup, migrations
├── web/                     # React frontend
│   ├── src/
│   │   ├── pages/           # Page components
│   │   ├── components/      # Shared components
│   │   ├── api/             # Axios API client
│   │   ├── hooks/           # Custom React hooks
│   │   ├── context/         # Auth context
│   │   └── App.tsx          # Router setup
│   ├── package.json
│   ├── tailwind.config.js
│   └── vite.config.ts       # Vite for fast dev builds
└── docs/
    └── superpowers/
        └── specs/
            └── 2026-03-19-sms-notifications-design.md
```

## Security

- Admin password hashed with bcrypt
- JWT tokens with 24-hour expiry
- All API routes behind auth middleware (except login)
- Semaphore API key stored in SQLite settings (server-side only, never exposed to frontend)
- Phone gateway accessible only on local network

## Configuration

Settings stored in SQLite `settings` table (key-value):

| Key                    | Default               | Description                    |
|------------------------|-----------------------|--------------------------------|
| phone_gateway_url      | http://192.168.100.165:8080 | Android phone gateway URL |
| semaphore_api_key      | (empty)               | Semaphore API key              |
| semaphore_enabled      | false                 | Enable Semaphore fallback      |
| semaphore_sender_name  | (empty)               | Semaphore sender name          |
| sms_delay_ms           | 1000                  | Delay between SMS in ms        |
| admin_username         | admin                 | Default admin username         |
