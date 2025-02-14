import Register from '@/pages/register'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/register')({
  component: Register,
})
