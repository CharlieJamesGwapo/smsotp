# SMS Notification System

A small self-hosted dashboard for sending SMS to your contacts in the Philippines. It sends **through an Android phone you own** (free, uses the SIM's load), with the [Semaphore](https://semaphore.co/) API as an optional fallback.

Built for small teams: one admin, up to a few hundred contacts, PH numbers.

## Features

- **Contacts and groups:** add contacts by hand or import a CSV, and organise them into groups
- **Templates:** save message templates and reuse them
- **Send or schedule:** send to contacts or whole groups now, or schedule for later
- **History and dashboard:** see what was sent, through which gateway, and what failed
- **Two gateways:** an Android phone gateway first, then Semaphore if it's enabled and the phone fails

## Stack

- **Backend:** Go, [chi](https://github.com/go-chi/chi) router, SQLite (`modernc.org/sqlite`, no CGO), JWT auth
- **Frontend:** React + TypeScript + Vite + Tailwind CSS (in `web/`)

## Run locally

```bash
# API (http://localhost:8080)
export ADMIN_PASSWORD='choose-a-strong-password'
export JWT_SECRET="$(openssl rand -hex 32)"
go run .

# Web app (http://localhost:5173), in another terminal
cd web
npm install
npm run dev
```

Log in as `admin` with your `ADMIN_PASSWORD`. If you don't set one, a random password is generated on first run and printed once in the server log. Change it from **Settings**.

Then open **Settings** and set:
- **Phone gateway URL:** the address of an SMS gateway app on your Android phone that accepts `POST /send-sms` with `{"phone": "...", "message": "..."}`. Give the phone a fixed IP on your Wi-Fi (DHCP reservation).
- **Semaphore** (optional): API key and sender name, and turn on the fallback.

## Environment variables

| Variable | What it does |
|---|---|
| `ADMIN_PASSWORD` | Password for the `admin` account, used only when the database is first created |
| `JWT_SECRET` | Secret for signing logins. **Set this in production**; without it a random one is used and everyone is logged out on restart |
| `PORT` | API port (default `8080`) |
| `FRONTEND_URL` | Allowed origin for the web app when it's hosted separately (e.g. on Vercel) |
| `VITE_API_URL` | (web) URL of the API when the web app is hosted separately |

## Deploy

`Dockerfile` builds the API into a small Alpine image. The `sms.db` SQLite file lives next to the binary, so mount a volume if your host restarts containers. `start.sh` builds both parts and can expose the app through a Cloudflare quick tunnel for testing.

## Use responsibly

Only text people who agreed to hear from you. Sending unsolicited or repeated messages can break the [Data Privacy Act of 2012](https://privacy.gov.ph/data-privacy-act/) and your network's terms.

## License

No license has been chosen yet, so all rights are reserved by the author.
