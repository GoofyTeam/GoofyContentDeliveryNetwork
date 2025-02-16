import Table from '@/components/Table';
import { Folder } from '@/types';
import { createFileRoute } from '@tanstack/react-router';
import { useNavigate } from '@tanstack/react-router';

const DriveComponent = () => {
  const { folders } = Route.useLoaderData();
  const { folderPath } = Route.useParams() as { folderPath: string };
  const navigate = useNavigate();

  const currentPath = folderPath ? `/${folderPath}` : '/';

  const handleFolderClick = (folder: Folder) => {
    const newPath = folder.path === '/' ? '' : folder.path;
    navigate({ to: `/drive${newPath}` });
  };

  const filteredFolders = folders.filter((folder: Folder) => {
    if (!folderPath) {
      return folder.depth === 0;
    }

    return (
      folder.path.startsWith(currentPath) &&
      folder.path !== currentPath &&
      !folder.path.slice(currentPath.length + 1).includes('/')
    );
  });

  return (
    <div>
      <div className='mb-4 text-sm breadcrumbs'>
        <span
          onClick={() => navigate({ to: '/drive' })}
          className='cursor-pointer hover:text-blue-500'
        >
          Root
        </span>
        {folderPath &&
          folderPath
            .split('/')
            .filter(Boolean)
            .map((segment, index, array) => (
              <span key={index}>
                {' / '}
                <span
                  className='cursor-pointer hover:text-blue-500'
                  onClick={() => {
                    const pathTo = array.slice(0, index + 1).join('/');
                    navigate({ to: `/drive/${pathTo}` });
                  }}
                >
                  {segment}
                </span>
              </span>
            ))}
      </div>
      <Table folders={filteredFolders} onFolderClick={handleFolderClick} />
    </div>
  );
};

export const Route = createFileRoute('/_auth/drive/')({
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
});
