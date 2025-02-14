import Drive from '@/pages/drive'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_auth/drive/')({
  component: () => <Drive />,
})
