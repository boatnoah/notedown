import {
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  redirect,
} from '@tanstack/react-router'

import { LoginPage } from './features/auth/components/LoginPage'
import { RegisterPage } from './features/auth/components/RegisterPage'
import { DocumentsPage } from './features/documents/components/DocumentsPage'
import { EditorPage } from './features/editor/EditorPage'
import { isAuthenticated } from './lib/auth'

const rootRoute = createRootRoute({
  component: Outlet,
})

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  beforeLoad: () => {
    throw redirect({ to: isAuthenticated() ? '/documents' : '/login' })
  },
})

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  validateSearch: (search: Record<string, unknown>) => {
    const raw = search.redirect
    // Accept only same-origin relative paths to prevent open-redirect attacks.
    const redirect =
      typeof raw === 'string' && raw.startsWith('/') && !raw.startsWith('//') ? raw : undefined
    return { redirect }
  },
  component: LoginPage,
})

const registerRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/register',
  component: RegisterPage,
})

// Pathless layout route — enforces auth for all child routes.
const authRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'auth',
  beforeLoad: ({ location }) => {
    if (!isAuthenticated()) {
      throw redirect({
        to: '/login',
        search: { redirect: location.href },
      })
    }
  },
  component: Outlet,
})

const documentsRoute = createRoute({
  getParentRoute: () => authRoute,
  path: '/documents',
  component: DocumentsPage,
})

const editorRoute = createRoute({
  getParentRoute: () => authRoute,
  path: '/editor',
  validateSearch: (search: Record<string, unknown>) => ({
    room: typeof search.room === 'string' ? search.room : undefined,
  }),
  component: EditorPage,
})

const routeTree = rootRoute.addChildren([
  indexRoute,
  loginRoute,
  registerRoute,
  authRoute.addChildren([documentsRoute, editorRoute]),
])

export const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
