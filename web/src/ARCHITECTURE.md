# Frontend Architecture

This document describes the architecture and organization of the Habibi-Go frontend application.

## Overview

The frontend is a React application built with TypeScript and Vite. It follows a feature-based architecture pattern where functionality is organized by feature modules rather than technical layers.

## Technology Stack

- **React 19** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **Zustand** - Primary state management
- **React Query** - Server state and caching
- **React Router** - Client-side routing
- **Tailwind CSS** - Styling
- **WebSocket** - Real-time communication

## Directory Structure

```
src/
├── api/                    # API clients and configurations
│   ├── client.ts          # HTTP client with axios
│   └── websocket.ts       # WebSocket client
├── app/                    # Application setup and configuration
│   ├── App.tsx            # Main application component
│   ├── store.ts           # Redux store setup (legacy)
│   ├── hooks.ts           # Redux hooks (legacy)
│   └── middleware/        # Redux middleware
│       └── websocket.ts   # WebSocket Redux middleware
├── components/            # Shared/common components
│   ├── Layout.tsx         # Main layout wrapper
│   └── ui/               # Generic UI components
│       ├── Modal.tsx
│       ├── DropdownMenu.tsx
│       └── RightDrawer.tsx
├── features/              # Feature-based modules
│   ├── assistant/         # Claude AI assistant
│   ├── auth/             # Authentication
│   ├── git/              # Git diff visualization
│   ├── projects/         # Project management
│   ├── sessions/         # Session management
│   ├── settings/         # App settings
│   ├── terminal/         # Terminal emulator
│   └── todos/            # Todo/task management
├── hooks/                 # Global custom hooks
│   └── useSessionActivity.ts
├── services/              # Service layers
│   └── api/
│       └── baseApi.ts    # RTK Query base setup
├── shared/                # Shared utilities and types
│   ├── components/       # Shared components
│   ├── hooks/           # Shared hooks
│   └── types/           # Shared type definitions
├── store/                 # Zustand store
│   └── index.ts          # Main app store
├── types/                 # Global type definitions
│   └── index.ts
├── utils/                 # Utility functions
│   ├── errorHandling.ts
│   └── notifications.ts
├── main.tsx              # Application entry point
└── index.css             # Global styles
```

## Feature Module Structure

Each feature module follows a consistent structure:

```
features/[feature-name]/
├── components/           # Feature-specific components
├── hooks/               # Feature-specific hooks
├── api/                 # Feature-specific API endpoints (if using RTK Query)
├── slice/               # Redux slice (legacy, being phased out)
└── types/               # Feature-specific types
```

## State Management

The application uses a hybrid approach for state management:

1. **Zustand** (Primary) - Used for global application state
   - Current project and session
   - UI preferences
   - Real-time updates

2. **React Query** - Used for server state
   - Data fetching and caching
   - Optimistic updates
   - Background refetching

3. **Redux Toolkit** (Legacy) - Being phased out
   - Still used in some features
   - Gradually migrating to Zustand

## API Layer

The application supports two API approaches:

1. **Direct API calls** - Using axios client
   - Most components use this approach
   - Simple and straightforward

2. **RTK Query** - Redux Toolkit Query
   - Used in some features
   - Provides automatic caching

## WebSocket Architecture

Real-time features use WebSocket connections:

- **Main WebSocket** - For application events and Claude responses
- **Terminal WebSockets** - Separate connections for each terminal session

## Key Features

### Assistant
- Claude AI chat interface
- Streaming responses
- Tool use visualization
- Task extraction

### Projects
- Project creation and management
- SSH project support
- Startup scripts

### Sessions
- Git worktree management
- Activity tracking
- Multiple concurrent sessions

### Terminal
- Full terminal emulator using xterm.js
- PTY backend support
- Session persistence

### Git Integration
- Diff visualization
- Branch comparison
- File change tracking

### Todos
- Task management
- Session-specific todos
- Progress tracking

## Component Patterns

### Container/Presentational
Some features use container components for logic and presentational components for UI.

### Direct State Access
Most components directly access Zustand store using the `useAppStore` hook.

### Error Boundaries
The app uses React error boundaries for graceful error handling.

## Styling

- **Tailwind CSS** for utility-first styling
- **Dark mode** support throughout
- **Responsive design** with mobile considerations

## Build and Development

- **Vite** for fast development and optimized builds
- **TypeScript** strict mode for type safety
- **Hot Module Replacement** for development

## Future Considerations

1. Complete migration from Redux to Zustand
2. Consolidate API approaches to a single pattern
3. Enhance WebSocket architecture for better performance
4. Add comprehensive testing setup