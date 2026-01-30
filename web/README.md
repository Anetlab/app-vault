# App Vault Web UI

Enterprise-grade Vue 3 web interface for SecureVault secret management system.

## Features

- ✅ **Modern Vue 3** with Composition API & TypeScript
- ✅ **Tailwind CSS** for beautiful, responsive design
- ✅ **Pinia** for state management
- ✅ **Vue Router** for navigation with guards
- ✅ **Axios** for API communication with interceptors
- ✅ **Double-layer encryption** support
- ✅ **Military-grade security** UI/UX

## Project Structure

```
web/
├── src/
│   ├── api/              # API clients
│   │   ├── client.ts     # Axios instance with interceptors
│   │   ├── auth.ts       # Authentication endpoints
│   │   ├── secrets.ts    # Secret management endpoints
│   │   ├── servicePrincipals.ts
│   │   └── system.ts     # Health checks, audit logs
│   ├── components/       # Reusable components
│   │   ├── Sidebar.vue
│   │   ├── Header.vue
│   │   ├── ToastContainer.vue
│   │   ├── AddSecretModal.vue
│   │   └── ViewSecretModal.vue
│   ├── layouts/          # Layout components
│   │   └── MainLayout.vue
│   ├── views/            # Page components
│   │   ├── Login.vue
│   │   ├── Register.vue
│   │   ├── Dashboard.vue
│   │   ├── Secrets.vue
│   │   ├── ServicePrincipals.vue
│   │   ├── KeyRotation.vue
│   │   ├── AuditLogs.vue
│   │   └── Settings.vue
│   ├── stores/           # Pinia stores
│   │   ├── auth.ts       # Authentication state
│   │   ├── secrets.ts    # Secrets management
│   │   └── toast.ts      # Toast notifications
│   ├── types/            # TypeScript types
│   │   └── index.ts
│   ├── router/           # Vue Router configuration
│   │   └── index.ts
│   ├── App.vue           # Root component
│   ├── main.ts           # Application entry point
│   └── style.css         # Global styles with Tailwind
├── public/               # Static assets
├── .env                  # Development environment variables
├── .env.production       # Production environment variables
├── vite.config.ts        # Vite configuration
├── tailwind.config.js    # Tailwind CSS configuration
├── tsconfig.json         # TypeScript configuration
└── package.json          # Dependencies
```

## Setup

### Prerequisites

- Node.js 18+ and npm/yarn
- Backend API running on port 8888

### Installation

```bash
cd web
npm install
```

### Development

```bash
npm run dev
```

Server runs on: http://localhost:5173

### Build for Production

```bash
npm run build
```

Output directory: `dist/`

### Preview Production Build

```bash
npm run preview
```

## Environment Variables

### Development (`.env`)
```env
VITE_API_URL=http://localhost:8888/api/v1
VITE_APP_NAME=SecureVault
VITE_VERSION=2.1.0
```

### Production (`.env.production`)
```env
VITE_API_URL=/api/v1
VITE_APP_NAME=SecureVault
VITE_VERSION=2.1.0
```

## API Integration

The frontend connects to the Go backend API:

- **Base URL**: `http://localhost:8888/api/v1` (dev) or `/api/v1` (production)
- **Authentication**: JWT tokens in `Authorization` header
- **Secret Key**: Custom `X-Vault-Secret-Key` header
- **Proxy**: Vite dev server proxies `/api` requests to backend

### Authentication Flow

1. **Register**: User provides email + password → Backend returns secret key (A3 format)
2. **Login**: User provides email + password + secret key → Backend returns JWT token
3. **Token Storage**: JWT in localStorage, Secret Key in sessionStorage
4. **Auto-attach**: Axios interceptors add headers to all requests
5. **Auto-redirect**: 401 responses trigger logout and redirect to login

### Security Features

- Tokens stored securely (localStorage for JWT, sessionStorage for secret key)
- Automatic token injection via interceptors
- CSRF protection through API design
- XSS protection through Vue's template escaping
- Content Security Policy headers

## Design System

### Colors

- **Background**: `#0f172a` (dark-bg)
- **Cards**: `#1e293b` (dark-card)
- **Borders**: `#334155` (dark-border)
- **Primary**: `#06b6d4` (cyan-500)
- **Text**: `#f8fafc` (white/gray-100)

### Typography

- **Font Family**: Inter (Google Fonts)
- **Headings**: Bold, 2xl-3xl
- **Body**: Regular, base size

### Icons

Font Awesome 6 (CDN)

## Features Implemented

### ✅ Phase 1: Core Infrastructure
- Vue 3 project setup with TypeScript
- Tailwind CSS configuration
- API client with interceptors
- Authentication store (Pinia)
- Router with guards
- Toast notifications

### ✅ Phase 2: Authentication
- Login page with email/password/secret key
- Register page with secret key generation
- Secure token management
- Auto-redirect on auth changes

### ✅ Phase 3: Main Application
- Dashboard with stats
- Secrets list with search & filters
- Add/Edit secret modals
- View secret modal with copy functionality
- Secret deletion
- Category filtering

### 🔄 Phase 4: Advanced Features (Coming Soon)
- Service Principals UI
- Key Rotation management
- Audit Logs viewer
- Settings page
- Export/Import functionality
- Bulk operations

## Development Guidelines

### Code Style

- Use Composition API (`<script setup>`)
- TypeScript for all new code
- Tailwind utility classes (avoid custom CSS)
- Component naming: PascalCase files, kebab-case in templates
- Props validation with TypeScript
- Emit events with `defineEmits`

### State Management

- **Local state**: `ref()` and `reactive()` in components
- **Shared state**: Pinia stores for auth, secrets, etc.
- **API calls**: Through dedicated API client modules

### Error Handling

- API errors caught in stores
- User-friendly toast notifications
- Error states displayed in UI
- Console logging for debugging

## Testing

```bash
# Type checking
npm run type-check

# Linting (when configured)
npm run lint
```

## Deployment

### Option 1: Static Files (Nginx/Apache)

Build and serve the `dist/` folder:

```bash
npm run build
# Copy dist/ to web server
```

### Option 2: Embedded in Go Binary

Serve static files from Go backend:

```go
//go:embed web/dist
var webUI embed.FS

http.Handle("/", http.FileServer(http.FS(webUI)))
```

### Option 3: Docker

```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
```

## Browser Support

- Chrome/Edge: Last 2 versions
- Firefox: Last 2 versions
- Safari: Last 2 versions

## Performance

- **Code Splitting**: Routes lazy-loaded
- **Tree Shaking**: Vite optimizes bundle size
- **CDN**: Font Awesome from CDN
- **Minification**: Terser for production builds

## Security Considerations

- ⚠️ **Secret Key in Memory**: Only stored in sessionStorage (cleared on tab close)
- ⚠️ **JWT Expiration**: Tokens expire after 24 hours
- ⚠️ **HTTPS Required**: Production must use HTTPS
- ⚠️ **CORS Configuration**: Backend must allow frontend origin

## Troubleshooting

### API Connection Issues

```bash
# Check backend is running
curl http://localhost:8888/health

# Check Vite proxy configuration
# vite.config.ts -> server.proxy
```

### Build Errors

```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install

# Clear Vite cache
rm -rf node_modules/.vite
```

## Contributing

See [CONTRIBUTING.md](../docs/CONTRIBUTING.md) for development workflow and guidelines.

## License

Same as parent project - see [LICENSE](../LICENSE).

---

**Status**: Phase 3 Complete ✅
**Next**: Service Principals UI, Key Rotation, Audit Logs
