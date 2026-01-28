┌─────────────────────────────────────────────────────────────┐
│                    HTTP Request                             │
└───────────────────┬─────────────────────────────────────────┘
                    │ HOST
        ┌───────────┴──────────┐
  api.sonnda.com.br     app.sonnda.com.br
        │                      │
   /api/*                   /web/*
        │                      │
        ▼                      ▼
┌───────────────┐      ┌───────────────┐
│ API Handler   │      │ Web Handler   │
└───────┬───────┘      └───────┬───────┘
        │                      │
        │ erro?                │ erro?
        ▼                      ▼
┌─────────────────┐    ┌─────────────────┐
│ APIError-       │    │ WebError-       │
│ Responder(...)  │    │ Responder(...)  │
└────────┬────────┘    └────────┬────────┘
         │                      │
         │   ┌──────────────────┘
         │   │
         ▼   ▼
    ┌──────────────────────┐
    │ BaseErrorResponder   │ ← Lógica compartilhada
    │ - ToHTTP()           │   (mapping, logging, etc)
    │ - logging            │
    │ - metadados          │
    └──────────┬───────────┘
               │
        ┌──────┴──────┐
        │  Presenter  │ ← Interface Strategy
        │  .Present() │
        └──────┬──────┘
               │
       ┌───────┴────────┐
       │                │
       ▼                ▼
┌──────────────┐  ┌──────────────┐
│JSON          │  │HTML          │
│Presenter     │  │Presenter     │
│              │  │ - HTMX?      │
│c.JSON(...)   │  │c.Data(...)   │
└──────────────┘  └──────────────┘