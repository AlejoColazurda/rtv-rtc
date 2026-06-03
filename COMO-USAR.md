# Cómo usar RTV-RTC

App para crear "invitaciones" con forma de remito/orden de compra, confirmarlas
(RSVP con firma) y descargarlas en PDF.

- **Backend**: Go (`backend/`) — API REST en `http://localhost:8080`
- **Frontend**: Angular (`frontend/`) — interfaz en `http://localhost:4200`

## Arranque rápido (un clic)

Hacé doble clic en **`start.bat`**. Se abren dos ventanas (backend y frontend)
y el navegador se abre solo en `http://localhost:4200`.

> La primera vez, la ventana del frontend corre `npm install` sola si falta
> `node_modules` (tarda ~1 min). Las siguientes veces arranca directo.

Para **detener** la app: cerrá las dos ventanas que se abrieron.

## Arranque manual

**Backend** (ventana 1):
```bat
cd backend
backend.exe
```

**Frontend** (ventana 2):
```bat
cd frontend
npm install        REM solo la primera vez
npx ng serve --open
```

## Modo de datos

- **Local (por defecto)**: con `DATABASE_URL` vacío en `backend/.env`, todo se
  guarda en `backend/data/invitations.json`. No necesita internet ni base de
  datos. El circuito completo (crear, editar, RSVP, PDF) funciona igual.
- **Supabase (opcional)**: completá `DATABASE_URL` en `backend/.env` con un host
  válido y reiniciá el backend. Si la conexión falla, el backend vuelve solo al
  modo local.

## Requisitos

- **Node.js ≥ 22.12** (el CLI de Angular lo exige). Probado con Node 24 LTS.
- **Go** solo si querés recompilar el backend (`go build -o backend.exe .` dentro
  de `backend/`). El `backend.exe` ya viene compilado.

## Notas

- El `backend/.env` ya no incluye credenciales de Supabase. Si reactivás esa base,
  poné un host/clave nuevos (el proyecto anterior fue dado de baja).
