import { LoginPage } from './features/auth/components/LoginPage'
import { DocumentsPage } from './features/documents/components/DocumentsPage'
import { EditorPage } from './features/editor/EditorPage'

export function App() {
  switch (window.location.pathname) {
    case '/editor':
      return <EditorPage />
    case '/documents':
      return <DocumentsPage />
    default:
      return <LoginPage />
  }
}
