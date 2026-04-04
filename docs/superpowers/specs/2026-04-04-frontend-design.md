# Frontend Design Spec — Room Booking Service

**Date:** 2026-04-04
**Status:** Draft

## Overview

React SPA фронтенд для Room Booking Service — сервиса бронирования переговорок.
Минималистичный дизайн с glassmorphism-акцентами.

### Version 1 (core — показ TA)
- LoginPage (dummyLogin)
- RoomsPage (список переговорок)
- BookingPage (просмотр слотов + бронирование)
- MyBookingsPage (свои брони + отмена)
- Базовый glassmorphism UI

### Version 2 (polish + deploy)
- AdminPage (создание переговорок, расписание, все брони)
- Улучшенные анимации, skeleton loaders, empty states
- Docker + Docker Compose integration
- Production deployment на university VM
- Screenshots для README

**Стек:**
- Vite + React 18 + TypeScript
- React Router v6 (роутинг)
- Zustand (auth state)
- TanStack Query (API caching, mutations)
- shadcn/ui + Tailwind CSS (UI компоненты)
- Framer Motion (анимации)
- Zod + React Hook Form (валидация форм)
- Sonner (toast уведомления)
- Axios (HTTP клиент)

## Architecture

### Структура проекта

```
frontend/
├── src/
│   ├── api/
│   │   ├── client.ts          # axios instance с JWT interceptor
│   │   └── hooks.ts           # React Query hooks для API
│   ├── components/
│   │   ├── ui/                # shadcn/ui компоненты
│   │   ├── layout/            # Layout, TabNav, AuthGuard
│   │   └── slots/             # CalendarGrid, SlotCard
│   ├── pages/
│   │   ├── LoginPage.tsx
│   │   ├── RoomsPage.tsx
│   │   ├── BookingPage.tsx
│   │   ├── MyBookingsPage.tsx
│   │   └── AdminPage.tsx
│   ├── store/
│   │   └── authStore.ts       # Zustand: token, user, role
│   ├── types/
│   │   └── index.ts           # TypeScript типы из API
│   ├── lib/
│   │   └── utils.ts           # cn() утилита
│   ├── App.tsx                # Роутинг + Layout
│   └── main.tsx               # Entry point
├── components.json
├── tailwind.config.ts
├── vite.config.ts
└── package.json
```

### API Integration

- Бэкенд на `http://localhost:8080`
- Все запросы через axios instance с baseURL
- JWT interceptor: берёт token из Zustand store, добавляет `Authorization: Bearer`
- При 401 → logout + redirect на `/login`
- React Query hooks: `useRooms`, `useSlots`, `useCreateBooking`, `useMyBookings`, `useCancelBooking`, `useCreateRoom`, `useCreateSchedule`, `useAllBookings`

### Auth Flow

- `dummyLogin`: POST `/dummyLogin` с `{role: "admin"|"user"}`
- Response: `{token: "jwt"}`
- Token сохраняется в Zustand store + `localStorage`
- Роуты защищены:
  - `/login` — публичный
  - `/rooms`, `/booking/:id`, `/my-bookings` — авторизованный
  - `/admin` — только admin role

### State Management

- **Zustand** (`authStore`): token, user (id, role), login(), logout()
- **React Query**: все серверные данные (rooms, slots, bookings)
- **Local state**: формы, UI toggles, date picker

## Visual Design

### Color Palette

| Token | Value | Usage |
|-------|-------|-------|
| bg-primary | `slate-50` | Фон страницы |
| glass-bg | `white/70` | Glassmorphism карточки |
| glass-border | `white/20` | Границы glassmorphism |
| accent | `indigo-500` | Primary actions |
| accent-hover | `indigo-500/20` | Hover на glassmorphism |
| success | `emerald-500` | Свободные слоты, успех |
| danger | `rose-500` | Занятые слоты, ошибки |
| warning | `amber-500` | Предупреждения |
| text-primary | `slate-900` | Заголовки |
| text-secondary | `slate-600` | Body текст |
| text-muted | `slate-400` | Secondary текст |

### Glassmorphism Pattern

```css
.glass {
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(16px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
}
```

Tailwind: `bg-white/70 backdrop-blur-lg border border-white/20 shadow-lg`

### Typography

- Font: Inter (Google Fonts)
- Headings: `font-semibold tracking-tight`
- Body: `font-normal leading-relaxed`

### Animations

- Framer Motion для page transitions и модалок
- Hover: `scale-[1.02] transition-transform duration-200`
- Loading: skeleton shimmer
- Toast: slide-in справа сверху

## Pages

### 1. LoginPage (`/login`)

**Layout:** Центрированная glassmorphism-карточка на градиентном фоне (`bg-gradient-to-br from-indigo-100 via-white to-purple-100`)

**Компоненты:**
- Заголовок: "Room Booking Service"
- Подзаголовок: "Выберите роль для входа"
- Кнопка "Войти как Admin" (indigo, full width)
- Кнопка "Войти как User" (outline, full width)

**Flow:**
1. Клик → POST `/dummyLogin` с `{role}`
2. Сохранить token + user в Zustand
3. Редирект: admin → `/admin`, user → `/rooms`

### 2. RoomsPage (`/rooms`)

**Layout:** Header с табами + сетка карточек

**Header (TabNav):**
- Табы: "Переговорки" (active) | "Мои брони" | "Админ" (только admin)
- Справа: аватар/роль + кнопка "Выход"
- Glassmorphism: `bg-white/50 backdrop-blur-md sticky top-0`

**Content:**
- Заголовок: "Переговорки"
- Сетка карточек: `grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6`
- Карточка переговорки (glassmorphism):
  - Название (h3)
  - Описание (muted text)
  - Вместимость: "до N человек"
  - Кнопка "Забронировать" (indigo)
- Клик на карточку → `/booking/:roomId`

**Empty state:** "Переговорки ещё не созданы. Обратитесь к администратору."

### 3. BookingPage (`/booking/:roomId`)

**Layout:** Header + Date picker + Calendar grid

**Header:**
- Кнопка "← Назад" → `/rooms`
- Название переговорки (h2)

**Date Picker:**
- shadcn Calendar компонент
- По умолчанию: сегодня
- При изменении → refetch слотов

**Calendar Grid:**
- Вертикальная timeline: 08:00–20:00
- Слоты по 30 минут
- Свободный слот: `bg-emerald-500/10 backdrop-blur-sm border border-emerald-500/20`
  - Время (09:00–09:30)
  - Кнопка "Забронировать" (emerald, small)
- Занятый слот: `bg-rose-500/10 backdrop-blur-sm border border-rose-500/20`
  - Время
  - Иконка замка + "Занято" (muted)

**Booking Modal (при клике "Забронировать"):**
- Glassmorphism overlay (`bg-black/20 backdrop-blur-sm`)
- Модалка: "Забронировать переговорку"
  - Дата и время слота
  - Чекбокс: "Создать ссылку на конференцию"
  - Кнопки: "Отмена" | "Подтвердить"
- Успех → toast "Бронь создана!" → редирект `/my-bookings`
- Ошибка (slot already booked) → toast "Слот уже занят" → refetch

### 4. MyBookingsPage (`/my-bookings`)

**Layout:** Header + список карточек

**Content:**
- Заголовок: "Мои брони"
- Список броней (вертикальный, gap-4)
- Карточка брони (glassmorphism):
  - Дата и время
  - Название переговорки
  - Статус: badge (active = emerald, cancelled = rose)
  - Conference link (если есть) — кликабельная ссылка
  - Кнопка "Отменить" (outline, rose) — только для active
- Клик "Отменить" → подтверждение → POST `/cancel` → toast "Бронь отменена" → refetch

**Empty state:** "У вас нет активных броней. Забронируйте переговорку!"

### 5. AdminPage (`/admin`)

**Layout:** Header + внутренние табы

**Header:**
- Табы: "Переговорки" | "Расписание" | "Все брони"

**Tab: Переговорки**
- Форма создания:
  - Input: "Название" (required)
  - Textarea: "Описание" (optional)
  - Input: "Вместимость" (number, optional)
  - Кнопка "Создать"
- Список существующих переговорок (таблица)

**Tab: Расписание**
- Select: выбор переговорки
- Чекбоксы дней недели: Пн, Вт, Ср, Чт, Пт, Сб, Вс
- Time picker: "Время начала" / "Время окончания"
- Кнопка "Создать расписание"
- Warning: "Расписание нельзя изменить после создания"

**Tab: Все брони**
- Таблица (shadcn Table):
  - ID, Пользователь, Переговорка, Дата/Время, Статус
- Пагинация (shadcn Pagination)
- page=1, pageSize=20 по умолчанию

## Error Handling & UX

### Loading States
- Skeleton loaders на всех страницах при загрузке данных
- Button spinner при мутациях
- Page-level loader при initial render

### Error Handling
- Toast (Sonner) для всех операций:
  - Success: зелёный toast
  - Error: красный toast с сообщением от API
- Inline form errors (Zod validation)
- Network error: "Нет соединения, попробуйте снова"

### Auth Guard
- При 401 от API → logout + redirect `/login`
- Protected routes: проверка token в Zustand
- Admin routes: проверка role === "admin"

### Form Validation
- Zod schemas для всех форм
- React Hook Form + Zod resolver
- Inline errors под полями
- Submit button disabled при invalid form

## API Endpoints Mapping

| Endpoint | Method | Hook | Pages |
|----------|--------|------|-------|
| `/dummyLogin` | POST | `useDummyLogin` | Login |
| `/rooms/list` | GET | `useRooms` | Rooms, Admin |
| `/rooms/create` | POST | `useCreateRoom` | Admin |
| `/rooms/{id}/schedule/create` | POST | `useCreateSchedule` | Admin |
| `/rooms/{id}/slots/list?date=` | GET | `useSlots` | Booking |
| `/bookings/create` | POST | `useCreateBooking` | Booking |
| `/bookings/my` | GET | `useMyBookings` | MyBookings |
| `/bookings/{id}/cancel` | POST | `useCancelBooking` | MyBookings |
| `/bookings/list?page=&pageSize=` | GET | `useAllBookings` | Admin |

## Deployment

### Docker (Task 4 requirement)

Фронтенд докеризуется через multi-stage build:

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

`nginx.conf` — SPA routing (все запросы → `index.html`), proxy `/api` на бэкенд `http://backend:8080`.

### Docker Compose integration

Фронтенд добавляется в `docker-compose.yaml` рядом с бэкендом:

```yaml
frontend:
  build: ./frontend
  ports:
    - "3000:80"
  depends_on:
    - backend
```

### Production build

- `npm run build` → `dist/`
- API URL настраивается через `VITE_API_URL` env var (build-time)
- Для university VM: nginx container, Ubuntu 24.04

### README requirements (Task 5)

Финальный README репозитория должен содержать:
- Product name: "Room Booking Service"
- One-line description
- Screenshots фронтенда (login, rooms, booking grid)
- End users: сотрудники компании, администраторы офисов
- Problem: нет единого сервиса бронирования переговорок
- Solution: web-приложение с визуальным календарём слотов
- Features: implemented (backend + frontend) и planned
- Usage: как запустить через `docker-compose up`
- Deployment: Ubuntu 24.04, Docker Compose, step-by-step
