import Table from '@/components/Table';
import { Folder } from '@/types';
import { createFileRoute } from '@tanstack/react-router';
import { useNavigate } from '@tanstack/react-router';

function DriveComponent() {
  const { folders } = Route.useLoaderData() as { folders: Folder[] };
  const { folderPath } = Route.useParams();
  const navigate = useNavigate();

  const filteredFolders = folders.filter((folder: Folder) => {
    if (!folderPath) {
      return folder.depth === 0;
    }
    const parentFolder = folders.find((f: Folder) => f.name === folderPath);
    return parentFolder && folder.parent_id === parentFolder.id;
  });

  const handleFolderClick = (folder: Folder) => {
    navigate({ to: `/auth/drive/${folder.name}` });
  };

  return <Table folders={filteredFolders} onFolderClick={handleFolderClick} />;
}

export const Route = createFileRoute('/_auth/drive/$folderPath')({
  loader: async ({ context }) => {
    try {
      const response = await fetch('http://localhost:8082/api/folders', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${context.auth.accessToken}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch folders');
      }

      const folders = await response.json();
      return { folders };
    } catch (error) {
      console.error('Error loading folders:', error);
      return { folders: [] };
    }
  },

  component: DriveComponent,

  parseParams: (params) => {
    return {
      folderPath: params.folderPath
        ? decodeURIComponent(params.folderPath)
        : '',
    };
  },
});
