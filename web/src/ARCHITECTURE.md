# Frontend Architecture Guide

## Overview

This frontend uses Redux Toolkit with RTK Query for state management and data fetching. The architecture follows a feature-based structure with clear separation of concerns.

## Folder Structure

```
src/
├── app/                 # App-wide setup and configuration
│   ├── store.ts        # Redux store configuration
│   ├── hooks.ts        # Typed Redux hooks
│   ├── middleware/     # Custom middleware (WebSocket)
│   └── App.tsx         # Root application component
│
├── features/           # Feature-based modules
│   ├── auth/          # Authentication feature
│   ├── projects/      # Projects management
│   ├── sessions/      # Sessions management
│   └── agents/        # Agents management
│
├── shared/            # Shared resources
│   ├── components/    # Reusable UI components
│   ├── hooks/        # Shared custom hooks
│   ├── utils/        # Utility functions
│   └── types/        # Shared TypeScript types and schemas
│
├── services/         # External service integrations
│   ├── api/         # API configuration and base setup
│   └── websocket/   # WebSocket utilities
│
└── components/       # Legacy components (to be migrated)
```

## Key Concepts

### 1. Redux Store Structure

The store is organized by feature with the following slices:
- `auth` - Authentication state and credentials
- `projects` - Current project and filters
- `sessions` - Current session and activity tracking
- `api` - RTK Query API state (managed automatically)

### 2. Data Flow

1. **UI Components** dispatch actions or call RTK Query hooks
2. **RTK Query** handles API calls with automatic caching
3. **Redux Slices** manage UI state and business logic
4. **WebSocket Middleware** handles real-time updates
5. **Selectors** provide derived state to components

### 3. Type Safety

- **Zod Schemas** validate API responses at runtime
- **TypeScript Types** are inferred from Zod schemas
- **Typed Hooks** ensure type safety throughout the app

### 4. WebSocket Integration

The WebSocket middleware:
- Manages connection lifecycle
- Handles reconnection with exponential backoff
- Dispatches messages as Redux actions
- Integrates with authentication

### 5. Authentication

- Credentials stored in Redux and localStorage
- Automatic auth header injection
- Auth modal for 401 responses
- WebSocket auth integration

## Best Practices

### Component Structure

```typescript
// Container Component (handles logic)
export function ProjectListContainer() {
  const dispatch = useAppDispatch()
  const projects = useGetProjectsQuery()
  
  const handleSelect = (project: Project) => {
    dispatch(setCurrentProject(project))
  }
  
  return <ProjectListView {...props} />
}

// View Component (handles presentation)
export function ProjectListView({ projects, onSelect }: Props) {
  return <div>...</div>
}
```

### API Calls

```typescript
// Use RTK Query hooks
const { data, isLoading, error } = useGetProjectsQuery()

// Mutations
const [createProject] = useCreateProjectMutation()

await createProject(data).unwrap()
```

### State Updates

```typescript
// Dispatch actions
dispatch(setCurrentProject(project))

// Use selectors
const currentProject = useAppSelector(selectCurrentProject)
```

### WebSocket Subscriptions

```typescript
useWebSocketSubscription({
  messageType: 'session_update',
  onMessage: (message) => {
    // Handle message
  }
})
```

## Migration Guide

To migrate a component:

1. Create feature folder structure
2. Define Zod schemas for types
3. Create RTK Query API endpoints
4. Create Redux slice if needed
5. Split into Container/View components
6. Update imports in parent components

## Performance Considerations

- RTK Query automatically deduplicates requests
- Use `skip` parameter to conditionally fetch
- Implement proper cache invalidation
- Use React.memo for expensive components
- Leverage RTK Query's optimistic updates

## Testing

The architecture supports easy testing:
- Mock the store for component tests
- Mock API endpoints for integration tests
- Test slices and selectors in isolation
- Use MSW for API mocking